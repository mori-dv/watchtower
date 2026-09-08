package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"watchtower/internal/alert"
	"watchtower/internal/api"
	"watchtower/internal/checker"
	"watchtower/internal/config"
	"watchtower/internal/logging"
	"watchtower/internal/metrics"
	"watchtower/internal/scheduler"
	"watchtower/internal/state"
	"watchtower/internal/storage"
	"watchtower/internal/worker"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logging.Init("info", "json")
		logging.Logger.Fatal().Err(err).Str("path", configPath).Msg("failed to load configuration")
	}

	logging.Init(cfg.LogLevel, cfg.LogFormat)
	metrics.Init()

	logging.Logger.Info().
		Int("workers", cfg.Workers).
		Int("queue_size", cfg.QueueSize).
		Int("targets", len(cfg.Targets)).
		Str("storage_type", cfg.Storage.Type).
		Msg("starting Watchtower monitoring service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Storage Layer (Memory or Redis)
	var store storage.Store
	if cfg.Storage.Type == "redis" || os.Getenv("REDIS_ADDRESS") != "" {
		addr := cfg.Storage.RedisAddr
		if envAddr := os.Getenv("REDIS_ADDRESS"); envAddr != "" {
			addr = envAddr
		}
		redisStore := storage.NewRedisStore(addr, cfg.Storage.RedisPassword)
		pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
		if err := redisStore.Ping(pingCtx); err != nil {
			logging.Logger.Warn().Err(err).Str("redis_addr", addr).Msg("redis ping failed, falling back to in-memory store")
			store = storage.NewMemoryStore()
		} else {
			logging.Logger.Info().Str("redis_addr", addr).Msg("connected to Redis state storage")
			store = redisStore
		}
		pingCancel()
	} else {
		logging.Logger.Info().Msg("using in-memory state storage")
		store = storage.NewMemoryStore()
	}

	evaluator := state.NewEvaluator(store)

	// Initialize Pluggable Alert Dispatchers
	var dispatchers []alert.Dispatcher
	if cfg.Alerts.Telegram.Enabled && cfg.Alerts.Telegram.BotToken != "" {
		dispatchers = append(dispatchers, alert.NewTelegramDispatcher(
			cfg.Alerts.Telegram.BotToken,
			cfg.Alerts.Telegram.ChatID,
		))
		logging.Logger.Info().Msg("telegram alert dispatcher registered")
	}
	if cfg.Alerts.Slack.Enabled && cfg.Alerts.Slack.WebhookURL != "" {
		dispatchers = append(dispatchers, alert.NewSlackDispatcher(
			cfg.Alerts.Slack.WebhookURL,
			cfg.Alerts.Slack.Channel,
		))
		logging.Logger.Info().Msg("slack alert dispatcher registered")
	}
	if cfg.Alerts.Webhook.Enabled && cfg.Alerts.Webhook.URL != "" {
		dispatchers = append(dispatchers, alert.NewWebhookDispatcher(
			cfg.Alerts.Webhook.URL,
			cfg.Alerts.Webhook.Headers,
		))
		logging.Logger.Info().Msg("webhook alert dispatcher registered")
	}

	alertManager := alert.NewManager(store, cfg.Alerts.Cooldown, dispatchers...)
	statusRegistry := api.NewStatusRegistry()

	// Initialize Worker Pool & Channels
	jobs := make(chan worker.Job, cfg.QueueSize)
	results := make(chan checker.Result, cfg.QueueSize)

	pool := worker.NewPool()
	pool.Start(ctx, cfg.Workers, jobs, results)

	// Initialize Dynamic Scheduler
	sched := scheduler.NewScheduler(jobs)
	sched.Start(ctx, cfg.Targets)

	// Results Processing Goroutine
	resultsDone := make(chan struct{})
	go func() {
		defer close(resultsDone)
		for result := range results {
			evaluated, err := evaluator.Evaluate(ctx, result)
			if err != nil {
				logging.Logger.Error().Err(err).Str("target", result.TargetName).Msg("failed to evaluate probe state")
				continue
			}

			// Update Prometheus Metrics
			metrics.ProbesTotal.WithLabelValues(result.TargetName, result.TargetType, string(evaluated.Status)).Inc()
			metrics.ProbeDuration.WithLabelValues(result.TargetName, result.TargetType).Observe(result.Latency.Seconds())

			if evaluated.Status == checker.StatusUp {
				metrics.TargetUp.WithLabelValues(result.TargetName, result.TargetType).Set(1)
			} else {
				metrics.TargetUp.WithLabelValues(result.TargetName, result.TargetType).Set(0)
			}

			if result.SSLExpiryDays != nil {
				metrics.SSLCertExpiryDays.WithLabelValues(result.TargetName, result.TargetType).Set(float64(*result.SSLExpiryDays))
			}
			metrics.ConsecutiveFailures.WithLabelValues(result.TargetName, result.TargetType).Set(float64(evaluated.ConsecutiveFailures))

			// Update Live Status Registry
			statusRegistry.Update(result, evaluated.ConsecutiveFailures)

			// Dispatch Alerts on outage or recovery
			if err := alertManager.Handle(
				ctx,
				result,
				evaluated.Status,
				evaluated.PreviousStatus,
				evaluated.ConsecutiveFailures,
				evaluated.IsRecovery,
			); err != nil {
				logging.Logger.Error().Err(err).Str("target", result.TargetName).Msg("alert dispatch encountered errors")
			}

			// Telemetry & Latency Spike Logging
			logging.LogProbeResult(result, evaluated.ConsecutiveFailures)
		}
	}()

	// Dynamic Reload Handler
	reloadConfig := func() error {
		logging.Logger.Info().Str("path", configPath).Msg("reloading targets from configuration...")
		newCfg, err := config.Load(configPath)
		if err != nil {
			metrics.ConfigReloadsTotal.WithLabelValues("failure").Inc()
			logging.Logger.Error().Err(err).Msg("failed to reload configuration")
			return err
		}

		sched.UpdateTargets(newCfg.Targets)
		metrics.ConfigReloadsTotal.WithLabelValues("success").Inc()
		logging.Logger.Info().Int("targets", len(newCfg.Targets)).Msg("target configuration successfully reloaded")
		return nil
	}

	// SIGHUP Listener for Unix configuration reload
	hupChan := make(chan os.Signal, 1)
	signal.Notify(hupChan, syscall.SIGHUP)
	go func() {
		for range hupChan {
			_ = reloadConfig()
		}
	}()

	// HTTP Server & Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", api.HealthHandler)
	mux.HandleFunc("/ready", api.ReadyHandler(func(ctx context.Context) error {
		if pingable, ok := store.(interface{ Ping(context.Context) error }); ok {
			return pingable.Ping(ctx)
		}
		return nil
	}))
	mux.HandleFunc("/api/v1/status", statusRegistry.Handler())
	mux.HandleFunc("/reload", api.ReloadHandler(reloadConfig))
	mux.Handle("/metrics", promhttp.Handler())

	serverPort := cfg.Server.Port
	if envPort := os.Getenv("MAIN_SERVER_PORT"); envPort != "" {
		if envPort[0] != ':' {
			serverPort = ":" + envPort
		} else {
			serverPort = envPort
		}
	}

	httpServer := &http.Server{
		Addr:         serverPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logging.Logger.Info().Str("addr", serverPort).Msg("HTTP management server listening")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Logger.Fatal().Err(err).Msg("HTTP server encountered fatal error")
		}
	}()

	// Graceful Shutdown on SIGINT or SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logging.Logger.Info().Str("signal", sig.String()).Msg("shutdown signal received, draining operations...")

	// 1. Stop scheduler so no new jobs are emitted
	sched.Stop()
	logging.Logger.Debug().Msg("scheduler stopped")

	// 2. Close worker jobs channel and drain active worker pool
	close(jobs)
	pool.Wait()
	logging.Logger.Debug().Msg("worker pool drained")

	// 3. Close results channel and wait for results processor to finish
	close(results)
	<-resultsDone
	logging.Logger.Debug().Msg("results processor finished")

	// 4. Gracefully terminate HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logging.Logger.Error().Err(err).Msg("error shutting down HTTP server")
	}

	// 5. Close storage connections
	_ = store.Close()

	cancel()
	logging.Logger.Info().Msg("Watchtower shutdown successfully completed")
}