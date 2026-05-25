# Watchtower

Lightweight production-oriented uptime and health monitoring service written in Go.

Watchtower was built as a portfolio project focused on DevOps, observability, reliability engineering, and backend infrastructure concepts — not as a large SaaS platform.

The goal of the project is to explore how real monitoring systems work internally while keeping the implementation intentionally lightweight, understandable, and operationally focused.

---

# Why I Built This

I built Watchtower to strengthen practical skills around:

* concurrent backend systems
* infrastructure-oriented Go development
* observability pipelines
* monitoring semantics
* reliability concepts
* production-style Docker environments
* metrics and logging ecosystems
* operational thinking used in SRE/DevOps roles

This project complements another infrastructure-focused project I built called Gatekeeper (distributed rate limiter).

---

# Features

## Monitoring

* HTTP health checks
* TCP endpoint monitoring
* configurable monitoring intervals
* configurable timeouts
* retry support
* latency measurement
* graceful shutdown

---

## Reliability Features

* consecutive failure tracking
* DEGRADED and DOWN states
* stateful monitoring semantics
* Redis-backed state persistence
* alert cooldown protection
* recovery handling

---

## Observability

* Prometheus metrics
* Grafana dashboards
* Loki log aggregation
* structured JSON logging
* uptime metrics
* latency metrics
* failure metrics

---

## Alerting

* Telegram alerts
* cooldown-based anti-spam logic
* state-aware notifications

---

## Infrastructure

* Docker-first setup
* Docker Compose stack
* multi-service observability environment
* production-oriented architecture
* GitHub Actions CI pipeline

---

# Architecture

```text
                +-------------+
                | Watchtower  |
                +-------------+
                     |
         +-----------+-----------+
         |                       |
         v                       v
   +-----------+          +-------------+
   | Prometheus|          | Structured  |
   | Metrics   |          | Logs        |
   +-----------+          +-------------+
         |                       |
         v                       v
   +-----------+          +-------------+
   | Grafana   |          | Loki        |
   +-----------+          +-------------+

                +-------------+
                | Redis State |
                +-------------+
                       |
                       v
                +-------------+
                | Telegram    |
                | Alerting    |
                +-------------+
```

---

# Internal Workflow

```text
Scheduler
   ↓
Worker Pool
   ↓
Checker Engine
   ↓
State Evaluator
   ↓
Metrics / Logs / Alerts
```

---

# Why Redis?

Redis is used as a lightweight state layer for operational reliability features.

It is responsible for:

* consecutive failure tracking
* alert cooldowns
* state persistence
* temporary monitoring state
* reliability semantics

Without Redis, monitoring would be stateless and unable to distinguish between transient failures and actual outages.

---

# Reliability Model

Watchtower intentionally avoids treating every single failed request as a full outage.

Example:

```text
1st failure  -> DEGRADED
2nd failure  -> DEGRADED
3rd failure  -> DOWN
```

This models real operational behavior more accurately and reduces noisy alerts caused by temporary network instability.

---

# Tech Stack

* Go
* Redis
* Docker
* Docker Compose
* Prometheus
* Grafana
* Loki
* Promtail
* Nginx
* GitHub Actions

---

# Project Structure

```text
watchtower/
├── cmd/
├── internal/
│   ├── alert/
│   ├── checker/
│   ├── config/
│   ├── logging/
│   ├── metrics/
│   ├── scheduler/
│   ├── state/
│   ├── storage/
│   └── worker/
├── configs/
├── observability/
│   ├── grafana/
│   ├── loki/
│   ├── prometheus/
│   └── promtail/
├── .github/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

---

# Quick Start

## Clone Repository

```bash
git clone https://github.com/yourusername/watchtower.git

cd watchtower
```

---

## Configure Targets

Edit:

```text
configs/config.yaml
```

Example:

```yaml
workers: 5

targets:
  - name: google
    type: http
    address: https://google.com
    interval: 10s
    timeout: 5s
    retries: 2

  - name: cloudflare-dns
    type: tcp
    address: 1.1.1.1:53
    interval: 5s
    timeout: 3s
    retries: 2
```

---

## Configure Telegram Alerts

Create `.env`

```env
TELEGRAM_BOT_TOKEN=YOUR_TOKEN
TELEGRAM_CHAT_ID=YOUR_CHAT_ID
```

---

## Start Stack

```bash
docker compose up --build
```

---

# Services

| Service    | Port |
| ---------- | ---- |
| Watchtower | 8080 |
| Grafana    | 3000 |
| Prometheus | 9090 |

---

# Metrics

Prometheus metrics available at:

```text
/metrics
```

Includes:

* uptime status
* latency
* check counts
* failure counts
* target availability

---

# Dashboards

Grafana dashboards visualize:

* service health
* uptime trends
* latency spikes
* failure rates
* operational state changes

---

# CI Pipeline

GitHub Actions pipeline includes:

* Go build validation
* automated tests
* golangci-lint
* Trivy security scanning

---

# Production Notes

This project is intentionally lightweight and optimized for learning and portfolio value.

In a real production environment, the following would typically be added:

* reverse proxy authentication
* HTTPS/TLS termination
* multi-region probes
* distributed monitoring agents
* persistent long-term storage
* HA monitoring stack
* RBAC and access control

---

# Future Improvements

Potential future enhancements:

* distributed regional agents
* Web UI
* historical incident timelines
* SLO/SLA tracking
* Kubernetes deployment
* dynamic service discovery
* WebSocket live status streaming

---

# What I Learned

Building Watchtower helped me better understand:

* observability systems
* production-style monitoring
* concurrency patterns in Go
* worker pool architecture
* reliability engineering concepts
* metrics-driven systems
* operational debugging workflows
* infrastructure-focused backend design

---

# License

MIT License
