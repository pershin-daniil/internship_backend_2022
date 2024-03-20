include .env
export

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf " \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f
.PHONY: compose-up

compose-down: ### Remove docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

docker-rm-volume: ### Remove docker volume
	docker volume rm pg-data
.PHONY: docker-rm-volume

linter-golangci: ### check by golangci linter
	golangci-lint run
.PHONY: linter-golangci

linter-handolint: ### check by handolint linter
	git ls-files --exclude='Dockerfile*' -c --ignored | xargs hadolint
.PHONY: linter-handolint

migrate-up: ### migration up
	migrate -path migrations -database '${PG_URL_LOCALHOST}?sslmode=disable' up
.PHONY: migrate-up

migrate-down: ### migrate down
	echo "y" | migrate -path migrations -database '$(PG_URL_LOCALHOST)?sslmode=disable' down
.PHONY: migrate-down

test: ### run test
	go test -v ./...

cover-html: ### runt test with coverage and open html report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out
.PHONY: coverage-html

cover: ### runt test with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out
.PHONY: coverage

mockgen: ### generate mock
	mockgen -source=internal/service/service.go -destination=internal/mocks/servicemocks.go -package=servicemocks
	mockgen -source=internal/repo/repo.go -destination=internal/mocks/repomocks.go -package=repomocks
	mockgen -source=internal/webapi/webapi.go -destination=internal/mocks/webapimocks.go -package=webapimocks
.PHONY: mockgen

swag: ### generate swagger docs
	swag init -g internal/app/app.go --parseInternal --parseDependency
