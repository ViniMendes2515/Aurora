package application_test

import (
	"testing"

	"aurora/services/notifications-service/internal/application"
	"aurora/services/notifications-service/internal/domain"
)

// ---------------------------------------------------------------------------
// Mock manual do repositório
// ---------------------------------------------------------------------------

// mockNotificationRepo implementa domain.NotificationRepository sem dependências
type mockNotificationRepo struct {
	saved            []*domain.Notification
	findNotification *domain.Notification
	findErr          error
	markAsReadCalled bool
}

func (m *mockNotificationRepo) Save(n *domain.Notification) error {
	m.saved = append(m.saved, n)
	return nil
}

func (m *mockNotificationRepo) FindByUserID(userID string) ([]*domain.Notification, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return []*domain.Notification{m.findNotification}, nil
}

func (m *mockNotificationRepo) FindByID(id string) (*domain.Notification, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.findNotification, nil
}

func (m *mockNotificationRepo) MarkAsRead(id string) error {
	m.markAsReadCalled = true
	return nil
}

// ---------------------------------------------------------------------------
// Testes
// ---------------------------------------------------------------------------

// TestGetByUserID_DelegaAoRepositorio verifica que o serviço delega corretamente
func TestGetByUserID_DelegaAoRepositorio(t *testing.T) {
	notif := &domain.Notification{ID: "n1", UserID: "user1"}
	repo := &mockNotificationRepo{findNotification: notif}
	svc := application.NewNotificationService(repo)

	resultado, err := svc.GetByUserID("user1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(resultado) != 1 || resultado[0].ID != "n1" {
		t.Errorf("esperava notificação n1, obteve: %v", resultado)
	}
}

// TestMarkAsRead_AcessoNegado garante que um usuário não consiga marcar como lida uma notificação que pertence a outro usuário.
func TestMarkAsRead_AcessoNegado(t *testing.T) {
	notif := &domain.Notification{ID: "n1", UserID: "user1"}
	repo := &mockNotificationRepo{findNotification: notif}
	svc := application.NewNotificationService(repo)

	err := svc.MarkAsRead("n1", "user2")
	if err != domain.ErrAccessDenied {
		t.Errorf("esperava ErrAccessDenied, obteve: %v", err)
	}
	if repo.markAsReadCalled {
		t.Error("MarkAsRead do repositório não deveria ter sido chamado")
	}
}

// TestMarkAsRead_BroadcastPodeSerMarcadoPorQualquer assegura que notificações do tipo broadcast (UserID="*") possam ser marcadas como lidas por qualquer usuário
func TestMarkAsRead_BroadcastPodeSerMarcadoPorQualquer(t *testing.T) {
	notif := &domain.Notification{ID: "n2", UserID: "*"}
	repo := &mockNotificationRepo{findNotification: notif}
	svc := application.NewNotificationService(repo)

	err := svc.MarkAsRead("n2", "qualquer-usuario")
	if err != nil {
		t.Fatalf("broadcast deve ser marcável por qualquer usuário, obteve erro: %v", err)
	}
	if !repo.markAsReadCalled {
		t.Error("MarkAsRead do repositório deveria ter sido chamado")
	}
}

// TestMarkAsRead_Sucesso verifica o fluxo feliz: o dono da notificação pode marcá-la como lida sem erros, e o repositório é chamado corretamente.
func TestMarkAsRead_Sucesso(t *testing.T) {
	notif := &domain.Notification{ID: "n3", UserID: "user1"}
	repo := &mockNotificationRepo{findNotification: notif}
	svc := application.NewNotificationService(repo)

	err := svc.MarkAsRead("n3", "user1")
	if err != nil {
		t.Fatalf("dono deve conseguir marcar como lida, obteve erro: %v", err)
	}
	if !repo.markAsReadCalled {
		t.Error("MarkAsRead do repositório deveria ter sido chamado")
	}
}

// TestHandleMotionDetected_CriaNotificacao verifica que o evento de movimento resulta em exatamente uma notificação persistida com o tipo correto.
func TestHandleMotionDetected_CriaNotificacao(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := application.NewNotificationService(repo)

	err := svc.HandleMotionDetected("sensor-01", "user1", "Sala")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("esperava 1 notificação salva, obteve %d", len(repo.saved))
	}
	if repo.saved[0].Type != domain.TypeMotionDetected {
		t.Errorf("tipo esperado: %s, obteve: %s", domain.TypeMotionDetected, repo.saved[0].Type)
	}
	if repo.saved[0].SensorID != "sensor-01" {
		t.Errorf("sensorID esperado: sensor-01, obteve: %s", repo.saved[0].SensorID)
	}
	if repo.saved[0].UserID != "user1" {
		t.Errorf("userID esperado: user1, obteve: %s", repo.saved[0].UserID)
	}
}

// TestHandleLightLow_AbaixoDoLimiar_CriaNotificacao confirma que valores de luminosidade abaixo de 30% geram uma notificação broadcast, alertando todos os usuários sobre a condição de luz baixa no ambiente.
func TestHandleLightLow_AbaixoDoLimiar_CriaNotificacao(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := application.NewNotificationService(repo)

	err := svc.HandleLightLow("sensor-02", 10.0)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("esperava 1 notificação salva, obteve %d", len(repo.saved))
	}
	if repo.saved[0].Type != domain.TypeLightLow {
		t.Errorf("tipo esperado: %s, obteve: %s", domain.TypeLightLow, repo.saved[0].Type)
	}
	if repo.saved[0].UserID != "*" {
		t.Errorf("notificação de luz baixa deve ser broadcast (*), obteve: %s", repo.saved[0].UserID)
	}
}

// TestHandleLightLow_AcimaDoLimiar_NaoCria garante que leituras normais (>= 30%) não gerem notificações desnecessárias, evitando ruído para os usuários.
func TestHandleLightLow_AcimaDoLimiar_NaoCria(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := application.NewNotificationService(repo)

	err := svc.HandleLightLow("sensor-03", 50.0)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Errorf("não deveria salvar notificação para valor acima do limiar, salvou %d", len(repo.saved))
	}
}

// TestHandleLightLow_Cooldown_NaoCriaDuplicata protege contra spam de notificações causado por leituras contínuas do ESP32
func TestHandleLightLow_Cooldown_NaoCriaDuplicata(t *testing.T) {
	repo := &mockNotificationRepo{}
	svc := application.NewNotificationService(repo)

	// Primeira chamada: deve criar a notificação normalmente
	err := svc.HandleLightLow("sensor-04", 10.0)
	if err != nil {
		t.Fatalf("erro na primeira chamada: %v", err)
	}

	// Segunda chamada imediata: deve ser bloqueada pelo cooldown de 5 minutos
	err = svc.HandleLightLow("sensor-04", 10.0)
	if err != nil {
		t.Fatalf("erro na segunda chamada: %v", err)
	}

	if len(repo.saved) != 1 {
		t.Errorf("cooldown deve impedir duplicata: esperava 1 notificação, obteve %d", len(repo.saved))
	}
}
