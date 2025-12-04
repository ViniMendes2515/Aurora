package domain

// UserRepository define o contrato para persistência de usuários
// Esta interface pertence ao domínio e será implementada na infraestrutura
type UserRepository interface {
	// Save persiste um novo usuário
	Save(user *User) error

	// FindByEmail busca um usuário pelo email
	FindByEmail(email string) (*User, error)

	// FindByID busca um usuário pelo ID
	FindByID(id string) (*User, error)

	// ExistsByEmail verifica se existe um usuário com o email
	ExistsByEmail(email string) bool
}
