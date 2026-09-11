.PHONY: dev dev-full stop build test lint frontend clean help

# Default target
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ---------- Development ----------

dev: frontend infra ## Build frontend, start infra, run server
	REDIS_ADDR=localhost:6379 \
	DATABASE_URL=postgres://grow:grow@localhost:5432/grow?sslmode=disable \
	go run .

dev-quick: infra ## Start server only (skip frontend build)
	REDIS_ADDR=localhost:6379 \
	DATABASE_URL=postgres://grow:grow@localhost:5432/grow?sslmode=disable \
	go run .

dev-bare: frontend ## Build frontend, run server without Redis/Postgres
	go run .

infra: ## Start Redis + Postgres in Docker
	@docker compose up -d --wait
	@echo "  Redis:    localhost:6379"
	@echo "  Postgres: localhost:5432 (grow/grow)"

stop: ## Stop Docker services
	@docker compose down

# ---------- Build ----------

frontend: ## Build frontend (Vite → internal/server/dist/)
	@cd frontend && pnpm install --frozen-lockfile && pnpm build

build: frontend ## Full production build
	go build -o grow .

# ---------- Quality ----------

lint: ## Format + vet
	gofmt -w .
	go vet ./...
	@cd frontend && npx tsc --noEmit

test: ## Run all tests
	go test -race -count=1 ./...

test-v: ## Run all tests (verbose)
	go test -race -count=1 -v ./...

# ---------- Load testing ----------

loadtest: ## Quick load test (10 concurrent, 100 requests)
	go run scripts/loadtest.go -conc 10 -n 100

benchmark: ## Concurrency sweep benchmark
	go run scripts/benchmark.go | tee results.csv

chart: ## Generate latency chart from results.csv
	python3 scripts/chart_latency.py results.csv -o docs/latency.svg

# ---------- Cleanup ----------

clean: stop ## Stop services, remove build artifacts
	rm -f grow results.csv
	rm -rf internal/server/dist
