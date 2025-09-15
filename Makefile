demo:
	go run cmd/all/main.go

dev:
	go run cmd/all/main.go

build:
	go build -o bin/server cmd/all/main.go

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

lint:
	go vet ./...
	go fmt ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

.PHONY: demo dev build test cover lint docker-up docker-down docker-logs
