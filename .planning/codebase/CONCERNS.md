# Codebase Concerns

**Analysis Date:** 2026-03-25

## Unimplemented Features

**TriggerSchedule Type - Critical Blocker:**
- Issue: `TriggerSchedule` is defined as a `TriggerType` constant in `/home/vini/Workspace/Aurora/services/rules-service/internal/domain/rule.go` (line 21), and a `ScheduleTime` value object exists with full validation and tests (`/home/vini/Workspace/Aurora/services/rules-service/internal/domain/schedule_time.go`), but there is **NO evaluation method** for schedule triggers in the rules engine.
- Files:
  - `/home/vini/Workspace/Aurora/services/rules-service/internal/domain/rule.go` (defines `TriggerSchedule`)
  - `/home/vini/Workspace/Aurora/services/rules-service/internal/domain/schedule_time.go` (value object exists)
  - `/home/vini/Workspace/Aurora/services/rules-service/internal/application/rules_engine.go` (missing `EvaluateScheduleTrigger()`)
- Impact: Users cannot create rules with schedule-based triggers. API accepts `trigger_type: "schedule"` but the engine has no handler for it. Rules with this trigger will be silently ignored.
- Fix approach:
  1. Add `ScheduleTime` field to `Rule` struct or create separate `ScheduledRule` type
  2. Implement `EvaluateScheduleTrigger(scheduleName string, scheduledTime domain.ScheduleTime)` in `RulesEngine`
  3. Add a background scheduler (cron-like) that periodically checks if any scheduled rules should fire
  4. Write comprehensive tests for schedule evaluation and timing edge cases
  5. Update migration to add schedule_time column to rules table

**Missing Schedule Trigger Tests:**
- Issue: No tests exist for `TriggerSchedule` evaluation. The `rules_engine_test.go` covers motion and light triggers but has zero coverage for schedule-based triggers.
- Files: `/home/vini/Workspace/Aurora/services/rules-service/internal/application/rules_engine_test.go` (lines 134-215 test motion/light only)
- Impact: Any schedule implementation will lack validation that scheduling works correctly across day boundaries, timezone handling, and race conditions.
- Fix approach: Add test cases for `EvaluateScheduleTrigger` including: exact time matches, off-by-one minute errors, multiple rules for same time, and day boundary cases.

## Test Coverage Gaps

**HTTP Handler Integration Tests Missing:**
- Issue: No tests exist for HTTP handlers at all services. Only application-layer tests exist for domain/application logic.
- Files affected:
  - `/home/vini/Workspace/Aurora/services/auth-service/internal/infrastructure/http/` (no *_test.go)
  - `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/http/` (no *_test.go)
  - `/home/vini/Workspace/Aurora/services/security-service/internal/infrastructure/http/` (no *_test.go)
  - `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/http/` (no *_test.go)
  - `/home/vini/Workspace/Aurora/services/rules-service/internal/infrastructure/http/` (no *_test.go)
  - `/home/vini/Workspace/Aurora/services/notifications-service/internal/infrastructure/http/` (no *_test.go)
- Impact:
  - JWT/auth middleware bugs won't be caught (e.g., malformed Authorization headers, missing X-Device-Key handling)
  - Request binding issues invisible (e.g., invalid JSON, missing required fields)
  - HTTP status codes could be wrong without detection
  - Auth bypass vulnerabilities could exist
- Fix approach: Create integration tests for each service testing GET/POST/PUT/DELETE endpoints with various request formats, auth scenarios, and error cases.

**WebSocket Connection Management - Potential Memory Leak:**
- Issue: In `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/ws/hub.go` and `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/ws/hub.go`, connections are only removed from the hub map when they actively read and encounter an error. If a client silently disconnects without being detected, the stale connection remains in `h.clients` indefinitely.
- Files:
  - `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/ws/hub.go` (lines 82-84)
  - `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/ws/hub.go` (lines 81-83)
- Impact: Long-running server could accumulate dead WebSocket connections, causing memory leak and failed broadcast attempts to stale clients.
- Fix approach:
  1. Implement heartbeat/ping-pong mechanism (WebSocket built-in support)
  2. Add write timeout to detect stale connections during `Broadcast()` writes
  3. Move failed write to a cleanup queue and remove from map
  4. Add tests for connection cleanup

**Frontend Test Coverage - Severely Limited:**
- Issue: Only 8 test files exist for ~8200 TypeScript files in the frontend. Multiple critical services have no tests.
- Files: `/home/vini/Workspace/Aurora/frontend/src/**/*.ts` (no spec.ts for most services)
- Impact: API service layer, interceptors, guards could have bugs that break at runtime in production.
- Fix approach: Prioritize tests for authentication service, HTTP interceptors, and route guards.

## Security Concerns

**JWT Secret Management - Default Insecure Value:**
- Issue: In `/home/vini/Workspace/Aurora/services/auth-service/cmd/api/main.go` (lines 17-20), if `JWT_SECRET` environment variable is not set, it falls back to `"default-dev-secret-key"`.
- Files: `/home/vini/Workspace/Aurora/services/auth-service/cmd/api/main.go` (lines 17-20)
- Impact: In production without proper env var setup, JWT tokens can be forged. Anyone knowing the default secret can generate valid tokens for any user.
- Mitigation present: Code requires explicit env var (though with unsafe fallback)
- Fix approach:
  1. Remove the default value entirely, panic if `JWT_SECRET` is not set
  2. Add startup validation that `JWT_SECRET` is at least 32 characters
  3. Document required environment variables prominently

**Device API Key in Source Code:**
- Issue: Multiple services use `DEVICE_API_KEY` environment variable with fallback defaults. In `/home/vini/Workspace/Aurora/services/security-service/cmd/api/main.go`, the default is `"aurora-device-key-2024"`.
- Files:
  - `/home/vini/Workspace/Aurora/services/security-service/cmd/api/main.go`
  - `/home/vini/Workspace/Aurora/services/lighting-service/cmd/api/main.go` (similar pattern)
- Impact: Any service to service calls using X-Device-Key can be spoofed if the default key is used in production.
- Fix approach:
  1. Remove hardcoded defaults for DEVICE_API_KEY
  2. Require it to be set in all environments
  3. Generate a strong random key during deployment setup

**WebSocket CheckOrigin Disabled:**
- Issue: In both WebSocket hubs (`/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/ws/hub.go` line 34 and `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/ws/hub.go` line 52), `CheckOrigin` returns `true` unconditionally.
- Files:
  - `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/ws/hub.go` (line 34: `CheckOrigin: func(r *http.Request) bool { return true }`)
  - `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/ws/hub.go` (line 52: same pattern)
- Impact: CSRF attacks possible - a malicious website can establish WebSocket connections to the Aurora service on behalf of authenticated users.
- Fix approach:
  1. Validate that `Origin` header matches expected domain
  2. Or use JWT in WebSocket upgrade handshake instead of relying on Origin

**Password Minimum Length Only 6 Characters:**
- Issue: In `/home/vini/Workspace/Aurora/services/auth-service/internal/application/auth_service.go` (line 48), password is validated as `len(req.Password) < 6`. No upper length limit, no character requirements.
- Files: `/home/vini/Workspace/Aurora/services/auth-service/internal/application/auth_service.go` (line 48)
- Impact: Weak passwords like `"123456"` are accepted. No NIST 800-63b compliance.
- Fix approach:
  1. Increase minimum to 12 characters or enforce passphrase approach
  2. Add zxcvbn entropy checking (optional but recommended)
  3. Add password history to prevent reuse

**Error Information Disclosure - Too Specific:**
- Issue: HTTP error responses may leak sensitive information. For example, in auth service handlers, distinguish between "user not found" vs "password incorrect" could enable email enumeration.
- Files: Check all handler files for error message specificity
- Impact: Attackers can enumerate valid email addresses in the system
- Mitigation: Code in `/home/vini/Workspace/Aurora/services/auth-service/internal/application/auth_service.go` already does this correctly (line 80 returns generic `ErrInvalidCredentials`)
- Remaining risk: HTTP handlers must properly translate domain errors to generic HTTP 401 responses

## Performance Bottlenecks

**Synchronous Event Publishing on Hot Path:**
- Issue: In `/home/vini/Workspace/Aurora/services/sensors-service/internal/application/light_service.go` (lines 58-61), if event publishing fails, the error is silently ignored (`_ = err`). This is good for resilience, but if publishing is slow, it blocks sensor registration.
- Files: `/home/vini/Workspace/Aurora/services/sensors-service/internal/application/light_service.go` (line 60)
- Impact: If NATS is slow or down, sensor readings will be delayed on the HTTP response path
- Fix approach: Make event publishing async with a queue, so HTTP response returns immediately

**Rules Engine Motion Auto-Off Uses Unbounded Goroutines:**
- Issue: In `/home/vini/Workspace/Aurora/services/rules-service/internal/application/rules_engine.go` (line 203), every motion trigger creates a new goroutine that sleeps for `motionOffTimeout`. With many motion sensors, this creates unbounded goroutine growth.
- Files: `/home/vini/Workspace/Aurora/services/rules-service/internal/application/rules_engine.go` (lines 196-225)
- Impact: Under high motion event frequency, goroutine count grows unbounded, eventually exhausting memory.
- Fix approach:
  1. Use a goroutine pool or worker pattern
  2. Or use a timer wheel/priority queue to batch delayed actions
  3. Add metrics for goroutine count

**Database Query Without Pagination Limits:**
- Issue: Multiple services query records without enforcing max limits. For example, sensor repository could return unlimited records.
- Files: `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/repository/sensor_repository.go` (check GetLightRecords)
- Impact: A sensor with millions of records could cause OOM if queried without limit
- Fix approach: Add hard maximum limits (e.g., max 10000 records per query) at repository layer

## Fragile Areas

**Rules Engine - Complex State Management:**
- Issue: `RulesEngine` maintains `lastMotionByLight` map with mutex synchronization. The logic for checking "did another motion happen?" (line 211 of rules_engine.go) is subtle and could have race conditions.
- Files: `/home/vini/Workspace/Aurora/services/rules-service/internal/application/rules_engine.go` (lines 196-225)
- Why fragile:
  - The goroutine closure captures `now` at dispatch time but checks `lastMotionByLight` later, creating a timing window
  - If a motion happens between scheduler creation and the `time.Sleep`, the logic might incorrectly prevent auto-off
  - No tests verify race conditions (e.g., 3 rapid motions in succession)
- Safe modification: Always add tests that simulate rapid event sequences before changing the auto-off logic

**Lighting Service Dependency on ESP32Client - Hard to Test:**
- Issue: In `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/device/esp32_client.go`, the `ESP32Client` makes real HTTP calls to physical devices. It's not mockable at the infrastructure layer.
- Files: `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/device/esp32_client.go`
- Impact: Integration tests require actual ESP32 devices or complex HTTP mocking setup
- Fix approach:
  1. Create a device abstraction interface in domain layer
  2. Have ESP32Client implement it
  3. Create a mock device for tests

**Database Migrations Not Versioned:**
- Issue: Migrations in all services use raw SQL executed once without version tracking. If two developers add migrations independently, there's no version control.
- Files: All `*_migrations.go` files
- Impact: Cannot safely rollback migrations, hard to track which migrations have run on production
- Fix approach: Use a migration tool like `golang-migrate` or `sql-migrate` with versioning

## Missing Critical Features

**No Audit Logging:**
- Issue: No audit trail of who created/modified/deleted rules, enabled/disabled devices, etc.
- Impact: Compliance gap (GDPR, SOX, etc.), can't investigate security incidents
- Fix approach: Add audit log tables and timestamp/user tracking to all mutation operations

**No API Rate Limiting:**
- Issue: No mention of rate limiting on HTTP endpoints
- Impact: Brute force attacks on auth endpoints, DDoS vulnerability
- Fix approach: Add rate limiting middleware (e.g., `go-ratelimit`)

**No Refresh Token Mechanism:**
- Issue: JWT tokens expire in 1 hour (auth_service.go line 35), but there's no refresh token endpoint
- Impact: Users get logged out after 1 hour with no way to stay logged in
- Fix approach:
  1. Add refresh token endpoint
  2. Issue both access (short-lived) and refresh (long-lived) tokens
  3. Store refresh tokens in database with rotation tracking

**No Input Validation - OWASP Concern:**
- Issue: Request payloads are only partially validated. For example, `CreateRuleRequest` in rules_engine.go only checks if Name is empty (line 86) but doesn't validate other fields like TriggerType.
- Files: All handler files
- Impact: Invalid enum values might be accepted and persisted
- Fix approach:
  1. Add field-level validation with `github.com/go-playground/validator`
  2. Validate enum values against allowed constants
  3. Validate string lengths, email formats, etc.

## Database Schema Issues

**No Foreign Key Constraints:**
- Issue: Rules table references action_id and trigger_id but has no foreign keys to validate they exist
- Files: `/home/vini/Workspace/Aurora/services/rules-service/internal/infrastructure/migrations/rule_migrations.go`
- Impact: Orphaned rules can be created pointing to non-existent devices/sensors
- Fix approach: Add FK constraints or at least document the contract

**Decimal(5,2) for Threshold May Be Insufficient:**
- Issue: `trigger_threshold DECIMAL(5,2)` limits values to 0-999.99, but light percentages should be 0-100. This is actually fine, but poorly chosen precision.
- Files: `/home/vini/Workspace/Aurora/services/rules-service/internal/infrastructure/migrations/rule_migrations.go` (line 18)
- Fix approach: Use `DECIMAL(3,1)` or just regular FLOAT if precision isn't critical

## Code Quality Issues

**Inconsistent Error Handling Pattern:**
- Issue: Some code ignores errors explicitly (`_ = err` in light_service.go), others propagate them, others silently fail
- Files: All services
- Impact: Inconsistent behavior makes debugging harder
- Fix approach: Establish error handling policy - define when to ignore, when to log, when to propagate

**Missing Repository Tests:**
- Issue: Repository layer (database) has no tests. Only mocks are used in application tests.
- Files: All `postgres_*_repository.go` files
- Impact: Schema changes or query bugs won't be caught until production
- Fix approach: Add integration tests with test database

**WebSocket Error Handling - Errors Silently Logged:**
- Issue: In both hub implementations, `WriteMessage` errors are only logged, not handled further. Stale connections aren't removed.
- Files:
  - `/home/vini/Workspace/Aurora/services/lighting-service/internal/infrastructure/ws/hub.go` (line 83-84)
  - `/home/vini/Workspace/Aurora/services/sensors-service/internal/infrastructure/ws/hub.go` (line 101-102)
- Impact: Dead connections remain in broadcast loop, consuming resources
- Fix approach: Track write failures and remove connection from map after N consecutive failures

---

*Concerns audit: 2026-03-25*
