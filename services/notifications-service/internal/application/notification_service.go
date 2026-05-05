package application

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"aurora/services/notifications-service/internal/domain"
)

const (
	lightLowThreshold = 30.0
	lightLowCooldown  = 5 * time.Minute
)

// NotificationService contem a logica de negocio de notificacoes
type NotificationService struct {
	repo         domain.NotificationRepository
	telegram     *TelegramService
	mu           sync.Mutex
	lastLightLow map[string]time.Time
}

// NewNotificationService cria uma nova instancia do servico
func NewNotificationService(repo domain.NotificationRepository, telegram *TelegramService) *NotificationService {
	return &NotificationService{
		repo:         repo,
		telegram:     telegram,
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
	if err := s.repo.Save(n); err != nil {
		return err
	}
	if s.telegram != nil {
		s.telegram.SendIfEnabled(context.Background(), userID, string(domain.TypeMotionDetected), location)
	}
	return nil
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

// HandleLightChanged processa eventos de luz ligada/desligada
func (s *NotificationService) HandleLightChanged(deviceID, userID, location string, isOn bool) error {
	notifType := domain.TypeLightOff
	title := "Luz desligada"
	if isOn {
		notifType = domain.TypeLightOn
		title = "Luz ligada"
	}
	n := domain.NewNotification(
		userID,
		notifType,
		title,
		fmt.Sprintf("%s em %s (dispositivo: %s)", title, location, deviceID),
		deviceID,
		location,
	)
	if err := s.repo.Save(n); err != nil {
		return err
	}
	if s.telegram != nil {
		s.telegram.SendIfEnabled(context.Background(), userID, string(notifType), location)
	}
	return nil
}

// HandleScheduleExecuted processa evento de execucao de agendamento
func (s *NotificationService) HandleScheduleExecuted(scheduleID, ownerID, actionType, scheduleName string) error {
	log.Printf("[Notif] HandleScheduleExecuted scheduleID=%s ownerID=%s action=%s name=%s telegram=%v", scheduleID, ownerID, actionType, scheduleName, s.telegram != nil)
	msg := fmt.Sprintf("Agendamento '%s' executado (ação: %s)", scheduleName, actionType)
	if scheduleName == "" {
		msg = fmt.Sprintf("Agendamento executado (ação: %s, id: %s)", actionType, scheduleID)
	}
	n := domain.NewNotification(
		ownerID,
		domain.TypeScheduleExecuted,
		"Ação agendada executada",
		msg,
		scheduleID,
		"",
	)
	if err := s.repo.Save(n); err != nil {
		log.Printf("[Notif] Erro ao salvar notificacao de agendamento: %v", err)
		return err
	}
	if s.telegram != nil {
		s.telegram.SendIfEnabled(context.Background(), ownerID, string(domain.TypeScheduleExecuted), scheduleName)
	} else {
		log.Printf("[Notif] telegram service é nil, notificacao nao enviada")
	}
	return nil
}

// HandleAlarmTriggered processa evento de alarme acionado
func (s *NotificationService) HandleAlarmTriggered(deviceID, userID, location string) error {
	n := domain.NewNotification(
		userID,
		domain.TypeAlarmTriggered,
		"Alarme acionado",
		fmt.Sprintf("Alarme acionado em %s (dispositivo: %s)", location, deviceID),
		deviceID,
		location,
	)
	if err := s.repo.Save(n); err != nil {
		return err
	}
	if s.telegram != nil {
		s.telegram.SendIfEnabled(context.Background(), userID, string(domain.TypeAlarmTriggered), location)
	}
	return nil
}
