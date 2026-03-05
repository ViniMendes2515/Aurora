FROM golang:1.21-alpine AS builder

# Diretório base do monorepo dentro da imagem
WORKDIR /app

# Copia TODO o monorepo (inclui services/ e pkg/)
COPY . .

# Entra na pasta do serviço de autenticação
WORKDIR /app/services/auth-service

# Baixa as dependências com base no go.mod do serviço
RUN go mod download

# Build do binário do serviço
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /auth-service .

EXPOSE 8080

CMD ["./auth-service"]
