# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod files first — cached unless deps change
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/notifications-service/go.mod services/notifications-service/go.sum ./services/notifications-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/notifications-service && go mod download

# Copy source and build
COPY pkg/ ./pkg/
COPY services/notifications-service/ ./services/notifications-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/notifications-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /notifications-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /notifications-service .

EXPOSE 8085

CMD ["./notifications-service"]
