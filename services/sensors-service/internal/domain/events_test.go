package domain_test

import (
	"testing"

	"aurora/services/sensors-service/internal/domain"
)

func TestNewMotionDetectedEvent_CamposPreenchidos(t *testing.T) {
	e := domain.NewMotionDetectedEvent("sensor1", "user1", "sala")
	if e.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if e.SensorID != "sensor1" {
		t.Errorf("SensorID esperado sensor1, obteve %s", e.SensorID)
	}
	if e.Intensity != 1.0 {
		t.Errorf("Intensity padrao esperado 1.0, obteve %v", e.Intensity)
	}
	if e.DetectedAt.IsZero() {
		t.Error("DetectedAt nao deve ser zero")
	}
}

func TestMotionDetectedEvent_Topic(t *testing.T) {
	e := domain.NewMotionDetectedEvent("s1", "u1", "sala")
	if got := e.Topic(); got != "sensors.motion.detected" {
		t.Errorf("Topic() esperado sensors.motion.detected, obteve %s", got)
	}
}

func TestNewMotionDetectedEvent_IDUnico(t *testing.T) {
	e1 := domain.NewMotionDetectedEvent("s1", "u1", "sala")
	e2 := domain.NewMotionDetectedEvent("s1", "u1", "sala")
	if e1.ID == e2.ID {
		t.Error("IDs de eventos distintos nao devem ser iguais")
	}
}

func TestNewLightLevelEvent_CamposPreenchidos(t *testing.T) {
	e := domain.NewLightLevelEvent("sensor1", 75.5, 3090)
	if e.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if e.Value != 75.5 {
		t.Errorf("Value esperado 75.5, obteve %v", e.Value)
	}
	if e.Raw != 3090 {
		t.Errorf("Raw esperado 3090, obteve %v", e.Raw)
	}
	if e.RecordedAt.IsZero() {
		t.Error("RecordedAt nao deve ser zero")
	}
}

func TestLightLevelEvent_Topic(t *testing.T) {
	e := domain.NewLightLevelEvent("s1", 50.0, 2048)
	if got := e.Topic(); got != "sensors.light.changed" {
		t.Errorf("Topic() esperado sensors.light.changed, obteve %s", got)
	}
}

func TestNewMotionRecord_CamposPreenchidos(t *testing.T) {
	r := domain.NewMotionRecord("sensor1", "user1")
	if r.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if r.SensorID != "sensor1" {
		t.Errorf("SensorID esperado sensor1, obteve %s", r.SensorID)
	}
	if r.DetectedAt.IsZero() {
		t.Error("DetectedAt nao deve ser zero")
	}
}

func TestNewLightRecord_CamposPreenchidos(t *testing.T) {
	r := domain.NewLightRecord("sensor1", 42.0, 1720)
	if r.ID == "" {
		t.Error("ID nao deve ser vazio")
	}
	if r.Value != 42.0 {
		t.Errorf("Value esperado 42.0, obteve %v", r.Value)
	}
	if r.Raw != 1720 {
		t.Errorf("Raw esperado 1720, obteve %v", r.Raw)
	}
	if r.RecordedAt.IsZero() {
		t.Error("RecordedAt nao deve ser zero")
	}
}
