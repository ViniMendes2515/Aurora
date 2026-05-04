.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "Aurora Home — comandos disponíveis"
	@echo ""
	@echo "  Docker"
	@echo "    make up            Sobe todos os serviços em background"
	@echo "    make up-build      Sobe rebuilding as imagens"
	@echo "    make debug         Sobe com DEBUG=true e portas expostas (Swagger ativo)"
	@echo "    make down          Derruba todos os containers"
	@echo "    make restart       Reinicia todos os containers"
	@echo ""
	@echo "  Logs"
	@echo "    make logs          Logs de todos os serviços"
	@echo "    make logs-<svc>    Logs de um serviço  ex: make logs-auth-service"
	@echo ""
	@echo "  Build / Teste local"
	@echo "    make build         go build em todos os serviços"
	@echo "    make test          go test em todos os serviços"
	@echo "    make test-<svc>    Testa um serviço     ex: make test-auth-service"
	@echo ""
	@echo "  Swagger"
	@echo "    make swagger       Regenera docs de todos os serviços"
	@echo "    make swagger-<svc> Regenera docs de um serviço  ex: make swagger-auth-service"
	@echo ""
	@echo "  Banco de dados"
	@echo "    make db-up         Sobe só o PostgreSQL"
	@echo "    make db-down       Para o PostgreSQL"
	@echo "    make db-reset      Derruba volumes e recria o banco"
	@echo ""
	@echo "  Limpeza"
	@echo "    make clean         down + remove imagens locais e volumes"
	@echo ""

COMPOSE      := docker compose -f infra/docker-compose.yml
COMPOSE_DEBUG := $(COMPOSE) -f infra/docker-compose.debug.yml
SERVICES     := auth-service sensors-service lighting-service security-service rules-service notifications-service schedule-service

# ── Subir / Parar ────────────────────────────────────────────────────────────

up:
	$(COMPOSE) up -d

up-build:
	$(COMPOSE) up -d --build

debug:
	$(COMPOSE_DEBUG) up -d --build

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

# ── Logs ─────────────────────────────────────────────────────────────────────

logs:
	$(COMPOSE) logs -f

logs-%:
	$(COMPOSE) logs -f $*

# ── Build ─────────────────────────────────────────────────────────────────────

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		(cd services/$$svc && go build ./...) && echo "  ✓ $$svc" || echo "  ✗ $$svc FAILED"; \
	done

# ── Testes ───────────────────────────────────────────────────────────────────

test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... 2>&1 | grep -v "^\?" && cd ../..; \
	done

test-%:
	cd services/$* && go test ./... 2>&1 | grep -v "^\?"

# ── Swagger ──────────────────────────────────────────────────────────────────

swagger:
	@for svc in $(SERVICES); do \
		echo "Generating swagger for $$svc..."; \
		(cd services/$$svc && swag init -g cmd/api/main.go --parseInternal -o ./docs 2>/dev/null) && echo "  ✓ $$svc" || echo "  ✗ $$svc FAILED"; \
	done

swagger-%:
	cd services/$* && swag init -g cmd/api/main.go --parseInternal -o ./docs

# ── Banco de dados ────────────────────────────────────────────────────────────

db-up:
	$(COMPOSE) up -d postgres

db-down:
	$(COMPOSE) stop postgres

db-reset:
	$(COMPOSE) down -v
	$(COMPOSE) up -d postgres

# ── Limpeza ───────────────────────────────────────────────────────────────────

clean:
	$(COMPOSE) down --rmi local --volumes --remove-orphans

.PHONY: help up up-build debug down restart logs build test swagger swagger-% test-% logs-% db-up db-down db-reset clean
