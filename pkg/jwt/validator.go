package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// ErrTokenValidation é retornado sempre que um token não puder ser validado
var ErrTokenValidation = errors.New("invalid or expired token")

// JWTValidator valida tokens JWT assinados com HMAC (HS256) e uma chave secreta compartilhada.
type JWTValidator struct {
	secretKey []byte
}

// TokenClaims contém apenas os dados da aplicação extraídos do token validado.
type TokenClaims struct {
	UserID string
	Email  string
}

// claims é a struct interna que mapeia o payload real do token JWT.
type claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// NewJWTValidator cria um JWTValidator usando a chave secreta fornecida.
func NewJWTValidator(secretKey string) *JWTValidator {
	return &JWTValidator{secretKey: []byte(secretKey)}
}

// ValidateToken parseia e valida um token JWT em formato string.
func (v *JWTValidator) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", t.Header["alg"])
		}
		return v.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrTokenValidation
	}

	c, ok := token.Claims.(*claims)
	if !ok {
		return nil, ErrTokenValidation
	}

	// GetSubject retorna o campo "sub" do token, que armazena o ID do usuário.
	sub, err := c.GetSubject()
	if err != nil {
		return nil, ErrTokenValidation
	}

	return &TokenClaims{
		UserID: sub,
		Email:  c.Email,
	}, nil
}
