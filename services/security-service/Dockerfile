# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod files first — cached unless deps change
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/security-service/go.mod services/security-service/go.sum ./services/security-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/security-service && go mod download

# Copy source and build
COPY pkg/ ./pkg/
COPY services/security-service/ ./services/security-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/security-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /security-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /security-service .

EXPOSE 8083

CMD ["./security-service"]
