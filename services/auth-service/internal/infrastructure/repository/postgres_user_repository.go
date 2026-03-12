package repository

import (
	"database/sql"

	"aurora/services/auth-service/internal/domain"
)

// PostgresUserRepository implementa domain.UserRepository usando PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository cria uma nova instância do repositório PostgreSQL
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// Save persiste um novo usuário
func (r *PostgresUserRepository) Save(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, user.ID, user.Email.String(), user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

// FindByEmail busca um usuário pelo email
func (r *PostgresUserRepository) FindByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var id, emailStr, passwordHash string
	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&id,
		&emailStr,
		&passwordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	e, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	user.ID = id
	user.Email = e
	user.PasswordHash = passwordHash
	return user, nil
}

// FindByID busca um usuário pelo ID
func (r *PostgresUserRepository) FindByID(id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var userID, emailStr, passwordHash string
	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&userID,
		&emailStr,
		&passwordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	e, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}
	user.ID = userID
	user.Email = e
	user.PasswordHash = passwordHash
	return user, nil
}

// ExistsByEmail verifica se existe um usuário com o email
func (r *PostgresUserRepository) ExistsByEmail(email string) bool {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
