# Requirements: Aurora — Schedule Service

**Defined:** 2026-03-25
**Core Value:** Agendamentos confiáveis que executam as ações certas no horário certo, mesmo após restart do serviço.

## v1 Requirements

### Domínio e Persistência

- [ ] **SCH-01**: Schedule entity com campos: `id`, `owner_id`, `name`, `description`, `schedule_type`, `cron_expr`, `run_at`, `time_of_day`, `days_of_week`, `timezone`, `enabled`, `action_type`, `action_target`, `action_payload`, `created_at`, `updated_at`
- [ ] **SCH-02**: Agendamentos persistidos no PostgreSQL e carregados ao iniciar o serviço (sobrevive a restarts)
- [ ] **SCH-03**: Multi-tenant: cada agendamento pertence a um `owner_id`; operações de leitura/escrita/exclusão validam propriedade
- [ ] **SCH-04**: Tabela `schedule_executions` com: `id`, `schedule_id`, `executed_at`, `status` (success/failure), `error_message`

### Tipos de Agendamento

- [ ] **SCH-05**: Agendamento por horário fixo diário (ex: todo dia às 16:00)
- [ ] **SCH-06**: Filtro por dias da semana (ex: só em dias úteis, só no fim de semana)
- [ ] **SCH-07**: Agendamento one-shot com data/hora absoluta futura (ex: daqui a 2h)
- [ ] **SCH-08**: Agendamento por expressão cron (ex: `0 8 * * 1` = toda segunda às 8h)
- [ ] **SCH-09**: Timezone por agendamento (campo IANA timezone string, ex: `America/Sao_Paulo`); execução convertida para UTC internamente

### Loop de Execução

- [ ] **SCH-10**: Scheduler loop via `gocron/v2` — registra e executa agendamentos habilitados automaticamente
- [ ] **SCH-11**: Recuperação no restart: ao subir, recarrega todos os agendamentos ativos do banco e os re-registra no scheduler
- [ ] **SCH-12**: Agendamentos one-shot já executados marcados como `completed` atomicamente (mesmo commit que o dispatch) para evitar re-execução no restart
- [ ] **SCH-13**: Goroutines de execução com timeout de contexto e worker pool limitado (evitar goroutine leak)

### Tipos de Ação

- [ ] **SCH-14**: Ação: controle de luzes via HTTP para `lighting-service` (ligar/desligar/intensidade)
- [ ] **SCH-15**: Ação: controle de alarme via HTTP para `security-service` (armar/desarmar)
- [ ] **SCH-16**: Ação: publicar evento NATS com topic e payload configuráveis
- [ ] **SCH-17**: Ação: chamada HTTP arbitrária para qualquer endpoint interno (método, URL, headers, body configuráveis)

### API REST

- [ ] **SCH-18**: `POST /schedules` — criar agendamento
- [ ] **SCH-19**: `GET /schedules` — listar agendamentos do usuário autenticado
- [ ] **SCH-20**: `GET /schedules/:id` — buscar agendamento por ID (valida propriedade)
- [ ] **SCH-21**: `PUT /schedules/:id` — atualizar agendamento
- [ ] **SCH-22**: `PATCH /schedules/:id/toggle` — habilitar/desabilitar sem deletar
- [ ] **SCH-23**: `DELETE /schedules/:id` — deletar agendamento e remover do scheduler
- [ ] **SCH-24**: `GET /schedules/:id/history` — histórico de execuções do agendamento
- [ ] **SCH-25**: `POST /schedules/preview` — retorna próximas N execuções sem criar (dry-run)

### Autenticação e Integração

- [ ] **SCH-26**: Todas as rotas protegidas por JWT via `pkg/jwt` middleware
- [ ] **SCH-27**: Serviço registrado no `infra/docker-compose.yml` com PostgreSQL e NATS
- [ ] **SCH-28**: NATS connection com reconnect handling (padrão `sensors-service`, não `rules-service`)

## v2 Requirements

### Observabilidade Avançada

- **SCH-V2-01**: Dashboard de agendamentos com próximas N execuções em tempo real
- **SCH-V2-02**: Alertas de execuções falhas via notifications-service

### Extensibilidade

- **SCH-V2-03**: Importar/exportar agendamentos em formato JSON
- **SCH-V2-04**: Templates de agendamento pré-configurados (ex: "rotina matinal")

## Out of Scope

| Feature | Reason |
|---------|--------|
| Frontend / UI de calendário | Fora do escopo — REST API é o contrato; frontend consome |
| Agendamento por geolocalização | Bounded context diferente (presence service), alta complexidade |
| Encadeamento de agendamentos (A depois de B) | Workflow orchestration, não scheduling — Temporal/Argo se necessário |
| Push notifications diretas | Delegado ao `notifications-service` via evento NATS `schedule.executed` |
| Sub-minuto precision | Use cases de smart home são todos em nível de minuto |
| Escalonamento horizontal multi-instância | Single instance + PostgreSQL advisory lock é suficiente para Aurora |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SCH-01, SCH-02, SCH-03, SCH-04 | Phase 1 — Domínio + Schema | Pending |
| SCH-05–SCH-09 | Phase 1 — Domínio + Schema | Pending |
| SCH-10–SCH-13 | Phase 2 — Scheduler Core | Pending |
| SCH-14–SCH-17 | Phase 2 — Scheduler Core | Pending |
| SCH-18–SCH-25 | Phase 3 — API REST + Repositório | Pending |
| SCH-26–SCH-28 | Phase 4 — Integração + Docker | Pending |
