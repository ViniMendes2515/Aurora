# External Integrations

**Analysis Date:** 2026-03-25

## APIs & External Services

**Hardware Device APIs:**
- ESP32 IoT Device - For lighting and security control
  - SDK/Client: Custom HTTP client implementation in `github.com/gorilla/websocket` and standard Go `net/http`
  - Implementation: `services/lighting-service/internal/infrastructure/device/esp32_client.go`, `services/security-service/internal/infrastructure/device/buzzer_client.go`
  - Authentication: `DEVICE_API_KEY` environment variable
  - Protocol: HTTP GET requests with 3-second timeout
  - Endpoints: IP address provided via `ESP32_IP` environment variable

## Data Storage

**Databases:**
- PostgreSQL 15-alpine
  - Provider: Docker container running `postgres:15-alpine`
  - Connection: Via `lib/pq` driver (golang PostgreSQL driver)
  - Client: Standard Go `database/sql` package
  - Connection Details:
    - Host: Environment variable `DB_HOST` (default: `postgres` in compose, `localhost` in dev)
    - Port: Environment variable `DB_PORT` (default: `5432`)
    - User: Environment variable `DB_USER` (default: `aurora`)
    - Password: Environment variable `DB_PASSWORD`
    - Database: Environment variable `DB_NAME` (default: `aurora_home`)
    - SSL Mode: Environment variable `DB_SSLMODE` (default: `disable`)
  - Location: `pkg/database/postgres.go` - Central connection management
  - Pool Configuration: 25 max open connections, 5 max idle, 5-minute lifetime
  - Migrations: Service-specific migrations in `services/*/internal/infrastructure/migrations/`
    - auth-service: User table and auth schema
    - sensors-service: Sensor and motion event tables
    - lighting-service: Light state tables
    - security-service: Alarm and alarm event tables
    - rules-service: Rule and trigger tables
    - notifications-service: Notification and event log tables

**File Storage:**
- Local filesystem only - No external file storage service configured
- Frontend assets served via Nginx from compiled build output
- No S3, cloud storage, or CDN integrations detected

**Caching:**
- None detected - No Redis, Memcached, or other caching layer configured

## Authentication & Identity

**Auth Provider:**
- Custom JWT implementation
  - Service: Internal auth-service (no third-party OAuth/OIDC)
  - Implementation: JWT HS256 symmetric signing (`github.com/golang-jwt/jwt/v5`)
  - Location: `services/auth-service/cmd/api/main.go`, `pkg/jwt/validator.go`, `pkg/jwt/middleware.go`
  - Token Secret: Environment variable `JWT_SECRET` (required for all services)
  - Token Expiration: 1 hour (hardcoded in auth-service)
  - Validation: Middleware in `pkg/jwt/middleware.go` validates tokens for protected endpoints
  - Password Security: Uses `golang.org/x/crypto` package for password hashing

**User Management:**
- Auth service stores users in PostgreSQL schema `aurora_home.users`
- No LDAP, Active Directory, or external identity provider integration
- Login endpoint: `/api/auth/login` (via Nginx API Gateway)
- Token refresh: Not yet implemented (1-hour expiration)

## Monitoring & Observability

**Error Tracking:**
- None detected - No Sentry, Rollbar, or error tracking service integration

**Logs:**
- Standard Go `log` package for console output
- Log locations:
  - Database connection retries: `pkg/database/postgres.go`
  - NATS connection status: `services/*/cmd/api/main.go`
  - Service startup messages: Each service main.go
  - WebSocket hub events: `services/lighting-service/internal/infrastructure/ws/hub.go`
- No log aggregation, no structured logging framework detected
- All logs output to stdout (Docker container logs)

**Metrics:**
- None detected - No Prometheus, DataDog, or metrics collection

**Tracing:**
- None detected - No distributed tracing (Jaeger, DataDog APM)

## CI/CD & Deployment

**Hosting:**
- Docker containerization - All services and dependencies containerized
- Docker Compose orchestration (`infra/docker-compose.yml`) for local/development deployment
- Multi-stage Docker builds: Frontend build stage (Node 20) → production stage (Nginx alpine)
- No cloud platform integration detected (AWS, GCP, Azure)

**CI Pipeline:**
- None detected - No GitHub Actions, GitLab CI, Jenkins, or other CI service configured
- Git repository present but no `.github/workflows/`, `.gitlab-ci.yml`, or similar

**Container Registry:**
- None detected - Services build locally from Dockerfile on each `docker compose up --build`
- No Docker Hub, ECR, GCR, or private registry configuration

## Environment Configuration

**Required env vars (all services):**
- `JWT_SECRET` - CRITICAL: Secret key for JWT signing/validation across all services
- `SERVER_PORT` - HTTP port (service-specific, 8080-8085)
- `DB_HOST` - PostgreSQL hostname
- `DB_PORT` - PostgreSQL port (5432)
- `DB_USER` - Database user (aurora)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (aurora_home)
- `DB_SSLMODE` - SSL mode for DB connection (disable for dev)
- `NATS_URL` - NATS message broker URL (nats://nats:4222 in compose)

**Service-Specific Required Vars:**
- Sensors/Lighting/Security/Rules: `DEVICE_API_KEY` - Device authentication key
- Lighting/Security: `ESP32_IP` - IP address of ESP32 device
- Security: `DEVICE_ID`, `BUZZER_DURATION_MS` - Device config
- Rules: `MOTION_OFF_TIMEOUT_SECONDS`, `LIGHTING_SERVICE_URL`, `SECURITY_SERVICE_URL`

**Secrets location:**
- Development: Environment variables in `infra/docker-compose.yml` (hardcoded for local testing)
- Production: Should be injected via environment or secrets management system
- `.env` file exists in `infra/` directory but not read during compose (vars in compose.yml directly)

**No external secrets management detected** (no HashiCorp Vault, AWS Secrets Manager, etc.)

## Message Broker & Event Streaming

**NATS Pub/Sub:**
- Service: NATS (deployed as `nats:latest` Docker container)
- Port: 4222 (messaging), 8222 (HTTP monitoring)
- SDK/Client: `github.com/nats-io/nats.go v1.31.0`
- Connection Details:
  - URL: Environment variable `NATS_URL` (default: `nats://localhost:4222`)
  - Retry logic: 10 reconnect attempts with exponential backoff
  - Location: `services/*/internal/infrastructure/messaging/nats_*.go`

**Published Events (Pub/Sub Topics):**
- `sensors.motion.detected` - Published by sensors-service, subscribed by rules-service and notifications-service
- `sensors.light.changed` - Published by sensors-service, subscribed by notifications-service
- `lighting.state.updated` - Published by lighting-service (implied in rules-service consumption)
- `security.alarm.triggered` - Published by security-service (implied in rules-service consumption)

**Event Producers:**
- `sensors-service`: Motion and light sensor events
  - Implementation: `services/sensors-service/internal/infrastructure/messaging/nats_publisher.go`
- `lighting-service`: Light state changes
- `security-service`: Alarm events
- `rules-service`: Potential automation triggers (acts as consumer primarily)

**Event Consumers:**
- `rules-service`: Subscribes to sensor events via NATS
  - Location: `services/rules-service/internal/application/rules_engine.go`
  - Makes HTTP calls to lighting-service and security-service based on rules
- `notifications-service`: Subscribes to motion and light events
  - Location: `services/notifications-service/internal/infrastructure/messaging/nats_subscriber.go`

## Inter-Service Communication

**Synchronous (HTTP/REST):**
- Rules-service calls lighting-service and security-service
  - URLs: `LIGHTING_SERVICE_URL` (http://lighting-service:8082), `SECURITY_SERVICE_URL` (http://security-service:8083)
  - Method: HTTP POST with JSON body
  - Timeout: 3 seconds per request
  - Authentication: JWT token in Authorization header
  - Location: `services/rules-service/internal/application/rules_engine.go`

**Asynchronous (NATS Pub/Sub):**
- Event-driven architecture for loose coupling
- No request-reply pattern detected (pure pub/sub only)

## Webhooks & Callbacks

**Incoming:**
- None detected - No webhook endpoints for external systems to trigger

**Outgoing:**
- None detected - No external service webhooks being called

**Device Callbacks:**
- ESP32 device integration: One-way HTTP GET requests from services to device
- No callbacks/webhooks from device to backend detected

## Frontend-Backend Communication

**REST API:**
- Angular frontend communicates with all services via HTTP
- Routed through Nginx API Gateway on port 80 at `/api/*`
- Path mapping: `/api/auth/*` → auth-service, `/api/sensors/*` → sensors-service, etc.
- Configuration: `infra/nginx/gateway.conf`

**WebSocket Real-Time Updates:**
- Endpoints: `/api/sensors/ws`, `/api/lighting/ws`
- Implementation: Gorilla WebSocket (`github.com/gorilla/websocket v1.5.3`)
- Location: `services/sensors-service/internal/infrastructure/ws/`, `services/lighting-service/internal/infrastructure/ws/`
- Hub pattern: Broadcast to all connected clients
- Used for: Real-time sensor readings, light state updates

---

*Integration audit: 2026-03-25*
