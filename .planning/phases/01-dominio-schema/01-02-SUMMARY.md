---
phase: 01-dominio-schema
plan: 02
subsystem: schedule-service/domain
tags: [domain, interfaces, errors, repository, dispatcher, registry]
dependency_graph:
  requires: []
  provides: [ScheduleRepository, ScheduleExecutionRepository, ActionDispatcher, SchedulerRegistry, sentinel-errors]
  affects: [01-03-PLAN, 01-04-PLAN, phase-2, phase-3, phase-4]
tech_stack:
  added: []
  patterns: [domain-interfaces, sentinel-errors, DDD-dependency-rule]
key_files:
  created:
    - services/schedule-service/internal/domain/errors.go
    - services/schedule-service/internal/domain/repository.go
    - services/schedule-service/internal/domain/dispatcher.go
    - services/schedule-service/internal/domain/scheduler_registry.go
    - services/schedule-service/internal/domain/schedule_execution.go
  modified:
    - services/schedule-service/internal/domain/schedule.go
    - services/schedule-service/internal/domain/action.go
decisions:
  - "Sentinel errors consolidated in errors.go; removed temporary vars from schedule.go and action.go to prevent redeclaration"
  - "schedule_execution.go created as Rule 3 deviation to unblock repository.go compilation in parallel wave 1 execution"
metrics:
  duration: "80s"
  completed_date: "2026-03-26"
  tasks_completed: 2
  files_created: 5
  files_modified: 2
---

# Phase 01 Plan 02: Interfaces e Erros do Dominio Summary

**One-liner:** Domain interfaces (4 repository/dispatcher/registry) and all 7 sentinel errors consolidated into errors.go with zero external imports.

## What Was Built

Plan 02 of Phase 01 created all domain interface contracts and consolidated sentinel errors for the schedule-service. The domain layer now has:

- `errors.go`: 7 sentinel error variables (ErrScheduleNotFound, ErrScheduleAccessDenied, ErrInvalidSchedule, ErrMissingCronExpression, ErrMissingRunAt, ErrInvalidAction, ErrPublishFailed)
- `repository.go`: ScheduleRepository (Save, FindByID, FindByOwnerID, FindAllEnabled, Delete) and ScheduleExecutionRepository (Save, FindByScheduleID) interfaces
- `dispatcher.go`: ActionDispatcher interface with Dispatch method accepting *Schedule
- `scheduler_registry.go`: SchedulerRegistry interface with Register and Unregister methods

All files use only stdlib (no external imports), fulfilling the DDD domain purity constraint.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create sentinel errors and repository interfaces | 140c309 | errors.go, repository.go, schedule.go (modified), action.go (modified), schedule_execution.go |
| 2 | Create ActionDispatcher and SchedulerRegistry interfaces | ae4a238 | dispatcher.go, scheduler_registry.go |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Created schedule_execution.go to unblock repository.go compilation**
- **Found during:** Task 1
- **Issue:** `repository.go` references `ScheduleExecution` type which plan 01-01 was supposed to create in `schedule_execution.go`. Since both plans run in parallel (wave 1) and plan 01-01 had not yet created this file, `go build ./internal/domain/...` failed with "undefined: ScheduleExecution"
- **Fix:** Created `schedule_execution.go` with `ScheduleExecution` entity, `ExecutionStatus` type, and `NewScheduleExecution` constructor — exactly matching the spec in plan 01-01
- **Files modified:** services/schedule-service/internal/domain/schedule_execution.go (created)
- **Commit:** 140c309

**2. [Rule 1 - Bug] Removed duplicate error variable declarations from schedule.go and action.go**
- **Found during:** Task 1
- **Issue:** Plan 01-01 had created temporary error vars in schedule.go (ErrInvalidSchedule, ErrMissingCronExpression, ErrMissingRunAt) and action.go (ErrInvalidAction). Creating errors.go with these same vars would cause redeclaration compile errors
- **Fix:** Removed temporary error vars from schedule.go and action.go; errors now live only in errors.go as intended
- **Files modified:** services/schedule-service/internal/domain/schedule.go, services/schedule-service/internal/domain/action.go
- **Commit:** 140c309

## Success Criteria Verification

- [x] errors.go has all 7 sentinel errors (SCH-03 access denied, SCH-07 missing run_at, SCH-08 missing cron)
- [x] ScheduleRepository has FindAllEnabled for SCH-02 (load on startup)
- [x] ScheduleExecutionRepository has FindByScheduleID for SCH-04
- [x] ActionDispatcher defined for Phase 2/4 dispatcher implementations
- [x] SchedulerRegistry defined for Phase 2/4 CronRunner
- [x] All files compile with zero external imports (`go build ./internal/domain/...` exits 0)
- [x] `go vet ./internal/domain/...` exits 0

## Known Stubs

None — all interfaces are pure contracts with no implementation stubs.

## Self-Check: PASSED
