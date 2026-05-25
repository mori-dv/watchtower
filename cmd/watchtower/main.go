package main

import (
	"context"
	"net/http"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/joho/godotenv"
)
func main() {

	// cfg := config.Config{
	// 	Workers: 5,
	// 	Targets: []config.Target{
	// 		{
	// 			Name:     "google",
	// 			Type:     "http",
	// 			Address:  "https://google.com",
	// 			Interval: 10 * time.Second,
	// 			Timeout:  5 * time.Second,
	// 			Retries:  2,
	// 		},
	// 	},
	// }

	_ = godotenv.Load()

	mainPort := os.Getenv("MAIN_SERVER_PORT")
	redisAddress := os.Getenv("REDIS_ADDRESS")

	cfg, err := config.Load("configs/config.yml")

	cfg.Alerts.TelegramBotToken=os.Getenv("TELEGRAM_BOT_TOKEN")
	cfg.Alerts.TelegramChatId=os.Getenv("TELEGRAM_CHAT_ID")

	if err != nil {
		panic(err)
	}

	logging.Init()
	
	metrics.Init()


	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan worker.Job, 100)

	results := make(chan checker.Result, 100)

	store := storage.NewRedisStore(
		redisAddress,
	)
	
	evaluator := state.NewEvaluator(
		store,
	)
	alertManager := alert.NewManager(
		store,
		cfg.Alerts.TelegramBotToken,
		cfg.Alerts.TelegramChatId,
		cfg.Alerts.Cooldown,
	)

	pool := worker.NewPool()

	pool.Start(
		ctx,
		cfg.Workers,
		jobs,
		results,
	)

	s := scheduler.NewScheduler(
		cfg.Targets,
		jobs,
	)

	s.Start(ctx)

	go func() {
		for result := range results {
			evaluated, err := evaluator.Evaluate(
				ctx,
				result,
			)
			if err != nil {
				continue
			}
			result.Status = evaluated.Status

			metrics.ChecksTotal.
				WithLabelValues(
					result.TargetName,
					result.TargetType,
					string(result.Status),
				).
				Inc()
				metrics.CheckLatency.
				WithLabelValues(
					result.TargetName,
					result.TargetType,
				).
				Observe(
					result.Latency.Seconds(),
				)
				if result.Status == checker.StatusUp {
				metrics.TargetUp.
					WithLabelValues(
						result.TargetName,
						result.TargetType,
					).
					Set(1)

			} else {
				metrics.TargetUp.
					WithLabelValues(
						result.TargetName,
						result.TargetType,
					).
					Set(0)
			}
			if err := alertManager.Handle(
				ctx,
				result,
			); err != nil {
				logging.Logger.Error().
					Err(err).
					Msg("failed to handle alert")
			}
			
			logging.Logger.Info().
			Str("target", result.TargetName).
			Str("type", result.TargetType).
			Str("status", string(result.Status)).
			Int(
				"consecutive_failures",
				evaluated.ConsecutiveFailures,
			).
			Dur("latency", result.Latency).
			Msg("check completed")
			if err := alertManager.Handle(
				ctx,
				result,
			); err != nil {

			logging.Logger.Error().
				Err(err).
				Msg("failed to handle alert")
		}
	}
}()

	http.HandleFunc("/healthz", api.HealthHandler)
	
	http.Handle("/metrics",	promhttp.Handler())

	go func() {
		if err := http.ListenAndServe(mainPort, nil); err != nil {
			log.Fatalf("couldn't serv the backend http server and have some issues: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-sigChan

	cancel()
}