package domain

import (
	"errors"
	"slices"
	"time"
)

var (
	ErrTelegramNotLinked    = errors.New("telegram account not linked")
	ErrTelegramTokenExpired = errors.New("link token expired or invalid")
)

// TelegramPreference armazena a vinculacao do usuario com o Telegram e quais tipos ele quer receber
type TelegramPreference struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ChatID       int64     `json:"chat_id"`
	EnabledTypes []string  `json:"enabled_types"`
	Active       bool      `json:"active"`
	LinkedAt     time.Time `json:"linked_at"`
}

// TelegramLinkToken e um token temporario para vincular a conta Telegram ao usuario Aurora
type TelegramLinkToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsTypeEnabled verifica se o tipo de notificacao esta habilitado para este usuario
func (p *TelegramPreference) IsTypeEnabled(notifType string) bool {
	return slices.Contains(p.EnabledTypes, notifType)
}

// DefaultEnabledTypes retorna os tipos habilitados por padrao ao vincular a conta
func DefaultEnabledTypes() []string {
	// TypeLightLow excluido — leitura continua do sensor, sem valor como alerta Telegram
	return []string{
		string(TypeMotionDetected),
		string(TypeLightOn),
		string(TypeLightOff),
		string(TypeAlarmTriggered),
		string(TypeScheduleExecuted),
	}
}

// TelegramPreferenceRepository define o contrato para persistencia das preferencias Telegram
type TelegramPreferenceRepository interface {
	Save(pref *TelegramPreference) error
	FindByUserID(userID string) (*TelegramPreference, error)
	FindByChatID(chatID int64) (*TelegramPreference, error)
	UpdateTypes(userID string, types []string) error
	SetActive(userID string, active bool) error
	Delete(userID string) error

	SaveLinkToken(token *TelegramLinkToken) error
	FindLinkToken(token string) (*TelegramLinkToken, error)
	DeleteLinkToken(token string) error
}
