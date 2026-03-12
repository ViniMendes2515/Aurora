package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidScheduleTime indica que o horário não está no formato HH:MM ou está fora do range válido.
var ErrInvalidScheduleTime = errors.New("invalid schedule time: expected HH:MM with hour 0-23 and minute 0-59")

// ScheduleTime é um Value Object que representa um horário no formato HH:MM.
// Usado em regras do tipo TriggerSchedule.
type ScheduleTime struct {
	value string
}

// NewScheduleTime valida e constrói um ScheduleTime.
// Retorna erro se o formato não for HH:MM ou se hora/minuto estiverem fora do range.
func NewScheduleTime(raw string) (ScheduleTime, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return ScheduleTime{}, ErrInvalidScheduleTime
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return ScheduleTime{}, ErrInvalidScheduleTime
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return ScheduleTime{}, ErrInvalidScheduleTime
	}

	if hour < 0 || hour > 23 {
		return ScheduleTime{}, ErrInvalidScheduleTime
	}
	if minute < 0 || minute > 59 {
		return ScheduleTime{}, ErrInvalidScheduleTime
	}

	return ScheduleTime{value: fmt.Sprintf("%02d:%02d", hour, minute)}, nil
}

// String retorna o valor string do horário no formato HH:MM.
func (s ScheduleTime) String() string { return s.value }

// Hour retorna a hora (0-23).
func (s ScheduleTime) Hour() int {
	hour, _ := strconv.Atoi(s.value[:2])
	return hour
}

// Minute retorna os minutos (0-59).
func (s ScheduleTime) Minute() int {
	minute, _ := strconv.Atoi(s.value[3:])
	return minute
}
