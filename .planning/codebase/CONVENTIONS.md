# Coding Conventions

**Analysis Date:** 2026-03-25

## Package Organization

**Structure:**
- `internal/domain` - Domain entities, value objects, and interfaces
- `internal/application` - Use cases and service orchestration
- `internal/infrastructure` - Implementation of domain interfaces
- `cmd/api` - Entry points

**Visibility:**
- Use `internal/` to hide implementation details from external packages
- Domain interfaces are defined in `internal/domain` and implemented in `internal/infrastructure`

## Naming Patterns

**Files:**
- `snake_case.go` - All files follow snake_case naming
- `*_test.go` - Test files mirror the package and use `_test` suffix
- Domain files: `entity_name.go`, `entity_name_value_object.go`
- Test files: `entity_name_test.go` in same package as domain code
- Interface files: `interface_name_contract.go` or embedded in domain files

Example: `email.go`, `email_test.go`, `jwt_contract.go`, `auth_errors.go`

**Functions and Methods:**
- PascalCase for exported functions and methods
- Constructor pattern: `NewEntityName(...) *EntityName` or `NewEntityName(...) (EntityName, error)`
- Getter methods: Short names without `Get` prefix: `String()`, `Float64()`, `Topic()`
- Domain validation methods: Explicit names like `BelongsTo(userID)`, `IsMotionSensor()`, `IsOn()`

Examples:
- `NewEmail(raw string) (Email, error)`
- `NewAlarmEvent(triggerType, sensorID, location string) (*AlarmEvent, error)`
- `Email.String() string`
- `Sensor.BelongsTo(userID string) bool`

**Variables:**
- camelCase for local variables and parameters
- UPPER_CASE for package-level constants
- Avoid single-letter variables except for loop indices

Example:
```go
var ErrInvalidEmail = errors.New("invalid email format")
var TriggerMotion = AlarmTriggerType{value: "motion"}

func someFunc(userID string) {
    existingUser := repo.FindByID(userID)
}
```

**Types:**
- PascalCase for all type names
- Value objects: Noun-based names (e.g., `Email`, `LightPercentage`)
- Entities: Noun-based names (e.g., `Alarm`, `User`, `Sensor`)
- Error types: Prefixed with `Err` (e.g., `ErrAlarmNotFound`, `ErrInvalidEmail`)
- Type aliases for enums: PascalCase constants (e.g., `AlarmStatus`, `LightState`, `SensorType`)

## Value Objects

**Pattern:**
Value objects encapsulate validation and represent domain concepts. They are small, immutable structs with a private field storing the value.

Location: `internal/domain/value_object_name.go`

Example structure:
```go
// Domain file: internal/domain/email.go
package domain

type Email struct {
    value string  // private field
}

// NewEmail validates and constructs an Email. Returns error if invalid.
func NewEmail(raw string) (Email, error) {
    if raw == "" {
        return Email{}, ErrInvalidEmail
    }
    // validation logic
    return Email{value: raw}, nil
}

// String returns the email value
func (e Email) String() string { return e.value }

// MarshalJSON and UnmarshalJSON for JSON serialization
func (e Email) MarshalJSON() ([]byte, error) {
    return json.Marshal(e.value)
}
```

Test file: `internal/domain/email_test.go` in package `domain_test`

**Value Objects in Codebase:**
- `Email` - `internal/domain/email.go` - Validates email format
- `AlarmTriggerType` - `internal/domain/alarm_trigger_type.go` - Constrains trigger types to "motion" or "manual"
- `LightPercentage` - `internal/domain/light_percentage.go` - Validates percentage 0-100 with JSON marshaling
- `SensorType` - `internal/domain/sensor.go` - Type alias for sensor categories

## Entity Definitions

**Pattern:**
Entities represent domain objects with identity. Defined as structs with exported fields in the domain layer.

Location: `internal/domain/entity_name.go`

**Key characteristics:**
- Constructor pattern: `NewEntityName(...) *EntityName`
- ID field for identity
- Timestamp fields: `CreatedAt`, `UpdatedAt` using `time.Time`
- State methods for domain logic (e.g., `Silence()`, `TurnOn()`, `TurnOff()`)
- Comments explaining purpose in Portuguese

Example: `internal/domain/alarm.go`
```go
type AlarmEvent struct {
    ID          string
    TriggerType AlarmTriggerType
    SensorID    string
    Status      AlarmStatus
    TriggeredAt time.Time
}

func NewAlarmEvent(triggerType, sensorID, location string) (*AlarmEvent, error) {
    tt, err := NewAlarmTriggerType(triggerType)
    if err != nil {
        return nil, err
    }
    return &AlarmEvent{
        ID:          newID(),
        TriggerType: tt,
        // ... rest of initialization
    }, nil
}

func (a *AlarmEvent) Silence() {
    now := time.Now()
    a.Status = AlarmStatusSilenced
    a.SilencedAt = &now
}
```

## Interfaces

**Location:** Interfaces are defined in `internal/domain/` - either in the entity file or dedicated contract files.

**Naming:**
- Suffixed with the type: `Repository`, `Client`, `Manager`, `Publisher`, `Service`
- Examples: `AlarmRepository`, `BuzzerClient`, `JWTManager`, `EventPublisher`

**Pattern:**
```go
// BuzzerClient defines the contract for ESP32 buzzer communication
type BuzzerClient interface {
    TurnOn(deviceID string) error
    TurnOff(deviceID string) error
}

// AlarmRepository defines the contract for alarm event persistence
type AlarmRepository interface {
    Save(event *AlarmEvent) error
    FindByID(id string) (*AlarmEvent, error)
    FindRecent(limit int) ([]*AlarmEvent, error)
}
```

**Documentation:**
- Include a comment explaining the interface purpose (in English or Portuguese)
- Keep interfaces small and focused (3-5 methods typical)
- Interface location: Define in domain layer, implement in infrastructure layer

## Error Handling

**Domain Errors:**

Location: `internal/domain/errors.go` or `internal/domain/entity_errors.go`

Pattern:
```go
package domain

import "errors"

var (
    ErrAlarmNotFound     = errors.New("alarm event not found")
    ErrInvalidEmail      = errors.New("invalid email format")
    ErrInvalidPassword   = errors.New("password must be at least 6 characters")
    ErrUserAlreadyExists = errors.New("user already exists")
)
```

**Error-based Validation:**
- Constructors return `error` as second return value: `(Entity, error)` or `(*Entity, error)`
- Use `errors.Is()` or `errors.New()` from standard library
- Wrap domain errors in application layer if needed
- Never expose internal implementation errors upward

Examples:
```go
// Domain layer
func NewAlarmEvent(triggerType, sensorID, location string) (*AlarmEvent, error) {
    tt, err := NewAlarmTriggerType(triggerType)
    if err != nil {
        return nil, err  // propagate domain error
    }
    return &AlarmEvent{...}, nil
}

// Application layer
func (s *AlarmService) TriggerAlarm(triggerType, sensorID, location string) (*AlarmResponse, error) {
    event, err := domain.NewAlarmEvent(triggerType, sensorID, location)
    if err != nil {
        return nil, err  // domain validation error
    }
    // ... continue with business logic
}

// Test
func TestLogin_CredentialsInvalid(t *testing.T) {
    _, err := svc.Login(LoginRequest{Email: "wrong", Password: "wrong"})
    if !errors.Is(err, domain.ErrInvalidCredentials) {
        t.Errorf("expected ErrInvalidCredentials, got: %v", err)
    }
}
```

## Application Layer DTOs

**Pattern:**
Request and response types are defined in the service file (application layer).

Location: `internal/application/service_name.go`

Naming: `*Request` and `*Response` suffixes

Example:
```go
// RegisterRequest represents input data for registration
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

// Conversion helper (lowercase, unexported)
func toResponse(user *domain.User) *RegisterResponse {
    return &RegisterResponse{
        ID:    user.ID,
        Email: user.Email.String(),
    }
}
```

**JSON Tags:**
- Use `json:"field_name"` for all response structs
- Use snake_case for JSON field names: `sensor_id`, `trigger_type`, `created_at`

## Comments

**Documentation Comments:**
- All exported types, functions, and methods have documentation comments
- Comments start with the identifier name: `// Email represents...`
- Use present tense for comments

Examples:
```go
// AlarmEvent represents an alarm triggering event
type AlarmEvent struct { ... }

// NewAlarmEvent creates a new alarm event. Returns error if triggerType is invalid.
func NewAlarmEvent(triggerType, sensorID, location string) (*AlarmEvent, error) { ... }

// BuzzerClient defines the contract for ESP32 buzzer communication
type BuzzerClient interface { ... }
```

**Inline Comments:**
- Explain business logic, constraints, or non-obvious decisions
- Used sparingly in Go style

**Test Comments:**
- Include test-specific comments explaining why the test matters
- Explain setup, assertions, and edge cases

Example:
```go
// TestRegister_SenhaNaoEArmazenadaEmPlainText verifies that passwords are hashed before persisting.
//
// Why is this test especially important?
// This is a SECURITY rule. If someone accidentally removes bcrypt,
// logic tests would still pass — but plain text passwords would be saved.
```

## Import Organization

**Order:** Imports are organized in three groups:

1. Standard library imports
2. External package imports (third-party)
3. Local/relative imports (aurora/...)

Separated by blank lines:

```go
package domain

import (
    "errors"      // Standard library
    "time"        // Standard library

    "github.com/google/uuid"  // External
    "golang.org/x/crypto"     // External

    "aurora/pkg/helpers"      // Local
)
```

**Aliases:**
- Use module path: `aurora/services/auth-service/internal/domain`
- No aliases in typical usage
- Use full import paths in source code

## Logging

**Pattern:**
- Use standard library `log` package: `import "log"`
- Log at operation boundaries (function entry/exit for errors)
- Use descriptive prefixes like `[Service] message`

Examples from codebase:
```go
log.Printf("[Alarm] Acionando buzzer por %v (trigger: %s, sensor: %s)",
    s.buzzerDuration, triggerType, sensorID)

log.Printf("[Alarm] Falha ao ligar buzzer: %v", err)
```

**When to Log:**
- External service failures (device unreachable, database errors)
- Async operations (goroutine completions)
- NOT for normal business logic (validation failures, not-found conditions)

## Function Design

**Size Guidelines:**
- Keep functions focused on a single responsibility
- Average 10-30 lines
- Service methods typically: validation → persistence → side effects → response

**Parameters:**
- Service constructors use dependency injection: pass repository, client, etc.
- Service methods pass request data and user context (userID, deviceID)
- Avoid optional parameters - use explicit method overloads if needed

Example:
```go
func NewAlarmService(alarmRepo domain.AlarmRepository, buzzerClient domain.BuzzerClient,
    buzzerDurationMs int, deviceID string) *AlarmService {
    return &AlarmService{
        alarmRepo:      alarmRepo,
        buzzerClient:   buzzerClient,
        buzzerDuration: time.Duration(buzzerDurationMs) * time.Millisecond,
        deviceID:       deviceID,
    }
}
```

**Return Values:**
- Functions returning entities/DTOs use pointer receivers: `(*Entity, error)`
- Value objects return by value: `(ValueObject, error)`
- Error always last: `(Result, error)`
- Always check error immediately after function call

## Module Layering

**Domain Layer** (`internal/domain/`):
- Pure domain logic, no external dependencies
- Entities and value objects
- Interface definitions (contracts)
- Error definitions
- Domain-level validation

**Application Layer** (`internal/application/`):
- Use case orchestration
- Request/Response DTOs
- Calls to domain layer for validation
- Calls to repositories and external services
- Business logic coordination
- Error handling and mapping

**Infrastructure Layer** (`internal/infrastructure/`):
- Implementations of domain interfaces
- HTTP routes and handlers
- Database repository implementations
- External service clients (HTTP, MQTT, etc.)
- JWT and authentication implementations

---

*Convention analysis: 2026-03-25*
