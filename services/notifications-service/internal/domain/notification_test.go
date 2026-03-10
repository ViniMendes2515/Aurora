package domain_test

import (
	"testing"
	"time"

	"aurora/services/notifications-service/internal/domain"
)

func TestNewNotification_CamposPreenchidos(t *testing.T) {
	n := domain.NewNotification("user1", domain.TypeMotionDetected,
		"Movimento detectado", "Movimento na sala", "sensor1", "sala")

	if n.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if n.UserID != "user1" {
		t.Errorf("UserID esperado user1, obteve %s", n.UserID)
	}
	if n.Type != domain.TypeMotionDetected {
		t.Errorf("Type esperado motion_detected, obteve %s", n.Type)
	}
	if n.Title != "Movimento detectado" {
		t.Errorf("Title inesperado: %s", n.Title)
	}
	if n.Message != "Movimento na sala" {
		t.Errorf("Message inesperado: %s", n.Message)
	}
	if n.Read {
		t.Error("Nova notificacao deve ser criada com Read = false")
	}
	if n.CreatedAt.IsZero() {
		t.Error("CreatedAt nao deve ser zero")
	}
}

func TestNewNotification_IDUnico(t *testing.T) {
	n1 := domain.NewNotification("user1", domain.TypeMotionDetected, "t", "m", "s1", "sala")
	n2 := domain.NewNotification("user1", domain.TypeMotionDetected, "t", "m", "s1", "sala")
	if n1.ID == n2.ID {
		t.Error("IDs de notificacoes distintas nao devem ser iguais")
	}
}

func TestNewNotification_BroadcastUserID(t *testing.T) {
	n := domain.NewNotification("*", domain.TypeLightLow, "Luz baixa", "Luminosidade 20%", "s1", "")
	if n.UserID != "*" {
		t.Errorf("UserID broadcast deve ser *, obteve %s", n.UserID)
	}
}

func TestNewNotification_CreatedAtRecente(t *testing.T) {
	antesDe := time.Now()
	n := domain.NewNotification("user1", domain.TypeMotionDetected, "t", "m", "s1", "sala")
	depoisDe := time.Now()

	if n.CreatedAt.Before(antesDe) || n.CreatedAt.After(depoisDe) {
		t.Errorf("CreatedAt deve estar entre %v e %v, obteve %v", antesDe, depoisDe, n.CreatedAt)
	}
}

func TestNotificationType_Constantes(t *testing.T) {
	if domain.TypeMotionDetected != "motion_detected" {
		t.Errorf("TypeMotionDetected esperado motion_detected, obteve %s", domain.TypeMotionDetected)
	}
	if domain.TypeLightLow != "light_low" {
		t.Errorf("TypeLightLow esperado light_low, obteve %s", domain.TypeLightLow)
	}
}
