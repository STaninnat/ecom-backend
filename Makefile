GO ?= go
GO_ENV ?= GOMODCACHE=/tmp/go-mod-cache GOCACHE=/tmp/go-build-cache
COMPOSE ?= $(shell docker compose version >/dev/null 2>&1 && echo docker compose || echo docker-compose)
ENV_FILE ?= .env.development
COVERPROFILE ?= coverage.out
BEHAVIOR_PACKAGES ?= $(shell GOMODCACHE=/tmp/go-mod-cache GOCACHE=/tmp/go-build-cache $(GO) list ./... | grep -Ev '(/docs|/internal/database|/internal/testutil|/models)$$|^github.com/STaninnat/ecom-backend$$')

.PHONY: dev-env services-up services-down services-logs migrate-up migrate-down migrate-status sqlc run smoke test cover cover-func cover-behavior tidy-check fmt-check vet build lint checks

dev-env:
	@test -f $(ENV_FILE) || cp .env.development.example $(ENV_FILE)
	@mkdir -p logs uploads
	@echo "Created $(ENV_FILE) if it was missing. Review it before running the server."

services-up:
	$(COMPOSE) up -d postgres mongodb redis

services-down:
	$(COMPOSE) down

services-logs:
	$(COMPOSE) logs -f postgres mongodb redis

migrate-up:
	./etc/scripts/migration.sh up

migrate-down:
	./etc/scripts/migration.sh down

migrate-status:
	./etc/scripts/migration.sh status

sqlc:
	sqlc generate

run: dev-env
	$(GO_ENV) $(GO) run .

smoke:
	./scripts/smoke_endpoints.sh

test:
	$(GO_ENV) $(GO) test ./...

cover:
	$(GO_ENV) $(GO) test ./... -cover

cover-func:
	$(GO_ENV) $(GO) test ./... -coverprofile=$(COVERPROFILE)
	$(GO) tool cover -func=$(COVERPROFILE)

cover-behavior:
	$(GO_ENV) $(GO) test $(BEHAVIOR_PACKAGES) -coverprofile=$(COVERPROFILE)
	$(GO) tool cover -func=$(COVERPROFILE)

tidy-check:
	$(GO_ENV) $(GO) mod tidy -diff

fmt-check:
	@gofmt -l $$(find . -name '*.go' -not -path './vendor/*')

vet:
	$(GO_ENV) $(GO) vet ./...

build:
	$(GO_ENV) $(GO) build ./...

lint:
	golangci-lint run --fix=false

checks: tidy-check fmt-check vet test cover build
