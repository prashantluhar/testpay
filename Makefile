.PHONY: build test test-integration run docker-up docker-down migrate lint

build:
	go build -o bin/testpay ./cmd/testpay

test:
	go test ./... -count=1

test-integration:
	TEST_DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable" \
	go test ./internal/store/postgres/... -v -count=1

run:
	DATABASE_URL="postgres://testpay:testpay@localhost:5432/testpay?sslmode=disable" \
	go run ./cmd/testpay start

docker-up:
	docker compose up -d

docker-down:
	docker compose down

lint:
	go vet ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
