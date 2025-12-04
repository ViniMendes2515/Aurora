package security

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrTokenValidation indica falha na validação do token
	ErrTokenValidation = errors.New("invalid or expired token")
)

// TokenClaims representa as claims extraídas do JWT
type TokenClaims struct {
	UserID string
	Email  string
}

// Claims representa as claims do JWT
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// JWTValidator valida tokens JWT
type JWTValidator struct {
	secretKey []byte
}

// NewJWTValidator cria uma nova instância de JWTValidator
func NewJWTValidator(secretKey string) *JWTValidator {
	return &JWTValidator{
		secretKey: []byte(secretKey),
	}
}

// ValidateToken valida um token JWT e retorna as claims
func (v *JWTValidator) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenValidation
		}
		return v.secretKey, nil
	})

	if err != nil {
		return nil, ErrTokenValidation
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &TokenClaims{
			UserID: claims.Subject,
			Email:  claims.Email,
		}, nil
	}

	return nil, ErrTokenValidation
}
