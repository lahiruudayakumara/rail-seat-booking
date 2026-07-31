.PHONY: up seed down reset logs migrate-up migrate-down test test-go build fmt

up:
	docker compose up --build

seed:
	docker compose run --rm seed

down:
	docker compose down

reset:
	docker compose down -v
	docker compose up --build

logs:
	docker compose logs -f api db migrate

migrate-up:
	docker compose run --rm --entrypoint /usr/local/bin/migrate api up

migrate-down:
	docker compose run --rm --entrypoint /usr/local/bin/migrate api down

test: test-go

test-go:
	go test -race ./...

build:
	go build ./apps/api/cmd/server

fmt:
	gofmt -w apps/api
