---
phase: 01-dominio-schema
plan: 01
subsystem: domain
tags: [go, ddd, domain, schedule, entity, value-object]

# Dependency graph
requires: []
provides:
  - Schedule entity with 15 fields, ScheduleType type, NewSchedule constructor with validation
  - BelongsTo, IsActiveOnDay, Disable, RecordRun domain methods
  - Action value object with ActionTarget (nats/http), ActionType, NewAction constructor with target validation
  - ScheduleExecution entity with ExecutionStatus type and NewScheduleExecution constructor
  - All domain sentinel errors (ErrInvalidSchedule, ErrMissingCronExpression, ErrMissingRunAt, ErrInvalidAction, ErrScheduleNotFound, ErrScheduleAccessDenied, ErrPublishFailed)
  - go.mod initializing aurora/services/schedule-service module at go 1.21
affects: [01-02, 01-03, 01-04, phase-2, phase-3, phase-4]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - DDD domain entity with exported fields, constructor validation returning (*T, error)
    - Domain methods for business logic (BelongsTo, IsActiveOnDay, Disable, RecordRun)
    - Sentinel errors with errors.New — no custom error types
    - Value object with target validation (Action.NewAction checks nats/http)
    - DaysFilter as integer bitmask (0=all days, bit per weekday using time.Weekday shift)
    - Zero external imports in domain package (only time, errors from stdlib)

key-files:
  created:
    - services/schedule-service/go.mod
    - services/schedule-service/internal/domain/schedule.go
    - services/schedule-service/internal/domain/action.go
    - services/schedule-service/internal/domain/schedule_execution.go
    - services/schedule-service/internal/domain/errors.go
  modified: []

key-decisions:
  - "BelongsTo does NOT use wildcard '*' unlike rules-service — schedules are always user-specific per SCH-03"
  - "errors.go centralizes all sentinel errors — no temporary duplicate vars in entity files"
  - "DaysFilter=0 means all days (not no days) — tested via bitmask 1<<uint(day)"
  - "Cron expression validation is presence-only in domain (empty check) — syntax validation deferred to infrastructure layer"
  - "Action.Payload is stored as string not []byte — consistent with existing Aurora service patterns"

patterns-established:
  - "Pattern: Domain entity: exported struct fields, no private fields, constructor returns (*T, error)"
  - "Pattern: Zero external imports in internal/domain/ — only time and errors from stdlib"
  - "Pattern: errors.go holds all sentinel errors for the package"
  - "Pattern: package domain (not package domain_test) in entity files"

requirements-completed: [SCH-01, SCH-02, SCH-03, SCH-04, SCH-05, SCH-06, SCH-07, SCH-08, SCH-09]

# Metrics
duration: 4min
completed: 2026-03-26
---

# Phase 1 Plan 1: Entidades do Dominio Summary

**Go domain layer for schedule-service: Schedule entity (15 fields, 4 methods), Action value object (nats/http target validation), ScheduleExecution entity, and sentinel errors — all compiling with zero external imports**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-03-26T11:33:40Z
- **Completed:** 2026-03-26T11:37:44Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Schedule entity with all 15 fields (ID, OwnerID, Name, Description, ScheduleType, CronExpr, RunAt, DaysFilter, Timezone, Action, Enabled, CreatedAt, UpdatedAt, LastRunAt, NextRunAt)
- NewSchedule constructor validates name, cron expression presence, run_at for one-shot; defaults timezone to "UTC"
- Action value object validates target (nats or http only), exposes 4 action type constants
- ScheduleExecution entity records execution history with ID, ScheduleID, ExecutedAt, Status, ErrorMessage
- `go build ./internal/domain/...` and `go vet ./internal/domain/...` both pass with zero errors

## Task Commits

Each task was committed atomically:

1. **Task 1: Initialize Go module and create Schedule entity** - `93d58e0` (feat)
   - go.mod, schedule.go, action.go, errors.go (schedule.go and action.go committed by parallel 01-02 agent as deviation Rule 3 unblock)
2. **Task 2: Create ScheduleExecution entity and Action value object** - `140c309` (feat by parallel agent)
   - schedule_execution.go committed in plan 01-02 context as it unblocked repository.go compilation

**Plan metadata:** pending (docs commit at end)

## Files Created/Modified
- `services/schedule-service/go.mod` - Module definition: aurora/services/schedule-service, go 1.21
- `services/schedule-service/internal/domain/schedule.go` - Schedule entity, ScheduleType, NewSchedule constructor, BelongsTo/IsActiveOnDay/Disable/RecordRun methods
- `services/schedule-service/internal/domain/action.go` - Action value object, ActionTarget/ActionType constants, NewAction constructor with target validation
- `services/schedule-service/internal/domain/schedule_execution.go` - ScheduleExecution entity, ExecutionStatus type, NewScheduleExecution constructor
- `services/schedule-service/internal/domain/errors.go` - All 7 sentinel errors for the schedule-service domain

## Decisions Made
- BelongsTo does NOT use wildcard `"*"` unlike rules-service (schedules are always user-specific, no system-owned schedules)
- DaysFilter bitmask uses 0 as "all days" sentinel — `1 << uint(day)` where `time.Sunday=0` maps to bit 1
- Cron expression validation deferred to infrastructure — domain only checks for empty string (ErrMissingCronExpression)
- All sentinel errors consolidated in errors.go immediately (no temporary duplicate vars in entity files)

## Deviations from Plan

### Context: Parallel Agent Execution

This plan (01-01) was executed in a parallel wave where another agent was running plan 01-02 concurrently. The parallel 01-02 agent committed schedule.go, action.go, errors.go, schedule_execution.go, repository.go, dispatcher.go, and scheduler_registry.go as part of its task commits (commits `140c309` and `ae4a238`). The plan 01-01 agent committed go.mod (commit `93d58e0`).

**[Rule 3 - Blocking] 01-02 agent pre-created domain entity files to unblock repository.go**
- **Found during:** Task 1 (parallel execution)
- **Issue:** repository.go (plan 01-02) referenced ScheduleExecution which required schedule_execution.go (plan 01-01 task 2) to exist for compilation
- **Fix:** The 01-02 agent created schedule_execution.go, schedule.go, action.go, errors.go as part of its work, correctly splitting the creation across the two plans' scope
- **Files modified:** All domain entity files
- **Verification:** `go build ./internal/domain/...` exits 0
- **Committed in:** 140c309 (parallel agent task commit)

---

**Total deviations:** 1 — parallel agent handled domain entity creation while 01-01 agent handled go.mod
**Impact on plan:** No scope creep. All 01-01 plan objectives were met. Files created correctly match acceptance criteria.

## Issues Encountered
- Parallel execution of plans 01-01 and 01-02 resulted in overlapping file creation. The parallel 01-02 agent's Rule 3 auto-fix pre-created the domain entity files before the 01-01 agent ran. The 01-01 agent verified all acceptance criteria were met and committed go.mod as its primary artifact.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All domain entities ready for use by plans 01-02 (already complete), 01-03, 01-04
- Domain package compiles with zero external imports — ready for Phase 2 application layer
- `go build ./internal/domain/...` and `go vet ./internal/domain/...` both pass

---
*Phase: 01-dominio-schema*
*Completed: 2026-03-26*
