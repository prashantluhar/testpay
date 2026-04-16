.PHONY: build test test-integration run docker-up docker-down migrate lint coverage

build:
	go build -o bin/testpay ./cmd/testpay

test:
	go test ./... -count=1

test-integration:
	TEST_DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable" \
	go test ./internal/store/postgres/... -v -count=1

run:
	DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable" \
	go run ./cmd/testpay start --config deploy/config/testpay.local.yaml

docker-up:
	docker compose -f deploy/docker/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker/docker-compose.yml down

lint:
	go vet ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
