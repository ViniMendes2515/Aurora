package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrNotificationNotFound indica que a notificacao nao foi encontrada
	ErrNotificationNotFound = errors.New("notification not found")

	// ErrAccessDenied indica que o usuario nao tem permissao para acessar a notificacao
	ErrAccessDenied = errors.New("access denied")
)

// NotificationType define o tipo de notificacao gerada pelo sistema
type NotificationType string

const (
	// TypeMotionDetected indica deteccao de movimento por sensor
	TypeMotionDetected NotificationType = "motion_detected"

	// TypeLightLow indica luminosidade abaixo do limiar configurado
	TypeLightLow NotificationType = "light_low"
)

// Notification representa uma notificacao gerada por um evento do sistema
type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`    // ID do destinatario, ou "*" para broadcast
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	SensorID  string           `json:"sensor_id"`
	Location  string           `json:"location"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
}

// NewNotification cria uma nova notificacao com ID unico e timestamp atual
func NewNotification(userID string, notifType NotificationType, title, message, sensorID, location string) *Notification {
	return &Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		SensorID:  sensorID,
		Location:  location,
		Read:      false,
		CreatedAt: time.Now(),
	}
}

// NotificationRepository define o contrato para persistencia de notificacoes
type NotificationRepository interface {
	Save(n *Notification) error
	FindByUserID(userID string) ([]*Notification, error)
	FindByID(id string) (*Notification, error)
	MarkAsRead(id string) error
}
