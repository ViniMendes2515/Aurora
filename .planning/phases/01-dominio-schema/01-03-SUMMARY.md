---
phase: 01-dominio-schema
plan: 03
subsystem: schedule-service/migrations
tags: [postgresql, migration, schema, go]
dependency_graph:
  requires: []
  provides:
    - RunScheduleMigrations function in package migrations
    - schedules table with 18 columns
    - schedule_executions table with FK ON DELETE CASCADE
    - 3 performance indexes
  affects:
    - services/schedule-service (new service foundation)
tech_stack:
  added: []
  patterns:
    - Single-string multi-statement db.Exec migration (rules-service pattern)
    - CREATE TABLE IF NOT EXISTS for idempotent migrations
key_files:
  created:
    - services/schedule-service/internal/infrastructure/migrations/schedule_migrations.go
  modified: []
decisions:
  - All 18 schedules columns baked in from day one — no ALTER TABLE needed in Phases 2-4
  - timezone VARCHAR(64) DEFAULT UTC present from initial migration per locked architectural decision
  - ON DELETE CASCADE on schedule_executions ensures history cleanup when schedule is deleted
  - Single db.Exec with multi-statement SQL matches rules-service migration pattern exactly
metrics:
  duration: ~5 minutes
  completed: 2026-03-26
  tasks_completed: 1
  tasks_total: 1
  files_created: 1
  files_modified: 0
---

# Phase 1 Plan 3: PostgreSQL Migration — schedules + schedule_executions Summary

**One-liner:** RunScheduleMigrations creates both tables with all v1 fields (timezone, schedule_type, days_filter, run_at, last_run_at, next_run_at) and 3 indexes using idempotent CREATE IF NOT EXISTS SQL.

## What Was Built

Created `services/schedule-service/internal/infrastructure/migrations/schedule_migrations.go` with `RunScheduleMigrations(db *sql.DB) error`.

The migration creates:

1. **schedules table** (18 columns): id, owner_id, name, description, schedule_type, cron_expr, run_at (nullable), days_filter (bitmask integer), timezone (default UTC), action_target, action_type, action_device_id, action_payload, enabled, created_at, updated_at, last_run_at (nullable), next_run_at (nullable).

2. **schedule_executions table** (5 columns): id, schedule_id (FK → schedules(id) ON DELETE CASCADE), executed_at, status, error_message.

3. **3 indexes**: idx_schedules_owner_id, idx_schedules_enabled, idx_executions_schedule_id.

## Commits

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Create PostgreSQL migration with schedules and schedule_executions tables | 49d9178 | services/schedule-service/internal/infrastructure/migrations/schedule_migrations.go |

## Decisions Made

- **All fields locked from day one:** timezone, schedule_type, days_filter, run_at, last_run_at, next_run_at are all present in the initial migration. This avoids painful ALTER TABLE migrations in Phases 2-4 when these fields are first consumed by business logic.

- **Single db.Exec with concatenated SQL:** Follows the rules-service pattern exactly — all statements in one multi-statement string passed to db.Exec. This is the established Aurora migration approach.

- **ON DELETE CASCADE:** Deleting a schedule automatically removes its execution history — correct behavior for multi-tenant cleanup.

- **no completed_at column:** One-shot atomicity is handled via enabled=false + last_run_at IS NOT NULL in a single Save(), per the ROADMAP architectural decision. No separate completed_at column needed.

## Deviations from Plan

None — plan executed exactly as written.

## Verification Results

- `go build ./internal/infrastructure/migrations/...` — PASSED (exit 0)
- `go vet ./internal/infrastructure/migrations/...` — PASSED (exit 0)
- All 22 acceptance criteria verified via grep checks — PASSED

## Known Stubs

None — this plan creates a pure SQL migration with no stub values. The migration is complete and will execute correctly against a PostgreSQL instance.

## Self-Check: PASSED

- File exists: services/schedule-service/internal/infrastructure/migrations/schedule_migrations.go — FOUND
- Commit exists: 49d9178 — FOUND
