package security

import (
	"time"

	"aurora/services/auth-service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager implementa a interface domain.JWTManager
type JWTManager struct {
	secretKey []byte
}

// NewJWTManager cria uma nova instância de JWTManager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
	}
}

// Claims representa as claims do JWT
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken gera um novo token JWT
func (m *JWTManager) GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "aurora-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken valida um token JWT e retorna as claims
func (m *JWTManager) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrTokenValidation
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, domain.ErrTokenValidation
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &domain.TokenClaims{
			UserID: claims.Subject,
			Email:  claims.Email,
		}, nil
	}

	return nil, domain.ErrTokenValidation
}
