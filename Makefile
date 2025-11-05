.PHONY: up down logs test test-go test-frontend migrate proto

up:
	docker compose up --build -d

down:
	docker compose down -v

logs:
	docker compose logs -f --tail=200

migrate:
	docker compose run --rm migrate

test: test-go test-frontend

test-go:
	cd services/orbit-service && go test ./...
	cd services/planning-service && go test ./...
	cd services/api-gateway && go test ./...

test-frontend:
	cd frontend && npm ci && npm test -- --run

proto:
	@echo "Proto stubs are checked in under services/orbit-service/internal/gen"
