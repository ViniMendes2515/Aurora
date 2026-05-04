# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod files first — cached unless deps change
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/auth-service && go mod download

# Copy source and build
COPY pkg/ ./pkg/
COPY services/auth-service/ ./services/auth-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/auth-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /auth-service .

EXPOSE 8080

CMD ["./auth-service"]
