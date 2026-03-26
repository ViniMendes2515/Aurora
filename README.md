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

- Docker
- Docker Compose

### Subindo o Sistema

```bash
cd infra
docker compose up --build
```

O frontend Angular estará disponível em `http://localhost` (porta 80, servido pelo Nginx).

### Acessar serviços individualmente (desenvolvimento)

Cada serviço pode ser acessado diretamente pela sua porta (8080–8086) ou via API Gateway na porta 80 com o prefixo `/api/<serviço>/`.

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

## Estrutura do Projeto

```
Aurora/
├── infra/
│   ├── docker-compose.yml
│   └── nginx/             # Configuração do API Gateway
├── services/
│   ├── auth-service/
│   ├── sensors-service/
│   ├── lighting-service/
│   ├── rules-service/
│   ├── security-service/
│   ├── notifications-service/
│   └── schedule-service/      # Agendamentos cron
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
