FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/schedule-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /schedule-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /schedule-service .

EXPOSE 8086

CMD ["./schedule-service"]
