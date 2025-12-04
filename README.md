# Aurora Home System

Sistema de automação residencial baseado em microsserviços, DDD (Domain-Driven Design) e arquitetura em camadas (Layered Architecture).

## 🏗️ Arquitetura

O projeto segue uma arquitetura de microsserviços onde cada serviço respeita rigorosamente a separação em camadas:

```
/domain         → Entidades, agregados, eventos, interfaces de repositório e contratos
/application    → Casos de uso (services), orquestra lógica de negócio
/infrastructure → Implementações concretas (HTTP, JWT, NATS, banco, repositórios)
/cmd/api        → Inicialização do serviço
```

### Serviços

| Serviço | Porta | Descrição |
|---------|-------|-----------|
| auth-service | 8080 | Autenticação e autorização (JWT) |
| sensors-service | 8081 | Gerenciamento de sensores e detecção de movimento |
| lighting-service | - | Controle de iluminação (em desenvolvimento) |
| rules-service | - | Motor de regras (Rust - em desenvolvimento) |
| security-service | - | Segurança residencial (Rust - em desenvolvimento) |
| notifications-service | - | Notificações (em desenvolvimento) |

## 🚀 Como Executar

### Pré-requisitos

- Docker
- Docker Compose

### Subindo o Sistema

```bash
cd aurora/infra
docker compose up --build
```

## 📡 API Endpoints

### Auth Service (porta 8080)

#### Registrar Usuário

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "usuario@exemplo.com", "password": "senha123"}'
```

Resposta (201 Created):
```json
{
  "id": "uuid-do-usuario",
  "email": "usuario@exemplo.com"
}
```

#### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "usuario@exemplo.com", "password": "senha123"}'
```

Resposta (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Sensors Service (porta 8081)

#### Registrar Movimento em Sensor

```bash
curl -X POST http://localhost:8081/sensors/sensor-001/motion \
  -H "Authorization: Bearer <seu-token-jwt>"
```

Resposta (200 OK):
```json
{
  "status": "motion registered"
}
```

## 🔧 Tecnologias

- **Go** - Linguagem principal dos microsserviços
- **Rust** - Para serviços de alta performance (rules, security)
- **NATS** - Message broker para comunicação entre serviços
- **PostgreSQL** - Banco de dados relacional
- **Docker** - Containerização
- **JWT** - Autenticação stateless

## 📁 Estrutura do Projeto

```
aurora/
├── infra/
│   └── docker-compose.yml
├── services/
│   ├── auth-service/
│   ├── sensors-service/
│   ├── lighting-service/
│   ├── rules-service/
│   ├── security-service/
│   └── notifications-service/
├── frontend/
│   └── aurora-web/
└── README.md
```

## 🔐 Segurança

- Senhas armazenadas com hash bcrypt
- Tokens JWT com expiração de 1 hora
- Algoritmo HS256 para assinatura de tokens
- Validação de propriedade de sensores por usuário

## 📋 Licença

MIT License
