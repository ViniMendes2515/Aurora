# Technology Stack: schedule-service

**Project:** Aurora smart home — schedule-service
**Researched:** 2026-03-25
**Dimension:** Stack research for Go scheduling microservice

---

## Context

The schedule-service is the 7th microservice in Aurora. It must fit the existing pattern:
Go 1.21 + Gin v1.9.1 + PostgreSQL (lib/pq v1.10.9) + NATS (nats.go v1.31.0) + DDD.
The core challenge is that a scheduler that only lives in memory is insufficient — jobs must
survive service restarts (PostgreSQL-persisted) and fire correctly after recovery.

---

## Recommended Stack

### Core Scheduling Engine

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/go-co-op/gocron/v2 | v2.19.1 | In-process scheduler (cron, interval, one-shot, weekly, monthly) | HIGH |

**Why gocron/v2 and not robfig/cron:**
- `robfig/cron v3` (v3.0.1, last released Jan 2020) only supports cron-expression-based recurring
  schedules. It has no one-shot ("run at time T") or day-of-week filtered jobs. Aurora requires
  both recurrence and one-shot delayed jobs ("turn off lights in 2h"). robfig/cron cannot satisfy
  the one-shot requirement without significant custom plumbing.
- gocron v2 ships `OneTimeJob`, `WeeklyJob`, `DailyJob`, `CronJob`, and `DurationJob` natively —
  all five schedule types required by the project spec are covered out of the box.
- gocron v2 is actively maintained (v2.19.1 released January 28, 2026 vs robfig last released 2020).
- gocron v2 provides a `WithIdentifier(uuid)` job option that maps in-memory jobs 1:1 to database
  rows by primary key — the cornerstone of the persistence recovery pattern.
- The `Scheduler` interface is testable (interface-typed), matches the DDD infrastructure layer
  abstraction style already used in Aurora.

**Why not go-quartz:**
- go-quartz v0.15.2 (Sep 2025) has a custom JobQueue interface for persistence but uses 6-field
  Quartz cron syntax (not standard 5-field Unix cron), adding a learning curve.
- Smaller community (7,943 gocron importers vs go-quartz's much lower count).
- The file-system-only persistence example in go-quartz is not a PostgreSQL integration — you'd
  write the same amount of custom code as with gocron but with a less common library.

### Persistence

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/lib/pq | v1.10.9 | PostgreSQL driver (matches existing services) | HIGH |
| database/sql (stdlib) | Go 1.21 | SQL abstraction layer | HIGH |

**Why lib/pq and not pgx:**
- All 6 existing services use lib/pq v1.10.9. Consistency is more valuable than marginal pgx
  performance gains for a home automation scheduler with low query volume.
- Switching a single service to pgx would create two different driver patterns, complicating
  `pkg/database` which is shared across services.
- lib/pq v1.12.0 exists and is still maintained (updated March 2026). Keep at v1.10.9 to match
  existing services unless a specific bug requires upgrading.

**Persistence pattern (no external job store library needed):**

gocron/v2 has no built-in persistence, which is actually an advantage here — it keeps the domain
model explicit. The correct pattern for Aurora's DDD style is:

1. Define a `Schedule` domain entity with full schedule definition (cron expression, day filters,
   next run time, action payload).
2. `ScheduleRepository` interface in domain layer; `postgres_schedule_repository.go` in
   infrastructure.
3. On service start (`main.go`), call `repo.FindAllActive()` and register each in gocron using
   `WithIdentifier(uuid.MustParse(schedule.ID))` so scheduler UUID == PostgreSQL primary key.
4. On create/update/delete via REST API, update PostgreSQL first, then call
   `scheduler.NewJob/Update/RemoveJob` to keep in-memory state consistent.
5. Use gocron's `AfterJobRuns` event listener to update `last_run_at` and `next_run_at` in
   PostgreSQL — enabling health monitoring and debugging.

This pattern requires zero extra libraries and fits Aurora's existing pattern exactly.

### Messaging

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/nats-io/nats.go | v1.31.0 | Publish schedule-triggered events to other services | HIGH |

**Why keep v1.31.0 and not upgrade to v1.50.0:**
- Existing services use v1.31.0. The schedule-service should match to avoid version skew in
  shared NATS connection patterns. Upgrading all services together is a separate concern.
- v1.31.0 fully covers the required use case: publish JSON payloads to topics like
  `schedule.action.triggered` which lighting-service and security-service already know how to
  consume for their respective actions.

**NATS integration pattern:**

The schedule-service publishes events on job execution; it does NOT subscribe. Execution flow:

```
gocron fires job → SchedulerService.executeAction() →
  NATSPublisher.Publish("schedule.action.triggered", payload) →
  lighting-service / security-service consume and act
```

For HTTP-direct actions (not all actions need NATS), keep an `HTTPActionClient` in infrastructure
(same pattern as rules-service calling lighting/security via HTTP).

### HTTP Layer

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| github.com/gin-gonic/gin | v1.9.1 | REST API for CRUD management of schedules | HIGH |

No change from existing stack. REST API exposes: create, read, update, delete, enable/disable
schedule endpoints. Authentication via shared `pkg/jwt` middleware.

### Supporting Libraries

| Library | Version | Purpose | Notes |
|---------|---------|---------|-------|
| github.com/google/uuid | v1.5.0 | UUID generation for schedule IDs | Match existing services; v1.6.0 is latest but consistency wins |
| github.com/golang-jwt/jwt/v5 | v5.3.0 | JWT validation (via pkg/jwt) | Shared, no change |

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Scheduler | gocron/v2 v2.19.1 | robfig/cron v3 | No one-shot jobs; last released 2020; cron-only |
| Scheduler | gocron/v2 v2.19.1 | go-quartz v0.15.2 | Quartz cron syntax; smaller community; same DIY persistence work |
| Scheduler | gocron/v2 v2.19.1 | Custom goroutine + time.AfterFunc | Race conditions on restart; no observable job state; hard to test |
| Persistence | lib/pq v1.10.9 | pgx v5 | Breaks consistency with 6 existing services using lib/pq |
| Job store | DIY PostgreSQL (recommended) | External job queue (Asynq, River) | Over-engineering for 1 service; adds Redis or separate DB dependency |
| Job store | DIY PostgreSQL (recommended) | gocron + Redis lock | Adds Redis infrastructure not in Aurora stack |

---

## What NOT to Use

### Do NOT use Asynq or River (Go job queue libraries)

Asynq and River are background job queue systems designed for high-throughput task processing
with at-least-once delivery guarantees. They solve a different problem (worker pools, retries,
priority queues) and require either Redis (Asynq) or a dedicated PostgreSQL schema with complex
migrations (River). Aurora's scheduling need is simpler: "at time T, do action A" — gocron covers
this without adding infrastructure.

### Do NOT use time.AfterFunc or raw goroutine timers

Manual goroutine-based scheduling loses all scheduled state on restart. Recreating timers from
the database at startup is exactly what gocron does — with a battle-tested implementation.
Using `time.AfterFunc` directly means rebuilding gocron's internals with worse reliability.

### Do NOT use robfig/cron as the primary scheduler

robfig/cron cannot handle one-shot jobs ("turn off the lights at 6pm today"). Using it alongside
manual timer management for one-shot cases produces two scheduling systems in one service.

---

## Dependency Versions Summary

```
go 1.21

require (
    aurora/pkg v0.0.0
    github.com/gin-gonic/gin v1.9.1
    github.com/go-co-op/gocron/v2 v2.19.1
    github.com/google/uuid v1.5.0
    github.com/golang-jwt/jwt/v5 v5.3.0
    github.com/lib/pq v1.10.9
    github.com/nats-io/nats.go v1.31.0
)
```

Note: gocron/v2 internally uses `github.com/google/uuid` as well — pin to v1.5.0 via go.mod
`replace` if version conflicts arise, or allow go mod to resolve to v1.6.0 (compatible with v1.5.0 API).

---

## Integration Points with Existing Infrastructure

### PostgreSQL

Use `pkg/database.NewPostgresConnection()` exactly as all other services do. The schedule-service
gets its own database tables (`schedules`) in the shared `aurora_home` database, following the
pattern where each service owns its tables but shares the database instance.

Schema considerations:
- Store `cron_expression` (nullable), `interval_seconds` (nullable), `one_shot_at` (nullable),
  `day_of_week_filter` (array or bitmask), `next_run_at`, `last_run_at`, `enabled`, `user_id`,
  `action_type`, `action_payload` (JSONB) — all needed to reconstruct a gocron job on restart.
- Primary key is a UUID stored as TEXT (matching existing services).
- On startup, `RunScheduleMigrations(db)` creates tables; no seeding needed.

### NATS

Initialize the NATS connection in `main.go` and inject `NATSPublisher` into the domain executor
(same pattern as sensors-service and rules-service). The schedule-service NATS publisher wraps
`nc.Publish(topic, jsonPayload)`.

Topic convention: `schedule.action.triggered` with a JSON payload containing `user_id`,
`schedule_id`, `action_type`, `action_payload`. Downstream services (lighting, security) can
filter by `action_type` in their existing NATS subscribers.

### gocron Lifecycle in main.go

```go
// main.go startup order:
// 1. Connect PostgreSQL
// 2. Run migrations
// 3. Create gocron.Scheduler
// 4. Load active schedules from repo, register each into gocron
// 5. scheduler.Start()
// 6. Start Gin HTTP server
// Shutdown: scheduler.Shutdown() on SIGTERM before HTTP server stops
```

The scheduler must be started before the HTTP server accepts requests so that any in-flight
schedule operations during startup do not race with handler registrations.

---

## Sources

- gocron/v2 pkg.go.dev (v2.19.1, retrieved 2026-03-25): https://pkg.go.dev/github.com/go-co-op/gocron/v2
- robfig/cron v3 pkg.go.dev (v3.0.1, retrieved 2026-03-25): https://pkg.go.dev/github.com/robfig/cron/v3
- go-quartz pkg.go.dev (v0.15.2, retrieved 2026-03-25): https://pkg.go.dev/github.com/reugn/go-quartz
- nats.go pkg.go.dev (v1.50.0 current, existing services use v1.31.0, retrieved 2026-03-25)
- lib/pq pkg.go.dev (v1.12.0 current, existing services use v1.10.9, retrieved 2026-03-25)
- Aurora existing go.mod: /services/rules-service/go.mod (cross-referenced)
- Aurora STACK.md: /.planning/codebase/STACK.md (cross-referenced)
