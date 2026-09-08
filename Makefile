.PHONY: all build mock run docker lint vet test test-race vulncheck fmt compose-up compose-down compose-logs clean

all: lint vet test-race build

build:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/watchtower ./cmd/watchtower
	go build -ldflags="-s -w" -o bin/mockserver ./cmd/mockserver

mock:
	go run ./cmd/mockserver

run:
	go run ./cmd/watchtower

test:
	go test -v ./...

test-race:
	go test -race -v -coverprofile=coverage.out ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

vulncheck:
	govulncheck ./...

fmt:
	go fmt ./...

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down -v

compose-logs:
	docker compose logs -f watchtower

clean:
	rm -rf bin/ coverage.out