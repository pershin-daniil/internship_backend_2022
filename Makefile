lint:
	gofumpt -w .
	go mod tidy
	golangci-lint run

up:
	docker-compose up -d

test-integration: up
	go test -v ./tests/integration_test.go
	docker-compose down