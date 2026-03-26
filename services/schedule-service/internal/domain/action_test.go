package domain_test

import (
	"testing"

	"aurora/services/schedule-service/internal/domain"
)

func TestNewAction_TargetHTTP(t *testing.T) {
	action, err := domain.NewAction("http", "turn_on_light", "light-001")
	if err != nil {
		t.Fatalf("esperado sem erro, obteve %v", err)
	}
	if action.Target != domain.ActionTargetHTTP {
		t.Errorf("Target esperado %s, obteve %s", domain.ActionTargetHTTP, action.Target)
	}
	if action.Type != domain.ActionType("turn_on_light") {
		t.Errorf("Type esperado turn_on_light, obteve %s", action.Type)
	}
	if action.DeviceID != "light-001" {
		t.Errorf("DeviceID esperado light-001, obteve %s", action.DeviceID)
	}
}

func TestNewAction_TargetNATS(t *testing.T) {
	action, err := domain.NewAction("nats", "turn_on_light", "light-001")
	if err != nil {
		t.Fatalf("esperado sem erro, obteve %v", err)
	}
	if action.Target != domain.ActionTargetNATS {
		t.Errorf("Target esperado %s, obteve %s", domain.ActionTargetNATS, action.Target)
	}
}

func TestNewAction_TargetInvalido(t *testing.T) {
	_, err := domain.NewAction("invalido", "turn_on_light", "light-001")
	if err != domain.ErrInvalidAction {
		t.Errorf("esperado ErrInvalidAction, obteve %v", err)
	}
}

func TestNewAction_TargetVazio(t *testing.T) {
	_, err := domain.NewAction("", "turn_on_light", "light-001")
	if err != domain.ErrInvalidAction {
		t.Errorf("esperado ErrInvalidAction, obteve %v", err)
	}
}
