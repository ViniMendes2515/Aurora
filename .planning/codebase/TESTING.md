# Testing Patterns

**Analysis Date:** 2026-03-25

## Test Framework

**Runner:**
- Go standard testing library (`testing` package)
- Built-in `go test` command
- No external test framework (testify, ginkgo, etc.)

**Assertion Library:**
- Manual assertions using `if` statements and `t.Error()`, `t.Fatalf()`, `t.Errorf()`
- No assertion library - idiomatic Go style

**Run Commands:**

```bash
# Run all tests in a package
go test ./internal/domain

# Run all tests in a service
go test ./...

# Run tests with verbose output
go test -v ./internal/domain

# Run specific test
go test -run TestNewEmail ./internal/domain

# Run tests with coverage
go test -cover ./internal/domain
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Watch mode (requires external tool like CompileDaemon or Watchman)
# Not currently configured in project
```

## Test File Organization

**Location:**
- Co-located with source files: `entity.go` and `entity_test.go` in same directory
- Test files in same package with `_test` suffix

**Naming:**
- Test files: `entity_name_test.go`
- Test functions: `TestFunctionName_Scenario(t *testing.T)`
- Test cases use underscore to separate test name from scenario

**Structure:**

```
services/auth-service/
├── internal/
│   ├── domain/
│   │   ├── user.go
│   │   ├── user_test.go          # Tests for User entity
│   │   ├── email.go
│   │   ├── email_test.go         # Tests for Email value object
│   │   ├── auth_errors.go
│   │   └── auth_repository.go
│   ├── application/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go  # Tests for AuthService
│   └── infrastructure/
│       └── repository/
│           └── postgres_user_repository.go  # No tests (integration layer)
```

**Test Package:**
- Use `package domain_test` for domain layer tests (separate test package)
- Use `package application_test` for application layer tests
- Allows testing public interface without exposing private implementation

Example:
```go
// File: internal/domain/email_test.go
package domain_test  // Note: _test suffix makes this a separate test package

import (
    "testing"
    "aurora/services/auth-service/internal/domain"
)

func TestNewEmail(t *testing.T) {
    e, err := domain.NewEmail("user@example.com")
    // assertions...
}
```

## Test Structure

**Suite Organization - Table-Driven Tests:**

Used for testing multiple scenarios with shared setup:

```go
func TestNewEmail(t *testing.T) {
    testes := []struct {  // "testes" = plural "tests" in Portuguese
        nome    string
        email   string
        comErro bool  // "comErro" = "with error"
    }{
        {"email valido", "user@example.com", false},
        {"email vazio", "", true},
        {"sem arroba", "userexample.com", true},
    }

    for _, tt := range testes {
        t.Run(tt.nome, func(t *testing.T) {
            e, err := domain.NewEmail(tt.email)
            if (err != nil) != tt.comErro {
                t.Errorf("NewEmail() erro = %v, esperado erro = %v", err, tt.comErro)
            }
            // more assertions...
        })
    }
}
```

**Individual Test Pattern:**

For single scenario tests or complex setups:

```go
func TestNewAlarmEvent_CamposPreenchidos(t *testing.T) {
    // Setup
    a, err := domain.NewAlarmEvent("motion", "sensor1", "sala")

    // Assertions
    if err != nil {
        t.Fatalf("NewAlarmEvent() erro inesperado: %v", err)
    }
    if a.ID == "" {
        t.Error("ID nao deve ser vazio")
    }
    if a.TriggerType.String() != "motion" {
        t.Errorf("TriggerType esperado motion, obteve %s", a.TriggerType.String())
    }
}
```

**Patterns:**

1. **Setup (Arrange):**
   - Create domain objects or mocks
   - Configure test data
   - Initialize service with mocks

2. **Execution (Act):**
   - Call the function being tested
   - Pass test parameters

3. **Assertion (Assert):**
   - Check return values
   - Verify error conditions using `errors.Is()` for domain errors
   - Verify side effects (mock calls, state changes)

## Mocking

**Framework:**
- Manual mock implementations using empty structs
- No mock library (testify/mock, golang/mock, etc.)
- Implements domain interfaces explicitly

**Why Manual Mocks?**

From codebase comments:
> "Go idiomático prefere mocks simples e explícitos. Uma struct que implementa a interface é suficiente — e mais fácil de ler e entender."
> (Idiomatic Go prefers simple, explicit mocks. A struct implementing the interface is enough — and easier to read and understand.)

**Pattern:**

```go
// Mock repository in memory
type mockAlarmRepo struct {
    saved        []*domain.AlarmEvent  // captured data
    findEvent    *domain.AlarmEvent    // returned on FindByID
    findErr      error                 // error to return
    recentEvents []*domain.AlarmEvent  // data for FindRecent
}

func (m *mockAlarmRepo) Save(event *domain.AlarmEvent) error {
    m.saved = append(m.saved, event)
    return nil
}

func (m *mockAlarmRepo) FindByID(id string) (*domain.AlarmEvent, error) {
    return m.findEvent, m.findErr
}

func (m *mockAlarmRepo) FindRecent(limit int) ([]*domain.AlarmEvent, error) {
    return m.recentEvents, nil
}
```

**Mock for Async Operations:**

When testing goroutines, use `atomic` for thread-safe flag checking:

```go
type mockBuzzerClient struct {
    turnOnCalled  atomic.Bool
    turnOffCalled atomic.Bool
}

func (m *mockBuzzerClient) TurnOn(deviceID string) error {
    m.turnOnCalled.Store(true)
    return nil
}

// In test:
time.Sleep(50 * time.Millisecond)  // wait for goroutine
if !buzzer.turnOnCalled.Load() {
    t.Error("esperava que TurnOn tivesse sido chamado")
}
```

**Mock with Callback:**

For intercepting specific method calls:

```go
type capturaUserRepo struct {
    mockUserRepo
    onSave func(*domain.User)  // callback field
}

func (c *capturaUserRepo) Save(user *domain.User) error {
    if c.onSave != nil {
        c.onSave(user)  // invoke callback
    }
    return nil
}

// In test:
var usuarioSalvo *domain.User
repoCaptura := &capturaUserRepo{
    onSave: func(u *domain.User) { usuarioSalvo = u },
}
```

**What to Mock:**
- Repositories (database access)
- External service clients (HTTP, MQTT)
- Device clients
- Publishers/message brokers
- JWT managers

**What NOT to Mock:**
- Domain entities and value objects
- Domain validators
- bcrypt (crypto operations) - use real crypto in tests for security checks

## Fixtures and Factories

**Test Data Pattern:**

Named helper functions create test instances:

```go
// Helper to create test light
func novaLuzDoUsuario() *domain.Light {
    return domain.NewLight("luz-1", "Sala", "Sala de Estar", "user-1", "device-1")
}

// In test:
func TestTurnOn_Sucesso(t *testing.T) {
    repo := newMockRepo()
    repo.lights["luz-1"] = novaLuzDoUsuario()  // use helper
    svc := application.NewLightService(repo, &mockDeviceClient{}, &mockStatePublisher{})
    // ... test
}
```

**Location:**
- Helper functions defined in same test file
- Lowercase names: `novaLuzDoUsuario()`, `newMockRepo()`
- Grouped with mocks at top of test file

## Coverage

**Requirements:**
- No explicit coverage enforced in CI/CD
- Coverage measurement available via `go test -cover`

**View Coverage:**

```bash
# Terminal output
go test -cover ./internal/domain

# HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Coverage per Service (18 test files found):**

| Service | Domain Tests | Application Tests | Infrastructure |
|---------|--------------|-------------------|-----------------|
| security-service | alarm_test.go, alarm_trigger_type_test.go | alarm_service_test.go | (none) |
| auth-service | email_test.go, user_test.go | auth_service_test.go | (none) |
| lighting-service | light_test.go | light_service_test.go | (none) |
| sensors-service | light_percentage_test.go, sensor_test.go, events_test.go | motion_service_test.go, light_service_test.go | (none) |
| notifications-service | notification_test.go | notification_service_test.go | (none) |
| rules-service | (none found) | rules_engine_test.go | (none) |

**Gap Analysis:**
- Infrastructure layer (repository, HTTP handlers) has no tests
- Integration tests not found
- E2E tests not found
- Only domain and application layers tested

## Test Types

**Unit Tests - Domain Layer:**

Test domain entities, value objects, and business rules in isolation.

Location: `internal/domain/*_test.go`

Examples:
- `TestNewEmail()` - Tests Email value object validation
- `TestNewAlarmEvent_CamposPreenchidos()` - Tests AlarmEvent constructor initialization
- `TestAlarmEvent_Silence()` - Tests domain state transition method
- `TestSensor_BelongsTo()` - Tests domain authorization logic

Characteristics:
- No external dependencies
- No mocks (except when testing against domain interfaces)
- Fast execution
- Test validation rules and state changes

**Unit Tests - Application Layer:**

Test use case orchestration, request handling, and coordination between domain and infrastructure.

Location: `internal/application/*_test.go`

Examples:
- `TestTriggerAlarm_TipoInvalido()` - Tests validation before persistence
- `TestTriggerAlarm_Sucesso()` - Tests complete happy path
- `TestRegister_SenhaNaoEArmazenadaEmPlainText()` - Tests security requirements
- `TestTurnOn_DispositivoInacessivel()` - Tests error handling from external service

Characteristics:
- Uses mocks for repositories and external services
- Tests error paths and edge cases
- Verifies side effects (database calls, broadcasts)
- Tests security rules and access control

**Integration Tests:**
- NOT found in codebase
- Would test: database operations, HTTP endpoints, message publishing

**E2E Tests:**
- NOT found in codebase
- Would require: full service startup, test database, orchestration

## Common Patterns

**Error Testing:**

```go
// Test expected error
func TestLogin_CredentialsInvalid(t *testing.T) {
    _, err := svc.Login(LoginRequest{Email: "wrong", Password: "wrong"})
    if !errors.Is(err, domain.ErrInvalidCredentials) {
        t.Errorf("esperado ErrInvalidCredentials, obteve: %v", err)
    }
}

// Test error propagation
func TestRegister_SenhaInvalida(t *testing.T) {
    _, err := svc.Register(RegisterRequest{
        Email:    "usuario@teste.com",
        Password: "123",  // too short
    })
    if !errors.Is(err, domain.ErrInvalidPassword) {
        t.Errorf("esperado ErrInvalidPassword, obteve: %v", err)
    }
}
```

**Async Testing - Goroutine Verification:**

```go
// Test async buzzer control in background
func TestTriggerAlarm_BuzzerEAcionadoEmBackground(t *testing.T) {
    repo := &mockAlarmRepo{}
    buzzer := &mockBuzzerClient{}
    svc := application.NewAlarmService(repo, buzzer, 10, "device-1")

    _, err := svc.TriggerAlarm("manual", "sensor-1", "entrada")
    if err != nil {
        t.Fatalf("não esperava erro: %v", err)
    }

    // Wait for goroutine to complete
    time.Sleep(50 * time.Millisecond)

    // Verify background operations completed
    if !buzzer.turnOnCalled.Load() {
        t.Error("esperava que TurnOn tivesse sido chamado")
    }
    if !buzzer.turnOffCalled.Load() {
        t.Error("esperava que TurnOff tivesse sido chamado")
    }
}
```

**State Verification - Testing Side Effects:**

```go
// Verify persistence
func TestTriggerAlarm_Sucesso(t *testing.T) {
    repo := &mockAlarmRepo{}
    buzzer := &mockBuzzerClient{}
    svc := application.NewAlarmService(repo, buzzer, 10000, "device-1")

    resp, err := svc.TriggerAlarm("motion", "sensor-42", "corredor")

    if err != nil {
        t.Fatalf("não esperava erro: %v", err)
    }

    // Verify event was persisted
    if len(repo.saved) != 1 {
        t.Errorf("esperava 1 evento salvo, mas foram %d", len(repo.saved))
    }

    // Verify response contains correct data
    if resp.ID == "" {
        t.Error("ID do evento não deve ser vazio")
    }
}
```

**Security Testing:**

```go
// Test password hashing (not plain text)
func TestRegister_SenhaNaoEArmazenadaEmPlainText(t *testing.T) {
    var usuarioSalvo *domain.User
    repoCaptura := &capturaUserRepo{
        onSave: func(u *domain.User) { usuarioSalvo = u },
    }
    svc := application.NewAuthService(repoCaptura, &mockJWTManager{})
    senhaOriginal := "minha-senha"

    _, err := svc.Register(RegisterRequest{
        Email:    "seguro@teste.com",
        Password: senhaOriginal,
    })
    if err != nil {
        t.Fatalf("Register() erro inesperado: %v", err)
    }

    // Verify password is NOT stored as plain text
    if usuarioSalvo.PasswordHash == senhaOriginal {
        t.Error("senha armazenada em plain text — isso é uma falha de segurança!")
    }

    // Verify it IS a valid bcrypt hash
    err = bcrypt.CompareHashAndPassword(
        []byte(usuarioSalvo.PasswordHash),
        []byte(senhaOriginal),
    )
    if err != nil {
        t.Errorf("PasswordHash não é um hash bcrypt válido: %v", err)
    }
}
```

**Access Control Testing:**

```go
// Test authorization boundaries
func TestTurnOn_AcessoNegado(t *testing.T) {
    repo := newMockRepo()
    luz := novaLuzDoUsuario()  // owner = "user-1"
    repo.lights["luz-1"] = luz

    svc := application.NewLightService(repo, &mockDeviceClient{}, &mockStatePublisher{})

    _, err := svc.TurnOn("luz-1", "user-outro")  // different user tries to control
    if !errors.Is(err, domain.ErrLightAccessDenied) {
        t.Errorf("esperado ErrLightAccessDenied, obtido: %v", err)
    }
}

// Test wildcard access for shared resources
func TestMarkAsRead_BroadcastPodeSerMarcadoPorQualquer(t *testing.T) {
    notif := &domain.Notification{ID: "n2", UserID: "*"}  // broadcast
    repo := &mockNotificationRepo{findNotification: notif}
    svc := application.NewNotificationService(repo)

    err := svc.MarkAsRead("n2", "qualquer-usuario")
    if err != nil {
        t.Fatalf("broadcast deve ser marcável por qualquer usuário: %v", err)
    }
    if !repo.markAsReadCalled {
        t.Error("MarkAsRead deveria ter sido chamado")
    }
}
```

**Comments on Test Importance:**

Tests include comments explaining WHY the test matters:

```go
// TestTriggerAlarm_TipoInvalido verifica que a camada de aplicação rejeita
// gatilhos desconhecidos antes mesmo de tentar persistir qualquer coisa.
// Isso é importante para garantir que a validação do domínio é respeitada
// e que eventos inválidos nunca chegam ao repositório.
func TestTriggerAlarm_TipoInvalido(t *testing.T) {
    // ... test code
}

// Por que este teste é especialmente importante?
// Esta é uma regra de SEGURANÇA. Se alguém remover o bcrypt por acidente,
// os testes de lógica continuariam passando — mas senhas em plain text
// estariam sendo salvas.
func TestRegister_SenhaNaoEArmazenadaEmPlainText(t *testing.T) {
    // ... test code
}
```

---

*Testing analysis: 2026-03-25*
