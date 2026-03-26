# Codebase Structure

**Analysis Date:** 2026-03-25

## Directory Layout

```
Aurora/
├── services/                          # Six microservices, each independent
│   ├── auth-service/                  # User authentication, JWT token generation
│   ├── sensors-service/               # Motion/light sensor management, event publishing
│   ├── lighting-service/              # Smart light control via ESP32
│   ├── security-service/              # Alarm management and buzzer control
│   ├── rules-service/                 # Automation engine, event evaluation
│   └── notifications-service/         # Event persistence and notification retrieval
│
├── pkg/                               # Shared packages across all services
│   ├── database/                      # Shared PostgreSQL connection utilities
│   ├── jwt/                           # Shared JWT middleware and validation
│   ├── go.mod                         # Go module definition for shared code
│   └── go.sum                         # Dependencies lock file
│
├── infra/                             # Infrastructure and deployment configuration
│   ├── docker-compose.yml             # Local development environment (PostgreSQL, NATS, services)
│   ├── nginx/                         # Nginx reverse proxy configuration
│   └── .env                           # Environment variables for local dev
│
├── firmware/                          # ESP32 firmware (embedded device code)
│   └── aurora_esp32/                  # Arduino/PlatformIO project for ESP32
│
├── frontend/                          # Web frontend
│   └── node_modules/                  # NPM dependencies (excluded from this analysis)
│
├── docs/                              # Documentation and diagrams
├── .planning/                         # GSD planning documents (auto-generated)
└── README.md                          # Project overview
```

## Directory Purposes

**`/services/`:**
- Purpose: Six independent microservices, each with own database, entry point, and bounded context
- Contains: All application code for the system
- Key Pattern: Each service follows `cmd/api/main.go` as entry point, `internal/` for implementation

**`/services/{service-name}/`:**
- Purpose: Standard Go service structure for each microservice
- Structure:
  ```
  {service-name}/
  ├── cmd/api/main.go            # Service entry point
  ├── internal/
  │   ├── domain/                # DDD domain layer
  │   ├── application/           # Use case implementations
  │   └── infrastructure/        # HTTP, database, messaging, hardware
  └── go.mod, go.sum             # (if service has own module - check actual structure)
  ```

**`/services/{service-name}/cmd/api/main.go`:**
- Purpose: Application entry point, dependency wiring, environment configuration
- Responsibilities: Load env vars, create repositories, create services, start HTTP server and background tasks
- Pattern: All initialization happens here, no initialization in other files

**`/services/{service-name}/internal/domain/`:**
- Purpose: Pure business logic, independent of frameworks
- Contains:
  - Entity structs: `User`, `Sensor`, `AlarmEvent`, `Light`, `Rule`, `Notification`
  - Value objects: `Email`, `AlarmTriggerType`, `LightPercentage`, `SensorType`
  - Domain errors: `var ErrXXX = errors.New("...")`
  - Repository interfaces: Contracts for data access (e.g., `UserRepository`, `SensorRepository`)
  - Event/publisher interfaces: `EventPublisher`, `BuzzerClient`
- Files: `{entity}.go`, `{entity}_test.go`, `{contract}.go`, `errors.go`
- No imports from: `infrastructure/`, `application/`, external packages except `time`, `errors`, `uuid`

**`/services/{service-name}/internal/application/`:**
- Purpose: Use case orchestration, application logic layer
- Contains:
  - Service structs: `AuthService`, `MotionService`, `LightService`, `RulesEngine`
  - Request DTOs: `RegisterRequest`, `LoginRequest`, `CreateRuleRequest`
  - Response DTOs: `RegisterResponse`, `LoginResponse`, `RuleResponse`
  - Use case methods: `Register()`, `Login()`, `TriggerAlarm()`, etc.
- Files: `{service_name}.go`, `{service_name}_test.go`
- Depends on: Domain interfaces only
- Called by: HTTP handlers
- Never directly accesses database (goes through domain repositories)

**`/services/{service-name}/internal/infrastructure/http/`:**
- Purpose: HTTP server, routes, handlers, middleware
- Files:
  - `server.go`: Gin HTTP server setup, port binding
  - `routes.go`: Route registration and middleware attachment
  - `handlers.go`: HTTP request handlers, JSON binding, error mapping
  - `auth_middleware.go`: JWT/device API key validation
- Pattern: Handler receives request, calls application service, returns response
- No business logic (calls application layer)

**`/services/{service-name}/internal/infrastructure/repository/`:**
- Purpose: Database implementations of domain repository interfaces
- Files: `postgres_{entity}_repository.go`
- Examples:
  - `postgres_user_repository.go` → implements `domain.UserRepository`
  - `postgres_alarm_repository.go` → implements `domain.AlarmRepository`
  - `postgres_sensor_repository.go` → implements `domain.SensorRepository`
- Pattern: Constructor `NewPostgres{Entity}Repository(db *sql.DB)`, methods execute SQL

**`/services/{service-name}/internal/infrastructure/migrations/`:**
- Purpose: Database schema creation and seeding
- Files: `{entity}_migrations.go`
- Pattern:
  - `Run{Service}Migrations(db *sql.DB)` creates tables
  - `SeedDemo{Entity}(db *sql.DB)` inserts demo data
- Called from: `main.go` before service starts

**`/services/{service-name}/internal/infrastructure/security/`:**
- Purpose: JWT validation, auth middleware
- Files: `jwt_validator.go`, (sometimes `jwt_manager.go` for token generation)
- Pattern:
  - `JWTValidator`: Validates tokens in HTTP middleware
  - `JWTManager` (auth-service only): Generates tokens after login

**`/services/{service-name}/internal/infrastructure/messaging/`:**
- Purpose: NATS message broker integration
- Files: `nats_subscriber.go`, `nats_publisher.go` (optional)
- Pattern:
  - Subscriber: Listens to topics, calls application service methods
  - Publisher: Takes domain events, publishes to NATS topics
- Services using: `sensors-service`, `security-service`, `rules-service`, `notifications-service`

**`/services/{service-name}/internal/infrastructure/device/`:**
- Purpose: Hardware device clients (ESP32)
- Files: `buzzer_client.go`, `esp32_client.go`
- Pattern: HTTP wrapper around device REST API, implements domain interfaces
- Services using: `lighting-service`, `security-service`

**`/services/{service-name}/internal/infrastructure/ws/`:**
- Purpose: WebSocket server for real-time updates
- Files: `hub.go`, (sometimes `handlers.go`)
- Services using: `sensors-service`, `lighting-service`
- Pattern: Hub maintains connections, broadcasts events to all connected clients

**`/pkg/`:**
- Purpose: Shared code across all services (no duplication)
- Structure:
  - `database/postgres.go`: PostgreSQL connection factory with retry logic, pool config
  - `jwt/middleware.go`: HTTP middleware for JWT validation (shared middleware)
  - `jwt/validator.go`: JWTValidator helper for validating tokens
- Go module: `module aurora/pkg` (shared as dependency)

**`/pkg/database/postgres.go`:**
- Exports: `Config` struct, `NewPostgresConnection(cfg)` function
- Used by: Every service's `main.go` to create database connection
- Features: Connection pooling (25 max open, 5 idle), retry logic (10 attempts, 2s delay)

**`/pkg/jwt/`:**
- `validator.go`: Exports `JWTValidator` for validating incoming tokens
- `middleware.go`: HTTP middleware for protecting routes

**`/infra/`:**
- Purpose: Deployment and local development environment
- `docker-compose.yml`: Defines PostgreSQL, NATS, all 6 services for local dev
- `nginx/`: Reverse proxy configuration (if used for production)
- `.env`: Environment variables (not committed, local only)

**`/firmware/`:**
- Purpose: ESP32 embedded code
- `aurora_esp32/`: Arduino/PlatformIO project
- Not analyzed in detail (embedded domain)

## Key File Locations

**Entry Points:**
- `services/auth-service/cmd/api/main.go`: Auth service startup
- `services/sensors-service/cmd/api/main.go`: Sensors service startup
- `services/lighting-service/cmd/api/main.go`: Lighting service startup
- `services/security-service/cmd/api/main.go`: Security service startup
- `services/rules-service/cmd/api/main.go`: Rules service startup
- `services/notifications-service/cmd/api/main.go`: Notifications service startup

**Configuration:**
- `infra/docker-compose.yml`: Service dependencies and port mappings
- `pkg/database/postgres.go`: Database connection config
- Environment variables: Loaded in each service's `main.go` (JWT_SECRET, NATS_URL, DB_HOST, etc.)

**Core Logic by Service:**

Auth Service:
- `services/auth-service/internal/domain/user.go`: User entity
- `services/auth-service/internal/domain/email.go`: Email value object (validation)
- `services/auth-service/internal/application/auth_service.go`: Register/Login use cases
- `services/auth-service/internal/infrastructure/repository/postgres_user_repository.go`: User persistence

Sensors Service:
- `services/sensors-service/internal/domain/sensor.go`: Sensor entity
- `services/sensors-service/internal/domain/motion_event.go`: Motion event domain object
- `services/sensors-service/internal/application/motion_service.go`: RegisterMotion use case
- `services/sensors-service/internal/infrastructure/messaging/nats_publisher.go`: Publishes motion events
- `services/sensors-service/internal/infrastructure/ws/hub.go`: Real-time sensor updates

Lighting Service:
- `services/lighting-service/internal/domain/light.go`: Light entity
- `services/lighting-service/internal/application/light_service.go`: TurnOnLight/TurnOffLight use cases
- `services/lighting-service/internal/infrastructure/device/esp32_client.go`: Hardware control
- `services/lighting-service/internal/infrastructure/ws/hub.go`: Real-time light state broadcasts

Security Service:
- `services/security-service/internal/domain/alarm.go`: AlarmEvent entity
- `services/security-service/internal/domain/alarm_trigger_type.go`: Value object for validation
- `services/security-service/internal/application/alarm_service.go`: TriggerAlarm/SilenceAlarm use cases
- `services/security-service/internal/infrastructure/device/buzzer_client.go`: Hardware control
- `services/security-service/internal/infrastructure/messaging/nats_subscriber.go`: Listens for motion events

Rules Service:
- `services/rules-service/internal/domain/rule.go`: Rule entity
- `services/rules-service/internal/application/rules_engine.go`: EvaluateRule/ExecuteAction use cases
- `services/rules-service/internal/infrastructure/messaging/nats_subscriber.go`: Event evaluation loop

Notifications Service:
- `services/notifications-service/internal/domain/notification.go`: Notification entity
- `services/notifications-service/internal/application/notification_service.go`: Persistence use cases
- `services/notifications-service/internal/infrastructure/messaging/nats_subscriber.go`: Listens to all events

**Testing:**
- `services/{service}/internal/{layer}/{entity}_test.go`: Unit tests for entities, services
- Tests co-located with implementation files
- Pattern: `TestXXX(t *testing.T)` functions in `_test.go` files

**Database Migrations:**
- `services/{service}/internal/infrastructure/migrations/{entity}_migrations.go`
- Example: `services/sensors-service/internal/infrastructure/migrations/sensor_migrations.go`

## Naming Conventions

**Files:**
- Entry point: `main.go` (under `cmd/api/`)
- Entities: `{entity_name}.go` (e.g., `user.go`, `sensor.go`, `alarm.go`)
- Value objects: `{value_object_name}.go` (e.g., `email.go`, `alarm_trigger_type.go`)
- Repositories: `{connection_type}_{entity_name}_repository.go` (e.g., `postgres_user_repository.go`)
- Services: `{domain_area}_service.go` (e.g., `auth_service.go`, `motion_service.go`)
- Engine: `rules_engine.go` (special naming for complex orchestrator)
- Tests: `{file_being_tested}_test.go` (e.g., `user_test.go`, `auth_service_test.go`)
- Infrastructure: `{layer}_{technology}.go` (e.g., `nats_subscriber.go`, `esp32_client.go`, `jwt_validator.go`)

**Directories:**
- Lowercase with hyphens: `auth-service`, `sensors-service`, `postgres_repository` (packages use underscore)
- Layer folders: `domain/`, `application/`, `infrastructure/`
- Sub-layers: `http/`, `repository/`, `migrations/`, `messaging/`, `device/`, `ws/`, `security/`

**Functions:**
- Constructor: `New{Type}(params) *Type` (e.g., `NewUser(email, hash)`, `NewAuthService(repo, jwtMgr)`)
- Handler: `(h *Handlers) {Action}(c *gin.Context)` (e.g., `Register`, `TriggerAlarm`)
- Service method: `(s *Service) {Action}(params) (response, error)` (e.g., `Register()`, `TriggerAlarm()`)
- Repository method: `(r *PostgresXRepository) {CRUD}(params) error` (e.g., `Save()`, `FindByID()`)

**Types:**
- Structs: PascalCase (e.g., `User`, `AlarmEvent`, `MotionService`)
- Interfaces: PascalCase ending in -er, -or (e.g., `UserRepository`, `EventPublisher`)
- Constants: SCREAMING_SNAKE_CASE (e.g., `DefaultLightThreshold`, `AlarmStatusTriggered`)
- Enum-like types: PascalCase (e.g., `SensorType`, `TriggerType`) with `const` blocks

## Where to Add New Code

**New Service:**
1. Create `services/{service-name}/` directory
2. Add `cmd/api/main.go` following the pattern of existing services
3. Create `internal/{domain,application,infrastructure}` directories
4. Domain layer: Define entities, value objects, interfaces, errors
5. Application layer: Implement services with use cases
6. Infrastructure: Implement repositories, HTTP handlers, migrations
7. Register routes in `infrastructure/http/routes.go`
8. Add service to `docker-compose.yml`

**New Use Case in Existing Service:**
1. Add domain entities/value objects if needed: `internal/domain/{entity}.go`
2. Add application service method: Add to existing service in `internal/application/{service}.go`
3. Add HTTP handler: Add to `internal/infrastructure/http/handlers.go`
4. Register route: Add to `internal/infrastructure/http/routes.go`
5. Add tests: `internal/{layer}/{feature}_test.go`

**New Domain Entity:**
1. Create `internal/domain/{entity}.go` with struct and factory method
2. Define value objects if applicable
3. Add repository interface if persistence needed
4. Implement repository: `internal/infrastructure/repository/postgres_{entity}_repository.go`
5. Add migration: `internal/infrastructure/migrations/{entity}_migrations.go`

**New External Integration:**
1. Define domain interface if dependency: `internal/domain/{integration_name}.go`
2. Create infrastructure implementation: `internal/infrastructure/{layer}/{integration_name}.go`
3. Inject into service via constructor
4. Update `cmd/api/main.go` to instantiate and wire

**Shared Code (pkg/):**
1. Add to `pkg/` if used by 2+ services
2. Update `pkg/go.mod` with dependencies
3. Import with full path: `import "aurora/pkg/database"`

## Special Directories

**`internal/`:**
- Purpose: Enforced Go package visibility - code here is not importable outside the service
- Generated: No
- Committed: Yes
- Pattern: All service code except `cmd/` lives under `internal/`

**`cmd/`:**
- Purpose: Executable entry points
- Generated: No
- Committed: Yes
- Contains: Only `api/main.go` per service

**`.planning/codebase/`:**
- Purpose: Auto-generated GSD planning documents
- Generated: Yes (by GSD commands)
- Committed: Yes (to share context across team)
- Contents: ARCHITECTURE.md, STRUCTURE.md, CONVENTIONS.md, TESTING.md, STACK.md, INTEGRATIONS.md, CONCERNS.md

**`infra/`:**
- Purpose: Deployment configuration and local environment
- Generated: No
- Committed: Yes (except `.env` which is in `.gitignore`)
- Contains: Docker compose, nginx config, environment templates

---

*Structure analysis: 2026-03-25*
