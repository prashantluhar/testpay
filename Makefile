.PHONY: build test test-integration run docker-up docker-down migrate lint coverage coverage-check

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

COVERAGE_THRESHOLD ?= 90

coverage-check:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total coverage: $$COVERAGE%"; \
	awk -v cov="$$COVERAGE" -v thr="$(COVERAGE_THRESHOLD)" 'BEGIN { if (cov+0 < thr+0) { exit 1 } }' || \
	(echo "Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; exit 1)
	@echo "Coverage gate passed"
