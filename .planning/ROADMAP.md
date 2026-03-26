# Roadmap: Aurora — schedule-service

**Projeto:** Aurora smart home — schedule-service (7º microsserviço)
**Atualizado:** 2026-03-25
**Granularidade:** Padrão (4 fases, 3–5 planos cada)
**Cobertura:** 28/28 requisitos v1 mapeados

---

## Fases

- [ ] **Fase 1: Domínio + Schema** — Entidades puras, interfaces de repositório, value objects e migração PostgreSQL com todos os campos obrigatórios
- [ ] **Fase 2: Camada de Aplicação** — Use cases CRUD, toggle, execução e recovery; interfaces ActionDispatcher e SchedulerRegistry
- [ ] **Fase 3: Infraestrutura — Repositório + API REST** — Implementação PostgreSQL, handlers Gin, rotas JWT e servidor HTTP na porta 8086
- [ ] **Fase 4: Scheduler Core + Integração** — CronRunner com gocron/v2, NATS publisher, HTTPActionClient, CompositeDispatcher e docker-compose

---

## Detalhes das Fases

### Fase 1: Domínio + Schema

**Goal**: A camada de domínio do schedule-service está completa, testada e estável — todas as entidades, interfaces e a migração PostgreSQL estão definidas com os campos que não podem ser adicionados retroativamente (timezone, schedule_type, completed_at, last_executed_at). Nenhuma outra camada depende de infraestrutura; o domínio compila e é testável de forma isolada.

**Depends on**: Nada — esta é a fase fundacional.

**Requirements**: SCH-01, SCH-02, SCH-03, SCH-04, SCH-05, SCH-06, SCH-07, SCH-08, SCH-09

**Success Criteria** (o que deve ser VERDADE ao término):
  1. Usuário consegue criar um `Schedule` válido com cada um dos quatro tipos (cron diário, days_of_week filter, one-shot absoluto, expressão cron) e recebe erro claro se os campos obrigatórios estiverem ausentes
  2. Um agendamento pertence exatamente a um `owner_id` e o método `BelongsTo()` rejeita acesso de outro usuário
  3. As tabelas `schedules` e `schedule_executions` são criadas na migração com todos os campos: `timezone`, `schedule_type`, `days_filter`, `last_run_at`, `next_run_at`, `run_at` — sem necessidade de migração adicional nas próximas fases
  4. As interfaces `ScheduleRepository`, `ScheduleExecutionRepository`, `ActionDispatcher` e `SchedulerRegistry` estão definidas no pacote domain sem importar nenhuma biblioteca externa
  5. Testes unitários cobrem: validação de cron expression inválida, one-shot sem `run_at`, ação com target inválido e filtragem de dias da semana

**Plans**: 4 planos

Plans:
- [x] 01-01-PLAN.md — Entidades do dominio (Schedule, ScheduleExecution, Action)
- [x] 01-02-PLAN.md — Interfaces e erros do dominio (repositories, dispatcher, registry, errors)
- [x] 01-03-PLAN.md — Migracao PostgreSQL (schedules + schedule_executions)
- [ ] 01-04-PLAN.md — Testes da camada de dominio

**Planos detalhados:**

**Plano 1.1 — Entidades do domínio**
Criar `internal/domain/schedule.go` com a entity `Schedule` (campos: ID, OwnerID, Name, Description, ScheduleType, CronExpr, RunAt, DaysFilter, Timezone, Action, Enabled, CreatedAt, UpdatedAt, LastRunAt, NextRunAt). Criar `internal/domain/schedule_execution.go` com `ScheduleExecution` (ID, ScheduleID, ExecutedAt, Status, ErrorMessage). Criar `internal/domain/action.go` com o value object `Action` (Target: nats|http, ActionType, DeviceID, Payload). Definir os tipos `ScheduleType` ("cron" | "one_shot"), `ExecutionStatus` ("success" | "failed" | "skipped"), `ActionTarget` e `ActionType`. Implementar `NewSchedule()`, `NewScheduleExecution()` e `NewAction()` com validação. Adicionar métodos de domínio: `BelongsTo()`, `IsActiveOnDay()`, `Disable()`, `RecordRun()`.

**Plano 1.2 — Interfaces e erros do domínio**
Criar `internal/domain/repository.go` com as interfaces `ScheduleRepository` (Save, FindByID, FindByOwnerID, FindAllEnabled, Delete) e `ScheduleExecutionRepository` (Save, FindByScheduleID). Criar `internal/domain/dispatcher.go` com a interface `ActionDispatcher` (Dispatch). Criar `internal/domain/scheduler_registry.go` com a interface `SchedulerRegistry` (Register, Unregister). Criar `internal/domain/errors.go` com as variáveis de erro: `ErrScheduleNotFound`, `ErrScheduleAccessDenied`, `ErrInvalidSchedule`, `ErrMissingCronExpression`, `ErrMissingRunAt`, `ErrInvalidAction`, `ErrPublishFailed`.

**Plano 1.3 — Migração PostgreSQL**
Criar `internal/infrastructure/migrations/schedule_migrations.go` com a função `RunScheduleMigrations(db *sql.DB) error`. A migração cria a tabela `schedules` com os campos: id (VARCHAR 36 PK), owner_id, name, description, schedule_type, cron_expr, run_at (TIMESTAMP nullable), days_filter (INTEGER bitmask), timezone (VARCHAR 64, default 'UTC'), action_target, action_type, action_device_id, action_payload (TEXT), enabled (BOOLEAN), created_at, updated_at, last_run_at (nullable), next_run_at (nullable). Cria a tabela `schedule_executions` com FK referenciando schedules(id) ON DELETE CASCADE. Cria os índices `idx_schedules_owner_id` e `idx_schedules_enabled` e `idx_executions_schedule_id`.

**Plano 1.4 — Testes da camada de domínio**
Criar `internal/domain/schedule_test.go` cobrindo: criação com cron expression válida, criação com cron expression inválida retorna `ErrMissingCronExpression`, criação one-shot sem `run_at` retorna `ErrMissingRunAt`, `BelongsTo()` com owner_id correto e incorreto, `IsActiveOnDay()` com DaysFilter = 0 (todos os dias) e com bitmask específico, `Disable()` muda Enabled para false. Criar `internal/domain/action_test.go` cobrindo: `NewAction()` com target inválido retorna `ErrInvalidAction`.

---

### Fase 2: Camada de Aplicação

**Goal**: Todos os use cases do schedule-service estão implementados e testáveis com mocks — criar, listar, buscar, atualizar, toggle, deletar, executar e recuperar histórico. O `ScheduleService` nunca importa bibliotecas externas, opera apenas sobre interfaces de domínio. A lógica de recovery pós-restart (recarregar agendamentos ativos) e a atomicidade do one-shot (completed_at no mesmo commit do dispatch) estão no nível de aplicação.

**Depends on**: Fase 1 (entidades, interfaces e erros de domínio definidos)

**Requirements**: SCH-10, SCH-11, SCH-12, SCH-13, SCH-14, SCH-15, SCH-16, SCH-17

**Success Criteria** (o que deve ser VERDADE ao término):
  1. Chamar `CreateSchedule` persiste o agendamento e registra imediatamente no scheduler (via `SchedulerRegistry.Register`) — o agendamento fica ativo sem necessitar restart
  2. Após restart simulado, `LoadAllEnabled()` retorna todos os agendamentos ativos do repositório, prontos para re-registro no CronRunner
  3. Executar um agendamento one-shot resulta em: ação despachada, `ScheduleExecution` gravado, agendamento marcado como disabled — tudo em uma única operação atômica
  4. Goroutines de dispatch têm timeout de contexto configurável e o pool está limitado a no máximo 10 goroutines simultâneas — sem goroutine leak possível
  5. `ListExecutions` retorna o histórico de execuções de um agendamento, filtrado por owner_id para garantir isolamento multi-tenant

**Plans**: TBD

**Planos detalhados:**

**Plano 2.1 — ScheduleService: use cases CRUD**
Criar `internal/application/schedule_service.go` com a struct `ScheduleService` e constructor `NewScheduleService(scheduleRepo, executionRepo, dispatcher, registry)`. Implementar os use cases: `CreateSchedule(req CreateScheduleRequest) (*ScheduleResponse, error)` — valida, cria entidade, persiste via repo, registra no SchedulerRegistry; `GetSchedule(id, ownerID string)` — busca e verifica propriedade; `ListSchedules(ownerID string)` — lista todos do usuário; `UpdateSchedule(id, ownerID string, req)` — busca, atualiza, persiste e re-registra no scheduler; `DeleteSchedule(id, ownerID string)` — verifica propriedade, deleta do repo e chama `SchedulerRegistry.Unregister`. Definir DTOs: `CreateScheduleRequest`, `UpdateScheduleRequest`, `ScheduleResponse` com tags json em snake_case.

**Plano 2.2 — ScheduleService: toggle e recovery**
Implementar `ToggleSchedule(id, ownerID string) (*ScheduleResponse, error)` — inverte o campo Enabled, persiste e chama Register ou Unregister conforme o novo estado. Implementar `LoadAllEnabled() ([]*domain.Schedule, error)` — delega para `ScheduleRepository.FindAllEnabled()`, usado pelo CronRunner na inicialização. Implementar lógica de recovery: ao subir, CronRunner chama `LoadAllEnabled()` e para cada agendamento due dentro de uma janela de graça configurável (padrão 5 minutos via env `MISSED_EXECUTION_GRACE_SECONDS`), chama `ExecuteSchedule` imediatamente antes de iniciar o loop gocron.

**Plano 2.3 — ScheduleService: execução e atomicidade one-shot**
Implementar `ExecuteSchedule(scheduleID string) error` — busca o Schedule, verifica `IsActiveOnDay(time.Now().Weekday())` (grava Skipped se false), chama `ActionDispatcher.Dispatch(schedule)`, grava `ScheduleExecution` com status success ou failed. Se `ScheduleType == one_shot`: chama `schedule.Disable()` e persiste em um único `Save()` antes de retornar — garantindo que restart não re-execute o one-shot. Chamar `schedule.RecordRun()` e persistir `last_run_at` / `next_run_at`. Implementar `ListExecutions(scheduleID, ownerID string, limit int) ([]*ExecutionResponse, error)` com verificação de propriedade.

**Plano 2.4 — Worker pool e controle de goroutines**
Implementar bounded worker pool no `ScheduleService` usando channel-based semaphore com capacidade máxima de 10 (configurável via constante `MaxConcurrentDispatches`). Cada chamada a `ExecuteSchedule` via CronRunner adquire um slot do semaphore com `context.WithTimeout` (padrão 30 segundos via env `DISPATCH_TIMEOUT_SECONDS`). Garantir que o contexto seja cancelado e o slot liberado mesmo em caso de panic. Adicionar log com prefixo `[Scheduler]` em falhas de dispatch.

**Plano 2.5 — Testes da camada de aplicação com mocks**
Criar `internal/application/schedule_service_test.go` com mocks das interfaces `ScheduleRepository`, `ScheduleExecutionRepository`, `ActionDispatcher` e `SchedulerRegistry`. Cobrir: `CreateSchedule` chama `Register`; `DeleteSchedule` chama `Unregister`; `ToggleSchedule` desabilita e chama `Unregister`; `ExecuteSchedule` one-shot chama `Disable()` e grava execution; `ExecuteSchedule` com IsActiveOnDay falso grava status `skipped`; `LoadAllEnabled` retorna lista do repositório.

---

### Fase 3: Infraestrutura — Repositório + API REST

**Goal**: O schedule-service tem uma API REST funcional acessível via HTTP — todas as 8 rotas respondem corretamente com autenticação JWT, e o repositório PostgreSQL persiste e recupera agendamentos e execuções com as queries corretas. Um usuário autenticado consegue criar, listar, buscar, atualizar, alternar e deletar agendamentos, e consultar histórico de execuções e preview de próximas execuções, sem nenhuma lógica de negócio no handler.

**Depends on**: Fase 2 (ScheduleService com todos os use cases implementados)

**Requirements**: SCH-18, SCH-19, SCH-20, SCH-21, SCH-22, SCH-23, SCH-24, SCH-25

**Success Criteria** (o que deve ser VERDADE ao término):
  1. `POST /schedules` com JWT válido e body correto retorna 201 com o agendamento criado; sem JWT retorna 401
  2. `GET /schedules` retorna apenas os agendamentos do usuário autenticado — outro usuário não vê os registros alheios
  3. `PATCH /schedules/:id/toggle` inverte o estado enabled/disabled sem deletar o agendamento
  4. `GET /schedules/:id/history` retorna a lista de execuções do agendamento, validando que o agendamento pertence ao usuário
  5. `POST /schedules/preview` retorna as próximas N execuções calculadas sem criar nenhum registro no banco

**Plans**: TBD

**Planos detalhados:**

**Plano 3.1 — PostgresScheduleRepository**
Criar `internal/infrastructure/repository/postgres_schedule_repository.go` implementando `domain.ScheduleRepository`. Constructor: `NewPostgresScheduleRepository(db *sql.DB)`. Implementar: `Save` com upsert `INSERT ... ON CONFLICT (id) DO UPDATE SET ...` (todos os campos); `FindByID` com SELECT completo; `FindByOwnerID` com WHERE owner_id = $1 ORDER BY created_at DESC; `FindAllEnabled` com WHERE enabled = true (usado pelo CronRunner no startup); `Delete` com DELETE WHERE id = $1. Criar também `PostgresScheduleExecutionRepository` no mesmo arquivo implementando `domain.ScheduleExecutionRepository`: `Save` com INSERT; `FindByScheduleID` com WHERE schedule_id = $1 ORDER BY executed_at DESC LIMIT $2.

**Plano 3.2 — Handlers Gin**
Criar `internal/infrastructure/http/handlers.go` com a struct `Handlers` e constructor `NewHandlers(service *application.ScheduleService)`. Implementar os handlers: `CreateSchedule`, `ListSchedules`, `GetSchedule`, `UpdateSchedule`, `ToggleSchedule`, `DeleteSchedule`, `ListExecutions`, `PreviewSchedule`. Cada handler extrai o `userID` do contexto JWT (via `c.Get("userID")`), faz bind do JSON de entrada, chama o método correspondente no `ScheduleService` e mapeia os erros de domínio para status HTTP: `ErrScheduleNotFound` → 404, `ErrScheduleAccessDenied` → 403, `ErrInvalidSchedule` → 400. Handler `PreviewSchedule` calcula as próximas N execuções usando a expressão cron sem persistir.

**Plano 3.3 — Rotas e servidor HTTP**
Criar `internal/infrastructure/http/routes.go` registrando todas as rotas no grupo protegido pelo middleware JWT: `POST /schedules`, `GET /schedules`, `GET /schedules/:id`, `PUT /schedules/:id`, `PATCH /schedules/:id/toggle`, `DELETE /schedules/:id`, `GET /schedules/:id/history`, `POST /schedules/preview`. Registrar rota pública `GET /health`. Criar `internal/infrastructure/http/server.go` com `NewServer(port string) *gin.Engine` seguindo o padrão dos outros serviços. Criar `internal/infrastructure/security/jwt_validator.go` usando `pkg/jwt` com o mesmo padrão dos outros 6 serviços.

**Plano 3.4 — Wiring parcial em main.go (sem scheduler)**
Criar `services/schedule-service/cmd/api/main.go` com o padrão de `getEnv()` e wiring parcial: carregar env vars (SERVER_PORT=8086, JWT_SECRET, DB_*), conectar PostgreSQL via `pkg/database.NewPostgresConnection()`, executar `RunScheduleMigrations(db)`, instanciar `PostgresScheduleRepository`, instanciar `ScheduleService` com dispatcher e registry placeholder (nil por ora), instanciar `Handlers`, registrar rotas, subir servidor Gin. Criar `go.mod` do serviço com as dependências: gin v1.9.1, lib/pq v1.10.9, uuid v1.5.0, jwt/v5 v5.3.0, gocron/v2 v2.19.1.

---

### Fase 4: Scheduler Core + Integração

**Goal**: O schedule-service está completo e rodando no docker-compose: o CronRunner com gocron/v2 registra e executa agendamentos persistidos, o NATS publisher envia eventos `schedule.executed` seguindo o padrão correto do sensors-service, o HTTPActionClient chama lighting-service e security-service, e o CompositeDispatcher roteia entre os dois. Após restart do container, todos os agendamentos ativos são recarregados e agendamentos missed dentro da janela de graça são re-executados.

**Depends on**: Fase 3 (main.go, repositórios e handlers funcionais)

**Requirements**: SCH-10, SCH-11, SCH-12, SCH-13, SCH-26, SCH-27, SCH-28

**Success Criteria** (o que deve ser VERDADE ao término):
  1. Criar um agendamento cron via `POST /schedules` e aguardar o horário — a ação é executada e um `ScheduleExecution` com status "success" fica visível em `GET /schedules/:id/history`
  2. Reiniciar o container `docker compose restart schedule-service` — todos os agendamentos ativos são recarregados e voltam a disparar no horário correto
  3. Um agendamento one-shot dispara uma única vez e aparece como `enabled: false` após execução — não re-executa após restart
  4. Agendamentos com `action_target: "nats"` publicam no tópico `schedule.executed`; agendamentos com `action_target: "http"` chamam lighting-service ou security-service corretamente
  5. O serviço `schedule-service` aparece em `docker compose ps` com estado "Up" e responde `GET /health` com 200

**Plans**: TBD

**Planos detalhados:**

**Plano 4.1 — NATSPublisher e HTTPActionClient**
Criar `internal/infrastructure/messaging/nats_publisher.go` seguindo o padrão de `sensors-service/infrastructure/messaging/nats_connection.go` (com `RetryOnFailedConnect`, `MaxReconnects(-1)`, disconnect/reconnect handlers). Implementar `NATSPublisher.Dispatch(schedule *domain.Schedule) error` que publica no tópico `schedule.executed` um JSON com campos: `schedule_id`, `owner_id`, `action_type`, `device_id`, `executed_at`. Adicionar retry com backoff exponencial (3 tentativas: 100ms / 500ms / 2s) em falhas de publish. Criar `internal/infrastructure/http/action_client.go` implementando `domain.ActionDispatcher` para ações HTTP: `turn_on_light` → `POST {LIGHTING_SERVICE_URL}/lights/{deviceID}/on`, `turn_off_light` → `.../off`, `trigger_alarm` → `POST {SECURITY_SERVICE_URL}/alarms/trigger`, `silence_alarm` → `.../silence`. Timeout de 3 segundos. Header `X-Device-Key` com `DEVICE_API_KEY`. Verificar rota de silence_alarm em `services/security-service/internal/infrastructure/http/routes.go` antes de implementar.

**Plano 4.2 — CompositeDispatcher**
Criar `internal/infrastructure/dispatcher/composite_dispatcher.go` com a struct `CompositeDispatcher` que implementa `domain.ActionDispatcher`. O método `Dispatch(schedule *domain.Schedule) error` roteia baseado em `schedule.Action.Target`: `ActionTargetNATS` → delega para `NATSPublisher`, `ActionTargetHTTP` → delega para `HTTPActionClient`, qualquer outro → retorna `domain.ErrInvalidAction`. Fallback: se NATS falhar após retries, tentar HTTPActionClient com log de warning `[Scheduler] NATS indisponível, fallback para HTTP`.

**Plano 4.3 — CronRunner com gocron/v2**
Criar `internal/infrastructure/scheduler/cron_runner.go` com a struct `CronRunner` (campos: scheduler gocron.Scheduler, service *application.ScheduleService, jobsByID map[string]uuid.UUID, mu sync.Mutex). Implementar `NewCronRunner(service)`. Implementar `Register(schedule *domain.Schedule) error`: para `ScheduleTypeOneShot` usa `gocron.OneTimeJob` com o `RunAt`; para `ScheduleTypeCron` usa `gocron.CronJob` com a expressão; ambos com `gocron.WithIdentifier(uuid.MustParse(schedule.ID))`. O callback de cada job: verifica `IsActiveOnDay()` e chama `service.ExecuteSchedule(scheduleID)` dentro do worker pool. Implementar `Unregister(scheduleID string)` que remove o job pelo UUID. Implementar `LoadAndStart(ctx context.Context) error`: carrega todos os enabled via `service.LoadAllEnabled()`, registra cada um, executa recovery de missed executions (schedules due dentro de `MISSED_EXECUTION_GRACE_SECONDS`), inicia o scheduler gocron. Implementar `Stop()` para graceful shutdown.

**Plano 4.4 — main.go final e docker-compose**
Atualizar `cmd/api/main.go` com o wiring completo na ordem correta: (1) env vars, (2) PostgreSQL, (3) migrações, (4) repositórios, (5) NATSPublisher com retry de conexão, (6) HTTPActionClient, (7) CompositeDispatcher, (8) CronRunner vazio, (9) ScheduleService com todos os componentes reais, (10) `runner.LoadAndStart(ctx)`, (11) JWT validator, (12) servidor Gin. Registrar signal handler para SIGTERM que chama `runner.Stop()` antes de encerrar o HTTP server. Adicionar `schedule-service` ao `infra/docker-compose.yml` com: build context, porta 8086:8086, variáveis de ambiente (SERVER_PORT, JWT_SECRET, NATS_URL, DB_*, LIGHTING_SERVICE_URL, SECURITY_SERVICE_URL, DEVICE_API_KEY, MISSED_EXECUTION_GRACE_SECONDS, DISPATCH_TIMEOUT_SECONDS), depends_on postgres e nats. Atualizar `infra/nginx/` se necessário para rotear `/api/schedules` para `schedule-service:8086`.

---

## Progresso

| Fase | Planos completos | Status | Concluída |
|------|-----------------|--------|-----------|
| 1. Domínio + Schema | 3/4 | In Progress |  |
| 2. Camada de Aplicação | 0/5 | Não iniciada | - |
| 3. Infraestrutura — Repositório + API REST | 0/4 | Não iniciada | - |
| 4. Scheduler Core + Integração | 0/4 | Não iniciada | - |

---

## Cobertura de Requisitos

| Requisito | Fase | Status |
|-----------|------|--------|
| SCH-01 | Fase 1 — Domínio + Schema | Pendente |
| SCH-02 | Fase 1 — Domínio + Schema | Pendente |
| SCH-03 | Fase 1 — Domínio + Schema | Pendente |
| SCH-04 | Fase 1 — Domínio + Schema | Pendente |
| SCH-05 | Fase 1 — Domínio + Schema | Pendente |
| SCH-06 | Fase 1 — Domínio + Schema | Pendente |
| SCH-07 | Fase 1 — Domínio + Schema | Pendente |
| SCH-08 | Fase 1 — Domínio + Schema | Pendente |
| SCH-09 | Fase 1 — Domínio + Schema | Pendente |
| SCH-10 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-11 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-12 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-13 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-14 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-15 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-16 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-17 | Fase 2 — Camada de Aplicação | Pendente |
| SCH-18 | Fase 3 — Repositório + API REST | Pendente |
| SCH-19 | Fase 3 — Repositório + API REST | Pendente |
| SCH-20 | Fase 3 — Repositório + API REST | Pendente |
| SCH-21 | Fase 3 — Repositório + API REST | Pendente |
| SCH-22 | Fase 3 — Repositório + API REST | Pendente |
| SCH-23 | Fase 3 — Repositório + API REST | Pendente |
| SCH-24 | Fase 3 — Repositório + API REST | Pendente |
| SCH-25 | Fase 3 — Repositório + API REST | Pendente |
| SCH-26 | Fase 4 — Scheduler Core + Integração | Pendente |
| SCH-27 | Fase 4 — Scheduler Core + Integração | Pendente |
| SCH-28 | Fase 4 — Scheduler Core + Integração | Pendente |

**Total v1: 28/28 requisitos mapeados. Nenhum órfão.**

---

## Decisões Registradas

| Decisão | Justificativa |
|---------|--------------|
| gocron/v2 como scheduler (não robfig/cron) | gocron/v2 cobre OneTimeJob nativamente; robfig/cron não suporta one-shot e está inativo desde 2020 |
| NATS segue padrão sensors-service | rules-service tem bug de reconexão; sensors-service usa RetryOnFailedConnect e MaxReconnects(-1) corretamente |
| Atomicidade one-shot: Disable() no mesmo Save() do dispatch | Evita re-execução após restart; completed_at e enabled=false na mesma transação |
| Timezone na migração inicial mesmo que UTC-only no v1 | Campo não pode ser adicionado retroativamente sem migration risk; validação IANA na borda da API desde a Fase 3 |
| Worker pool limitado a 10 goroutines com timeout de contexto | Evita goroutine leak documentado em CONCERNS.md do rules-service |
| PostgreSQL advisory lock antes de iniciar o loop gocron | Evita duplicate execution em rolling deploys (dois containers sobem simultaneamente) |

---

*Roadmap criado: 2026-03-25*
