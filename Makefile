.PHONY: dev down logs migrate-up gen-sqlc seed test-go lint-go sec-go test-py lint-py sec-py fmt

dev:
	docker compose -f infra/docker-compose.yml up -d --build

down:
	docker compose -f infra/docker-compose.yml down

logs:
	docker compose -f infra/docker-compose.yml logs -f

gen-sqlc:
	cd backend-go && go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

migrate-up:
	set -a; [ -f infra/.env ] && . infra/.env || . infra/.env.example; set +a; \
	cd backend-go && go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/db/migrations postgres "$$DB_URL" up

seed:
	set -a; [ -f infra/.env ] && . infra/.env || . infra/.env.example; set +a; \
	cd backend-go && go run ./cmd/seed

test-go:
	cd backend-go && go test ./...

lint-go:
	cd backend-go && gofmt -w . && go vet ./...

sec-go:
	cd backend-go && go run github.com/securego/gosec/v2/cmd/gosec@latest ./...

test-py:
	cd python-analytics && python -m pytest

lint-py:
	cd python-analytics && python -m ruff check app

sec-py:
	cd python-analytics && python -m bandit -r app -x app/test_main.py

fmt:
	cd backend-go && gofmt -w .
	cd python-analytics && python -m ruff format app
