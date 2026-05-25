build:
	go build -o bin/watchtower ./cmd/watchtower

run:
	go run ./cmd/watchtower

docker:
	docker compose up --build

lint:
	golangci-lint run

test:
	go test ./...

fmt:
	go fmt ./...