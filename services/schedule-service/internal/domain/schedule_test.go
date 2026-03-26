package domain_test

import (
	"testing"
	"time"

	"aurora/services/schedule-service/internal/domain"
)

// validAction retorna um Action válido para uso nos testes de Schedule
func validAction() domain.Action {
	return domain.Action{
		Target:   domain.ActionTargetHTTP,
		Type:     domain.ActionTurnOnLight,
		DeviceID: "light-001",
	}
}

func TestNewSchedule_CamposPreenchidos(t *testing.T) {
	s, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "Apagar luzes da sala",
		domain.ScheduleTypeCron, "0 16 * * *", nil, 0, "UTC",
		validAction(),
	)
	if err != nil {
		t.Fatalf("esperado sem erro, obteve %v", err)
	}
	if s.ID != "sched-1" {
		t.Errorf("ID esperado sched-1, obteve %s", s.ID)
	}
	if s.OwnerID != "owner-1" {
		t.Errorf("OwnerID esperado owner-1, obteve %s", s.OwnerID)
	}
	if s.Name != "Luzes sala" {
		t.Errorf("Name esperado 'Luzes sala', obteve %s", s.Name)
	}
	if !s.Enabled {
		t.Error("novo Schedule deve ser criado habilitado")
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt nao deve ser zero")
	}
	if s.Timezone != "UTC" {
		t.Errorf("Timezone esperado UTC, obteve %s", s.Timezone)
	}
}

func TestNewSchedule_NomeVazio(t *testing.T) {
	_, err := domain.NewSchedule(
		"sched-1", "owner-1", "", "",
		domain.ScheduleTypeCron, "0 16 * * *", nil, 0, "UTC",
		validAction(),
	)
	if err != domain.ErrInvalidSchedule {
		t.Errorf("esperado ErrInvalidSchedule, obteve %v", err)
	}
}

func TestNewSchedule_TipoCronSemExpressao(t *testing.T) {
	_, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeCron, "", nil, 0, "UTC",
		validAction(),
	)
	if err != domain.ErrMissingCronExpression {
		t.Errorf("esperado ErrMissingCronExpression, obteve %v", err)
	}
}

func TestNewSchedule_OneShotSemRunAt(t *testing.T) {
	_, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeOneShot, "", nil, 0, "UTC",
		validAction(),
	)
	if err != domain.ErrMissingRunAt {
		t.Errorf("esperado ErrMissingRunAt, obteve %v", err)
	}
}

func TestNewSchedule_OneShotComRunAt(t *testing.T) {
	runAt := time.Now().Add(2 * time.Hour)
	s, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeOneShot, "", &runAt, 0, "UTC",
		validAction(),
	)
	if err != nil {
		t.Fatalf("esperado sem erro, obteve %v", err)
	}
	if s.RunAt == nil {
		t.Error("RunAt nao deve ser nil para one_shot com runAt fornecido")
	}
}

func TestNewSchedule_TimezoneDefault(t *testing.T) {
	s, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeCron, "0 16 * * *", nil, 0, "",
		validAction(),
	)
	if err != nil {
		t.Fatalf("esperado sem erro, obteve %v", err)
	}
	if s.Timezone != "America/Sao_Paulo" {
		t.Errorf("Timezone vazio deve defaultar para America/Sao_Paulo, obteve %s", s.Timezone)
	}
}

func TestNewSchedule_TimezonePersonalizado(t *testing.T) {
	s, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeCron, "0 16 * * *", nil, 0, "America/Sao_Paulo",
		validAction(),
	)
	if err != nil {
		t.Fatalf("esperado sem erro, obteve %v", err)
	}
	if s.Timezone != "America/Sao_Paulo" {
		t.Errorf("Timezone esperado America/Sao_Paulo, obteve %s", s.Timezone)
	}
}

func TestSchedule_BelongsTo(t *testing.T) {
	testes := []struct {
		nome     string
		ownerID  string
		userID   string
		esperado bool
	}{
		{"dono correto", "user1", "user1", true},
		{"dono errado", "user1", "user2", false},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			s := &domain.Schedule{OwnerID: tt.ownerID}
			if got := s.BelongsTo(tt.userID); got != tt.esperado {
				t.Errorf("BelongsTo(%s) = %v, esperado %v", tt.userID, got, tt.esperado)
			}
		})
	}
}

func TestSchedule_IsActiveOnDay(t *testing.T) {
	testes := []struct {
		nome       string
		daysFilter int
		dia        time.Weekday
		esperado   bool
	}{
		{"filtro zero = todos os dias, segunda", 0, time.Monday, true},
		{"filtro zero = todos os dias, domingo", 0, time.Sunday, true},
		{"bitmask dias uteis, segunda", 0b0111110, time.Monday, true},
		{"bitmask dias uteis, domingo", 0b0111110, time.Sunday, false},
		{"bitmask fim de semana, sabado", 0b1000001, time.Saturday, true},
		{"bitmask fim de semana, segunda", 0b1000001, time.Monday, false},
	}
	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			s := &domain.Schedule{DaysFilter: tt.daysFilter}
			if got := s.IsActiveOnDay(tt.dia); got != tt.esperado {
				t.Errorf("IsActiveOnDay(%v) = %v, esperado %v", tt.dia, got, tt.esperado)
			}
		})
	}
}

func TestSchedule_Disable(t *testing.T) {
	s, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeCron, "0 16 * * *", nil, 0, "UTC",
		validAction(),
	)
	if err != nil {
		t.Fatalf("esperado sem erro ao criar schedule, obteve %v", err)
	}
	if !s.Enabled {
		t.Fatal("Schedule deve ser habilitado após criação")
	}
	updatedBefore := s.UpdatedAt
	time.Sleep(time.Millisecond) // garante diferença de tempo
	s.Disable()
	if s.Enabled {
		t.Error("Enabled deve ser false após Disable()")
	}
	if !s.UpdatedAt.After(updatedBefore) {
		t.Error("UpdatedAt deve ser atualizado após Disable()")
	}
}

func TestSchedule_RecordRun(t *testing.T) {
	s, err := domain.NewSchedule(
		"sched-1", "owner-1", "Luzes sala", "",
		domain.ScheduleTypeCron, "0 16 * * *", nil, 0, "UTC",
		validAction(),
	)
	if err != nil {
		t.Fatalf("esperado sem erro ao criar schedule, obteve %v", err)
	}
	runAt := time.Now()
	nextRun := runAt.Add(24 * time.Hour)
	s.RecordRun(runAt, &nextRun)
	if s.LastRunAt == nil {
		t.Error("LastRunAt nao deve ser nil após RecordRun")
	}
	if !s.LastRunAt.Equal(runAt) {
		t.Errorf("LastRunAt esperado %v, obteve %v", runAt, s.LastRunAt)
	}
	if s.NextRunAt == nil {
		t.Error("NextRunAt nao deve ser nil após RecordRun com next fornecido")
	}
	if !s.NextRunAt.Equal(nextRun) {
		t.Errorf("NextRunAt esperado %v, obteve %v", nextRun, s.NextRunAt)
	}
}

func TestNewScheduleExecution_CamposPreenchidos(t *testing.T) {
	exec := domain.NewScheduleExecution("exec-1", "sched-1", domain.ExecutionStatusSuccess, "")
	if exec.ID != "exec-1" {
		t.Errorf("ID esperado exec-1, obteve %s", exec.ID)
	}
	if exec.ScheduleID != "sched-1" {
		t.Errorf("ScheduleID esperado sched-1, obteve %s", exec.ScheduleID)
	}
	if exec.Status != domain.ExecutionStatusSuccess {
		t.Errorf("Status esperado success, obteve %s", exec.Status)
	}
	if exec.ExecutedAt.IsZero() {
		t.Error("ExecutedAt nao deve ser zero")
	}
}
