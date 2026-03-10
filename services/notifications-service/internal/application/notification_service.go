package application

import (
	"fmt"
	"sync"
	"time"

	"aurora/services/notifications-service/internal/domain"
)

const (
	// lightLowThreshold e o limiar de luminosidade (%) abaixo do qual uma notificacao e criada
	lightLowThreshold = 30.0

	// lightLowCooldown evita notificacoes repetidas para o mesmo sensor em curto intervalo
	lightLowCooldown = 5 * time.Minute
)

// NotificationService contem a logica de negocio de notificacoes
type NotificationService struct {
	repo         domain.NotificationRepository
	mu           sync.Mutex
	lastLightLow map[string]time.Time // sensorID -> ultima notificacao enviada
}

// NewNotificationService cria uma nova instancia do servico
func NewNotificationService(repo domain.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo:         repo,
		lastLightLow: make(map[string]time.Time),
	}
}

// GetByUserID retorna todas as notificacoes do usuario (proprias + broadcast com user_id="*")
func (s *NotificationService) GetByUserID(userID string) ([]*domain.Notification, error) {
	return s.repo.FindByUserID(userID)
}

// MarkAsRead marca uma notificacao como lida, verificando se o usuario tem acesso
func (s *NotificationService) MarkAsRead(id, userID string) error {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	// Notificacoes broadcast (user_id="*") podem ser marcadas por qualquer usuario
	if n.UserID != "*" && n.UserID != userID {
		return domain.ErrAccessDenied
	}
	return s.repo.MarkAsRead(id)
}

// HandleMotionDetected processa evento NATS de movimento e persiste notificacao
func (s *NotificationService) HandleMotionDetected(sensorID, userID, location string) error {
	n := domain.NewNotification(
		userID,
		domain.TypeMotionDetected,
		"Movimento detectado",
		fmt.Sprintf("Movimento detectado em %s (sensor: %s)", location, sensorID),
		sensorID,
		location,
	)
	return s.repo.Save(n)
}

// HandleLightLow processa evento NATS de luminosidade e cria notificacao se abaixo do limiar.
// Um cooldown de 5 minutos por sensor evita notificacoes repetidas em leituras continuas do ESP32.
func (s *NotificationService) HandleLightLow(sensorID string, value float64) error {
	if value >= lightLowThreshold {
		return nil
	}

	s.mu.Lock()
	last, seen := s.lastLightLow[sensorID]
	if seen && time.Since(last) < lightLowCooldown {
		s.mu.Unlock()
		return nil
	}
	s.lastLightLow[sensorID] = time.Now()
	s.mu.Unlock()

	n := domain.NewNotification(
		"*",
		domain.TypeLightLow,
		"Luz baixa detectada",
		fmt.Sprintf("Luminosidade baixa: %.1f%% no sensor %s", value, sensorID),
		sensorID,
		"",
	)
	return s.repo.Save(n)
}
