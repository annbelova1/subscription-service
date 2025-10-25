.PHONY: up up-all fill build-fill logs down migrate test-unit test-integration test-all clean setup

up:
	docker compose up -d

up-all: build-fill
	docker compose --profile tools up -d

fill: build-fill
	docker compose run --rm fill_db

migrate:
	docker compose run --rm migrations

logs:
	docker compose logs -f app

down:
	docker compose down

clear:
	docker compose down --remove-orphans --volumes --rmi all

build-fill:
	mkdir -p bin
	go build -o bin/fill_db ./cmd/fill_db

run-fill: build-fill
	./bin/fill_db -api-url http://localhost:8080

test-unit:
	docker compose -f docker-compose.test.yml up unit-test-runner --build --abort-on-container-exit

test-integration:
	docker compose -f docker-compose.test.yml up integration-test-runner --build --abort-on-container-exit

test-all:
	docker compose -f docker-compose.test.yml up unit-test-runner integration-test-runner --build --abort-on-container-exit

test-local-unit:
	go test -v -short ./internal/handlers ./internal/service

test-local-integration:
	DB_HOST=localhost DB_PORT=5433 go test -v -tags=integration ./internal/repository

clean:
	docker compose -f docker-compose.test.yml down -v
	rm -rf test-reports/

setup:
	mkdir -p test-reports