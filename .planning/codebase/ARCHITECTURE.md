# Architecture

**Analysis Date:** 2026-03-25

## Pattern Overview

**Overall:** Microservices with Domain-Driven Design (DDD) and Event-Driven Architecture

**Key Characteristics:**
- Each service implements clean layered architecture (Domain → Application → Infrastructure)
- Inter-service communication via NATS message broker (event-driven)
- Service-to-service HTTP calls for synchronous operations (rules-service to lighting/security)
- Shared database per service (no cross-service DB access)
- JWT-based authentication with shared secret across all services

## Layers

**Domain Layer:**
- Purpose: Core business logic, entities, value objects, domain contracts/interfaces
- Location: `services/{service-name}/internal/domain/`
- Contains: Entity structs (e.g., `User`, `AlarmEvent`, `Sensor`, `Rule`), value objects (e.g., `Email`, `AlarmTriggerType`, `LightPercentage`), domain errors, interface contracts (`Repository`, `Publisher`, `BuzzerClient`)
- Depends on: Only Go stdlib (no external dependencies)
- Used by: Application layer exclusively

**Application Layer:**
- Purpose: Use case implementations, orchestration of domain logic, request/response handling
- Location: `services/{service-name}/internal/application/`
- Contains: Service classes (e.g., `AuthService`, `AlarmService`, `MotionService`), request/response DTOs, use case logic
- Depends on: Domain layer contracts, external clients for cross-service calls
- Used by: Infrastructure/HTTP handlers

**Infrastructure Layer:**
- Purpose: Technical implementations, external integrations, database access, HTTP servers
- Location: `services/{service-name}/internal/infrastructure/`
- Contains:
  - `http/`: HTTP server, routes, handlers, middleware
  - `repository/`: Database implementations (PostgreSQL)
  - `migrations/`: Database schema and seeding
  - `security/`: JWT validation, auth middleware
  - `messaging/`: NATS publisher/subscriber implementations
  - `device/`: Hardware clients (ESP32, buzzer)
  - `ws/`: WebSocket hubs for real-time updates
- Depends on: Domain, Application, external packages (Gin, NATS, PostgreSQL driver)
- Used by: Entry point only

**Entry Point:**
- Location: `services/{service-name}/cmd/api/main.go`
- Responsibilities: Load environment, instantiate all dependencies, wire layers together, start HTTP server and background listeners

## Bounded Contexts (Services)

**Auth Service:**
- Port: 8080
- Purpose: User registration, login, JWT token generation
- Domain: User entities, Email value object, JWT contracts
- Key Use Cases: Register, Login
- External Dependencies: PostgreSQL
- Messaging: None (request-response only)

**Sensors Service:**
- Port: 8081
- Purpose: Sensor management, motion detection recording, light level tracking
- Domain: Sensor entities, MotionDetectedEvent, LightChangeEvent, LightPercentage value object
- Key Use Cases: RegisterMotion, GetLightLevel, ListSensors
- External Dependencies: PostgreSQL, NATS (publish motion/light events), WebSocket
- Messaging: Publishes `sensors.motion.detected`, `sensors.light.level_changed` events

**Lighting Service:**
- Port: 8082
- Purpose: Smart light control and state management
- Domain: Light entities, LightState value objects
- Key Use Cases: TurnOnLight, TurnOffLight, SetBrightness
- External Dependencies: PostgreSQL, ESP32 device API, WebSocket
- Messaging: Subscribes to motion/light events for automation
- Hardware: ESP32-based LED controller via HTTP API

**Security Service:**
- Port: 8083
- Purpose: Alarm management and triggering
- Domain: AlarmEvent entities, AlarmTriggerType value object, BuzzerClient interface
- Key Use Cases: TriggerAlarm, SilenceAlarm, GetRecentAlarms
- External Dependencies: PostgreSQL, ESP32 buzzer device, NATS
- Messaging: Subscribes to motion events (auto-triggers alarm if enabled)
- Hardware: ESP32 buzzer device

**Rules Service:**
- Port: 8084
- Purpose: Automation engine, evaluates triggers and executes actions
- Domain: Rule entities, TriggerType, ActionType enums
- Key Use Cases: CreateRule, EvaluateRule, ExecuteAction
- External Dependencies: PostgreSQL, NATS, HTTP to lighting-service and security-service
- Messaging: Subscribes to motion and light events, publishes rule execution events
- Pattern: Implements motion-based auto-light-off with timeout tracking

**Notifications Service:**
- Port: 8085
- Purpose: Alarm and event notifications persistence and retrieval
- Domain: Notification entities
- Key Use Cases: StoreNotification, GetNotifications
- External Dependencies: PostgreSQL, NATS
- Messaging: Subscribes to all domain events (motion, light, alarm)

## Data Flow

**User Registration & Login Flow:**

1. POST `/auth/register` → Auth Service handler
2. AuthService validates email/password via domain rules
3. Saves User to PostgreSQL via UserRepository
4. Returns user ID and email

**Motion Detection to Alarm Flow:**

1. Device detects motion via ESP32 physical sensor
2. POST `/motion` to Sensors Service (from ESP32 or user)
3. MotionService validates sensor ownership
4. Creates MotionDetectedEvent domain object
5. Saves MotionRecord to PostgreSQL
6. Publishes `sensors.motion.detected` to NATS
7. Rules Service receives event, matches motion trigger rules
8. Rules Service calls Security Service HTTP API: POST `/alarms/trigger`
9. AlarmService triggers alarm (if AutoAlarmOnMotion enabled)
10. Buzzer client turns on ESP32 buzzer
11. Notifications Service receives event, persists notification

**Motion Detected to Light Auto-On Flow:**

1. Motion event published to NATS
2. Rules Service evaluates motion-triggered rules (TriggerMotion → ActionTurnOnLight)
3. Rules Service calls Lighting Service HTTP API: POST `/lights/{id}/on`
4. LightService turns on light via ESP32 client
5. Publishes light state update to WebSocket hub
6. Rules Service sets timeout, will auto-turn-off after motionOffTimeout

**State Management:**

- **Rules Service**: In-memory map of `lastMotionByLight` tracks when motion-triggered lights were turned on (for auto-off)
- **Sensors Service**: WebSocket hub broadcasts real-time sensor updates to frontend
- **Lighting Service**: WebSocket hub broadcasts light state changes
- **Repositories**: Persistent state in PostgreSQL per service

## Key Abstractions

**Repository Pattern:**
- Purpose: Abstract database access, allow domain layer to work with interfaces
- Examples: `domain.UserRepository`, `domain.SensorRepository`, `domain.AlarmRepository`, `domain.RuleRepository`
- Pattern: Interface defined in domain, PostgreSQL implementation in infrastructure

**Event Publisher:**
- Purpose: Abstract message publication, decouple event sources from consumers
- Examples: `domain.EventPublisher` (NATS implementation in sensors-service)
- Pattern: Interface in domain, NATS wrapper in infrastructure

**Device Clients:**
- Purpose: Abstract hardware communication, allow testing without physical devices
- Examples: `domain.BuzzerClient`, `ESP32Client` (interfaces defined in domain)
- Pattern: HTTP clients wrapping ESP32 REST API

**Service Layers:**
- Purpose: Orchestrate domain logic without exposing domain to HTTP layer
- Examples: `AuthService`, `MotionService`, `LightService`, `AlarmService`
- Relationship: Handlers inject service into themselves, call service methods

## Entry Points

**Auth Service: `/cmd/api/main.go`**
- Loads JWT secret, server port, database config from environment
- Creates UserRepository (PostgreSQL)
- Creates JWTManager (shared from pkg/security)
- Creates AuthService
- Creates HTTP Server with Gin
- Routes: POST `/auth/register`, POST `/auth/login`, GET `/health`

**Sensors Service: `/cmd/api/main.go`**
- Loads JWT secret, NATS URL, device API key from environment
- Runs database migrations and seeds demo sensors
- Creates SensorRepository (PostgreSQL)
- Creates WebSocket Hub
- Creates NATS Publisher (event publisher implementation)
- Creates MotionService and LightService
- Starts HTTP server with routes for motion/light registration
- Starts WebSocket server for real-time updates

**Lighting Service: `/cmd/api/main.go`**
- Loads ESP32 IP, device API key, JWT secret from environment
- Runs database migrations and seeds demo lights
- Creates LightRepository (PostgreSQL)
- Creates ESP32Client (hardware abstraction)
- Creates WebSocket Hub
- Creates LightService
- Starts HTTP server with light control routes

**Security Service: `/cmd/api/main.go`**
- Loads ESP32 IP, NATS URL, buzzer duration from environment
- Creates AlarmRepository (PostgreSQL)
- Creates BuzzerClient (ESP32 hardware abstraction)
- Creates AlarmService
- Starts NATS subscriber to listen for motion events
- Starts HTTP server with alarm routes

**Rules Service: `/cmd/api/main.go`**
- Loads NATS URL, lighting/security service URLs, motion timeout from environment
- Creates RuleRepository (PostgreSQL)
- Creates RulesEngine with HTTP client references to other services
- Starts NATS subscriber to listen for motion/light events
- Starts HTTP server with rule management routes
- RulesEngine maintains in-memory state for motion timeouts

**Notifications Service: `/cmd/api/main.go`**
- Loads NATS URL from environment
- Creates NotificationRepository (PostgreSQL)
- Creates NotificationService
- Starts NATS subscriber to listen for all domain events
- Starts HTTP server with notification retrieval routes

## Error Handling

**Strategy:** Layered error handling with domain-defined errors

**Patterns:**
- Domain layer defines error variables (e.g., `ErrAlarmNotFound`, `ErrInvalidEmail`, `ErrSensorAccessDenied`)
- Application layer propagates domain errors or wraps infrastructure errors
- HTTP handlers catch application errors and map to HTTP status codes
- Database errors logged but handled gracefully (return domain error to caller)

**HTTP Status Mapping:**
- 400 Bad Request: Validation errors (invalid email, invalid input)
- 401 Unauthorized: Missing/invalid JWT token
- 403 Forbidden: Access denied (e.g., sensor doesn't belong to user)
- 404 Not Found: Entity not found
- 409 Conflict: Duplicate user email
- 500 Internal Server Error: Unexpected database/service errors

## Cross-Cutting Concerns

**Logging:**
- Uses Go standard library `log` package
- Entry points log service startup, migrations
- Application layer logs significant operations (alarm trigger, rule execution)
- Infrastructure layer logs database/NATS connections

**Validation:**
- Domain: Email validation (Email value object), AlarmTriggerType validation, rule field validation
- Application: Request validation (email/password presence, sensor ID presence)
- HTTP: JSON binding validation via Gin (required field checks)

**Authentication:**
- JWT tokens generated by Auth Service
- JWTValidator used by all services to validate incoming Authorization header
- Device API Key (X-Device-Key header) used for device-to-service calls (ESP32)
- Both JWT and API Key validation in HTTP middleware

**WebSocket Hubs:**
- Sensors Service and Lighting Service maintain Hub instances
- Hub broadcasts state changes to connected frontend clients in real-time
- No persistence of WebSocket connections (clients reconnect on close)

---

*Architecture analysis: 2026-03-25*
