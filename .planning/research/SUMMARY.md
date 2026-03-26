# Project Research Summary

**Project:** Aurora — schedule-service
**Domain:** Go scheduling microservice (smart home automation)
**Researched:** 2026-03-25
**Confidence:** HIGH

## Executive Summary

The schedule-service is Aurora's 7th microservice and adds a distinct bounded context to the system: time-triggered automation. Unlike the existing rules-service (which reacts to sensor events), the schedule-service fires actions at predetermined times. The core value proposition, as stated in PROJECT.md, is "reliable schedules that execute even after service restart" — which means in-memory-only scheduling is not acceptable. The recommended approach is gocron/v2 as the in-process scheduler (which natively supports all five required schedule types), backed by PostgreSQL persistence for durability, and NATS publishing for decoupled action dispatch. The service follows the identical DDD layered structure of the six existing Aurora services, with one net-new component: a CronRunner background goroutine in the infrastructure layer.

The recommended stack requires zero new infrastructure components. gocron/v2 provides `OneTimeJob`, `WeeklyJob`, `DailyJob`, `CronJob`, and `DurationJob` out of the box, covering every schedule type in the project spec. PostgreSQL persistence uses the existing `pkg/database` pattern; the scheduler UUID maps 1:1 to database primary keys via `WithIdentifier(uuid)`. NATS publishing follows the `sensors-service` nats_connection.go pattern (not the `rules-service` pattern, which has known reconnection deficiencies). The HTTP layer uses Gin v1.9.1 with JWT auth via `pkg/jwt`, consistent with all existing services.

The top risks are: (1) missed executions during downtime if a startup recovery check is not built in from day one; (2) goroutine leaks if action dispatch uses unbounded goroutines without timeouts — a known issue already flagged in the existing codebase's CONCERNS.md; (3) timezone bugs if per-schedule IANA timezone is not added to the schema in the initial migration; (4) duplicate executions on rolling restarts if a PostgreSQL advisory lock is not used. All four risks must be addressed in Phase 1 at the schema and architecture level — they cannot be retrofitted cheaply.

---

## Key Findings

### Recommended Stack

The schedule-service slots directly into Aurora's existing stack without adding new infrastructure. gocron/v2 (v2.19.1) is the clear choice over `robfig/cron` (which cannot handle one-shot jobs and was last released in 2020) and over raw goroutine timers (no observable state, no restart recovery). lib/pq v1.10.9 is kept for consistency with the six existing services. NATS v1.31.0 matches the existing services; the schedule-service publishes but does not subscribe. Port 8086 is the next available after notifications-service at 8085.

**Core technologies:**
- `github.com/go-co-op/gocron/v2 v2.19.1`: in-process scheduler — covers all five schedule types natively, actively maintained, testable interface, UUID-keyed job identity maps to PK
- `github.com/lib/pq v1.10.9`: PostgreSQL driver — matches all 6 existing services; consistency over marginal pgx gains
- `github.com/nats-io/nats.go v1.31.0`: NATS messaging — publish-only (schedule.executed topic); matches existing services
- `github.com/gin-gonic/gin v1.9.1`: HTTP REST API — identical to all existing services
- `github.com/google/uuid v1.5.0`: UUID generation for schedule IDs — matches existing services

**What NOT to use:**
- Asynq or River (job queues): over-engineering, add Redis/complex schema, solve a different problem
- `time.AfterFunc` or raw goroutines as the persistence source of truth: state is lost on restart
- `robfig/cron` alone: cannot handle one-shot "fire at time T" schedules

### Expected Features

The features split cleanly between what must ship in v1 and what can be deferred. Every table-stakes feature is well-understood and maps directly to patterns already proven in the codebase (rules-service, sensors-service).

**Must have (table stakes):**
- Fixed daily time schedules with day-of-week filter — primary stated use case ("lights off at 4pm")
- One-shot absolute and relative schedules — "turn off lights in 2 hours"
- Recurring cron-expression schedules — arbitrary recurrence
- PostgreSQL persistence with restart survival — core value of the service
- CRUD REST API with JWT auth — POST/GET/PUT/PATCH toggle/DELETE /schedules
- Enable/disable toggle — usability baseline
- Action: turn on/off light (HTTP to lighting-service) — proven in rules-service
- Action: trigger/silence alarm (HTTP to security-service) — proven in rules-service
- Action: publish NATS event — extensibility and decoupling
- Multi-tenant isolation by user_id — mandatory, every other service enforces this

**Should have (differentiators):**
- Named schedules with description field — trivial to add, makes list readable
- Next-execution timestamp in API response — frontend UX without client-side computation
- Execution history / audit log — `GET /schedules/:id/executions`, NATS publish for notifications-service
- Schedule preview endpoint — `POST /schedules/preview`, returns next N fire times

**Defer (v2+):**
- Timezone-aware scheduling — adds real complexity; start UTC-only with a migration path (but add the `timezone` column to the schema in v1 even if unused initially)
- Arbitrary HTTP action target — defer until a concrete use case beyond lighting/security exists
- Geo-triggered schedules — different bounded context entirely, explicitly out of scope in PROJECT.md

### Architecture Approach

The schedule-service follows the same DDD layered structure as all six existing Aurora services: `domain/` (pure Go, no external imports), `application/` (use cases, interfaces only), `infrastructure/` (HTTP, repository, messaging, scheduler). The only net-new component is `CronRunner` in `infrastructure/scheduler/`, which wraps gocron/v2 and runs as a background goroutine — analogous to how the NATS subscriber goroutine already runs in `rules-service`, `security-service`, and `notifications-service`. The domain defines `ActionDispatcher` and `SchedulerRegistry` interfaces; infrastructure provides `NATSPublisher`, `HTTPActionClient`, `CompositeDispatcher`, and `CronRunner` as implementations. Application layer (`ScheduleService`) never imports any external library.

**Major components:**
1. `Schedule` entity + `ScheduleExecution` entity + `Action` value object — domain layer, zero external imports
2. `ScheduleService` (application) — all use cases: CRUD, toggle, execute, list executions
3. `CronRunner` (infrastructure/scheduler) — wraps gocron/v2, loads schedules on startup, registers/unregisters dynamically, fires `ScheduleService.ExecuteSchedule()`
4. `PostgresScheduleRepository` (infrastructure/repository) — upsert on Save, `FindAllEnabled()` for startup reload
5. `CompositeDispatcher` (infrastructure) — routes to `NATSPublisher` or `HTTPActionClient` based on `Action.Target`
6. `Handlers` (infrastructure/http) — Gin handlers, routes, JWT middleware; port 8086

**Key data flows:**
- CRUD: HTTP handler → ScheduleService.Create/Update/Delete → repo.Save + CronRunner.Register/Unregister
- Execution: gocron fires → CronRunner callback → ScheduleService.ExecuteSchedule → CompositeDispatcher → NATS or HTTP → save ScheduleExecution record
- Startup: main.go loads active schedules from repo → registers each in gocron with `WithIdentifier(uuid)` → gocron.Start()

**Strict dependency direction:**
`main.go` → infrastructure → application → domain → stdlib only

### Critical Pitfalls

1. **Missed executions on restart** — Store `last_executed_at` per schedule in PostgreSQL. On startup, run a recovery check: for each enabled schedule due within a configurable grace window (default 5 minutes, via `MISSED_EXECUTION_GRACE_SECONDS` env var), re-execute before starting the cron loop. This must be built in from day one — retrofitting requires schema migration and behavioral changes.

2. **Goroutine leak in action executor** — Always wrap every dispatch call in `context.WithTimeout`. Use a bounded worker pool (channel-based semaphore, max 10 concurrent job goroutines). Never put NATS publish or HTTP calls in a cron callback without a timeout. This is a known existing issue in rules-service (logged in CONCERNS.md) — do not repeat it.

3. **Timezone bugs** — Store `timezone` (IANA string, e.g., `"America/Sao_Paulo"`) per schedule. Always evaluate day-of-week filters using `time.Now().In(loc)`. Use `time/tzdata` embed package so tzdata is available in Alpine Docker images without `apk add tzdata`. Add the `timezone` column to the initial schema even if UTC-only in v1.

4. **Duplicate execution on restart overlap** — Use a PostgreSQL advisory lock (`pg_try_advisory_lock`) before starting the cron loop. This prevents two instances of schedule-service during a rolling deploy from both firing the same job.

5. **NATS reconnection silent failures** — Follow `sensors-service/infrastructure/messaging/nats_connection.go` as the NATS connection template (uses `RetryOnFailedConnect`, `MaxReconnects(-1)`, disconnect/reconnect handlers). Do NOT follow the `rules-service` NATS pattern, which lacks these handlers. Add retry with exponential backoff (3 attempts: 100ms/500ms/2s) on publish failures; fall back to HTTP dispatch if NATS is unavailable.

---

## Implications for Roadmap

Based on the combined research, three phases are recommended. The dependency chain is clear: domain and schema must be stable before any other layer is written; the scheduler core must work correctly (including recovery) before NATS is wired; NATS and execution polish come last.

### Phase 1: Domain, Schema, and Scheduler Core

**Rationale:** Everything else depends on a correct domain model, a complete database schema, and a working scheduler loop with restart recovery. The three pitfalls that cannot be retrofitted (timezone column, one-shot `completed_at`, `last_executed_at` for recovery) must all be in the initial migration. The bounded worker pool and graceful shutdown must also be built as the foundation, not added later.

**Delivers:** A fully working scheduler that creates, persists, and executes schedules across restarts — with no goroutine leaks and correct one-shot lifecycle. All schedule types (cron, fixed daily, one-shot) functional via HTTP API. HTTP actions to lighting-service and security-service working.

**Addresses (from FEATURES.md):** Schedule entity + PostgreSQL persistence, CRUD REST API with JWT auth, fixed daily time + day-of-week, one-shot schedules, cron expression schedules, enable/disable toggle, action turn on/off light, action trigger/silence alarm.

**Avoids (from PITFALLS.md):** Pitfall 1 (missed executions — startup recovery), Pitfall 3 (timezone — schema field from day one), Pitfall 4 (goroutine leak — bounded worker pool), Pitfall 6 (one-shot re-firing — atomic completed_at), Pitfall 7 (cron validation — parse at domain construction), Pitfall 8 (in-memory timer — DB-poll for one-shots), Pitfall 9 (graceful shutdown), Pitfall 10 (day-of-week TZ), Pitfall 11 (user isolation).

**Build order (strict):** domain layer → repository interface + errors → ScheduleService (application) → PostgresScheduleRepository + migrations → HTTPActionClient → CronRunner → HTTP handlers → main.go wiring → docker-compose integration.

### Phase 2: NATS Integration and Execution Observability

**Rationale:** NATS integration is kept separate because it introduces connection lifecycle complexity (reconnect handling, publish retries, HTTP fallback) that should be built on top of a stable scheduler core. Execution history is the natural complement: once the scheduler fires reliably, the audit log and NATS publish happen in the same `ExecuteSchedule` method extension.

**Delivers:** `schedule.executed` NATS events consumed by notifications-service. Execution history table (`schedule_executions`) with `GET /schedules/:id/executions` endpoint. Broken schedule detection (404 from target service marks schedule as `status = 'broken'`). PostgreSQL advisory lock for duplicate execution prevention.

**Addresses (from FEATURES.md):** Action publish NATS event, execution history / audit log, next-execution timestamp in API response.

**Avoids (from PITFALLS.md):** Pitfall 2 (duplicate execution — advisory lock), Pitfall 5 (NATS reconnection — sensors-service pattern), Pitfall 12 (orphaned schedules referencing deleted devices).

**Uses (from STACK.md):** `nats.go v1.31.0` with `RetryOnFailedConnect`, `MaxReconnects(-1)`, disconnect/reconnect handlers (sensors-service template).

### Phase 3: UX Enhancements and Timezone Support

**Rationale:** Once the scheduler is correct and observable, invest in developer/user experience features. Timezone support is the one feature that requires a schema change (the `timezone` column is already in the schema from Phase 1, just unused), so it is the least disruptive to add here. Preview endpoint and description field are low-complexity additions.

**Delivers:** Per-schedule timezone support (IANA tzdata). `POST /schedules/preview` dry-run endpoint returning next N execution times. Named schedule descriptions surfaced in API responses. Potential v2 roadmap items: arbitrary HTTP action target, cross-service device deletion webhooks.

**Addresses (from FEATURES.md):** Timezone-aware scheduling, schedule preview, named schedules with descriptions.

**Avoids (from PITFALLS.md):** Pitfall 3 full resolution (the column exists from Phase 1; Phase 3 activates the logic).

### Phase Ordering Rationale

- Domain and schema must be complete and immutable-ish before other layers are written — adding `timezone`, `completed_at`, or `last_executed_at` after initial deployment requires migrations and is risky.
- CronRunner depends on ScheduleService which depends on the repository — the build order within Phase 1 is strict and well-defined.
- NATS integration is isolated in Phase 2 because its failure modes (reconnect, duplicate publish) are distinct from HTTP action failures and need focused testing.
- Timezone activation is deferred to Phase 3 because it requires validating IANA string input across all existing schedules (migration concern) even though the column exists from day one.

### Research Flags

Phases where standard patterns apply (no additional research needed):
- **Phase 1 (domain/scheduler core):** Well-documented. gocron/v2 API is clear, DDD layer pattern is identical to existing services. Build order is explicit in ARCHITECTURE.md.
- **Phase 2 (NATS + execution history):** The `sensors-service` NATS pattern is the direct template. No unknowns.
- **Phase 3 (timezone/UX):** Go's `time.LoadLocation` and `time/tzdata` embed are standard library concerns. No research needed.

Potential decision point before coding:
- **One-shot implementation mechanism:** ARCHITECTURE.md suggests `time.AfterFunc` for one-shots; PITFALLS.md (Pitfall 8) argues against it in favor of a DB-poll loop. These are in tension. A decision is needed before Phase 1 coding starts (see Open Questions below).

---

## Open Questions (Decisions Needed Before Coding)

These are unresolved tensions across research files that need a decision before Phase 1 begins:

1. **One-shot scheduling mechanism: `time.AfterFunc` vs DB-poll loop**
   - ARCHITECTURE.md proposes `time.AfterFunc` for one-shots (simpler, consistent with existing rules-service pattern)
   - PITFALLS.md (Pitfall 8) argues DB-poll is more reliable because `time.AfterFunc` timers are lost on restart
   - The PITFALLS.md argument is stronger given the core value ("survives restarts") — but DB-poll every 30-60s adds a query loop
   - **Recommendation:** Use DB-poll for one-shots. Store `next_run_at` in PostgreSQL and include one-shots in the same `FindDueSchedules()` query. Eliminates `time.AfterFunc` entirely. Simpler recovery logic.

2. **gocron/v2 vs robfig/cron as the gocron runner**
   - STACK.md recommends gocron/v2; ARCHITECTURE.md references `robfig/cron` in the CronRunner component description (the architecture file was written before the stack decision was locked)
   - **Decision:** gocron/v2 is correct per STACK.md rationale. The CronRunner wraps gocron/v2, not robfig/cron. ARCHITECTURE.md's CronRunner description applies identically to gocron/v2 — the library swap is internal to that component.

3. **Startup timing: scheduler before or after HTTP server?**
   - STACK.md says scheduler must start before HTTP server to avoid race conditions on handler registration
   - ARCHITECTURE.md main.go wiring order lists HTTP server last (step 12), which is consistent
   - **Decision:** Confirmed. Scheduler starts at step 10 (LoadAndStart), HTTP server at step 12. No conflict.

4. **Should `timezone` be active in Phase 1 or just stored?**
   - Storing the column from day one is non-negotiable (PITFALLS.md Pitfall 3)
   - Activating timezone-aware evaluation in Phase 1 vs Phase 3 is a scope decision
   - **Recommendation:** Add the column and validate IANA strings at API boundary in Phase 1. Defer timezone-aware cron expression conversion to Phase 3. This prevents invalid data from entering the system while keeping Phase 1 scope manageable.

5. **`silence_alarm` endpoint on security-service — does it exist?**
   - FEATURES.md notes "MEDIUM confidence — endpoint needs verification against security-service routes"
   - The `trigger_alarm` action is HIGH confidence (proven in rules-service), but `silence_alarm` HTTP endpoint URL needs to be verified before wiring `HTTPActionClient`
   - **Action:** Check `services/security-service/internal/infrastructure/http/routes.go` before implementing `silence_alarm` in `HTTPActionClient`.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Directly derived from existing go.mod files and pkg.go.dev verification. gocron/v2 API confirmed at v2.19.1. All other dependencies match existing services exactly. |
| Features | HIGH | Authoritative source is PROJECT.md (direct requirement). Action patterns verified in rules-service codebase. One uncertainty: `silence_alarm` endpoint URL needs verification. |
| Architecture | HIGH | Derived from direct analysis of all 6 existing services. DDD layer structure is mechanical replication. CronRunner is the only net-new component and its pattern (background goroutine injected with service) is already in rules-service/security-service. |
| Pitfalls | HIGH | Pitfalls 1, 3, 4, 5, 9 directly evidenced by codebase CONCERNS.md and existing service patterns. Pitfalls 2, 6, 7, 8 are fundamental scheduling correctness properties. |

**Overall confidence:** HIGH

### Gaps to Address

- **`silence_alarm` endpoint URL:** Verify `services/security-service/internal/infrastructure/http/routes.go` before implementing `HTTPActionClient.Dispatch` for that action type.
- **gocron/v2 one-shot implementation:** gocron/v2's `OneTimeJob` handles the "fire at absolute time T" case natively — confirm whether it persists `next_run_at` correctly on registration so the DB-poll recovery check works, or whether `next_run_at` must be computed and stored manually before registering.
- **PostgreSQL advisory lock key:** Choose a stable integer key for `pg_try_advisory_lock`. Convention is to hash the service name or use a hardcoded project-specific constant. Document the chosen key in a constant in `main.go`.
- **`MISSED_EXECUTION_GRACE_SECONDS` default value:** 5 minutes is suggested in PITFALLS.md. Confirm this is appropriate for Aurora's deploy cadence (Docker Compose restart typically takes < 30 seconds, so 5 minutes provides a large safety margin).

---

## Sources

### Primary (HIGH confidence)
- `.planning/PROJECT.md` — authoritative requirements, core value, out-of-scope list
- `.planning/research/STACK.md` — technology decisions with alternatives analysis
- `.planning/research/FEATURES.md` — feature breakdown, schedule types, action types, MVP order
- `.planning/research/ARCHITECTURE.md` — DDD layer structure, component boundaries, data flows, build order
- `.planning/research/PITFALLS.md` — 12 pitfalls with phase assignments and prevention strategies
- `services/rules-service/go.mod` — existing dependency versions (cross-referenced)
- `services/sensors-service/internal/infrastructure/messaging/nats_connection.go` — correct NATS pattern
- `.planning/codebase/CONCERNS.md` — goroutine leak precedent, NATS inconsistency documented
- `pkg.go.dev/github.com/go-co-op/gocron/v2` — gocron/v2 v2.19.1 API (verified 2026-03-25)

### Secondary (MEDIUM confidence)
- gocron/v2 community patterns for DB-backed persistence — no built-in persistence means DIY, but the pattern is well-established
- robfig/cron v3 (pkg.go.dev) — alternative evaluated and rejected for one-shot limitation
- go-quartz v0.15.2 — alternative evaluated and rejected for community size and Quartz cron syntax

---
*Research completed: 2026-03-25*
*Ready for roadmap: yes*
