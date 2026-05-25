FROM golang:1.25-alpine3.23 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o watchtower ./cmd/watchtower

FROM alpine:3.23 AS runner

WORKDIR /app
RUN adduser -D -H -s /sbin/nologin appuser
USER appuser

COPY --from=builder /app/watchtower .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

ENTRYPOINT ["/app/watchtower"]