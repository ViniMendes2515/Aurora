package application_test

import (
	"errors"
	"testing"

	"aurora/services/auth-service/internal/application"
	"aurora/services/auth-service/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

// =============================================================================
// MOCKS
// Por que criar mocks manualmente em vez de usar uma lib como `testify/mock`?
//
// Go idiomático prefere mocks simples e explícitos. Uma struct que implementa
// a interface é suficiente — e mais fácil de ler e entender.
// =============================================================================

// mockUserRepo implementa domain.UserRepository em memória.
// Campos configuráveis permitem controlar o comportamento em cada teste.
type mockUserRepo struct {
	existsByEmail bool         // controla o retorno de ExistsByEmail
	findUser      *domain.User // usuário que FindByEmail vai retornar
	findErr       error        // erro que FindByEmail vai retornar
	saveErr       error        // erro que Save vai retornar
}

func (m *mockUserRepo) Save(user *domain.User) error { return m.saveErr }
func (m *mockUserRepo) FindByEmail(email string) (*domain.User, error) {
	return m.findUser, m.findErr
}
func (m *mockUserRepo) FindByID(id string) (*domain.User, error) { return nil, nil }
func (m *mockUserRepo) ExistsByEmail(email string) bool          { return m.existsByEmail }

// mockJWTManager implementa domain.JWTManager em memória.
type mockJWTManager struct {
	token    string
	tokenErr error
}

func (m *mockJWTManager) GenerateToken(userID, email string) (string, error) {
	return m.token, m.tokenErr
}
func (m *mockJWTManager) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	return nil, nil
}

// =============================================================================
// TESTES — Register
// =============================================================================

// TestRegister_SenhaInvalida verifica que senhas com menos de 6 caracteres são rejeitadas.
// Por que este teste importa?
// A validação de senha mínima é uma regra de negócio da camada de application.
// Ela deve acontecer ANTES de qualquer I/O (banco, bcrypt). Este teste garante isso.
func TestRegister_SenhaInvalida(t *testing.T) {
	svc := application.NewAuthService(&mockUserRepo{}, &mockJWTManager{})

	_, err := svc.Register(application.RegisterRequest{
		Email:    "usuario@teste.com",
		Password: "123", // menos de 6 caracteres
	})

	if !errors.Is(err, domain.ErrInvalidPassword) {
		t.Errorf("esperado ErrInvalidPassword, obteve: %v", err)
	}
}

// TestRegister_EmailDuplicado verifica que cadastrar um email já existente retorna ErrUserAlreadyExists.
// Por que este teste importa?
// Demonstra que o AuthService coordena corretamente com o repositório:
// pergunta se o email existe antes de tentar salvar. Sem este teste,
// poderíamos quebrar essa lógica sem perceber.
func TestRegister_EmailDuplicado(t *testing.T) {
	repo := &mockUserRepo{existsByEmail: true} // simula email já cadastrado
	svc := application.NewAuthService(repo, &mockJWTManager{})

	_, err := svc.Register(application.RegisterRequest{
		Email:    "existente@teste.com",
		Password: "senha123",
	})

	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Errorf("esperado ErrUserAlreadyExists, obteve: %v", err)
	}
}

// TestRegister_Sucesso verifica o caminho feliz: cadastro retorna ID e Email corretos.
// Por que este teste importa?
// Valida que o AuthService: cria o usuário, persiste e retorna os dados corretos.
func TestRegister_Sucesso(t *testing.T) {
	repo := &mockUserRepo{existsByEmail: false}
	svc := application.NewAuthService(repo, &mockJWTManager{})

	resp, err := svc.Register(application.RegisterRequest{
		Email:    "novo@teste.com",
		Password: "senha123",
	})

	if err != nil {
		t.Fatalf("Register() erro inesperado: %v", err)
	}
	if resp.ID == "" {
		t.Error("ID da resposta não deve ser vazio")
	}
	if resp.Email != "novo@teste.com" {
		t.Errorf("Email esperado novo@teste.com, obteve: %s", resp.Email)
	}
}

// TestRegister_SenhaNaoEArmazenadaEmPlainText verifica que a senha é hasheada antes de persistir.
//
// Por que este teste é especialmente importante?
//
// Esta é uma regra de SEGURANÇA. Se alguém remover o bcrypt por acidente,
// os testes de lógica continuariam passando — mas senhas em plain text
// estariam sendo salvas. Este teste captura exatamente isso.
//
// Como funciona: interceptamos o Save() para capturar o User que foi salvo,
// e então tentamos verificar se o hash bate com a senha original usando bcrypt.
// Se bater, a senha estava hasheada corretamente.
func TestRegister_SenhaNaoEArmazenadaEmPlainText(t *testing.T) {
	// Mock com Save customizado para capturar o usuário
	var usuarioSalvo *domain.User
	repoCaptura := &capturaUserRepo{
		onSave: func(u *domain.User) { usuarioSalvo = u },
	}

	svc := application.NewAuthService(repoCaptura, &mockJWTManager{})
	senhaOriginal := "minha-senha"

	_, err := svc.Register(application.RegisterRequest{
		Email:    "seguro@teste.com",
		Password: senhaOriginal,
	})
	if err != nil {
		t.Fatalf("Register() erro inesperado: %v", err)
	}

	// A senha salva NÃO deve ser igual ao plain text
	if usuarioSalvo.PasswordHash == senhaOriginal {
		t.Error("senha armazenada em plain text — isso é uma falha de segurança!")
	}

	// Deve ser um hash bcrypt válido
	err = bcrypt.CompareHashAndPassword([]byte(usuarioSalvo.PasswordHash), []byte(senhaOriginal))
	if err != nil {
		t.Errorf("PasswordHash não é um hash bcrypt válido da senha: %v", err)
	}
}

// capturaUserRepo permite interceptar chamadas ao Save via callback.
// Por que usar callback em vez de herança?
// Go não tem herança. Callbacks (funções como campos) são o padrão idiomático
// para personalizar comportamento de mocks sem duplicar código.
type capturaUserRepo struct {
	mockUserRepo
	onSave func(*domain.User)
}

func (c *capturaUserRepo) Save(user *domain.User) error {
	if c.onSave != nil {
		c.onSave(user)
	}
	return nil
}

// =============================================================================
// TESTES — Login
// =============================================================================

// TestLogin_CredenciaisInvalidas_UsuarioNaoEncontrado verifica que login com email
// desconhecido retorna ErrInvalidCredentials (e NÃO ErrUserNotFound).
//
// Por que não retornar ErrUserNotFound?
// Segurança: se retornássemos erros diferentes para "email não existe" vs
// "senha errada", um atacante consegue enumerar quais emails existem no sistema.
// Retornar sempre ErrInvalidCredentials evita esse vazamento de informação.
func TestLogin_CredenciaisInvalidas_UsuarioNaoEncontrado(t *testing.T) {
	repo := &mockUserRepo{findErr: domain.ErrUserNotFound}
	svc := application.NewAuthService(repo, &mockJWTManager{})

	_, err := svc.Login(application.LoginRequest{
		Email:    "fantasma@teste.com",
		Password: "qualquer",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("esperado ErrInvalidCredentials para email inexistente, obteve: %v", err)
	}
}

// TestLogin_CredenciaisInvalidas_SenhaErrada verifica que senha incorreta retorna ErrInvalidCredentials.
// Este teste requer criar um usuário com hash bcrypt real, pois o AuthService
// usa bcrypt.CompareHashAndPassword — que não pode ser mockado.
func TestLogin_CredenciaisInvalidas_SenhaErrada(t *testing.T) {
	// Cria o hash bcrypt de "senha-correta" para simular o banco
	hashCorreto, _ := bcrypt.GenerateFromPassword([]byte("senha-correta"), bcrypt.MinCost)

	user, _ := domain.NewUser("usuario@teste.com", string(hashCorreto))
	repo := &mockUserRepo{findUser: user, findErr: nil}
	svc := application.NewAuthService(repo, &mockJWTManager{token: "token-jwt"})

	_, err := svc.Login(application.LoginRequest{
		Email:    "usuario@teste.com",
		Password: "senha-ERRADA", // senha diferente
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("esperado ErrInvalidCredentials para senha errada, obteve: %v", err)
	}
}

// TestLogin_Sucesso verifica o caminho feliz: credenciais corretas retornam um token.
func TestLogin_Sucesso(t *testing.T) {
	senhaCorreta := "senha-correta"
	hash, _ := bcrypt.GenerateFromPassword([]byte(senhaCorreta), bcrypt.MinCost)

	user, _ := domain.NewUser("usuario@teste.com", string(hash))
	repo := &mockUserRepo{findUser: user, findErr: nil}
	jwt := &mockJWTManager{token: "token-jwt-gerado"}
	svc := application.NewAuthService(repo, jwt)

	resp, err := svc.Login(application.LoginRequest{
		Email:    "usuario@teste.com",
		Password: senhaCorreta,
	})

	if err != nil {
		t.Fatalf("Login() erro inesperado: %v", err)
	}
	if resp.Token != "token-jwt-gerado" {
		t.Errorf("Token esperado 'token-jwt-gerado', obteve: %s", resp.Token)
	}
}
