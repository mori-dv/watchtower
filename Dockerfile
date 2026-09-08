# Stage 1: Build static binaries
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Leverage Docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/watchtower ./cmd/watchtower

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-s -w -extldflags '-static'" \
    -o /bin/mockserver ./cmd/mockserver

# Stage 2: Hardened, minimal runtime container
FROM alpine:3.21 AS runner

RUN apk --no-cache add ca-certificates tzdata iputils

# Create least-privilege non-root user and group
RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup -D -H -s /sbin/nologin

WORKDIR /app

COPY --from=builder --chown=10001:10001 /bin/watchtower /bin/watchtower
COPY --from=builder --chown=10001:10001 /bin/mockserver /bin/mockserver
COPY --from=builder --chown=10001:10001 /src/configs ./configs

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O - http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/bin/watchtower"]