# Architecture Patterns: schedule-service

**Domain:** Go scheduling microservice with DDD, NATS, PostgreSQL
**Researched:** 2026-03-25
**Confidence:** HIGH — derived directly from the six existing services in this codebase

---

## Recommended Architecture

The `schedule-service` follows the identical layered structure used by every other
Aurora service. The only net-new concern is the scheduler loop, which runs as a
background goroutine started in `main.go` — analogous to how the NATS subscriber
goroutine runs today in `rules-service`, `security-service`, and
`notifications-service`.

```
services/schedule-service/
├── cmd/api/main.go                               # Entry point, wiring
└── internal/
    ├── domain/
    │   ├── schedule.go                           # Schedule entity + ScheduleType value object
    │   ├── schedule_execution.go                 # ScheduleExecution entity
    │   ├── action.go                             # Action value object
    │   ├── errors.go                             # Domain error variables
    │   └── repository.go                         # ScheduleRepository interface
    ├── application/
    │   └── schedule_service.go                   # Use cases + DTOs
    └── infrastructure/
        ├── http/
        │   ├── server.go
        │   ├── routes.go
        │   └── handlers.go
        ├── repository/
        │   └── postgres_schedule_repository.go
        ├── migrations/
        │   └── schedule_migrations.go
        ├── messaging/
        │   └── nats_publisher.go                 # Publishes schedule.executed events
        ├── scheduler/
        │   └── cron_runner.go                    # Wraps robfig/cron, calls ScheduleService
        └── security/
            └── jwt_validator.go
```

---

## Component Boundaries

| Component | Layer | Responsibility | Communicates With |
|-----------|-------|---------------|-------------------|
| `Schedule` entity | Domain | Core business object, owns cron expression validation, enabled state | Nothing (pure) |
| `ScheduleExecution` entity | Domain | Immutable audit record of a single run attempt | Nothing (pure) |
| `Action` value object | Domain | Encapsulates target service + action type + target device ID | Nothing (pure) |
| `ScheduleRepository` interface | Domain | Contract for persistence | Implemented by infra |
| `ScheduleService` | Application | All use cases — CRUD, toggle, trigger execution | Domain interfaces only |
| `PostgresScheduleRepository` | Infrastructure/repository | SQL implementations of `ScheduleRepository` | PostgreSQL via `pkg/database` |
| `NATSPublisher` | Infrastructure/messaging | Publishes `schedule.executed` events after execution | NATS |
| `CronRunner` | Infrastructure/scheduler | Owns the `robfig/cron` scheduler, reloads schedules on startup, fires `ScheduleService.Execute()` | `ScheduleService` |
| `Handlers` | Infrastructure/http | REST handlers, JSON binding, error mapping | `ScheduleService` |
| `main.go` | Entry point | Dependency wiring, starts HTTP server + CronRunner goroutine | All |

**Dependency direction (strict):**

```
main.go
  └─► infrastructure (http, repository, messaging, scheduler, security)
        └─► application (ScheduleService)
              └─► domain (Schedule, ScheduleExecution, Action, ScheduleRepository)
                    └─► stdlib only (time, errors)
```

Domain has zero imports outside Go stdlib. Application imports domain interfaces only.
Infrastructure imports application and domain — never the reverse.

---

## Domain Layer

### `Schedule` entity — `internal/domain/schedule.go`

```go
package domain

import (
    "errors"
    "time"
)

// ScheduleType represents the recurrence model of a schedule
type ScheduleType string

const (
    ScheduleTypeCron    ScheduleType = "cron"     // standard cron expression e.g. "0 16 * * *"
    ScheduleTypeOneShot ScheduleType = "one_shot"  // fires once at a specific time, then disables
)

// DayOfWeek bitmask — Sunday=1, Monday=2, ..., Saturday=64 (0 = all days)
type DayOfWeek int

const DayOfWeekAll DayOfWeek = 0

// Schedule represents a user-defined time-triggered automation
type Schedule struct {
    ID          string
    UserID      string
    Name        string
    Description string
    ScheduleType ScheduleType
    CronExpr    string       // e.g. "0 16 * * *" for 4pm every day
    RunAt       *time.Time   // only for one_shot; nil for cron
    DaysFilter  DayOfWeek    // 0 = all days; bitmask for weekday filtering
    Action      Action
    Enabled     bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
    LastRunAt   *time.Time
    NextRunAt   *time.Time
}

// NewSchedule creates a new Schedule. Returns error if required fields are missing.
func NewSchedule(id, userID, name string, scheduleType ScheduleType,
    cronExpr string, runAt *time.Time, daysFilter DayOfWeek, action Action) (*Schedule, error) {
    if name == "" {
        return nil, ErrInvalidSchedule
    }
    if scheduleType == ScheduleTypeCron && cronExpr == "" {
        return nil, ErrMissingCronExpression
    }
    if scheduleType == ScheduleTypeOneShot && runAt == nil {
        return nil, ErrMissingRunAt
    }
    now := time.Now()
    return &Schedule{
        ID:           id,
        UserID:       userID,
        Name:         name,
        ScheduleType: scheduleType,
        CronExpr:     cronExpr,
        RunAt:        runAt,
        DaysFilter:   daysFilter,
        Action:       action,
        Enabled:      true,
        CreatedAt:    now,
        UpdatedAt:    now,
    }, nil
}

// BelongsTo checks ownership.
func (s *Schedule) BelongsTo(userID string) bool {
    return s.UserID == userID
}

// IsActiveOnDay reports whether the schedule should fire on the given weekday.
// Sunday=0 per time.Weekday; DayOfWeek bitmask: Sunday=1, Monday=2, Tuesday=4, ...
func (s *Schedule) IsActiveOnDay(day time.Weekday) bool {
    if s.DaysFilter == DayOfWeekAll {
        return true
    }
    mask := DayOfWeek(1 << uint(day))
    return s.DaysFilter&mask != 0
}

// Disable marks the schedule inactive (used after one_shot fires).
func (s *Schedule) Disable() {
    s.Enabled = false
    s.UpdatedAt = time.Now()
}

// RecordRun updates last/next run timestamps.
func (s *Schedule) RecordRun(at time.Time, next *time.Time) {
    s.LastRunAt = &at
    s.NextRunAt = next
    s.UpdatedAt = at
}
```

### `ScheduleExecution` entity — `internal/domain/schedule_execution.go`

```go
package domain

import "time"

// ExecutionStatus represents the outcome of a schedule execution attempt
type ExecutionStatus string

const (
    ExecutionStatusSuccess ExecutionStatus = "success"
    ExecutionStatusFailed  ExecutionStatus = "failed"
    ExecutionStatusSkipped ExecutionStatus = "skipped" // day-filter blocked execution
)

// ScheduleExecution is an immutable audit record of one execution attempt
type ScheduleExecution struct {
    ID          string
    ScheduleID  string
    ExecutedAt  time.Time
    Status      ExecutionStatus
    ErrorMessage string  // empty on success
}

// NewScheduleExecution creates an execution record.
func NewScheduleExecution(id, scheduleID string, executedAt time.Time,
    status ExecutionStatus, errMsg string) *ScheduleExecution {
    return &ScheduleExecution{
        ID:           id,
        ScheduleID:   scheduleID,
        ExecutedAt:   executedAt,
        Status:       status,
        ErrorMessage: errMsg,
    }
}
```

### `Action` value object — `internal/domain/action.go`

```go
package domain

// ActionTarget indicates which integration channel to use
type ActionTarget string

const (
    ActionTargetNATS ActionTarget = "nats" // publish event on a NATS topic
    ActionTargetHTTP ActionTarget = "http" // call an HTTP endpoint directly
)

// ActionType mirrors the action vocabulary already in rules-service
type ActionType string

const (
    ActionTurnOnLight  ActionType = "turn_on_light"
    ActionTurnOffLight ActionType = "turn_off_light"
    ActionTriggerAlarm ActionType = "trigger_alarm"
    ActionSilenceAlarm ActionType = "silence_alarm"
)

// Action is an immutable value object describing what to do when a schedule fires
type Action struct {
    Target     ActionTarget // nats or http
    ActionType ActionType   // what action to perform
    DeviceID   string       // target device/light/sensor ID
    Payload    string       // optional JSON payload override (empty = derive from ActionType)
}

// NewAction validates and constructs an Action.
func NewAction(target ActionTarget, actionType ActionType, deviceID string) (Action, error) {
    if target != ActionTargetNATS && target != ActionTargetHTTP {
        return Action{}, ErrInvalidAction
    }
    return Action{
        Target:     target,
        ActionType: actionType,
        DeviceID:   deviceID,
    }, nil
}
```

### `ScheduleRepository` interface — `internal/domain/repository.go`

```go
package domain

// ScheduleRepository defines persistence contracts for schedules.
// Defined in domain; implemented in infrastructure/repository.
type ScheduleRepository interface {
    Save(schedule *Schedule) error
    FindByID(id string) (*Schedule, error)
    FindByUserID(userID string) ([]*Schedule, error)
    FindAllEnabled() ([]*Schedule, error)  // used by CronRunner on startup reload
    Delete(id string) error
}

// ScheduleExecutionRepository defines persistence for execution audit records.
type ScheduleExecutionRepository interface {
    Save(execution *ScheduleExecution) error
    FindByScheduleID(scheduleID string, limit int) ([]*ScheduleExecution, error)
}
```

### `errors.go` — `internal/domain/errors.go`

```go
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

---

## Application Layer

### `ScheduleService` — `internal/application/schedule_service.go`

All DTOs and use cases live in this one file (consistent with how `rules_engine.go`
collocates `CreateRuleRequest`, `RuleResponse`, and every use case method).

**Key methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `CreateSchedule` | `(req CreateScheduleRequest) (*ScheduleResponse, error)` | Validates, creates entity, persists |
| `GetSchedule` | `(id, userID string) (*ScheduleResponse, error)` | Ownership-checked lookup |
| `ListSchedules` | `(userID string) ([]*ScheduleResponse, error)` | All schedules for user |
| `UpdateSchedule` | `(id, userID string, req UpdateScheduleRequest) (*ScheduleResponse, error)` | Ownership-checked update |
| `ToggleSchedule` | `(id, userID string) (*ScheduleResponse, error)` | Flip enabled flag |
| `DeleteSchedule` | `(id, userID string) error` | Ownership-checked delete |
| `ExecuteSchedule` | `(scheduleID string) error` | Called by CronRunner — dispatches action, records execution |
| `ListExecutions` | `(scheduleID, userID string, limit int) ([]*ExecutionResponse, error)` | Execution audit history |
| `LoadAllEnabled` | `() ([]*domain.Schedule, error)` | Called by CronRunner at startup |

**Constructor:**

```go
func NewScheduleService(
    scheduleRepo    domain.ScheduleRepository,
    executionRepo   domain.ScheduleExecutionRepository,
    actionDispatcher domain.ActionDispatcher,  // interface defined in domain
) *ScheduleService
```

**`ActionDispatcher` interface** (defined in domain so application stays decoupled):

```go
// ActionDispatcher is the domain contract for executing an Action at runtime.
// Implemented in infrastructure (NATS publisher or HTTP client).
type ActionDispatcher interface {
    Dispatch(schedule *Schedule) error
}
```

This mirrors the pattern of `domain.EventPublisher` in sensors-service and
`domain.BuzzerClient` in security-service: the domain defines the interface,
infrastructure provides the implementation, application holds only the interface.

---

## Infrastructure Layer

### `CronRunner` — `internal/infrastructure/scheduler/cron_runner.go`

This is the net-new component with no direct precedent in the six existing services.

**Library:** `github.com/robfig/cron/v3`

Rationale: It is the canonical Go cron library (3,000+ GitHub stars, stable v3 API
since 2018, used by Kubernetes job scheduler internally). It accepts standard 5-field
cron expressions, runs each job in its own goroutine, and is thread-safe. Using a
ticker loop (`time.Ticker`) or `time.Sleep` would require reimplementing expression
parsing — `robfig/cron` already does this correctly and is well-tested.

**Startup flow:**

```
main.go
  1. Create CronRunner(scheduleService)
  2. Call runner.LoadAndStart()
      a. Call scheduleService.LoadAllEnabled()
      b. For each schedule: runner.Register(schedule)
      c. Call cron.Start() — launches background goroutine
  3. CronRunner.Register() is also called after CreateSchedule / UpdateSchedule
     from the HTTP handlers so new schedules become active without restart
```

**What `CronRunner` does when a schedule fires:**

```
cron callback fires (in its own goroutine per robfig/cron)
  1. Check IsActiveOnDay(time.Now().Weekday())  → if false, record Skipped execution, return
  2. Call scheduleService.ExecuteSchedule(scheduleID)
      a. application fetches Schedule from repository
      b. application calls actionDispatcher.Dispatch(schedule)
      c. application records ScheduleExecution (success or failed)
      d. if one_shot: application calls schedule.Disable(), saves
      e. application calls schedule.RecordRun(), saves
```

**One-shot implementation:**

One-shot schedules are stored with `schedule_type = "one_shot"` and a `run_at`
timestamp. `CronRunner` converts `run_at` into a `@` cron spec at registration:
`cron.AddFunc("@" + runAtUnix, ...)` is not valid — instead, use a `time.AfterFunc`
goroutine for one-shots, or convert `run_at` to the exact cron expression
`"MM HH DD month *"`. The `time.AfterFunc` approach is simpler and avoids cron
expression construction:

```go
// For one_shot: schedule via time.AfterFunc, not robfig/cron
delay := schedule.RunAt.Sub(time.Now())
if delay > 0 {
    time.AfterFunc(delay, func() {
        runner.service.ExecuteSchedule(schedule.ID)
    })
}
```

This is consistent with the existing `scheduleMotionAutoOff` pattern in
`rules-service` which already uses `go func() { time.Sleep(timeout); ... }()`.

**DayOfWeek filtering:**

The filter is checked inside the cron callback (step 1 above), not in the cron
expression itself. This keeps the cron expression simple (`"0 16 * * *"`) while
the business rule (weekdays only) lives in the domain entity's `IsActiveOnDay()`
method. The check is at execution time, not at scheduling time.

**Dynamic reload:**

```go
type CronRunner struct {
    cron         *cron.Cron
    service      *application.ScheduleService
    jobsByID     map[string]cron.EntryID  // scheduleID → cron entry
    mu           sync.Mutex
}

// Register adds or replaces a schedule in the cron runner.
func (r *CronRunner) Register(schedule *domain.Schedule) error

// Unregister removes a schedule from the cron runner (called on delete or disable).
func (r *CronRunner) Unregister(scheduleID string)
```

`Register` and `Unregister` are called from `ScheduleService` via an interface so
the application layer can notify the runner without importing the infrastructure
package:

```go
// SchedulerRegistry is the domain contract for registering/unregistering schedules
// with the underlying scheduler. Defined in domain, implemented by CronRunner.
type SchedulerRegistry interface {
    Register(schedule *Schedule) error
    Unregister(scheduleID string)
}
```

`NewScheduleService` receives `SchedulerRegistry` as a dependency.

### `NATSPublisher` — `internal/infrastructure/messaging/nats_publisher.go`

Implements `domain.ActionDispatcher` for the `nats` target path. Follows the exact
same pattern as `sensors-service/infrastructure/messaging/nats_publisher.go`.

**NATS topic emitted:** `schedule.executed`

**Payload published:**

```json
{
  "schedule_id": "...",
  "user_id": "...",
  "action_type": "turn_off_light",
  "device_id": "light-001",
  "executed_at": "2026-03-25T16:00:00Z"
}
```

Other services can subscribe to `schedule.executed` to react (e.g.,
`notifications-service` already subscribes to all domain events and would pick
this up with a wildcard or explicit subscription to log the execution).

### `HTTPActionClient` — `internal/infrastructure/http/action_client.go`

Implements `domain.ActionDispatcher` for the `http` target path. Follows the
exact same pattern as `executeAction()` in `rules-service/application/rules_engine.go`
(copied into its own infrastructure file to keep application layer clean):

- `POST {lightingServiceURL}/lights/{deviceID}/on` for `turn_on_light`
- `POST {lightingServiceURL}/lights/{deviceID}/off` for `turn_off_light`
- `POST {securityServiceURL}/alarms/trigger` for `trigger_alarm`
- `POST {securityServiceURL}/alarms/silence` for `silence_alarm`
- Sets `X-Device-Key` header (same `deviceAPIKey` env var used by rules-service)
- 3-second HTTP timeout

### Composite `ActionDispatcher`

Since a schedule can target NATS or HTTP, `main.go` wires a composite dispatcher:

```go
// CompositeDispatcher routes to NATS publisher or HTTP client based on Action.Target
type CompositeDispatcher struct {
    nats domain.ActionDispatcher
    http domain.ActionDispatcher
}

func (d *CompositeDispatcher) Dispatch(schedule *domain.Schedule) error {
    switch schedule.Action.Target {
    case domain.ActionTargetNATS:
        return d.nats.Dispatch(schedule)
    case domain.ActionTargetHTTP:
        return d.http.Dispatch(schedule)
    }
    return domain.ErrInvalidAction
}
```

### `PostgresScheduleRepository` — `internal/infrastructure/repository/postgres_schedule_repository.go`

Implements both `domain.ScheduleRepository` and `domain.ScheduleExecutionRepository`.
Constructor: `NewPostgresScheduleRepository(db *sql.DB)`.

Key queries:
- `FindAllEnabled()`: `SELECT ... FROM schedules WHERE enabled = true` — called
  once at startup by `CronRunner.LoadAndStart()`
- `Save()`: upsert via `INSERT ... ON CONFLICT (id) DO UPDATE SET ...`

### Migrations — `internal/infrastructure/migrations/schedule_migrations.go`

```sql
CREATE TABLE IF NOT EXISTS schedules (
    id              VARCHAR(36) PRIMARY KEY,
    user_id         VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    schedule_type   VARCHAR(20) NOT NULL,    -- 'cron' | 'one_shot'
    cron_expr       VARCHAR(100) NOT NULL DEFAULT '',
    run_at          TIMESTAMP,               -- NULL for cron type
    days_filter     INTEGER NOT NULL DEFAULT 0, -- 0 = all days; bitmask
    action_target   VARCHAR(10) NOT NULL,    -- 'nats' | 'http'
    action_type     VARCHAR(50) NOT NULL,
    action_device_id VARCHAR(255) NOT NULL DEFAULT '',
    action_payload  TEXT NOT NULL DEFAULT '',
    enabled         BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    last_run_at     TIMESTAMP,
    next_run_at     TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_schedules_user_id ON schedules(user_id);
CREATE INDEX IF NOT EXISTS idx_schedules_enabled ON schedules(enabled);

CREATE TABLE IF NOT EXISTS schedule_executions (
    id              VARCHAR(36) PRIMARY KEY,
    schedule_id     VARCHAR(36) NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    executed_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL,  -- 'success' | 'failed' | 'skipped'
    error_message   TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_executions_schedule_id ON schedule_executions(schedule_id);
```

### HTTP Server — `internal/infrastructure/http/`

`server.go` / `routes.go` / `handlers.go` follow the identical structure of all
other services. Proposed port: **8086** (next available after notifications-service
at 8085).

**REST routes:**

```
POST   /schedules              CreateSchedule    (JWT required)
GET    /schedules              ListSchedules     (JWT required)
GET    /schedules/:id          GetSchedule       (JWT required)
PUT    /schedules/:id          UpdateSchedule    (JWT required)
PATCH  /schedules/:id/toggle   ToggleSchedule    (JWT required)
DELETE /schedules/:id          DeleteSchedule    (JWT required)
GET    /schedules/:id/executions  ListExecutions (JWT required)
GET    /health                 Health check      (no auth)
```

---

## Data Flow: Schedule Execution

```
CronRunner (background goroutine, robfig/cron)
  │
  │ every minute, cron fires matching entries
  ▼
CronRunner callback
  ├─ check schedule.IsActiveOnDay(today)
  │    └─ if false → ScheduleService.RecordExecution(skipped) → return
  │
  └─ ScheduleService.ExecuteSchedule(scheduleID)
        ├─ fetch Schedule from PostgresScheduleRepository
        ├─ CompositeDispatcher.Dispatch(schedule)
        │    ├─ if nats → NATSPublisher.Dispatch()
        │    │    └─ publish "schedule.executed" to NATS
        │    │         └─ (other services subscribe and act)
        │    └─ if http → HTTPActionClient.Dispatch()
        │         └─ POST to lighting-service or security-service
        │              └─ service acts (turn on light, trigger alarm)
        │
        ├─ ScheduleExecutionRepository.Save(execution record)
        ├─ if one_shot → schedule.Disable() → ScheduleRepository.Save()
        └─ schedule.RecordRun() → ScheduleRepository.Save()
```

## Data Flow: Schedule CRUD

```
HTTP Handler (Gin)
  │
  │ POST /schedules  { cron_expr, action, ... }
  ▼
ScheduleService.CreateSchedule(req)
  ├─ domain.NewSchedule(...) → validates cron_expr, action
  ├─ ScheduleRepository.Save(schedule)
  └─ SchedulerRegistry.Register(schedule)  → CronRunner adds entry
       └─ robfig/cron.AddFunc(cronExpr, callback)
            └─ schedule now active without restart
```

---

## Integration with Existing Services

### How schedule-service integrates

| Target | Mechanism | When | Topic / Endpoint |
|--------|-----------|------|-----------------|
| lighting-service | HTTP POST | Action `turn_on_light` / `turn_off_light` | `/lights/{id}/on`, `/lights/{id}/off` |
| security-service | HTTP POST | Action `trigger_alarm` / `silence_alarm` | `/alarms/trigger`, `/alarms/silence` |
| notifications-service | NATS publish | Every execution | `schedule.executed` |
| Any future service | NATS publish | Generic events | `schedule.executed` |

### Environment variables needed

```
SERVER_PORT=8086
JWT_SECRET=...              # shared with all services
NATS_URL=nats://nats:4222
DB_HOST=postgres
DB_PORT=5432
DB_USER=aurora
DB_PASSWORD=aurora_secret
DB_NAME=aurora_home
DB_SSLMODE=disable
LIGHTING_SERVICE_URL=http://lighting-service:8082
SECURITY_SERVICE_URL=http://security-service:8083
DEVICE_API_KEY=aurora-device-key-2024
```

### `main.go` wiring order

```go
// 1. Load env vars (same getEnv helper pattern)
// 2. Connect PostgreSQL via pkg/database.NewPostgresConnection()
// 3. Run migrations (RunScheduleMigrations + SeedDemoSchedules)
// 4. Create repositories (PostgresScheduleRepository)
// 5. Create NATSPublisher (connect to NATS with retry loop)
// 6. Create HTTPActionClient (lightingURL, securityURL, deviceAPIKey)
// 7. Create CompositeDispatcher(nats, http)
// 8. Create CronRunner (empty, before service so it can be passed as SchedulerRegistry)
// 9. Create ScheduleService(scheduleRepo, executionRepo, compositeDispatcher, cronRunner)
// 10. Load enabled schedules into CronRunner: runner.LoadAndStart(scheduleService)
// 11. Create JWT validator (security.NewJWTValidator)
// 12. Create HTTP server, start serving
```

Step 8-10 requires a two-phase init where `CronRunner` is created before
`ScheduleService` but receives a back-reference to the service. This is the same
pattern `rules-service` uses when the NATS subscriber is created before but then
subscribes via a method that holds a reference to `RulesEngine`.

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Querying the DB on every cron tick
**What:** Loading all enabled schedules from PostgreSQL on each cron wake-up
**Why bad:** Unnecessary DB load; robfig/cron already maintains in-process schedule state
**Instead:** Load once at startup via `FindAllEnabled()`. Update the in-process
cron entry on Create/Update/Toggle/Delete via `SchedulerRegistry.Register/Unregister`.

### Anti-Pattern 2: Polling with `time.Ticker` or `time.Sleep`
**What:** Building a custom every-minute polling loop instead of using robfig/cron
**Why bad:** Requires reimplementing cron expression parsing, handling DST, missed
ticks on restart, concurrent execution guards — all solved by robfig/cron
**Instead:** Use `robfig/cron/v3` for expression-based schedules. Use `time.AfterFunc`
only for one-shot schedules (no cron expression needed).

### Anti-Pattern 3: Domain entity importing robfig/cron
**What:** Putting cron expression parsing inside the `Schedule` entity
**Why bad:** Violates DDD — domain would depend on an external library
**Instead:** Domain entity stores `CronExpr string` as a plain string. CronRunner
in infrastructure layer calls `cron.AddFunc(schedule.CronExpr, ...)` which is
where the external library import belongs.

### Anti-Pattern 4: HTTP action calls inside application layer
**What:** Putting `net/http` calls directly in `ScheduleService.ExecuteSchedule()`
**Why bad:** This is what `rules-service` currently does (http calls inside
`executeAction` in the application layer) — it is a known DDD violation in the
existing codebase. The schedule-service has the opportunity to do this correctly.
**Instead:** Define `domain.ActionDispatcher` interface; inject the HTTP client
implementation from infrastructure. Application layer only calls the interface.

### Anti-Pattern 5: Shared state without mutex
**What:** `CronRunner.jobsByID` map accessed from multiple goroutines without locking
**Why bad:** robfig/cron fires each job in its own goroutine. `Register` and
`Unregister` called from HTTP handlers are on different goroutines. Race condition.
**Instead:** Wrap all `jobsByID` reads and writes in `sync.Mutex` — same pattern
`rules-service` uses for `lastMotionByLight` with `sync.RWMutex`.

---

## Scalability Considerations

| Concern | At current scale (household) | At multi-tenant SaaS scale |
|---------|-------------------------------|---------------------------|
| In-process cron | Fine — hundreds of schedules fit in memory | Replace CronRunner with distributed job queue (e.g., asynq/Redis) |
| One-shot via `time.AfterFunc` | Fine — goroutines are cheap for small counts | Loses one-shots on restart; persist "next_run_at" and reload on startup |
| Single service instance | Fine | Requires distributed locking (e.g., Redis SETNX) to prevent duplicate execution |
| Execution history table | Fine — low write rate | Add TTL / partition by month |

For Aurora's current scope (residential, single household), the in-process cron
approach is correct and consistent with how the rest of the system handles
time-based state (rules-service in-memory `lastMotionByLight` map).

---

## Build Order Implications

1. **Domain layer first** — zero external imports, can be written and tested in isolation
2. **Repository interface + errors** — stable before any other layer touches persistence
3. **Application `ScheduleService`** — depends only on domain interfaces, testable with mock repositories
4. **Infrastructure: PostgresScheduleRepository + migrations** — integrates with `pkg/database`, testable with real DB
5. **Infrastructure: NATSPublisher + HTTPActionClient** — separate, testable independently
6. **Infrastructure: CronRunner** — depends on `robfig/cron` + `ScheduleService`; add last since it ties everything together
7. **Infrastructure: HTTP handlers + routes** — depends on `ScheduleService` only
8. **`main.go` wiring** — last, wires all components together
9. **docker-compose.yml integration** — add service definition, env vars, port 8086

Each step compiles and has meaningful unit tests before the next step begins.
This mirrors the build order followed by every other Aurora service.

---

## Sources

- Derived from direct analysis of existing codebase (`rules-service`, `sensors-service`,
  `security-service`) — HIGH confidence
- `robfig/cron/v3`: canonical Go cron library, stable API since 2018
  https://github.com/robfig/cron — HIGH confidence (training data, widely used)
- DDD patterns applied: Interface-in-domain / implementation-in-infrastructure — same
  as `domain.BuzzerClient`, `domain.EventPublisher` already in this codebase
- NATS topic naming: follows existing `sensors.{entity}.{event}` convention
  → `schedule.executed`
- PostgreSQL schema: follows existing `VARCHAR(36) PRIMARY KEY`, `TIMESTAMP`, index
  patterns from `rule_migrations.go` and others

*Architecture analysis: 2026-03-25*
