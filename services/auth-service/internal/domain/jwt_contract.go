package domain

// TokenClaims representa as claims do JWT
type TokenClaims struct {
	UserID string
	Email  string
}

// JWTManager define o contrato para operações com JWT
// Esta interface pertence ao domínio e será implementada na infraestrutura
type JWTManager interface {
	// GenerateToken gera um novo token JWT para o usuário
	GenerateToken(userID, email string) (string, error)
	ValidateToken(tokenString string) (*TokenClaims, error) // ValidateToken valida um token e retorna as claims
}
