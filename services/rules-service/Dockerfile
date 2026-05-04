# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod files first — cached unless deps change
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/rules-service/go.mod services/rules-service/go.sum ./services/rules-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/rules-service && go mod download

# Copy source and build
COPY pkg/ ./pkg/
COPY services/rules-service/ ./services/rules-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/rules-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /rules-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /rules-service .

EXPOSE 8084

CMD ["./rules-service"]
