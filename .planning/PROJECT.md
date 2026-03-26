# Aurora — Smart Home Automation

## What This Is

Aurora é um sistema de automação residencial inteligente composto por microsserviços em Go seguindo DDD. Controla iluminação, alarmes e sensores via ESP32, com comunicação entre serviços via NATS message broker. O próximo passo é o `schedule-service`, que permitirá agendar ações automáticas no sistema (como desligar luzes às 4pm ou desarmar alarme às 5pm).

## Core Value

Agendamentos confiáveis que executam as ações certas no horário certo, mesmo após restart do serviço.

## Requirements

### Validated

- ✓ Autenticação de usuários com JWT — serviço `auth-service` em produção
- ✓ Controle de iluminação via ESP32 — `lighting-service` com TurnOn/TurnOff/AdjustIntensity
- ✓ Controle de alarme e buzzer — `security-service` com TriggerAlarm/SilenceAlarm
- ✓ Detecção de movimento com sensores — `sensors-service` publicando eventos no NATS
- ✓ Motor de regras de automação — `rules-service` avaliando eventos e executando ações
- ✓ Persistência de notificações — `notifications-service` ouvindo todos os eventos

### Active

- [ ] Schedule-service com API REST CRUD para gerenciar agendamentos
- [ ] Agendamentos por horário fixo (ex: todos os dias às 4pm)
- [ ] Agendamentos recorrentes tipo cron (ex: toda segunda às 8h)
- [ ] Filtro por dias da semana (ex: só em dias úteis)
- [ ] Agendamentos únicos one-shot (ex: apagar luzes daqui a 2h)
- [ ] Execução de ações via NATS (publicação de eventos para outros serviços reagirem)
- [ ] Execução de ações via HTTP direto para outros serviços
- [ ] Controle de luzes como ação agendável
- [ ] Controle de alarme como ação agendável
- [ ] Agendamentos por usuário (multi-tenant, isolados por user_id)
- [ ] Persistência no PostgreSQL (sobrevive a restarts)
- [ ] Integração no docker-compose.yml

### Out of Scope

- Interface gráfica de calendário — frontend é fora do escopo deste serviço
- Agendamento por geolocalização ("quando eu chegar em casa") — complexidade alta, fora do v1
- Dependências entre agendamentos ("executar A depois de B") — workflow engine, não scheduling
- Notificações push de agendamentos executados — delegado ao notifications-service via NATS

## Context

**Codebase existente:** 6 microsserviços em Go com DDD, cada um com `domain/`, `application/`, `infrastructure/` e entry point em `cmd/api/main.go`. Compartilham `pkg/database` (PostgreSQL) e `pkg/jwt` (autenticação).

**Padrão de integração:** Serviços se comunicam via NATS (eventos assíncronos) e HTTP direto quando necessário. O schedule-service seguirá o mesmo padrão — publicar eventos NATS e/ou chamar endpoints HTTP dos serviços alvo.

**Infraestrutura:** PostgreSQL via `pkg/database.NewPostgresConnection()`, NATS para mensageria, nginx como reverse proxy, tudo orquestrado via `infra/docker-compose.yml`.

**Tecnologia:** Go, Gin framework, PostgreSQL, NATS, DDD estrito (domain layer não importa nada externo).

## Constraints

- **Tech stack**: Go + Gin + PostgreSQL + NATS — mesma stack dos outros 6 serviços
- **Arquitetura**: DDD com domain/application/infrastructure — consistência com o projeto
- **Isolamento**: `internal/` Go package — sem compartilhamento de código entre serviços além de `pkg/`
- **Autenticação**: JWT via `pkg/jwt` — mesma validação dos outros serviços

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Novo microsserviço separado | Bounded context próprio, não polui rules-service | — Pending |
| PostgreSQL para persistência | Padrão do projeto, agendamentos sobrevivem a restarts | — Pending |
| Publicação via NATS como mecanismo principal | Desacoplamento, outros serviços reagem aos eventos | — Pending |
| Multi-tenant por user_id | Sistema multi-usuário, agendamentos isolados por usuário | — Pending |

## Evolution

Este documento evolui nas transições de fase e marcos de milestone.

**Após cada transição de fase** (via `/gsd:transition`):
1. Requisitos invalidados? → Mover para Out of Scope com motivo
2. Requisitos validados? → Mover para Validated com referência à fase
3. Novos requisitos emergiram? → Adicionar em Active
4. Decisões a registrar? → Adicionar em Key Decisions
5. "What This Is" ainda preciso? → Atualizar se necessário

**Após cada milestone** (via `/gsd:complete-milestone`):
1. Revisão completa de todas as seções
2. Core Value check — ainda é a prioridade certa?
3. Auditar Out of Scope — motivos ainda válidos?
4. Atualizar Context com estado atual

---
*Last updated: 2026-03-25 after initialization*
