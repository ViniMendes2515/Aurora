package application

import (
	"context"
	"log"
	"time"

	"aurora/services/schedule-service/internal/domain"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// MaxConcurrentDispatches é o número máximo de goroutines de execução simultâneas.
const MaxConcurrentDispatches = 10

// ScheduleService implementa todos os use cases de agendamento.
// Depende apenas de interfaces de domínio — sem imports de infraestrutura.
type ScheduleService struct {
	scheduleRepo    domain.ScheduleRepository
	executionRepo   domain.ScheduleExecutionRepository
	dispatcher      domain.ActionDispatcher
	registry        domain.SchedulerRegistry
	semaphore       chan struct{}
	dispatchTimeout time.Duration
}

// NewScheduleService cria uma nova instância do ScheduleService.
func NewScheduleService(
	scheduleRepo domain.ScheduleRepository,
	executionRepo domain.ScheduleExecutionRepository,
	dispatcher domain.ActionDispatcher,
	registry domain.SchedulerRegistry,
	dispatchTimeout time.Duration,
) *ScheduleService {
	if dispatchTimeout <= 0 {
		dispatchTimeout = 30 * time.Second
	}
	return &ScheduleService{
		scheduleRepo:    scheduleRepo,
		executionRepo:   executionRepo,
		dispatcher:      dispatcher,
		registry:        registry,
		semaphore:       make(chan struct{}, MaxConcurrentDispatches),
		dispatchTimeout: dispatchTimeout,
	}
}

// --- DTOs ---

// CreateScheduleRequest representa os dados de entrada para criar um agendamento.
type CreateScheduleRequest struct {
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	OwnerID        string              `json:"-"`
	ScheduleType   domain.ScheduleType `json:"schedule_type"`
	CronExpr       string              `json:"cron_expr"`
	RunAt          *time.Time          `json:"run_at"`
	DaysFilter     int                 `json:"days_filter"`
	Timezone       string              `json:"timezone"`
	ActionTarget   string              `json:"action_target"`
	ActionType     string              `json:"action_type"`
	ActionDeviceID string              `json:"action_device_id"`
	ActionPayload  string              `json:"action_payload"`
}

// UpdateScheduleRequest representa os dados de entrada para atualizar um agendamento.
type UpdateScheduleRequest struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CronExpr       string     `json:"cron_expr"`
	RunAt          *time.Time `json:"run_at"`
	DaysFilter     int        `json:"days_filter"`
	Timezone       string     `json:"timezone"`
	ActionTarget   string     `json:"action_target"`
	ActionType     string     `json:"action_type"`
	ActionDeviceID string     `json:"action_device_id"`
	ActionPayload  string     `json:"action_payload"`
}

// ScheduleResponse representa a resposta de um agendamento.
type ScheduleResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	OwnerID        string              `json:"owner_id"`
	ScheduleType   domain.ScheduleType `json:"schedule_type"`
	CronExpr       string              `json:"cron_expr"`
	RunAt          *time.Time          `json:"run_at,omitempty"`
	DaysFilter     int                 `json:"days_filter"`
	Timezone       string              `json:"timezone"`
	ActionTarget   domain.ActionTarget `json:"action_target"`
	ActionType     domain.ActionType   `json:"action_type"`
	ActionDeviceID string              `json:"action_device_id"`
	ActionPayload  string              `json:"action_payload"`
	Enabled        bool                `json:"enabled"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	LastRunAt      *time.Time          `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time          `json:"next_run_at,omitempty"`
}

// ExecutionResponse representa a resposta de uma execução.
type ExecutionResponse struct {
	ID           string                 `json:"id"`
	ScheduleID   string                 `json:"schedule_id"`
	ExecutedAt   time.Time              `json:"executed_at"`
	Status       domain.ExecutionStatus `json:"status"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// PreviewResponse representa as próximas execuções calculadas.
type PreviewResponse struct {
	ScheduleID string      `json:"schedule_id"`
	Next       []time.Time `json:"next"`
}

func toScheduleResponse(s *domain.Schedule) *ScheduleResponse {
	return &ScheduleResponse{
		ID:             s.ID,
		Name:           s.Name,
		Description:    s.Description,
		OwnerID:        s.OwnerID,
		ScheduleType:   s.ScheduleType,
		CronExpr:       s.CronExpr,
		RunAt:          s.RunAt,
		DaysFilter:     s.DaysFilter,
		Timezone:       s.Timezone,
		ActionTarget:   s.Action.Target,
		ActionType:     s.Action.Type,
		ActionDeviceID: s.Action.DeviceID,
		ActionPayload:  s.Action.Payload,
		Enabled:        s.Enabled,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		LastRunAt:      s.LastRunAt,
		NextRunAt:      s.NextRunAt,
	}
}

func toExecutionResponse(e *domain.ScheduleExecution) *ExecutionResponse {
	return &ExecutionResponse{
		ID:           e.ID,
		ScheduleID:   e.ScheduleID,
		ExecutedAt:   e.ExecutedAt,
		Status:       e.Status,
		ErrorMessage: e.ErrorMessage,
	}
}

// --- Use Cases CRUD ---

// CreateSchedule cria um novo agendamento, persiste e registra no scheduler.
func (s *ScheduleService) CreateSchedule(req CreateScheduleRequest) (*ScheduleResponse, error) {
	ctx := context.Background()

	action, err := domain.NewAction(req.ActionTarget, req.ActionType, req.ActionDeviceID)
	if err != nil {
		return nil, err
	}
	action.Payload = req.ActionPayload

	schedule, err := domain.NewSchedule(
		uuid.New().String(),
		req.OwnerID,
		req.Name,
		req.Description,
		req.ScheduleType,
		req.CronExpr,
		req.RunAt,
		req.DaysFilter,
		req.Timezone,
		*action,
	)
	if err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Save(ctx, schedule); err != nil {
		return nil, err
	}

	if s.registry != nil {
		if err := s.registry.Register(schedule); err != nil {
			log.Printf("[ScheduleService] Failed to register schedule %s in scheduler: %v", schedule.ID, err)
		}
	}

	return toScheduleResponse(schedule), nil
}

// GetSchedule busca um agendamento por ID, validando propriedade.
func (s *ScheduleService) GetSchedule(id, ownerID string) (*ScheduleResponse, error) {
	ctx := context.Background()

	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !schedule.BelongsTo(ownerID) {
		return nil, domain.ErrScheduleAccessDenied
	}
	return toScheduleResponse(schedule), nil
}

// ListSchedules lista os agendamentos de um usuário.
func (s *ScheduleService) ListSchedules(ownerID string) ([]*ScheduleResponse, error) {
	ctx := context.Background()

	schedules, err := s.scheduleRepo.FindByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	responses := make([]*ScheduleResponse, 0, len(schedules))
	for _, sch := range schedules {
		responses = append(responses, toScheduleResponse(sch))
	}
	return responses, nil
}

// UpdateSchedule atualiza um agendamento existente e re-registra no scheduler.
func (s *ScheduleService) UpdateSchedule(id, ownerID string, req UpdateScheduleRequest) (*ScheduleResponse, error) {
	ctx := context.Background()

	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !schedule.BelongsTo(ownerID) {
		return nil, domain.ErrScheduleAccessDenied
	}

	action, err := domain.NewAction(req.ActionTarget, req.ActionType, req.ActionDeviceID)
	if err != nil {
		return nil, err
	}
	action.Payload = req.ActionPayload

	schedule.Name = req.Name
	schedule.Description = req.Description
	schedule.CronExpr = req.CronExpr
	schedule.RunAt = req.RunAt
	schedule.DaysFilter = req.DaysFilter
	if req.Timezone != "" {
		schedule.Timezone = req.Timezone
	}
	schedule.Action = *action
	schedule.UpdatedAt = time.Now()

	if err := s.scheduleRepo.Save(ctx, schedule); err != nil {
		return nil, err
	}

	if s.registry != nil && schedule.Enabled {
		s.registry.Unregister(schedule.ID)
		if err := s.registry.Register(schedule); err != nil {
			log.Printf("[ScheduleService] Failed to re-register schedule %s: %v", schedule.ID, err)
		}
	}

	return toScheduleResponse(schedule), nil
}

// DeleteSchedule remove um agendamento e o cancela no scheduler.
func (s *ScheduleService) DeleteSchedule(id, ownerID string) error {
	ctx := context.Background()

	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !schedule.BelongsTo(ownerID) {
		return domain.ErrScheduleAccessDenied
	}

	if s.registry != nil {
		s.registry.Unregister(id)
	}

	return s.scheduleRepo.Delete(ctx, id)
}

// --- Toggle e Recovery ---

// ToggleSchedule inverte o estado enabled do agendamento.
func (s *ScheduleService) ToggleSchedule(id, ownerID string) (*ScheduleResponse, error) {
	ctx := context.Background()

	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !schedule.BelongsTo(ownerID) {
		return nil, domain.ErrScheduleAccessDenied
	}

	schedule.Enabled = !schedule.Enabled
	schedule.UpdatedAt = time.Now()

	if err := s.scheduleRepo.Save(ctx, schedule); err != nil {
		return nil, err
	}

	if s.registry != nil {
		if schedule.Enabled {
			if err := s.registry.Register(schedule); err != nil {
				log.Printf("[ScheduleService] Failed to register schedule %s after toggle: %v", schedule.ID, err)
			}
		} else {
			s.registry.Unregister(schedule.ID)
		}
	}

	return toScheduleResponse(schedule), nil
}

// SetRegistry define o registry de scheduler após a construção do serviço.
// Necessário para resolver a dependência circular entre ScheduleService e CronRunner.
func (s *ScheduleService) SetRegistry(r domain.SchedulerRegistry) {
	s.registry = r
}

// LoadAllEnabled retorna todos os agendamentos habilitados — usado pelo CronRunner na inicialização.
func (s *ScheduleService) LoadAllEnabled(ctx context.Context) ([]*domain.Schedule, error) {
	return s.scheduleRepo.FindAllEnabled(ctx)
}

// --- Execução e Atomicidade ---

// ExecuteSchedule executa um agendamento dentro do worker pool com timeout de contexto.
// Para agendamentos one_shot: desabilita atomicamente na mesma operação de Save.
// Para agendamentos com DaysFilter: verifica IsActiveOnDay e grava Skipped se inativo no dia.
func (s *ScheduleService) ExecuteSchedule(scheduleID string) error {
	// Adquire slot do worker pool com timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.dispatchTimeout)
	defer cancel()

	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-ctx.Done():
		log.Printf("[Scheduler] Timeout aguardando slot do worker pool para schedule %s", scheduleID)
		return ctx.Err()
	}

	return s.executeSchedule(ctx, scheduleID)
}

func (s *ScheduleService) executeSchedule(ctx context.Context, scheduleID string) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Scheduler] Panic na execução do schedule %s: %v", scheduleID, r)
			retErr = domain.ErrPublishFailed
		}
	}()

	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return err
	}

	// Verifica filtro de dia da semana
	loc := loadLocation(schedule.Timezone)
	now := time.Now().In(loc)

	if !schedule.IsActiveOnDay(now.Weekday()) {
		exec := domain.NewScheduleExecution(uuid.New().String(), scheduleID, domain.ExecutionStatusSkipped, "schedule not active on this day")
		if saveErr := s.executionRepo.Save(ctx, exec); saveErr != nil {
			log.Printf("[Scheduler] Failed to save skipped execution for %s: %v", scheduleID, saveErr)
		}
		return nil
	}

	// Executa a ação
	dispatchErr := s.dispatcher.Dispatch(ctx, schedule)

	status := domain.ExecutionStatusSuccess
	errorMsg := ""
	if dispatchErr != nil {
		status = domain.ExecutionStatusFailed
		errorMsg = dispatchErr.Error()
		log.Printf("[Scheduler] Dispatch falhou para schedule %s: %v", scheduleID, dispatchErr)
	}

	// Salva registro de execução
	exec := domain.NewScheduleExecution(uuid.New().String(), scheduleID, status, errorMsg)
	if saveErr := s.executionRepo.Save(ctx, exec); saveErr != nil {
		log.Printf("[Scheduler] Failed to save execution record for %s: %v", scheduleID, saveErr)
	}

	// Atualiza last_run_at
	schedule.RecordRun(now, nil)

	// Atomicidade one-shot: desabilita no mesmo Save que atualiza last_run_at
	if schedule.ScheduleType == domain.ScheduleTypeOneShot {
		schedule.Disable()
		if s.registry != nil {
			s.registry.Unregister(scheduleID)
		}
	}

	if saveErr := s.scheduleRepo.Save(ctx, schedule); saveErr != nil {
		log.Printf("[Scheduler] Failed to update schedule %s after execution: %v", scheduleID, saveErr)
	}

	return dispatchErr
}

// --- Histórico de Execuções ---

// ListExecutions retorna o histórico de execuções de um agendamento, validando propriedade.
func (s *ScheduleService) ListExecutions(scheduleID, ownerID string, limit int) ([]*ExecutionResponse, error) {
	ctx := context.Background()

	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	if !schedule.BelongsTo(ownerID) {
		return nil, domain.ErrScheduleAccessDenied
	}

	if limit <= 0 {
		limit = 50
	}

	executions, err := s.executionRepo.FindByScheduleID(ctx, scheduleID, limit)
	if err != nil {
		return nil, err
	}

	responses := make([]*ExecutionResponse, 0, len(executions))
	for _, e := range executions {
		responses = append(responses, toExecutionResponse(e))
	}
	return responses, nil
}

// PreviewSchedule calcula as próximas execuções de um agendamento cron.
func (s *ScheduleService) PreviewSchedule(scheduleID, ownerID string, count int) (*PreviewResponse, error) {
	ctx := context.Background()

	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	if !schedule.BelongsTo(ownerID) {
		return nil, domain.ErrScheduleAccessDenied
	}
	if schedule.ScheduleType != domain.ScheduleTypeCron {
		return &PreviewResponse{ScheduleID: scheduleID, Next: []time.Time{}}, nil
	}

	loc := loadLocation(schedule.Timezone)

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	cronSchedule, err := parser.Parse(schedule.CronExpr)
	if err != nil {
		return nil, domain.ErrInvalidSchedule
	}

	if count <= 0 || count > 20 {
		count = 5
	}
	next := make([]time.Time, 0, count)
	t := time.Now().In(loc)
	for i := 0; i < count; i++ {
		t = cronSchedule.Next(t)
		if t.IsZero() {
			break
		}
		next = append(next, t)
	}

	return &PreviewResponse{ScheduleID: scheduleID, Next: next}, nil
}

// loadLocation carrega um timezone IANA com fallback para UTC quando tzdata não está disponível.
func loadLocation(timezone string) *time.Location {
	if loc, err := time.LoadLocation(timezone); err == nil {
		return loc
	}
	if loc, err := time.LoadLocation(domain.DefaultTimezone); err == nil {
		return loc
	}
	return time.UTC
}
