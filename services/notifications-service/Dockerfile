FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

WORKDIR /app/services/notifications-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /notifications-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /notifications-service .

EXPOSE 8085

CMD ["./notifications-service"]
