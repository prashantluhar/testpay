.PHONY: build test migrate run

build:
	go build -o bin/testpay ./cmd/testpay

test:
	go test ./...

run:
	go run ./cmd/testpay start

migrate:
	go run ./cmd/testpay migrate
