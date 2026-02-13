.PHONY: dev down logs migrate-up migrate-down migrate-status seed gen-sqlc test-go lint-go sec-go test-py lint-py sec-py security sbom fmt

dev:
	docker compose -f infra/docker-compose.yml up -d --build

down:
	docker compose -f infra/docker-compose.yml down

logs:
	docker compose -f infra/docker-compose.yml logs -f

migrate-up:
	cd backend-go && go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/db/migrations postgres "$$DATABASE_URL" up

migrate-down:
	cd backend-go && go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/db/migrations postgres "$$DATABASE_URL" down

migrate-status:
	cd backend-go && go run github.com/pressly/goose/v3/cmd/goose@latest -dir internal/db/migrations postgres "$$DATABASE_URL" status

seed:
	@echo "seed placeholder: create admin/roles/permissions"

gen-sqlc:
	cd backend-go && go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate -f internal/db/sqlc/sqlc.yaml

test-go:
	cd backend-go && go test ./...

lint-go:
	cd backend-go && go vet ./...

sec-go:
	cd backend-go && go run github.com/securego/gosec/v2/cmd/gosec@latest ./...

test-py:
	cd python-analytics && python -m pytest

lint-py:
	cd python-analytics && python -m ruff check app

sec-py:
	cd python-analytics && python -m bandit -r app -x app/test_main.py

security: sec-go sec-py

sbom:
	@echo "SBOM placeholder - integrate CycloneDX in CI"

fmt:
	cd backend-go && go fmt ./...
	cd python-analytics && python -m ruff format app
