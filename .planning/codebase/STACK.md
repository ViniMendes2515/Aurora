# Technology Stack

**Analysis Date:** 2026-03-25

## Languages

**Primary:**
- Go 1.21 - All microservices backend implementation (`/services/**/cmd/api/main.go`, `/pkg/**/*.go`)
- TypeScript 5.4 - Angular 18 frontend (`/frontend/tsconfig.json`)
- JavaScript - Build and configuration tooling (`/frontend/postcss.config.js`, `/frontend/tailwind.config.js`)

**Secondary:**
- Embedded C/C++ - ESP32 firmware for IoT device control (`/firmware/`)
- SQL - PostgreSQL migrations and queries (`/services/**/internal/infrastructure/migrations/`)

## Runtime

**Environment:**
- Go 1.21 runtime for all backend services
- Node.js 20-alpine - Used in frontend Docker build stage (`/frontend/Dockerfile`)
- Docker container runtime for all services

**Package Manager:**
- Go modules - Primary Go dependency management via `go.mod`
- npm/npm ci - Frontend dependency management (`/frontend/package.json`)

## Frameworks

**Core Backend:**
- Gin v1.9.1 - HTTP web framework used in all microservices (`/services/*/go.mod`)
  - Provides REST API routing and middleware
  - Used across: auth-service, sensors-service, lighting-service, security-service, rules-service, notifications-service

**Frontend:**
- Angular 18.0.0 - Frontend SPA framework
  - Package: `@angular/core`, `@angular/common`, `@angular/router`, `@angular/forms`, `@angular/animations`
  - Location: `/frontend/package.json`

**Testing:**
- Angular Testing - Built with `@angular/core` test utilities (`/frontend/package.json` includes testing scripts)
- No explicit test framework detected in backend (no jest.config, vitest.config, etc.)

**Build/Dev:**
- @angular/cli 18.0.0 - Angular development CLI (`/frontend/package.json`)
- @angular-devkit/build-angular 18.0.0 - Angular build tooling
- TypeScript 5.4.0 - TypeScript compiler and language support
- Tailwind CSS 3.4.3 - Utility-first CSS framework (`/frontend/tailwind.config.js`)
- PostCSS 8.4.38 - CSS transformation tool with Autoprefixer 10.4.19
- Docker - Container runtime for all services
- Docker Compose - Local orchestration for multi-service development (`/infra/docker-compose.yml`)
- Nginx alpine - API Gateway and frontend server (`/infra/docker-compose.yml`)

## Key Dependencies

**Critical Backend:**
- github.com/gin-gonic/gin v1.9.1 - HTTP routing and middleware across all services
- github.com/golang-jwt/jwt/v5 v5.3.0 - JWT token validation and creation (`/services/auth-service/go.mod`)
- github.com/lib/pq v1.10.9 - PostgreSQL driver for database connections (`/pkg/database/postgres.go`)
- github.com/nats-io/nats.go v1.31.0 - NATS client for async messaging (`/services/sensors-service/go.mod`, `/services/rules-service/go.mod`, `/services/security-service/go.mod`, `/services/notifications-service/go.mod`)
- github.com/gorilla/websocket v1.5.3 - WebSocket implementation for real-time updates (`/services/sensors-service/go.mod`, `/services/lighting-service/go.mod`)
- google.golang.org/protobuf v1.30.0 - Protocol Buffer support for serialization

**Frontend UI:**
- primeng v17.18.0 - PrimeNG component library for Angular
- primeicons v7.0.0 - Icon library for PrimeNG
- rxjs v7.8.0 - Reactive programming library for Angular
- zone.js v0.14.0 - Angular zone management
- tslib v2.3.0 - TypeScript runtime library

**Security:**
- golang.org/x/crypto v0.17.0 - Cryptographic utilities for password hashing (`/services/auth-service/go.mod`)

**Utilities:**
- github.com/google/uuid v1.5.0 - UUID generation (`/services/auth-service/go.mod`, `/services/sensors-service/go.mod`, `/services/rules-service/go.mod`, `/services/security-service/go.mod`, `/services/notifications-service/go.mod`)

## Configuration

**Environment:**
Configuration via environment variables read in each service's main.go:

**Backend Services Common Vars:**
- `JWT_SECRET` - Secret key for JWT token signing/validation
- `SERVER_PORT` - HTTP server port (default: service-specific 8080-8085)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` - PostgreSQL connection (`/services/*/cmd/api/main.go`)
- `NATS_URL` - NATS message broker connection URL (default: `nats://localhost:4222`)

**Service-Specific Vars:**
- `DEVICE_API_KEY` - API key for device authentication (sensors, lighting, security services)
- `ESP32_IP` - IP address of ESP32 device for lighting and security control
- `DEVICE_ID` - Device identifier (e.g., `esp32-main` in security-service)
- `BUZZER_DURATION_MS` - Buzzer duration in milliseconds for security alerts
- `MOTION_OFF_TIMEOUT_SECONDS` - Motion detection timeout for rules-service
- `LIGHTING_SERVICE_URL` - HTTP endpoint for lighting-service (rules-service only)
- `SECURITY_SERVICE_URL` - HTTP endpoint for security-service (rules-service only)

**Frontend Build:**
- Angular TypeScript configuration: `/frontend/tsconfig.json` (strict mode, ES2022 target)
- Tailwind CSS configuration: `/frontend/tailwind.config.js` (custom Aurora color scheme)
- PostCSS configuration: `/frontend/postcss.config.js` (Tailwind + Autoprefixer)

**Database:**
- PostgreSQL 15-alpine via Docker Compose
- Connection pooling: 25 max open, 5 max idle, 5-minute lifetime
- SSL mode: disabled for development
- Database name: `aurora_home`
- User: `aurora`

## Platform Requirements

**Development:**
- Docker and Docker Compose - For running entire stack
- Go 1.21 - For local backend development
- Node.js 20+ - For frontend local development
- npm - For managing frontend dependencies

**Production:**
- Docker/Kubernetes orchestration capable
- PostgreSQL 15+ database
- NATS message broker
- Nginx or reverse proxy for API Gateway
- ESP32 IoT device with firmware (optional, for full home automation)

**Build Output:**
- Frontend: Angular SPA compiled to `/dist/aurora-frontend/browser/` (served by Nginx)
- Backend: Go binaries in Docker images (multi-stage builds)

---

*Stack analysis: 2026-03-25*
