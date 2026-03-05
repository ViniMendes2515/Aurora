FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/lighting-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /lighting-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /lighting-service .

EXPOSE 8082

CMD ["./lighting-service"]
