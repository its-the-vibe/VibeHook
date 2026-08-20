BINARY_NAME=vibehook

.PHONY: build test run docker-build docker-up docker-down tidy

build:
	go build -o bin/$(BINARY_NAME) ./cmd/vibehook

test:
	go test ./...

run:
	go run ./cmd/vibehook --config config.yaml

docker-build:
	docker build -t its-the-vibe/vibehook:local .

docker-up:
	docker compose up --build

docker-down:
	docker compose down

tidy:
	go mod tidy
