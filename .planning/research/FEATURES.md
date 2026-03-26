# Feature Landscape: Schedule Service

**Domain:** Smart home automation scheduling microservice
**Researched:** 2026-03-25
**Project:** Aurora — subsequent milestone, adds scheduling to existing lighting, alarm, sensors, and rules-engine services

---

## Context

Aurora already has a rules engine (`rules-service`) that reacts to real-time events (motion, light level) and executes actions via HTTP to `lighting-service` and `security-service`. The `schedule-service` is a separate bounded context: instead of reacting to sensor events, it fires actions at predetermined times. The action execution mechanism (HTTP to existing services, or NATS publish) is already proven in the codebase.

The core value proposition is: **reliable scheduled execution that survives service restarts**, with per-user isolation.

---

## Table Stakes

Features that users of a scheduling service expect. Missing any of these makes the service feel incomplete or untrustworthy.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Fixed daily time schedules | "Turn off lights at 4pm every day" is the canonical use case | Low | Simple: store HH:MM + days-of-week bitmask, tick every minute |
| Day-of-week filter | Weekday vs weekend behavior is a basic scheduling need | Low | Represented as a bitmask or array on the schedule entity |
| One-shot (future datetime) schedules | "Turn off lights in 2 hours" — time-bound automation | Medium | Requires storing absolute execution time + auto-delete or mark as executed after firing |
| Recurring cron-like schedules | Arbitrary recurrence not expressible as daily-at-X | Medium | Standard cron expression string (e.g., `"0 8 * * 1"`) enables full flexibility; requires a Go cron parser |
| Persist schedules across restarts | Core value of the service — in-memory-only is not acceptable | Low | PostgreSQL persistence already established via `pkg/database`; schedules loaded and re-registered on startup |
| Multi-tenant isolation by user_id | Every other service enforces `BelongsTo(userID)` ownership; schedule must too | Low | `owner_id` column + `BelongsTo` check on all read/write/delete operations |
| CRUD REST API | Create, list, get by ID, update, delete schedules | Low | Consistent with rules-service handlers pattern: POST, GET, DELETE, PUT, PATCH toggle |
| Enable/disable toggle | Users need to temporarily disable without deleting | Low | `enabled` boolean on entity; checked before firing; already proven in rules-service |
| Action: turn on/off light | Lighting control is the primary action target | Low | HTTP POST to `lighting-service` — same mechanism already in `rules-service.executeAction` |
| Action: trigger/silence alarm | Security control is the secondary action target | Low | HTTP POST to `security-service` — same mechanism |
| Action: publish NATS event | Generic event target for extensibility and decoupling | Medium | Requires `schedule-service` to hold a NATS connection and publish a configurable topic + payload |
| JWT authentication | All services require JWT; schedule-service is no exception | Low | Use `pkg/jwt` middleware — already shared |

---

## Differentiators

Features that go beyond what users expect. Not required for the service to be useful, but provide meaningful value once table stakes are met.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Action: HTTP call to arbitrary URL | Extensibility beyond known services — any future service (thermostat, media) can be a target without changing schedule-service code | Medium | Store `method`, `url`, `headers`, `body` on the action. Dangerous if open to external URLs; scope to internal network / known service hosts only in v1 |
| Named schedules with descriptions | Makes the list of schedules readable and manageable by users | Low | Just a `description` string field on the entity — trivial to add |
| Next-execution timestamp in API response | Frontend can show "next runs at 4:00 PM today" without the user computing it | Low | Compute at read time from schedule expression, no persistence needed |
| Execution history / audit log | "Did my 4pm schedule actually run?" — observable and debuggable system | Medium | Insert a row into `schedule_executions` table on each fire. Queryable via GET `/schedules/:id/history`. Integrates naturally with `notifications-service` via NATS publish on execution |
| Schedule dry-run / preview | "When would this schedule next fire?" without creating it | Low | A `POST /schedules/preview` endpoint that parses the expression and returns the next N execution times |
| Timezone-aware scheduling | Different users in different timezones; "at 8am" means different UTC offsets | Medium | Store timezone string (IANA e.g., `"America/Sao_Paulo"`) per schedule. All existing timestamps in Aurora appear to use server local time; UTC + offset conversion adds complexity. Flag for later milestone |

---

## Anti-Features

Features to deliberately NOT build in v1. Each has a clear reason and an alternative that already exists or is simpler.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Calendar / UI frontend | Out of scope per PROJECT.md. Frontend is not this service's responsibility | The REST API surface is the contract; frontend team consumes it |
| Geo-triggered schedules ("when I arrive home") | Location tracking requires a new bounded context (presence service), GPS data pipeline, geofence calculations — entirely different domain | Explicitly deferred in PROJECT.md as high complexity |
| Schedule chaining / dependencies ("run A then B after C completes") | This is workflow orchestration (e.g., Temporal, Argo), not scheduling. Conflating the two turns schedule-service into a workflow engine | If needed later, add a dedicated workflow-service or use rules-service for chaining via NATS events |
| Push notifications on schedule execution | Delegated explicitly in PROJECT.md: `notifications-service` subscribes to NATS events. Schedule-service publishes `schedule.executed` to NATS; notifications-service handles delivery | Publish a `schedule.executed` NATS event on fire; notifications-service already listens for all domain events |
| Rate limiting / throttle per user | Premature optimization; no evidence Aurora has abuse concern at this scale | Standard JWT auth provides per-user scope; add if needed in a later phase |
| Complex cron validation UI | Cron expression validation is a string parsing concern, not a feature | Validate on ingest (return 400 if expression is invalid); reject at API boundary |
| Cross-schedule coordination (mutual exclusion, locks) | Not a scheduling primitive. A schedule is a fire-and-forget timer | Each schedule fires independently; action idempotency is the responsibility of the target service |
| Sub-minute precision (second-level schedules) | Smart home use cases are all minute-level (4:00 PM, not 4:00:30 PM). Sub-minute requires a high-frequency tick loop and increases CPU load | Minimum granularity: 1 minute. Enforce in domain validation |

---

## Feature Dependencies

```
JWT auth middleware
    → required by: all CRUD endpoints

PostgreSQL persistence
    → required by: CRUD, enable/disable toggle, restart survival, execution history

Schedule entity (id, owner_id, name, type, expression, enabled, action_type, action_target, action_payload)
    → required by: CRUD, scheduler tick loop, execution history

Scheduler tick loop (goroutine that polls every minute or uses cron library)
    → required by: all schedule types (fixed, cron, one-shot)
    → requires: schedule entity + persistence

Action executor (HTTP client to lighting/security, NATS publisher)
    → required by: schedule firing
    → requires: scheduler tick loop

Day-of-week filter
    → required by: fixed time schedules
    → is a sub-feature of the schedule entity, not independent

One-shot auto-cleanup / mark-as-executed
    → required by: one-shot schedule type
    → requires: scheduler tick loop + persistence

Execution history table
    → requires: scheduler tick loop (writes on each fire)
    → enhances: observability, audit

Next-execution computation
    → requires: schedule expression parsing (cron library)
    → enables: preview endpoint, API response enrichment
```

---

## Schedule Type Breakdown

This project requires four schedule types. Each has distinct domain modeling needs.

| Type | Expression Model | Fire Logic | Cleanup |
|------|-----------------|------------|---------|
| Fixed daily time | `time_of_day` (HH:MM) + `days_of_week` (bitmask or array) | Fire when current time matches HH:MM and current weekday is in bitmask | Never — recurs indefinitely until disabled/deleted |
| Cron expression | `cron_expression` string (5-field standard) | Delegate to Go cron library (e.g., `robfig/cron`) for next-fire computation | Never — recurs per expression |
| One-shot (absolute) | `execute_at` timestamp (UTC) | Fire when `now >= execute_at` and not yet fired | Mark `fired = true` after execution; optionally auto-delete or retain for history |
| One-shot (relative, "in N minutes") | Convert at creation time to absolute `execute_at` | Same as absolute one-shot | Same as absolute one-shot |

Recommendation: model all as a single `Schedule` entity with a `schedule_type` enum. Fixed-daily and cron share the recurrence logic; one-shot is just a degenerate case where recurrence does not apply.

---

## Action Type Breakdown

| Action Type | Target | Payload Needed | Integration Mechanism | Confidence |
|-------------|--------|---------------|----------------------|------------|
| `turn_on_light` | `lighting-service` POST `/lights/:id/on` | `light_id` | HTTP with `X-Device-Key` header | HIGH — proven in rules-service |
| `turn_off_light` | `lighting-service` POST `/lights/:id/off` | `light_id` | HTTP with `X-Device-Key` header | HIGH — proven in rules-service |
| `trigger_alarm` | `security-service` POST `/alarms/trigger` | `trigger_type`, `sensor_id`, `location` | HTTP with `X-Device-Key` header | HIGH — proven in rules-service |
| `silence_alarm` | `security-service` POST `/alarms/silence` (or equivalent) | none or alarm_id | HTTP | MEDIUM — endpoint needs verification against security-service routes |
| `publish_nats_event` | NATS broker | `topic`, `payload` (JSON) | NATS publish | MEDIUM — NATS publisher already in sensors-service pattern |
| `http_call` | arbitrary internal URL | `method`, `url`, `body` | HTTP | LOW confidence for v1 — scope unclear, defer to later phase |

---

## MVP Recommendation

Prioritize in this order:

1. **Schedule entity + PostgreSQL persistence** — everything depends on this
2. **CRUD REST API with JWT auth** — makes the service usable
3. **Fixed daily time + day-of-week** — covers the primary stated use case ("turn off lights at 4pm")
4. **One-shot schedules** — covers "in 2 hours" use case
5. **Action: turn on/off light, trigger alarm** — HTTP execution to known services
6. **Enable/disable toggle** — needed immediately for usability
7. **Cron expression schedules** — adds full flexibility, low marginal cost once scheduler is wired
8. **Action: publish NATS event** — makes the service extensible and decouples it from hardcoded target services

Defer to a follow-on phase:
- **Execution history** — valuable but not blocking; `notifications-service` via NATS already provides some audit trail
- **Next-execution in API response** — nice UX but not blocking
- **Timezone support** — adds complexity; reasonable to start UTC-only with a migration path
- **Arbitrary HTTP action** — defer until a concrete use case exists beyond known services

---

## Sources

- Project requirements: `.planning/PROJECT.md` — HIGH confidence (authoritative)
- Existing action execution patterns: `services/rules-service/internal/application/rules_engine.go` — HIGH confidence (verified codebase)
- Integration patterns: `.planning/codebase/INTEGRATIONS.md` — HIGH confidence (verified codebase audit)
- Architecture conventions: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/CONVENTIONS.md` — HIGH confidence (verified codebase audit)
- Schedule type modeling: Standard scheduling system design (cron, one-shot, recurring) — MEDIUM confidence (established patterns, no external verification due to search unavailability)
- `robfig/cron` Go library patterns: Training knowledge — MEDIUM confidence (well-known library, needs version verification in STACK.md)
