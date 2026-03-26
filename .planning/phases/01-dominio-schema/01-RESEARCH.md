# Phase 1: Domínio + Schema — Research

**Researched:** 2026-03-25
**Domain:** Go DDD domain layer — entities, value objects, interfaces, PostgreSQL migration
**Confidence:** HIGH — derived directly from existing Aurora codebase patterns and pre-existing project research

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SCH-01 | Schedule entity com campos: id, owner_id, name, description, schedule_type, cron_expr, run_at, time_of_day, days_of_week, timezone, enabled, action_type, action_target, action_payload, created_at, updated_at | Roadmap Plan 1.1 details exact struct fields; ARCHITECTURE.md has full entity code example |
| SCH-02 | Agendamentos persistidos no PostgreSQL e carregados ao iniciar o serviço | Plan 1.3 migration; FindAllEnabled() interface method in ScheduleRepository |
| SCH-03 | Multi-tenant: cada agendamento pertence a um owner_id; operações validam propriedade | BelongsTo() method + ErrScheduleAccessDenied error; mirrors rules-service Rule.BelongsTo() pattern |
| SCH-04 | Tabela schedule_executions com: id, schedule_id, executed_at, status, error_message | Plan 1.3 migration; ScheduleExecution entity in Plan 1.1 |
| SCH-05 | Agendamento por horário fixo diário (ex: todo dia às 16:00) | ScheduleType "cron" com expressão "0 16 * * *"; DaysFilter=0 (todos os dias) |
| SCH-06 | Filtro por dias da semana (ex: só em dias úteis) | DaysFilter bitmask (integer); IsActiveOnDay() method using time.Weekday bitmask shift |
| SCH-07 | Agendamento one-shot com data/hora absoluta futura | ScheduleType "one_shot" + RunAt *time.Time field; ErrMissingRunAt validation |
| SCH-08 | Agendamento por expressão cron (ex: "0 8 * * 1") | ScheduleType "cron" + CronExpr string; ErrMissingCronExpression validation |
| SCH-09 | Timezone por agendamento (campo IANA timezone string); execução convertida para UTC internamente | Timezone string field in entity + DB column (VARCHAR 64, default 'UTC'); validation via time.LoadLocation at API boundary in later phases |
</phase_requirements>

---

## Summary

Phase 1 builds the foundational domain layer for the schedule-service — the 7th microservice in the Aurora smart home system. This phase is pure Go with zero external library imports in the domain package. It produces: domain entities (Schedule, ScheduleExecution), a value object (Action), repository/dispatcher/registry interfaces, sentinel errors, a PostgreSQL migration, and unit tests.

The Aurora codebase has six existing services all following the same DDD structure: domain layer (entities + interfaces, stdlib only) → application layer (use cases) → infrastructure layer (PostgreSQL, NATS, HTTP, Gin). The schedule-service domain layer strictly follows this established pattern. The key structural reference is `services/rules-service/internal/domain/` which shows both the entity pattern (rule.go), value object pattern (schedule_time.go), and test conventions (package domain_test, table-driven tests in Portuguese).

The most critical design decision for this phase is that the migration must include all fields that cannot be added retroactively without migration risk: `timezone`, `schedule_type`, `days_filter`, `last_run_at`, `next_run_at`, `run_at`. These fields are locked from day one per the roadmap's architectural decisions.

**Primary recommendation:** Implement entities and interfaces strictly following the rules-service domain pattern (package domain, package domain_test for tests, stdlib only, no external imports), with the full PostgreSQL schema baked into migration from the start.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`time`, `errors`) | Go 1.26 (project uses 1.21+) | Entity fields, error variables, time handling | Domain layer must have zero external imports — this is an Aurora architectural rule |
| `database/sql` | stdlib | Migration function signature `RunScheduleMigrations(db *sql.DB) error` | Matches all existing migration files in the codebase |
| `github.com/lib/pq` | v1.10.9 | PostgreSQL driver (used in migration file's import side-effect) | All 6 existing services use v1.10.9; consistency required |
| `github.com/google/uuid` | v1.5.0 | UUID generation for Schedule IDs in constructors | All 6 existing services use v1.5.0 |

### Supporting (test layer only)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `testing` | stdlib | Unit test framework | All domain tests; no external test library used in existing services |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| stdlib `errors.New` | `pkg/errors` or `github.com/cockroachdb/errors` | Richer stack traces, but Aurora's existing services all use `errors.New`; no justification to diverge in domain layer |
| Integer bitmask for DaysFilter | `[]time.Weekday` slice | Bitmask fits as a single INTEGER column in PostgreSQL; matches roadmap specification |
| `*time.Time` for RunAt | `sql.NullTime` | Domain entities in Aurora use Go native types; SQL-specific types belong in repository layer |

**Installation:** No new dependencies needed in go.mod for the domain and migration layer. uuid is already a dependency of gocron/v2 (which will be added in Phase 4). Add to the new module's go.mod:

```bash
# In services/schedule-service/
go mod init aurora/services/schedule-service
# dependencies will be specified at go.mod creation time (Plan 3.4)
```

---

## Architecture Patterns

### Recommended Project Structure (Phase 1 scope)

```
services/schedule-service/
└── internal/
    ├── domain/
    │   ├── schedule.go              # Schedule entity + ScheduleType type
    │   ├── schedule_execution.go    # ScheduleExecution entity + ExecutionStatus type
    │   ├── action.go                # Action value object + ActionTarget/ActionType types
    │   ├── repository.go            # ScheduleRepository + ScheduleExecutionRepository interfaces
    │   ├── dispatcher.go            # ActionDispatcher interface
    │   ├── scheduler_registry.go    # SchedulerRegistry interface
    │   ├── errors.go                # Sentinel error variables
    │   ├── schedule_test.go         # Tests for Schedule entity
    │   └── action_test.go           # Tests for Action value object
    └── infrastructure/
        └── migrations/
            └── schedule_migrations.go  # RunScheduleMigrations(db *sql.DB) error
```

### Pattern 1: Domain Entity with Constructor Validation

The established Aurora pattern: exported struct fields (no private fields), constructor function that returns `(*Entity, error)`, methods for business logic. Tests use `package domain_test` (external test package) for black-box testing.

```go
// Source: derived from services/rules-service/internal/domain/rule.go
package domain

import (
    "errors"
    "time"
)

type Schedule struct {
    ID           string
    OwnerID      string
    Name         string
    Description  string
    ScheduleType ScheduleType
    CronExpr     string
    RunAt        *time.Time
    DaysFilter   int        // bitmask: 0=all days, Sunday=1,Monday=2,Tuesday=4,...,Saturday=64
    Timezone     string     // IANA timezone string, default "UTC"
    Action       Action
    Enabled      bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
    LastRunAt    *time.Time
    NextRunAt    *time.Time
}

func NewSchedule(id, ownerID, name, description string, scheduleType ScheduleType,
    cronExpr string, runAt *time.Time, daysFilter int, timezone string,
    action Action) (*Schedule, error) {
    if name == "" {
        return nil, ErrInvalidSchedule
    }
    if scheduleType == ScheduleTypeCron && cronExpr == "" {
        return nil, ErrMissingCronExpression
    }
    if scheduleType == ScheduleTypeOneShot && runAt == nil {
        return nil, ErrMissingRunAt
    }
    if timezone == "" {
        timezone = "UTC"
    }
    now := time.Now()
    return &Schedule{
        ID:           id,
        OwnerID:      ownerID,
        Name:         name,
        Description:  description,
        ScheduleType: scheduleType,
        CronExpr:     cronExpr,
        RunAt:        runAt,
        DaysFilter:   daysFilter,
        Timezone:     timezone,
        Action:       action,
        Enabled:      true,
        CreatedAt:    now,
        UpdatedAt:    now,
    }, nil
}
```

### Pattern 2: Domain Methods for Business Logic

```go
// Source: derived from services/rules-service/internal/domain/rule.go (BelongsTo pattern)
func (s *Schedule) BelongsTo(ownerID string) bool {
    return s.OwnerID == ownerID
}

// IsActiveOnDay checks whether schedule should run on a given weekday.
// DaysFilter bitmask: 0 = all days. Sunday=1, Monday=2, Tuesday=4, Wednesday=8,
// Thursday=16, Friday=32, Saturday=64.
// time.Weekday: Sunday=0, Monday=1, ..., Saturday=6
func (s *Schedule) IsActiveOnDay(day time.Weekday) bool {
    if s.DaysFilter == 0 {
        return true
    }
    mask := 1 << uint(day)
    return s.DaysFilter&mask != 0
}

func (s *Schedule) Disable() {
    s.Enabled = false
    s.UpdatedAt = time.Now()
}

func (s *Schedule) RecordRun(at time.Time, next *time.Time) {
    s.LastRunAt = &at
    s.NextRunAt = next
    s.UpdatedAt = at
}
```

### Pattern 3: Repository Interface in Domain Package

```go
// Source: derived from services/rules-service/internal/domain/rule.go (RuleRepository pattern)
// File: internal/domain/repository.go
package domain

type ScheduleRepository interface {
    Save(schedule *Schedule) error
    FindByID(id string) (*Schedule, error)
    FindByOwnerID(ownerID string) ([]*Schedule, error)
    FindAllEnabled() ([]*Schedule, error)
    Delete(id string) error
}

type ScheduleExecutionRepository interface {
    Save(execution *ScheduleExecution) error
    FindByScheduleID(scheduleID string, limit int) ([]*ScheduleExecution, error)
}
```

### Pattern 4: Table-Driven Tests in Portuguese

```go
// Source: services/rules-service/internal/domain/rule_test.go and schedule_time_test.go
package domain_test

import (
    "testing"
    "time"
    "aurora/services/schedule-service/internal/domain"
)

func TestNewSchedule_TipoCronSemExpressao(t *testing.T) {
    _, err := domain.NewSchedule("id1", "owner1", "Luzes sala", "",
        domain.ScheduleTypeCron, "", nil, 0, "UTC",
        domain.Action{Target: domain.ActionTargetHTTP})
    if err != domain.ErrMissingCronExpression {
        t.Errorf("esperado ErrMissingCronExpression, obteve %v", err)
    }
}

func TestSchedule_IsActiveOnDay(t *testing.T) {
    testes := []struct {
        nome       string
        daysFilter int
        dia        time.Weekday
        esperado   bool
    }{
        {"filtro zero = todos os dias, segunda", 0, time.Monday, true},
        {"filtro zero = todos os dias, domingo", 0, time.Sunday, true},
        {"bitmask dias úteis, segunda", 0b0111110, time.Monday, true},
        {"bitmask dias úteis, domingo", 0b0111110, time.Sunday, false},
        {"bitmask fim de semana, sabado", 0b1000001, time.Saturday, true},
    }
    for _, tt := range testes {
        t.Run(tt.nome, func(t *testing.T) {
            s := &domain.Schedule{DaysFilter: tt.daysFilter}
            if got := s.IsActiveOnDay(tt.dia); got != tt.esperado {
                t.Errorf("IsActiveOnDay(%v) = %v, esperado %v", tt.dia, got, tt.esperado)
            }
        })
    }
}
```

### Pattern 5: Migration Function

```go
// Source: derived from existing migration files in other Aurora services
// File: internal/infrastructure/migrations/schedule_migrations.go
package migrations

import "database/sql"

func RunScheduleMigrations(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS schedules (
            id              VARCHAR(36) PRIMARY KEY,
            owner_id        VARCHAR(255) NOT NULL,
            name            VARCHAR(255) NOT NULL,
            description     TEXT NOT NULL DEFAULT '',
            schedule_type   VARCHAR(20) NOT NULL,
            cron_expr       VARCHAR(100) NOT NULL DEFAULT '',
            run_at          TIMESTAMP,
            days_filter     INTEGER NOT NULL DEFAULT 0,
            timezone        VARCHAR(64) NOT NULL DEFAULT 'UTC',
            action_target   VARCHAR(10) NOT NULL,
            action_type     VARCHAR(50) NOT NULL,
            action_device_id VARCHAR(255) NOT NULL DEFAULT '',
            action_payload  TEXT NOT NULL DEFAULT '',
            enabled         BOOLEAN NOT NULL DEFAULT true,
            created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
            last_run_at     TIMESTAMP,
            next_run_at     TIMESTAMP
        )`,
        `CREATE INDEX IF NOT EXISTS idx_schedules_owner_id ON schedules(owner_id)`,
        `CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled)`,
        `CREATE TABLE IF NOT EXISTS schedule_executions (
            id            VARCHAR(36) PRIMARY KEY,
            schedule_id   VARCHAR(36) NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
            executed_at   TIMESTAMP NOT NULL DEFAULT NOW(),
            status        VARCHAR(20) NOT NULL,
            error_message TEXT NOT NULL DEFAULT ''
        )`,
        `CREATE INDEX IF NOT EXISTS idx_executions_schedule_id ON schedule_executions(schedule_id)`,
    }
    for _, q := range queries {
        if _, err := db.Exec(q); err != nil {
            return err
        }
    }
    return nil
}
```

### Anti-Patterns to Avoid

- **External imports in domain package:** The domain package must only import `time` and `errors` from stdlib. Never import `lib/pq`, `uuid`, `gin`, or gocron in domain files. Cron expression parsing (if added for validation) should be in infrastructure, not domain.
- **SQL types in domain entities:** Fields like `sql.NullString` or `sql.NullTime` belong in the repository implementation, not the domain entity. Use `*string` and `*time.Time` in domain structs.
- **Repository implementation in domain:** The `repository.go` file defines interfaces only. No struct implementing the interface should exist in the domain package.
- **Package `domain` for test files:** Use `package domain_test` (external test package). This is the established pattern in all six existing services and prevents domain from accidentally exposing internal state through tests.
- **time.Local in IsActiveOnDay:** The day check must use `time.Now().In(loc)` where `loc` comes from `time.LoadLocation(s.Timezone)`. Never call `time.Now().Weekday()` directly — this is Pitfall 10 from project research.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| UUID generation | Custom ID generator | `github.com/google/uuid` v1.5.0 | Collision risk; all 6 existing services already use uuid v1.5.0 |
| Error type hierarchy | Custom error types with wrapping | `errors.New` sentinel variables | Aurora's pattern is sentinel errors (ErrNotFound, ErrAccessDenied); `errors.Is` comparison works correctly; no need for custom types |
| DaysFilter encoding | JSON array `["monday","tuesday"]` | Integer bitmask | Single INTEGER PostgreSQL column; O(1) day check with bitwise AND; roadmap already specifies bitmask |
| Timezone validation | Custom regex or list lookup | `time.LoadLocation(tz)` | stdlib validates IANA names correctly; returns error for invalid strings |
| Migration idempotency | Version tracking table | `CREATE TABLE IF NOT EXISTS` | Established pattern in all existing Aurora service migrations; simpler and sufficient |

**Key insight:** The domain layer is intentionally free of library dependencies. The only "don't hand-roll" guidance here is at the infrastructure boundary (uuid for IDs, migration SQL for schema). All business logic (validation, bitmask math, timezone default) is custom Go code — that is the correct DDD approach.

---

## Common Pitfalls

### Pitfall 1: Missing Fields in Initial Migration

**What goes wrong:** A field that "might be needed later" (timezone, schedule_type, last_run_at, next_run_at) is left out of the Phase 1 migration. In Phase 3 or 4, when it's needed, adding it requires a new ALTER TABLE migration with default values, or data backfill.

**Why it happens:** Premature optimization — "we don't need timezone in v1, we'll add it later."

**How to avoid:** The roadmap explicitly locks these fields as Phase 1 requirements. The migration must include ALL columns from the roadmap specification: `timezone` (VARCHAR 64, default 'UTC'), `schedule_type`, `days_filter`, `last_run_at` (nullable), `next_run_at` (nullable), `run_at` (nullable), `completed_at` equivalent (handled via `enabled=false` for one-shot, not a separate column per roadmap).

**Warning signs:** Phase 3 or 4 plan needing to ALTER TABLE schedules — this means Phase 1 was incomplete.

### Pitfall 2: DaysFilter Bitmask Off-by-One

**What goes wrong:** `time.Weekday` in Go uses Sunday=0, Monday=1, ..., Saturday=6. A bitmask where Sunday=1 means `mask = 1 << uint(day)` works correctly (Sunday: 1<<0=1, Monday: 1<<1=2, etc.). If you instead define Sunday=0 in the bitmask (using day value directly), `1<<0` is 1 which is correct, but `days_filter=0` becomes "no days" instead of "all days."

**How to avoid:** Use 0 as the sentinel for "all days" (no filter). For any non-zero filter, `mask = 1 << uint(day)` and `daysFilter & mask != 0`. Test explicitly with Sunday (day=0, mask=1) and Saturday (day=6, mask=64).

**Warning signs:** Tests passing for Monday–Friday but failing for Sunday/Saturday edge cases.

### Pitfall 3: owner_id vs user_id Field Naming

**What goes wrong:** The roadmap, requirements, and ARCHITECTURE.md use `owner_id` consistently. The ARCHITECTURE.md code examples use `UserID` in the Go struct (field name) and `user_id` in PostgreSQL. The REQUIREMENTS.md uses `owner_id`.

**How to avoid:** Use `OwnerID` as the Go struct field name (matches `BelongsTo(ownerID string)` parameter) and `owner_id` as the PostgreSQL column name. This is what the Roadmap Plan 1.1 specifies. Do NOT use `UserID`/`user_id` — the existing ARCHITECTURE.md code examples contain this inconsistency; trust the Roadmap over ARCHITECTURE.md on field names.

**Warning signs:** Grep for `UserID` vs `OwnerID` — must be `OwnerID` throughout.

### Pitfall 4: Cron Validation in Domain vs Infrastructure

**What goes wrong:** The roadmap specifies that Phase 1 domain layer has zero external imports. But validating a cron expression (e.g., `"0 25 * * *"` with hour=25) requires a cron parser. If you try to validate the cron expression inside `NewSchedule()`, you must import `robfig/cron` or `gocron/v2` — violating the domain purity constraint.

**How to avoid:** Phase 1 domain validation is limited to presence checks: `if scheduleType == ScheduleTypeCron && cronExpr == ""`. Deep cron expression validation (parsing the expression for validity) is deferred to the infrastructure/application layer. The ROADMAP plan does not specify cron expression parsing in Phase 1 — only `ErrMissingCronExpression` for the empty case.

**Warning signs:** Any `import "github.com/..."` appearing in `internal/domain/*.go` files.

### Pitfall 5: Interfaces Not in Separate Files

**What goes wrong:** Putting all interfaces (ScheduleRepository, ScheduleExecutionRepository, ActionDispatcher, SchedulerRegistry) in a single `interfaces.go` file. This makes the domain less navigable and harder to mock.

**How to avoid:** Follow the roadmap's plan-by-plan file assignment: `repository.go` for repository interfaces, `dispatcher.go` for ActionDispatcher, `scheduler_registry.go` for SchedulerRegistry. This is the same approach rules-service takes with `RuleRepository` embedded in `rule.go`.

---

## Code Examples

### Sentinel Errors Pattern (errors.go)

```go
// Source: services/rules-service/internal/domain/rule.go (var block pattern)
package domain

import "errors"

var (
    ErrScheduleNotFound      = errors.New("schedule not found")
    ErrScheduleAccessDenied  = errors.New("access denied to schedule")
    ErrInvalidSchedule       = errors.New("invalid schedule")
    ErrMissingCronExpression = errors.New("cron expression required for cron schedule type")
    ErrMissingRunAt          = errors.New("run_at required for one_shot schedule type")
    ErrInvalidAction         = errors.New("invalid action target")
    ErrPublishFailed         = errors.New("failed to publish schedule event")
)
```

### ActionDispatcher Interface (dispatcher.go)

```go
// Source: derived from domain.BuzzerClient in security-service and domain.EventPublisher in sensors-service
// Interface defined in domain, implemented in infrastructure — standard Aurora DDD pattern
package domain

type ActionDispatcher interface {
    Dispatch(schedule *Schedule) error
}
```

### SchedulerRegistry Interface (scheduler_registry.go)

```go
// Source: ARCHITECTURE.md — defined in domain so application layer can call Register/Unregister
// without importing infrastructure package
package domain

type SchedulerRegistry interface {
    Register(schedule *Schedule) error
    Unregister(scheduleID string)
}
```

### Test Structure Pattern

```go
// Source: services/rules-service/internal/domain/rule_test.go
// Package name: domain_test (external test package — mandatory in Aurora)
// Language: Portuguese for test names
// Style: table-driven for multi-case, direct assertion for single-case
package domain_test

import (
    "testing"
    "aurora/services/schedule-service/internal/domain"
)

func TestNewSchedule_OneShotSemRunAt(t *testing.T) {
    _, err := domain.NewSchedule("id", "owner", "Nome", "",
        domain.ScheduleTypeOneShot, "", nil, 0, "UTC",
        domain.Action{Target: domain.ActionTargetHTTP})
    if err != domain.ErrMissingRunAt {
        t.Errorf("esperado ErrMissingRunAt, obteve %v", err)
    }
}

func TestNewAction_TargetInvalido(t *testing.T) {
    _, err := domain.NewAction("invalido", "turn_on_light", "light-001")
    if err != domain.ErrInvalidAction {
        t.Errorf("esperado ErrInvalidAction, obteve %v", err)
    }
}
```

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | All domain compilation and tests | Yes | 1.26.1 | — |
| PostgreSQL | Plan 1.3 migration (test requires running DB) | Not checked in CI for unit tests | — | Domain and migration code can be written without a running DB; migration tested manually or via integration test in later phases |
| `go test` | Plan 1.4 domain tests | Yes (stdlib) | Go 1.26.1 | — |

**Missing dependencies with no fallback:** None for Phase 1. Domain layer tests are pure unit tests with no external dependencies. Migration code requires a PostgreSQL instance only when executed — not at compile time.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | `testing` (stdlib) — no external test library in any existing Aurora service |
| Config file | None — `go test ./...` from service root |
| Quick run command | `go test ./internal/domain/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SCH-01 | NewSchedule creates entity with all required fields | unit | `go test ./internal/domain/... -run TestNewSchedule` | Wave 0 |
| SCH-02 | FindAllEnabled interface method exists on ScheduleRepository | compile-time check | `go build ./internal/domain/...` | Wave 0 |
| SCH-03 | BelongsTo() returns true for owner, false for other user | unit | `go test ./internal/domain/... -run TestSchedule_BelongsTo` | Wave 0 |
| SCH-04 | NewScheduleExecution creates execution record with required fields | unit | `go test ./internal/domain/... -run TestNewScheduleExecution` | Wave 0 |
| SCH-05 | NewSchedule with cron type and valid expression succeeds | unit | `go test ./internal/domain/... -run TestNewSchedule_TipoCron` | Wave 0 |
| SCH-06 | IsActiveOnDay returns correct result for bitmask filter and all-days (0) | unit | `go test ./internal/domain/... -run TestSchedule_IsActiveOnDay` | Wave 0 |
| SCH-07 | NewSchedule one_shot without run_at returns ErrMissingRunAt | unit | `go test ./internal/domain/... -run TestNewSchedule_OneShotSemRunAt` | Wave 0 |
| SCH-08 | NewSchedule cron without cron_expr returns ErrMissingCronExpression | unit | `go test ./internal/domain/... -run TestNewSchedule_TipoCronSemExpressao` | Wave 0 |
| SCH-09 | Timezone field defaults to "UTC" when empty; stored as string | unit | `go test ./internal/domain/... -run TestNewSchedule_TimezoneDefault` | Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/domain/...`
- **Per wave merge:** `go test ./...` (covers domain + any future pkg imports)
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/domain/schedule_test.go` — covers SCH-01, SCH-03, SCH-05, SCH-06, SCH-07, SCH-08, SCH-09
- [ ] `internal/domain/action_test.go` — covers SCH-01 (Action field), ErrInvalidAction
- [ ] `internal/domain/schedule_execution_test.go` — covers SCH-04 (can be minimal)
- [ ] `go.mod` for `aurora/services/schedule-service` — required before `go test` works
- [ ] Module path: `aurora/services/schedule-service` (matching all other services)

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| robfig/cron for all schedule types | gocron/v2 for OneTimeJob + CronJob | Decision in ROADMAP (2026-03-25) | One-shot jobs no longer need time.AfterFunc workaround |
| Single RuleRepository interface in rule.go | Separate repository.go, dispatcher.go, scheduler_registry.go files | Phase 1 plan | Cleaner separation; each interface in its own file |

**Deprecated/outdated:**
- ARCHITECTURE.md references `robfig/cron` as the scheduler — this was overridden by the final roadmap decision to use `gocron/v2`. Phase 1 domain layer is unaffected (no scheduler library in domain), but the CronRunner in Phase 4 will use gocron/v2, not robfig/cron.
- ARCHITECTURE.md uses `UserID` in some struct examples — the roadmap and requirements consistently use `OwnerID`. Use `OwnerID`.

---

## Open Questions

1. **Cron expression validation in domain vs application layer**
   - What we know: Phase 1 domain has no external imports; gocron/v2 parser belongs in infrastructure
   - What's unclear: Should `NewSchedule` return an error for a syntactically invalid cron string (e.g., `"0 25 * * *"`)? The roadmap test plan only specifies `ErrMissingCronExpression` for empty strings.
   - Recommendation: Phase 1 validates presence only (empty check). Phase 3 (API handler) validates cron expression syntax using gocron/v2's parser before calling `NewSchedule`. Document this clearly in code comments.

2. **`completed_at` column vs `enabled=false` for one-shot atomicity**
   - What we know: ROADMAP says "completed_at and enabled=false in the same Save()" — but the migration spec does not include a `completed_at` column; it uses `enabled=false` + `last_run_at`
   - What's unclear: Is `completed_at` a separate column or is `last_run_at` used as the completion indicator for one-shot?
   - Recommendation: Follow the migration spec verbatim (no `completed_at` column). Use `enabled=false` + `last_run_at IS NOT NULL` as the atomic one-shot completion check. The recovery logic in Phase 2 checks `WHERE enabled = true` which naturally excludes completed one-shots.

---

## Sources

### Primary (HIGH confidence)

- `/home/vini/Workspace/Aurora/.planning/research/ARCHITECTURE.md` — full entity code examples, interface definitions, migration SQL, build order
- `/home/vini/Workspace/Aurora/.planning/research/PITFALLS.md` — 12 pitfalls with phase assignments; Pitfalls 3, 6, 7, 10 directly affect Phase 1
- `/home/vini/Workspace/Aurora/.planning/research/STACK.md` — library versions, rationale for choices
- `/home/vini/Workspace/Aurora/.planning/ROADMAP.md` — locked field names, migration column spec, test scenarios
- `/home/vini/Workspace/Aurora/services/rules-service/internal/domain/rule.go` — authoritative entity pattern
- `/home/vini/Workspace/Aurora/services/rules-service/internal/domain/rule_test.go` — authoritative test pattern (package domain_test, table-driven, Portuguese names)
- `/home/vini/Workspace/Aurora/services/rules-service/internal/domain/schedule_time.go` — value object validation pattern

### Secondary (MEDIUM confidence)

- `/home/vini/Workspace/Aurora/.planning/research/FEATURES.md` — feature landscape confirming what is table stakes vs differentiator
- `/home/vini/Workspace/Aurora/.planning/research/SUMMARY.md` — confirmed decisions list

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — directly derived from 6 existing services with identical stack
- Architecture: HIGH — ARCHITECTURE.md provides entity code; rules-service provides file-by-file precedent
- Pitfalls: HIGH — PITFALLS.md is project-specific research based on codebase analysis
- Test patterns: HIGH — 18 existing test files across 6 services all follow the same conventions

**Research date:** 2026-03-25
**Valid until:** 2026-04-25 (stable domain; Go and lib/pq versions unlikely to change)
