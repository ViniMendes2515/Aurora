module aurora/services/auth-service

go 1.21

require (
	aurora/pkg v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.5.0
	golang.org/x/crypto v0.17.0
)

require github.com/lib/pq v1.10.9 // indirect

replace aurora/pkg => ../../pkg
