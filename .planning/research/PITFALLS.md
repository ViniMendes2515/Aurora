# Domain Pitfalls: Go Scheduling Microservice

**Domain:** Cron/scheduling microservice in Go with PostgreSQL persistence and NATS messaging
**Researched:** 2026-03-25
**Project:** Aurora schedule-service (new microservice, fits into existing 6-service DDD ecosystem)

---

## Critical Pitfalls

Mistakes that cause missed executions, duplicate fires, silent data corruption, or full rewrites.

---

### Pitfall 1: Missed Executions on Restart (No Recovery Window)

**What goes wrong:**
The scheduler loads schedules from PostgreSQL on startup and registers them in-memory. If the service is down during a scheduled window (e.g., crash at 3:59pm, restart at 4:01pm), the `4:00pm lights off` schedule fires with no one running. On restart, the cron library simply waits for the next occurrence — the missed execution is silently lost forever.

**Why it happens:**
Standard cron libraries (`robfig/cron`, `gocron`) only look forward. They calculate "next run time" from `time.Now()` at registration time. No concept of "did I miss something while I was down?" exists unless you explicitly build it.

**Consequences:**
- Lights stay on overnight because the `turn off at 10pm` schedule was missed during a deploy
- User-facing promise ("reliable schedules that execute even after restart") is broken — the core value of the service per PROJECT.md
- Failures are invisible: no error, no log, no notification

**Prevention:**
1. Store `last_executed_at` (nullable timestamp) per schedule in PostgreSQL.
2. On startup, after loading schedules, run a recovery check: for each enabled schedule where `last_executed_at` is older than the most recent due time (and the schedule was due in the last N minutes — a configurable grace window), re-execute it.
3. The grace window (e.g., 5 minutes) prevents firing a schedule that was "missed" an hour ago during a long maintenance window. Make this configurable via `MISSED_EXECUTION_GRACE_SECONDS` env var.
4. Update `last_executed_at` atomically with a PostgreSQL UPDATE (not in-memory only) immediately before dispatching the action.

**Detection (warning signs):**
- Users reporting "schedule didn't run" after a deploy
- `last_executed_at` in the DB being older than `next_run_at` by more than the cron interval

**Phase:** Address in Phase 1 (core scheduler implementation). The `last_executed_at` column and startup recovery check must be baked in from the beginning — retrofitting after the schema is in production requires a migration and behavioral change users may notice.

---

### Pitfall 2: Duplicate Execution on Concurrent Startup

**What goes wrong:**
If two instances of schedule-service start simultaneously (e.g., rolling deploy in Docker, a crash-restart race, or future horizontal scaling), both load the same schedules and both fire the same job at the same second. The lighting-service and security-service receive two identical commands within milliseconds.

**Why it happens:**
Each service instance independently loads from PostgreSQL and independently runs its cron loop. There is no coordination between instances.

**Consequences:**
- Alarm triggered twice: the buzzer fires, silences, then fires again within seconds — confusing and noisy
- Light commands: `turn on` then `turn off` in rapid succession leaves the light in an undefined state
- Aurora is a single-household system today, but the pattern is fragile and breaks as soon as docker-compose restart overlaps

**Prevention (single-instance, Aurora's current constraint):**
1. Use an advisory lock in PostgreSQL via `pg_try_advisory_lock(key)` as a startup guard. The service acquires the lock before starting the cron loop. If a second instance starts, it detects the lock and either exits or waits.
2. Store `is_running_on` (hostname + pid) in a `schedule_service_lock` table with a TTL. Refresh the TTL heartbeat every 30 seconds. If the lock is stale (no refresh for > 60s), a new instance may claim it.
3. For idempotent actions (lights): ensure the action payload is idempotent (`turn_on` is safe to send twice). For non-idempotent actions (alarm trigger): the lock approach is mandatory.

**Detection (warning signs):**
- Duplicate NATS messages with identical `schedule_id` within a 1-second window
- Users reporting double-actions ("alarm went off twice")

**Phase:** Address in Phase 2 (NATS integration). The lock mechanism must be in place before NATS publishing is wired up, because that is when duplicates become user-visible. PostgreSQL advisory lock is the simplest implementation that fits the existing `pkg/database` pattern.

---

### Pitfall 3: Timezone Handling — `time.Local` vs Stored TZ

**What goes wrong:**
Schedules are stored as `HH:MM` strings in PostgreSQL (matching the existing `ScheduleTime` value object in `rules-service/internal/domain/schedule_time.go`). When the cron library evaluates "is it time to run?", it compares against `time.Now()`. If the Go process TZ (`time.Local`) differs from the user's intended timezone, every schedule fires at the wrong time.

**Why it happens:**
Docker containers default to UTC. A Brazilian user creating a schedule for `16:00` expects it at 4pm Brasilia time (UTC-3), but the container fires it at 4pm UTC (1pm local time — 3 hours early).

**Consequences:**
- "Turn off lights at 10pm" fires at 7pm local time — lights go off while family is still active
- Alarms arm/disarm at wrong times — security risk
- Bug is invisible in development if developer's machine is also UTC

**Prevention:**
1. Store a `timezone` field per schedule (e.g., `"America/Sao_Paulo"`). Use Go's `time.LoadLocation(tz)` to parse it. Default to `"UTC"` if not provided.
2. When registering the schedule with the cron library, convert `HH:MM` to a `time.Location`-aware next-run calculation. Do NOT use the library's raw string cron parsing with `time.Local`.
3. Validate timezone strings at API boundary using `time.LoadLocation` — return 400 if the value is not a valid IANA timezone name.
4. Add a `TZ` environment variable to the docker-compose entry for schedule-service and document that it should match the household's local timezone. But treat this as a fallback — per-schedule TZ is the correct solution.

**Warning signs:**
- Schedules firing consistently N hours early or late
- `time.LoadLocation("America/Sao_Paulo")` returning an error in the container (tzdata not installed) — this is a Docker image concern: use `golang:1.21-alpine` with `apk add tzdata` or the `time/tzdata` embed package

**Phase:** Address in Phase 1 (domain model). The `timezone` field must be in the domain model and the database schema from day one. Adding it later requires a migration and API version change.

---

### Pitfall 4: Goroutine Leak in the Cron Executor

**What goes wrong:**
Each scheduled job spawns a goroutine to execute the action (publish to NATS or call HTTP). If the action blocks — NATS is temporarily disconnected, HTTP target service is slow, or there's a network partition — the goroutine hangs indefinitely. The next tick of the same schedule spawns another blocked goroutine. Over time, goroutines accumulate.

**Why it happens:**
This is already observed in the existing codebase: `rules-service/internal/application/rules_engine.go` line 203 uses unbounded `go func()` with `time.Sleep`, noted in `CONCERNS.md` as "unbounded goroutine growth." The schedule-service would repeat this pattern if written naively.

**Consequences:**
- Memory grows linearly with accumulated blocked goroutines
- Eventually, the service OOMs and Docker restarts it — causing a restart-and-missed-execution cycle
- HTTP client calls without timeouts block forever (`rules-service` uses `Timeout: 3 * time.Second`, but NATS publish calls currently have no timeout)

**Prevention:**
1. Always use `context.WithTimeout` for every action dispatch. For HTTP: `http.NewRequestWithContext(ctx, ...)`. For NATS publish: use `nc.PublishMsg` wrapped in a goroutine with a `select` on `ctx.Done()`.
2. Use a **bounded worker pool** for executing scheduled jobs. A channel-based semaphore pattern limits concurrent job goroutines to N (e.g., 10). If the pool is full, log a warning and skip (or queue with a short deadline).
3. Log goroutine count periodically using `runtime.NumGoroutine()`. Alert if it exceeds a threshold.
4. Never put NATS Publish or HTTP calls directly in a cron callback without a timeout context.

**Warning signs:**
- `runtime.NumGoroutine()` climbing in logs
- CPU at 0% but memory growing over hours
- NATS disconnect causing the service to appear "frozen" from the outside

**Phase:** Address in Phase 1 (scheduler core). The worker pool pattern and timeout-wrapped dispatch must be the foundation, not a later optimization. The CONCERNS.md note about unbounded goroutines in rules-service is the direct precedent to not repeat.

---

### Pitfall 5: NATS Reconnection — Subscriptions Not Re-registered

**What goes wrong:**
The `rules-service` NATS subscriber pattern (in `nats_subscriber.go`) uses `nats.Connect` with a manual retry loop but NO `nats.RetryOnFailedConnect`, NO `DisconnectErrHandler`, and NO `ReconnectHandler`. If NATS disconnects after initial connection, subscriptions are lost silently. The `sensors-service` pattern (in `nats_connection.go`) uses the correct options but is not followed consistently.

For the schedule-service, the concern is different but related: after a NATS reconnect, any pending publish call that was in-flight during the disconnect fails. If the scheduler just logs the error and moves on, the scheduled action silently never fires.

**Why it happens:**
NATS client automatically reconnects the TCP connection with the correct options, but in-flight `Publish` calls during the disconnect window return `nats.ErrConnectionClosed` or `nats.ErrDisconnected`. Without explicit handling, these errors are discarded.

**Consequences:**
- "Turn off lights at 10pm" fires, publishes to NATS, gets `ErrDisconnected`, logs it, and exits — lights stay on
- No retry, no dead-letter queue, no user notification
- The bug is especially dangerous because it's intermittent: NATS is usually up, so the failure only happens during brief reconnect windows

**Prevention:**
1. Use the `sensors-service` pattern (`nats_connection.go`) as the template for schedule-service NATS setup, not the `rules-service` pattern. Use `nats.RetryOnFailedConnect(true)`, `nats.MaxReconnects(-1)` (infinite), `DisconnectErrHandler`, and `ReconnectHandler`.
2. For NATS publish failures in the scheduler's action executor: implement a retry with exponential backoff (max 3 attempts, 100ms/500ms/2s delays) before marking the execution as failed.
3. After max retries are exhausted, fall back to HTTP direct call (the schedule-service already plans to support both NATS and HTTP dispatch — use HTTP as the fallback when NATS is unavailable).
4. Store action execution status in PostgreSQL: `pending`, `dispatched`, `failed`. A background sweep job retries `pending` actions older than 30 seconds.

**Warning signs:**
- Logs showing `nats: connection closed` followed by no retry
- `last_executed_at` updated but no downstream effect (means dispatch failed silently)
- NATS reconnect logs appearing frequently (NATS is flapping — investigate NATS health)

**Phase:** Address in Phase 2 (NATS integration). The connection setup must follow `sensors-service` pattern from the start. The retry/fallback logic is a Phase 2 concern, not a Phase 1 concern.

---

## Moderate Pitfalls

---

### Pitfall 6: One-Shot Schedule Left Enabled After Firing

**What goes wrong:**
A one-shot schedule ("turn off lights in 2 hours") fires once, but the record in PostgreSQL remains `enabled = true`. On the next service restart, the recovery check (Pitfall 1 prevention) finds it "overdue" and fires it again.

**Why it happens:**
Recurring and one-shot schedules share the same `enabled` field. One-shot schedules need to be disabled atomically when they fire.

**Consequences:**
- One-shot actions fire on every restart after they were originally due
- User deletes the schedule — but between creation and deletion, it fires multiple times

**Prevention:**
1. Add a `schedule_type` field to the domain model: `recurring` vs `one_shot`.
2. For `one_shot` schedules: after successful dispatch, set `enabled = false` and `completed_at = now()` in the same PostgreSQL transaction that updates `last_executed_at`.
3. The recovery check must skip schedules where `completed_at IS NOT NULL`.
4. Expose `completed_at` in the list API response so users can see which one-shots have fired.

**Phase:** Phase 1 (domain model). The distinction between recurring and one-shot must be in the schema from day one.

---

### Pitfall 7: Cron Expression Accepted but Never Validated

**What goes wrong:**
The API accepts a cron expression (e.g., `"0 16 * * 1-5"` for weekdays at 4pm) but does not validate it at the HTTP handler layer. An invalid expression (`"0 25 * * *"` — hour 25) is stored in PostgreSQL. The scheduler silently fails to parse it on startup and skips the schedule with no user notification.

**Why it happens:**
Validation in the domain layer (like the existing `ScheduleTime` value object) requires calling the cron parser. If validation is deferred to the scheduler initialization rather than the API handler, errors are invisible to users.

**Consequences:**
- User creates a schedule, it appears in the list as "active", but never fires
- No error log visible to the user
- Support burden: "why isn't my schedule running?"

**Prevention:**
1. Validate cron expressions at domain model construction time using `robfig/cron/v3`'s `cron.NewParser().Parse(expr)`. Return a domain error if invalid.
2. The HTTP handler must return 422 with a descriptive error message that includes the invalid expression and what was wrong about it.
3. Test validation with known-bad expressions: 6-field expressions in a 5-field parser, out-of-range values, unsupported syntax.

**Phase:** Phase 1 (domain model and API layer). Follow the `ScheduleTime` value object pattern already established in the codebase.

---

### Pitfall 8: `time.AfterFunc` / `time.NewTimer` for One-Shot Schedules

**What goes wrong:**
One-shot schedules ("in 2 hours") are implemented with `time.AfterFunc(2*time.Hour, action)`. This timer lives in-memory only. On service restart, the timer is gone — and if `last_executed_at` tracking is absent, the one-shot never fires.

**Why it happens:**
`time.AfterFunc` is the most natural Go solution for a delayed action. It works in tests and simple scenarios but is fundamentally unreliable for a persistence-backed scheduler.

**Consequences:**
- "Turn off lights in 2 hours" never fires if the service restarts in that 2-hour window
- No recovery is possible because there is no `next_run_at` stored in PostgreSQL

**Prevention:**
Store all schedules (both recurring and one-shot) in PostgreSQL with a computed `next_run_at` timestamp. The scheduler loop polls PostgreSQL every 30-60 seconds for due schedules (where `next_run_at <= now() AND enabled = true AND completed_at IS NULL`). Never use in-memory timers as the source of truth for one-shot schedules.

**Phase:** Phase 1 (scheduling architecture). This is a fundamental design decision: poll-from-DB is more reliable than in-memory timer for this use case.

---

### Pitfall 9: Missing Graceful Shutdown for the Cron Loop

**What goes wrong:**
The cron scheduler starts a background goroutine at startup. On `SIGTERM` (Docker stop, deploy restart), the main function exits but the cron goroutine and any in-flight action goroutines are killed abruptly. An action that was 100ms from completing (e.g., an HTTP call to lighting-service) is interrupted mid-flight.

**Why it happens:**
The existing services all use `defer db.Close()` and `defer subscriber.Close()` but there is no pattern in the codebase for waiting on background goroutines before exit. The cron goroutine is not tracked.

**Consequences:**
- Partially-executed actions: the HTTP request to `lighting-service` was sent but the `last_executed_at` update was not committed — on restart, the action fires again (double-execution)
- NATS publish was in-flight: message dropped

**Prevention:**
1. Use a `context.Context` for the cron loop: `ctx, cancel := context.WithCancel(context.Background())`. Pass this context to the cron scheduler and to every action goroutine.
2. On `SIGTERM`/`SIGINT` (caught via `signal.NotifyContext`), call `cancel()`. Wait for in-flight jobs to complete using a `sync.WaitGroup` with a deadline (e.g., 10 seconds).
3. Only update `last_executed_at` in PostgreSQL AFTER the action completes (or fails with max retries exhausted). Never mark as executed before dispatch succeeds.

**Phase:** Phase 1 (scheduler implementation). Add graceful shutdown at the same time as the cron loop — not as a later improvement. Follow the signal-handling pattern that is currently absent in all existing services but which the schedule-service should establish as the project standard.

---

## Minor Pitfalls

---

### Pitfall 10: Day-of-Week Filter Applied in Wrong Timezone

**What goes wrong:**
A schedule set to "weekdays only" checks `time.Now().Weekday()`. If the service is running in UTC but the user is in UTC-5, midnight in the user's timezone is 5am UTC. A `Monday 8am` schedule in UTC-5 sees `time.Now()` as `Monday 1pm UTC` — which is correct. But a `Sunday 11pm` schedule in UTC-5 sees `time.Now()` as `Monday 4am UTC`, making the weekday check wrong.

**Prevention:**
Always evaluate day-of-week filters using the schedule's stored timezone via `time.Now().In(loc).Weekday()`, not `time.Now().Weekday()`. The timezone fix from Pitfall 3 automatically resolves this if implemented correctly.

**Phase:** Phase 1 (domain model, part of timezone handling).

---

### Pitfall 11: User Isolation Not Enforced at Scheduler Layer

**What goes wrong:**
Schedule CRUD endpoints correctly scope queries by `user_id` (multi-tenant requirement per PROJECT.md). But the scheduler background loop loads ALL enabled schedules (no `user_id` filter needed for execution). If the loop is ever changed to use a user-scoped query by mistake, schedules for users who haven't logged in recently get silently dropped.

**Prevention:**
Keep the execution loop user-agnostic: load all `enabled = true` schedules regardless of `user_id`. User isolation is a read/write concern (who can create/modify/delete), not an execution concern. Document this explicitly in code comments.

**Phase:** Phase 1 (repository layer). Ensure `FindAllEnabled()` and `FindByOwnerID()` are distinct repository methods with distinct purposes.

---

### Pitfall 12: Actions Referencing Deleted Devices

**What goes wrong:**
A schedule targets `light_id = "abc123"`. The user deletes the light from the lighting-service. The schedule still fires and calls `POST /lights/abc123/on`, which returns 404. The scheduler logs the 404 and continues — no cleanup.

**Why this matters:**
Orphaned schedules accumulate over time. The lighting-service has no foreign key constraints (noted in CONCERNS.md), and neither will the schedule-service unless explicitly designed.

**Prevention:**
1. When a schedule action returns 404, mark the schedule as `status = 'broken'` with a `broken_reason` field. Stop retrying.
2. Expose `broken` schedules in the list API so users can see and clean them up.
3. Optionally: implement a webhook/event from lighting-service on device deletion to proactively invalidate schedules — but this adds coupling. The broken-status approach is simpler and fits the existing architecture.

**Phase:** Phase 2 (action execution). This is an operational polish concern, not a core correctness concern.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|----------------|------------|
| Domain model & schema | Missing `timezone`, `schedule_type`, `last_executed_at` fields | Bake all fields into the initial migration — retrofitting is costly |
| Scheduler core loop | In-memory timer for one-shots (Pitfall 8), goroutine leak (Pitfall 4) | Use DB-poll pattern, bounded worker pool, always use context with timeout |
| Startup recovery | Missed executions during downtime (Pitfall 1) silently lost | Implement recovery check at startup before starting cron loop |
| NATS integration | Wrong NATS connection pattern (Pitfall 5), silent publish failure | Copy `sensors-service` nats_connection.go pattern, add retry with HTTP fallback |
| One-shot schedules | Left enabled after firing (Pitfall 6) | Atomic `enabled = false` + `completed_at` in same DB transaction |
| API validation | Invalid cron expressions accepted (Pitfall 7) | Validate at domain construction using cron parser, return 422 |
| Docker/deploy | Duplicate execution on restart overlap (Pitfall 2) | PostgreSQL advisory lock before starting cron loop |
| Graceful shutdown | Abrupt kill leaving action half-executed (Pitfall 9) | `signal.NotifyContext` + `sync.WaitGroup` with 10s deadline |
| Day-of-week filtering | Filter evaluated in wrong timezone (Pitfall 10) | Always evaluate using `time.Now().In(loc)` from stored schedule TZ |

---

## Sources

- Direct codebase analysis: `/home/vini/Workspace/Aurora/.planning/codebase/CONCERNS.md` — goroutine leak precedent in rules-engine, NATS connection pattern inconsistency
- Codebase pattern analysis: `services/sensors-service/internal/infrastructure/messaging/nats_connection.go` vs `services/rules-service/internal/infrastructure/messaging/nats_subscriber.go` — two NATS connection patterns, sensors-service is correct model
- Existing domain model: `services/rules-service/internal/domain/schedule_time.go` — value object pattern established for time validation
- Infrastructure context: `infra/docker-compose.yml` — NATS runs as plain `nats:latest` with no persistence (`--js` JetStream not enabled), confirming at-most-once delivery semantics
- Module versions: `services/rules-service/go.mod` — `nats.go v1.31.0`, `gin v1.9.1`, `go 1.21`
- Architecture constraints: `.planning/PROJECT.md` — single-household, not currently horizontally scaled; core value is "reliable schedules that execute even after restart"

**Confidence levels:**
- Pitfalls 1, 3, 4, 5, 9: HIGH — directly evidenced by codebase patterns and PROJECT.md requirements
- Pitfalls 2, 6, 7, 8: HIGH — fundamental scheduling correctness properties, not speculative
- Pitfalls 10, 11, 12: MEDIUM — follow logically from the above but are secondary execution concerns
