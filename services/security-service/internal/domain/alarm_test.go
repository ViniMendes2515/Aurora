package domain_test

import (
	"testing"

	"aurora/services/security-service/internal/domain"
)

func TestNewAlarmEvent_CamposPreenchidos(t *testing.T) {
	a, err := domain.NewAlarmEvent("user-1", "motion", "sensor1", "sala")
	if err != nil {
		t.Fatalf("NewAlarmEvent() erro inesperado: %v", err)
	}
	if a.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if a.TriggerType.String() != "motion" {
		t.Errorf("TriggerType esperado motion, obteve %s", a.TriggerType.String())
	}
	if a.Status != domain.AlarmStatusTriggered {
		t.Errorf("Status inicial deve ser triggered, obteve %s", a.Status)
	}
	if a.SilencedAt != nil {
		t.Error("SilencedAt deve ser nil em novo alarme")
	}
	if a.TriggeredAt.IsZero() {
		t.Error("TriggeredAt nao deve ser zero")
	}
}

func TestNewAlarmEvent_IDUnico(t *testing.T) {
	a1, _ := domain.NewAlarmEvent("user-1", "motion", "s1", "sala")
	a2, _ := domain.NewAlarmEvent("user-1", "motion", "s2", "quarto")
	if a1.ID == a2.ID {
		t.Error("IDs de alarmes distintos nao devem ser iguais")
	}
}

func TestAlarmEvent_Silence(t *testing.T) {
	a, _ := domain.NewAlarmEvent("user-1", "manual", "sensor1", "quarto")
	a.Silence()
	if a.Status != domain.AlarmStatusSilenced {
		t.Errorf("Status deve ser silenced apos Silence(), obteve %s", a.Status)
	}
	if a.SilencedAt == nil {
		t.Error("SilencedAt nao deve ser nil apos Silence()")
	}
}

func TestAlarmEvent_Silence_PreservaTriggerType(t *testing.T) {
	a, _ := domain.NewAlarmEvent("user-1", "motion", "sensor1", "sala")
	a.Silence()
	if a.TriggerType.String() != "motion" {
		t.Errorf("Silence nao deve alterar TriggerType, obteve %s", a.TriggerType.String())
	}
}
