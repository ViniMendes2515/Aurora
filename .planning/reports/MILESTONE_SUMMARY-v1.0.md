# Milestone v1.0 — Project Summary: Aurora schedule-service

**Generated:** 2026-03-26
**Purpose:** Team onboarding and project review
**Status:** In-progress — planning complete, execution not yet started

---

## 1. Project Overview

**Aurora** is a smart home automation system composed of Go microservices following Domain-Driven Design (DDD). It controls lighting, alarms, and sensors via ESP32 devices, with inter-service communication via NATS message broker.

The **schedule-service** is the 7th microservice in the system. Its mission: let users schedule automated actions (e.g., "turn off lights at 4pm every day" or "disarm alarm at 5pm on weekdays") with the guarantee that schedules survive service restarts.

**Core value:** Reliable schedules that execute the right actions at the right time, even after service restarts.

**Target users:** Smart home owners who want time-based automation, managed via REST API and isolated per user (multi-tenant by `owner_id`).

### System Context

Six existing services are already in production:
- `auth-service` — JWT authentication
- `lighting-service` — ESP32 light control (TurnOn/TurnOff/AdjustIntensity)
- `security-service` — Alarm/buzzer control (TriggerAlarm/SilenceAlarm)
- `sensors-service` — Motion detection, publishes to NATS
- `rules-service` — Automation engine, evaluates events and executes actions
- `notifications-service` — Listens to all events, persists notifications

The schedule-service adds time-based triggering to complement the event-based `rules-service`.

---

## 2. Architecture & Technical Decisions

The schedule-service follows the identical DDD structure used by all six existing services:

```
services/schedule-service/
└── internal/
    ├── domain/          # Entities + interfaces — stdlib only, zero external imports
    ├── application/     # Use cases, operates on domain interfaces
    └── infrastructure/  # PostgreSQL, NATS, Gin HTTP, gocron scheduler
```

### Key Technical Decisions

- **`gocron/v2` as scheduler (not `robfig/cron`)**
  - **Why:** gocron/v2 has native `OneTimeJob` support; robfig/cron doesn't support one-shot and has been inactive since 2020. Chosen in roadmap planning.
  - **Phase:** Decision made during roadmap creation (pre-Phase 1)

- **NATS follows `sensors-service` pattern (not `rules-service`)**
  - **Why:** `rules-service` has a reconnection bug; `sensors-service` correctly uses `RetryOnFailedConnect` and `MaxReconnects(-1)`.
  - **Phase:** Decision from research phase

- **Atomic one-shot completion: `Disable()` in the same `Save()` as dispatch**
  - **Why:** Prevents re-execution after restart. `enabled=false` + `last_run_at IS NOT NULL` serves as the atomic one-shot completion check. No separate `completed_at` column needed.
  - **Phase:** Locked in Phase 1 migration spec

- **Timezone field in initial migration (even though UTC-only in v1)**
  - **Why:** `timezone` field cannot be added retroactively without migration risk and data backfill. IANA timezone string (`America/Sao_Paulo`) stored from day one.
  - **Phase:** Phase 1 (migration)

- **Worker pool limited to 10 goroutines with context timeout**
  - **Why:** Prevents goroutine leak documented in `rules-service/CONCERNS.md`. Each dispatch has a configurable timeout (default 30s via `DISPATCH_TIMEOUT_SECONDS`).
  - **Phase:** Phase 2 (application layer)

- **PostgreSQL advisory lock before starting gocron loop**
  - **Why:** Prevents duplicate execution in rolling deploys (two containers starting simultaneously both registering and firing the same schedule).
  - **Phase:** Phase 4 (CronRunner startup)

- **DaysFilter as integer bitmask (not JSON array)**
  - **Why:** Single `INTEGER` PostgreSQL column; O(1) day check with bitwise AND. `0` = all days (no filter). Sunday=1, Monday=2, Tuesday=4, ..., Saturday=64.
  - **Phase:** Phase 1 (domain entity)

### Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.1 |
| HTTP Framework | Gin | v1.9.1 |
| Database | PostgreSQL via `lib/pq` | v1.10.9 |
| Scheduler | gocron/v2 | v2.19.1 |
| Messaging | NATS | — |
| ID Generation | google/uuid | v1.5.0 |
| Auth | pkg/jwt (shared) | JWT v5.3.0 |
| HTTP Port | 8086 | — |

---

## 3. Phases Delivered

| Phase | Name | Status | Goal |
|-------|------|--------|------|
| 1 | Domínio + Schema | Not started (planned) | Domain entities, interfaces, PostgreSQL migration — pure Go, zero external imports |
| 2 | Camada de Aplicação | Not started (planned) | CRUD use cases, toggle, execution, recovery, worker pool |
| 3 | Infraestrutura — Repositório + API REST | Not started (planned) | PostgreSQL repository, Gin handlers, 8 REST routes, JWT middleware |
| 4 | Scheduler Core + Integração | Not started (planned) | CronRunner (gocron/v2), NATS publisher, HTTPActionClient, docker-compose |

**Phase 1 is the most detailed** — it has full PLAN.md files for all 4 plans and a RESEARCH.md with code examples, pitfalls, and validation strategy.

### Phase 1 Plans (Ready to Execute)

| Plan | File | Description |
|------|------|-------------|
| 1.1 | `01-01-PLAN.md` | Entities: Schedule, ScheduleExecution, Action value object |
| 1.2 | `01-02-PLAN.md` | Interfaces and domain errors |
| 1.3 | `01-03-PLAN.md` | PostgreSQL migration (schedules + schedule_executions) |
| 1.4 | `01-04-PLAN.md` | Domain unit tests |

---

## 4. Requirements Coverage

All 28 requirements are **pending** (planning phase — no code written yet).

### Phase 1 — Domain + Schema (SCH-01 to SCH-09)

- ⬜ **SCH-01**: Schedule entity with all fields (id, owner_id, name, description, schedule_type, cron_expr, run_at, days_of_week, timezone, enabled, action fields, timestamps)
- ⬜ **SCH-02**: Schedules persisted in PostgreSQL, loaded on startup (survives restarts)
- ⬜ **SCH-03**: Multi-tenant: each schedule belongs to `owner_id`; read/write/delete validate ownership
- ⬜ **SCH-04**: `schedule_executions` table with id, schedule_id, executed_at, status, error_message
- ⬜ **SCH-05**: Fixed daily time schedule (e.g., every day at 16:00)
- ⬜ **SCH-06**: Days-of-week filter (e.g., weekdays only, weekends only)
- ⬜ **SCH-07**: One-shot schedule with absolute future datetime
- ⬜ **SCH-08**: Cron expression schedule (e.g., `0 8 * * 1`)
- ⬜ **SCH-09**: Per-schedule timezone (IANA string); execution converted to UTC internally

### Phase 2 — Application Layer (SCH-10 to SCH-17)

- ⬜ **SCH-10**: Scheduler loop via gocron/v2
- ⬜ **SCH-11**: Restart recovery — reload all active schedules on startup
- ⬜ **SCH-12**: One-shot atomically marked completed (same commit as dispatch)
- ⬜ **SCH-13**: Goroutine timeout + bounded worker pool
- ⬜ **SCH-14**: Light control action via HTTP → lighting-service
- ⬜ **SCH-15**: Alarm control action via HTTP → security-service
- ⬜ **SCH-16**: NATS event publish action
- ⬜ **SCH-17**: Arbitrary HTTP action (configurable method, URL, headers, body)

### Phase 3 — REST API (SCH-18 to SCH-25)

- ⬜ **SCH-18**: `POST /schedules` — create
- ⬜ **SCH-19**: `GET /schedules` — list (authenticated user only)
- ⬜ **SCH-20**: `GET /schedules/:id` — get by ID (validates ownership)
- ⬜ **SCH-21**: `PUT /schedules/:id` — update
- ⬜ **SCH-22**: `PATCH /schedules/:id/toggle` — enable/disable without deleting
- ⬜ **SCH-23**: `DELETE /schedules/:id` — delete and remove from scheduler
- ⬜ **SCH-24**: `GET /schedules/:id/history` — execution history
- ⬜ **SCH-25**: `POST /schedules/preview` — next N executions dry-run

### Phase 4 — Integration + Docker (SCH-26 to SCH-28)

- ⬜ **SCH-26**: All routes protected by JWT via `pkg/jwt` middleware
- ⬜ **SCH-27**: Service registered in `infra/docker-compose.yml` with PostgreSQL + NATS
- ⬜ **SCH-28**: NATS connection with reconnect handling (sensors-service pattern)

### Out of Scope (v1)

| Feature | Reason |
|---------|--------|
| Frontend / calendar UI | REST API is the contract; frontend consumes it |
| Geolocation-based scheduling | Different bounded context (presence service), high complexity |
| Schedule chaining (A then B) | Workflow orchestration (Temporal/Argo), not scheduling |
| Direct push notifications | Delegated to `notifications-service` via `schedule.executed` NATS event |
| Sub-minute precision | Smart home use cases are all at the minute level |
| Horizontal scaling (multi-instance) | Single instance + PostgreSQL advisory lock is sufficient for Aurora |

---

## 5. Key Decisions Log

| ID | Decision | Phase | Rationale |
|----|----------|-------|-----------|
| D-01 | gocron/v2 as scheduler | Roadmap | Native OneTimeJob; robfig/cron inactive since 2020, no one-shot support |
| D-02 | NATS pattern: follow sensors-service | Roadmap | rules-service has reconnection bug; sensors-service uses RetryOnFailedConnect + MaxReconnects(-1) |
| D-03 | One-shot atomicity: Disable() in same Save() as dispatch | Phase 1 | Prevents re-execution on restart; no separate completed_at column |
| D-04 | Timezone in Phase 1 migration (UTC-only v1) | Phase 1 | Can't add retroactively without migration risk; validate IANA at API boundary from Phase 3 |
| D-05 | Worker pool capped at 10 goroutines | Phase 2 | Prevents goroutine leak documented in rules-service |
| D-06 | PostgreSQL advisory lock before gocron loop | Phase 4 | Prevents duplicate execution in rolling deploys |
| D-07 | DaysFilter as integer bitmask | Phase 1 | Single INTEGER column; O(1) bitwise AND check; 0 = all days sentinel |
| D-08 | Cron validation: presence-only in domain, syntax in infrastructure | Phase 1 | Domain must have zero external imports; cron parser belongs in infra/application |
| D-09 | New separate microservice (not extending rules-service) | Roadmap | Own bounded context; time-based vs event-based are distinct trigger models |

---

## 6. Tech Debt & Deferred Items

### Known Open Questions (from Phase 1 Research)

1. **`silence_alarm` HTTP endpoint URL** — needs to be verified in `security-service/internal/infrastructure/http/routes.go` before implementing `HTTPActionClient` in Phase 4.

2. **gocron/v2 `OneTimeJob` persistence API** — verify the exact API for registering a one-shot job with a specific `next_run_at` before coding Phase 4. API may differ from docs.

3. **PostgreSQL advisory lock key value** — a specific integer key needs to be chosen and documented before Phase 4 implementation.

### Deferred to v2

- Dashboard with real-time next-N-executions view
- Failure alerts via notifications-service
- Import/export schedules as JSON
- Pre-configured schedule templates ("morning routine")

### Anti-patterns to Watch

- **External imports in domain package**: Any `import "github.com/..."` in `internal/domain/*.go` is a violation.
- **SQL types in domain entities**: `sql.NullString`/`sql.NullTime` belong in the repository layer, not domain structs.
- **`time.Now().Weekday()` in IsActiveOnDay**: Must use `time.Now().In(loc)` where `loc = time.LoadLocation(s.Timezone)` — otherwise UTC timezone mismatch breaks Brazilian users (UTC-3).
- **`UserID` vs `OwnerID`**: ARCHITECTURE.md uses `UserID` in some examples — this is incorrect. Use `OwnerID` throughout (field name, DB column: `owner_id`).

---

## 7. Getting Started

### Project Structure

```
/home/vini/Workspace/Aurora/
├── services/
│   ├── rules-service/         # Reference service — follow its domain/ pattern
│   ├── sensors-service/       # Reference for NATS connection pattern
│   └── schedule-service/      # (to be created — this milestone)
├── pkg/
│   ├── database/              # NewPostgresConnection() — shared
│   └── jwt/                   # JWT validation middleware — shared
└── infra/
    ├── docker-compose.yml     # Add schedule-service here in Phase 4
    └── nginx/                 # Route /api/schedules → schedule-service:8086
```

### Reference Files for Implementers

| What to look at | Why |
|----------------|-----|
| `services/rules-service/internal/domain/rule.go` | Authoritative entity pattern (exported fields, constructor with validation, BelongsTo) |
| `services/rules-service/internal/domain/rule_test.go` | Authoritative test pattern (package domain_test, table-driven, Portuguese names) |
| `services/sensors-service/internal/infrastructure/messaging/nats_connection.go` | NATS reconnect pattern to follow |
| `.planning/phases/01-dominio-schema/01-RESEARCH.md` | Phase 1 code examples, pitfalls, validation strategy |
| `.planning/phases/01-dominio-schema/01-01-PLAN.md` through `01-04-PLAN.md` | Ready-to-execute implementation plans |

### Run Tests (once Phase 1 is implemented)

```bash
cd services/schedule-service
go test ./internal/domain/...    # After every task commit
go test ./...                    # After every plan wave
```

### Start Executing

The project is ready to begin execution. Phase 1 has complete plans:

```bash
/gsd:execute-phase   # Execute Phase 1 plans
```

---

## Stats

- **Timeline:** 2026-03-25 → in progress
- **Phases:** 0 / 4 complete (planning done, execution not started)
- **Commits:** 12 (10 planning/docs + 2 pre-existing)
- **Files changed:** 30 files (+7,497 lines)
- **Contributors:** vinimendes
- **Requirements:** 28 v1 requirements mapped, 0 implemented, 0 out of scope
