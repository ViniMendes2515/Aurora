package domain

import (
	"time"
)

const DefaultTimezone = "America/Sao_Paulo"

// ScheduleType define o tipo de agendamento
type ScheduleType string

const (
	ScheduleTypeCron    ScheduleType = "cron"
	ScheduleTypeOneShot ScheduleType = "one_shot"
)

// Schedule representa um agendamento de automação
// DaysFilter bitmask: 0=todos os dias, Sunday=1<<0=1, Monday=1<<1=2,
// Tuesday=1<<2=4, Wednesday=1<<3=8, Thursday=1<<4=16, Friday=1<<5=32, Saturday=1<<6=64
type Schedule struct {
	ID           string
	OwnerID      string
	Name         string
	Description  string
	ScheduleType ScheduleType
	CronExpr     string
	RunAt        *time.Time
	DaysFilter   int    // bitmask: 0=all days, Sunday=1, Monday=2, ..., Saturday=64
	Timezone     string // IANA timezone string, default "UTC"
	Action       Action
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastRunAt    *time.Time
	NextRunAt    *time.Time
}

// NewSchedule cria um novo agendamento com validação.
// Validação de presença apenas — validação sintática de cron expression é feita na camada de infrastructure/API.
func NewSchedule(id, ownerID, name, description string, scheduleType ScheduleType,
	cronExpr string, runAt *time.Time, daysFilter int, timezone string,
	action Action) (*Schedule, error) {
	if name == "" {
		return nil, ErrInvalidSchedule
	}
	if scheduleType == ScheduleTypeCron && cronExpr == "" {
		return nil, ErrMissingCronExpression
	}
	if scheduleType == ScheduleTypeOneShot && runAt == nil {
		return nil, ErrMissingRunAt
	}
	if timezone == "" {
		timezone = DefaultTimezone
	}
	now := time.Now()
	return &Schedule{
		ID:           id,
		OwnerID:      ownerID,
		Name:         name,
		Description:  description,
		ScheduleType: scheduleType,
		CronExpr:     cronExpr,
		RunAt:        runAt,
		DaysFilter:   daysFilter,
		Timezone:     timezone,
		Action:       action,
		Enabled:      true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// BelongsTo verifica se o agendamento pertence ao usuário indicado.
// Schedules são sempre user-specific (sem wildcard "*") — diferente do rules-service.
func (s *Schedule) BelongsTo(ownerID string) bool {
	return s.OwnerID == ownerID
}

// IsActiveOnDay verifica se o agendamento deve executar no dia da semana indicado.
// Se DaysFilter == 0, retorna true para qualquer dia.
// Bitmask: time.Sunday=0 → mask=1, time.Monday=1 → mask=2, ..., time.Saturday=6 → mask=64.
func (s *Schedule) IsActiveOnDay(day time.Weekday) bool {
	if s.DaysFilter == 0 {
		return true
	}
	mask := 1 << uint(day)
	return s.DaysFilter&mask != 0
}

// Disable desativa o agendamento e atualiza o timestamp de modificação.
func (s *Schedule) Disable() {
	s.Enabled = false
	s.UpdatedAt = time.Now()
}

// RecordRun registra a execução do agendamento com o horário e próxima execução.
func (s *Schedule) RecordRun(at time.Time, next *time.Time) {
	s.LastRunAt = &at
	s.NextRunAt = next
	s.UpdatedAt = at
}
