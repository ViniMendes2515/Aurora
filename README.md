# Aurora Home System

Sistema de automação residencial baseado em microsserviços, DDD (Domain-Driven Design) e arquitetura em camadas.

## Arquitetura

O projeto segue uma arquitetura de microsserviços onde cada serviço respeita rigorosamente a separação em camadas:

```
/domain         → Entidades, agregados, eventos, interfaces de repositório e contratos
/application    → Casos de uso (services), orquestra lógica de negócio
/infrastructure → Implementações concretas (HTTP/Gin, JWT, NATS, banco, repositórios, WebSocket)
/cmd/api        → Inicialização do serviço
```

### Serviços e Portas

| Serviço | Porta | Responsabilidade |
|---|---|---|
| auth-service | 8080 | Autenticação JWT, cadastro e login de usuários |
| sensors-service | 8081 | Sensores de movimento/temperatura/umidade + WebSocket |
| lighting-service | 8082 | Controle de iluminação + WebSocket |
| security-service | 8083 | Sistema de alarme residencial |
| rules-service | 8084 | Motor de regras de automação |
| notifications-service | 8085 | Notificações de eventos do sistema |
| schedule-service | 8086 | Agendamentos de automação (cron e one-shot) |
| **Nginx (API Gateway)** | **80** | Roteamento `/api/*` → serviços + serve o frontend Angular |

### Bounded Contexts

Cada serviço representa um Bounded Context isolado com domínio próprio:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Aurora Home System                             │
│                                                                         │
│  ┌──────────────┐   ┌──────────────────┐   ┌──────────────────────┐     │
│  │ Auth Context │   │ Sensors Context  │   │  Lighting Context    │     │
│  │              │   │                  │   │                      │     │
│  │ - User       │   │ - Sensor         │   │ - Light              │     │
│  │ - JWT tokens │   │ - MotionEvent    │   │ - LightState         │     │
│  └──────────────┘   └────────┬─────────┘   └──────────┬───────────┘     │
│                              │ NATS                    │ NATS           │
│  ┌──────────────┐   ┌────────▼─────────┐   ┌──────────▼───────────┐     │
│  │ Rules Context│   │Security Context  │   │Notifications Context │     │
│  │              │   │                  │   │                      │     │
│  │ - Rule       │◄──│ - Alarm          │   │ - Notification       │     │
│  │ - Trigger    │   │ - AlarmEvent     │   │ - EventLog           │     │
│  └──────────────┘   └──────────────────┘   └──────────────────────┘     │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                     Schedule Context                              │  │
│  │                                                                   │  │
│  │  - Schedule (cron / one-shot)    - ScheduleExecution              │  │
│  │  - CronRunner (gocron/v2)        - HTTPActionClient               │  │
│  │  Dispara ações HTTP em lighting-service e security-service        │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

### Comunicação entre Serviços

- **HTTP/REST**: Via Gin, roteado pelo Nginx API Gateway
- **Assíncrono (NATS PubSub)**: Eventos de domínio publicados entre serviços (ex: `sensor.motion.detected` → rules-service e notifications-service)
- **Real-time (WebSocket)**: Endpoints `/api/sensors/ws` e `/api/lighting/ws` para o frontend Angular
- **Agendamento (gocron/v2)**: schedule-service dispara ações HTTP para lighting-service e security-service nos horários configurados

## Como Executar

### Pré-requisitos

- Docker e Docker Compose
- Go 1.21+
- [swag CLI](https://github.com/swaggo/swag) (`go install github.com/swaggo/swag/cmd/swag@latest`)

### Makefile

O projeto tem um Makefile na raiz com os principais comandos. Para ver todos:

```bash
make help
```

#### Comandos mais usados

```bash
make up            # sobe todos os serviços em background
make up-build      # sobe rebuilding as imagens
make debug         # sobe em modo debug (Swagger ativo + portas expostas)
make down          # derruba tudo
make logs          # logs de todos os serviços
make logs-<svc>    # logs de um serviço  ex: make logs-auth-service
make test          # roda os testes de todos os serviços
make build         # compila todos os serviços
make swagger       # regenera a documentação Swagger de todos os serviços
```

### Subindo o Sistema

```bash
make up-build
```

O frontend Angular estará disponível em `http://localhost` (porta 80, servido pelo Nginx).

### Modo Debug (Swagger)

```bash
make debug
```

Sobe todos os serviços com `DEBUG=true` e expõe as portas diretamente no host. A documentação Swagger fica disponível em:

| Serviço | URL |
|---|---|
| auth-service | http://localhost:8080/swagger/index.html |
| sensors-service | http://localhost:8081/swagger/index.html |
| lighting-service | http://localhost:8082/swagger/index.html |
| security-service | http://localhost:8083/swagger/index.html |
| rules-service | http://localhost:8084/swagger/index.html |
| notifications-service | http://localhost:8085/swagger/index.html |
| schedule-service | http://localhost:8086/swagger/index.html |

Para regenerar os docs após alterar anotações nos handlers:

```bash
make swagger          # todos os serviços
make swagger-<svc>    # ex: make swagger-auth-service
```

## Tecnologias

- **Go** — Linguagem de todos os microsserviços backend
- **Gin** — Framework HTTP usado em todos os serviços
- **Angular 18** — SPA frontend (standalone components, lazy loading)
- **Tailwind CSS 3.4 + PrimeNG 17** — UI do frontend
- **NATS** — Message broker para comunicação assíncrona entre serviços
- **PostgreSQL** — Banco de dados relacional (cada serviço com seu schema)
- **Nginx** — API Gateway e servidor do frontend em produção
- **Docker / Docker Compose** — Containerização e orquestração local
- **JWT HS256** — Autenticação stateless com expiração de 1 hora
- **Swagger (swaggo/swag)** — Documentação interativa das APIs, disponível em modo debug

## Estrutura do Projeto

```
Aurora/
├── Makefile                   # Comandos principais do projeto
├── infra/
│   ├── docker-compose.yml
│   ├── docker-compose.debug.yml   # Override para modo debug (Swagger + portas)
│   └── nginx/                 # Configuração do API Gateway
├── services/
│   ├── auth-service/
│   │   ├── docs/              # Swagger gerado (swag init)
│   │   └── ...
│   ├── sensors-service/
│   ├── lighting-service/
│   ├── rules-service/
│   ├── security-service/
│   ├── notifications-service/
│   └── schedule-service/
├── pkg/
│   ├── jwt/               # JWTValidator, JWTManager, AuthMiddleware compartilhados
│   └── database/          # Conexão e health check do PostgreSQL
├── frontend/              # Angular 18 SPA
├── firmware/              # Código ESP32 (dispositivos IoT)
└── docs/
```

## Autenticação

- **Usuários**: JWT HS256, header `Authorization: Bearer <token>`, expiração de 1 hora
- **Dispositivos ESP32**: Header `X-Device-Key` com chave de API (bypass do JWT via `AuthMiddlewareWithDeviceKey`)
- Senhas armazenadas com hash bcrypt

## Segurança

- Senhas armazenadas com hash bcrypt
- Tokens JWT com expiração de 1 hora
- Algoritmo HS256 para assinatura de tokens
- Validação de propriedade de recursos por usuário

## Licença

MIT License
