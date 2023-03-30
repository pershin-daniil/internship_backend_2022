lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run

up:
	docker compose up -d

down:
	docker compose down

test: up
	go test -v ./tests/integration_test.go
	docker compose down

run: up
	go run ./cmd/main.go
