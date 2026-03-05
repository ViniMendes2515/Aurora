module aurora/services/lighting-service

go 1.21

require (
	aurora/pkg v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/lib/pq v1.10.9 // indirect
)

replace aurora/pkg => ../../pkg
