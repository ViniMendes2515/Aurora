# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/schedule-service/go.mod services/schedule-service/go.sum ./services/schedule-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/schedule-service && go mod download

COPY pkg/ ./pkg/
COPY services/schedule-service/ ./services/schedule-service/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd services/schedule-service && \
    CGO_ENABLED=0 GOOS=linux go build -o /schedule-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /schedule-service .

EXPOSE 8086

CMD ["./schedule-service"]
