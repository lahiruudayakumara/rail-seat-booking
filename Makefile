.PHONY: up seed down reset logs migrate-up migrate-down migration-status test test-go test-web build fmt verify smoke integration load

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

migration-status:
	./scripts/check-migrations.sh

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

verify:
	./scripts/verify.sh

smoke:
	./scripts/smoke-test.sh

integration:
	./tests/integration/booking-flow.sh

load:
	k6 run tests/load/booking-contention.js
