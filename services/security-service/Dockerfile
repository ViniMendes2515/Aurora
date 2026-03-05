FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/security-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /security-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /security-service .

EXPOSE 8083

CMD ["./security-service"]
