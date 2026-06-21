# ============================================================================
# Tinta Backend · Makefile
# Comandos principales del monorepo. Ejecutar `make help` para ver la lista.
# ============================================================================

.PHONY: help up down restart logs build keys \
        migrate-identity migrate-community migrate-recommendations migrate-all \
        sqlc-identity sqlc-community sqlc-recommendations sqlc-all \
        seed test tidy clean

DATABASE_URL_IDENTITY        ?= postgres://tinta:tinta_dev_pass@localhost:5432/tinta?sslmode=disable&search_path=identity
DATABASE_URL_COMMUNITY       ?= postgres://tinta:tinta_dev_pass@localhost:5432/tinta?sslmode=disable&search_path=community
DATABASE_URL_RECOMMENDATIONS ?= postgres://tinta:tinta_dev_pass@localhost:5432/tinta?sslmode=disable&search_path=recommendations

# Default: muestra ayuda
help: ## Muestra esta lista de comandos
	@echo "Tinta Backend - comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}'

# ---------- Infraestructura ----------
up: ## Levanta toda la infraestructura (Postgres, Redis, servicios)
	docker compose up -d
	@echo ""
	@echo "Servicios disponibles:"
	@echo "  Postgres:        localhost:5432"
	@echo "  Redis:           localhost:6379"
	@echo "  Identity API:    http://localhost:8001"
	@echo "  Community API:   http://localhost:8002"
	@echo "  Recommendations: http://localhost:8003"

down: ## Detiene y elimina los contenedores
	docker compose down

restart: down up ## Reinicia toda la infraestructura

logs: ## Muestra logs en vivo (usar SERVICE=identity para filtrar)
	@if [ -z "$(SERVICE)" ]; then docker compose logs -f; else docker compose logs -f $(SERVICE); fi

build: ## Reconstruye las imágenes de los servicios Go
	docker compose build

# ---------- Llaves JWT ----------
keys: ## Genera el par de llaves JWT (RS256) si no existen
	@mkdir -p keys
	@if [ ! -f keys/jwt_private.pem ]; then \
		openssl genrsa -out keys/jwt_private.pem 2048 && \
		openssl rsa -in keys/jwt_private.pem -pubout -out keys/jwt_public.pem && \
		echo "Llaves JWT generadas en keys/"; \
	else \
		echo "Las llaves JWT ya existen, no se sobrescriben"; \
	fi

# ---------- Migraciones ----------
migrate-identity: ## Aplica migraciones al esquema 'identity'
	migrate -path services/identity/migrations -database "$(DATABASE_URL_IDENTITY)" up

migrate-community: ## Aplica migraciones al esquema 'community'
	migrate -path services/community/migrations -database "$(DATABASE_URL_COMMUNITY)" up

migrate-recommendations: ## Aplica migraciones al esquema 'recommendations'
	migrate -path services/recommendations/migrations -database "$(DATABASE_URL_RECOMMENDATIONS)" up

migrate-all: migrate-identity migrate-community migrate-recommendations ## Aplica todas las migraciones

# ---------- sqlc ----------
sqlc-identity: ## Regenera código Go desde queries SQL de identity
	cd services/identity && sqlc generate

sqlc-community: ## Regenera código Go desde queries SQL de community
	cd services/community && sqlc generate

sqlc-recommendations: ## Regenera código Go desde queries SQL de recommendations
	cd services/recommendations && sqlc generate

sqlc-all: sqlc-identity sqlc-community sqlc-recommendations ## Regenera todo sqlc

# ---------- Seed (datos iniciales) ----------
seed: ## Inserta los 4 admins iniciales (adrian, diego, gael, system)
	@echo "Insertando admins iniciales..."
	@cd scripts/seed && go run main.go

# ---------- Calidad ----------
test: ## Corre los tests de todos los servicios
	cd services/identity && go test ./...
	cd services/community && go test ./...
	cd services/recommendations && go test ./...

tidy: ## go mod tidy en todos los servicios
	cd services/identity && go mod tidy
	cd services/community && go mod tidy
	cd services/recommendations && go mod tidy
	cd shared && go mod tidy
	cd scripts/seed && go mod tidy

clean: ## Limpia binarios y caches
	rm -rf services/*/bin
	go clean -cache
