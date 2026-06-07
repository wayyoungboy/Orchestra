.PHONY: help verify verify-focused build run test reset-data clean backend-run backend-test backend-reset frontend-install frontend-dev frontend-build frontend-test

help:
	@echo "Orchestra development commands"
	@echo ""
	@echo "  make verify         Run backend tests, frontend build/unit tests, and focused spec typecheck"
	@echo "  make verify-focused Run the full local focused MVP browser gate with a temporary backend"
	@echo "  make run           Start the backend API server (alias: backend-run)"
	@echo "  make test          Run backend tests (alias: backend-test)"
	@echo "  make build         Build the backend server binary"
	@echo "  make reset-data    Reset backend SQLite data (alias: backend-reset)"
	@echo "  make backend-run    Start the backend API server"
	@echo "  make backend-test   Run backend tests"
	@echo "  make backend-reset  Reset backend SQLite data"
	@echo "  make frontend-dev   Start the frontend dev server"
	@echo "  make frontend-build Build the frontend"
	@echo "  make frontend-test  Run frontend unit tests"

verify:
	./scripts/verify-mvp.sh

verify-focused:
	./scripts/run-focused-e2e-local.sh

build:
	cd backend && $(MAKE) build

run: backend-run

test: backend-test

reset-data: backend-reset

clean:
	cd backend && $(MAKE) clean

backend-run:
	cd backend && $(MAKE) run

backend-test:
	cd backend && $(MAKE) test

backend-reset:
	cd backend && $(MAKE) reset-data

frontend-install:
	cd frontend && pnpm install

frontend-dev:
	cd frontend && pnpm dev

frontend-build:
	cd frontend && pnpm build

frontend-test:
	cd frontend && pnpm test
