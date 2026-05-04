# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod files first — cached unless deps change
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/lighting-service/go.mod services/lighting-service/go.sum ./services/lighting-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/lighting-service && go mod download

# Copy source and build
COPY pkg/ ./pkg/
COPY services/lighting-service/ ./services/lighting-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/lighting-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /lighting-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /lighting-service .

EXPOSE 8082

CMD ["./lighting-service"]
