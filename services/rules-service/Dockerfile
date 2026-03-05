FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/rules-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /rules-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /rules-service .

EXPOSE 8084

CMD ["./rules-service"]
