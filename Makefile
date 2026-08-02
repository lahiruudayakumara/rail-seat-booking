=.PHONY: up seed down reset logs migrate-up migrate-down test test-go test-web build fmt

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
	docker compose logs -f api web db migrate

migrate-up:
	docker compose run --rm --entrypoint /usr/local/bin/migrate api up

migrate-down:
	docker compose run --rm --entrypoint /usr/local/bin/migrate api down

test: test-go test-web

test-go:
	go test -race ./...

test-web:
	pnpm --filter @rail/web test

build:
	go build ./apps/api/cmd/server
	pnpm --filter @rail/web build

fmt:
	gofmt -w apps/api
