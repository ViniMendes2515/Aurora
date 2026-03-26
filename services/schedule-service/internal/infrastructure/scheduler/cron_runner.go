package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/domain"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

// CronRunner implementa domain.SchedulerRegistry usando gocron/v2.
type CronRunner struct {
	scheduler   gocron.Scheduler
	service     *application.ScheduleService
	jobsByID    map[string]uuid.UUID
	mu          sync.Mutex
	gracePeriod time.Duration
}

// NewCronRunner cria uma nova instância do CronRunner.
func NewCronRunner(service *application.ScheduleService, gracePeriod time.Duration) (*CronRunner, error) {
	if gracePeriod <= 0 {
		gracePeriod = 5 * time.Minute
	}
	location, err := time.LoadLocation(domain.DefaultTimezone)
	if err != nil {
		location = time.Local
		log.Printf("[CronRunner] Falha ao carregar timezone %s, usando time.Local: %v", domain.DefaultTimezone, err)
	}

	s, err := gocron.NewScheduler(gocron.WithLocation(location))
	if err != nil {
		return nil, err
	}
	return &CronRunner{
		scheduler:   s,
		service:     service,
		jobsByID:    make(map[string]uuid.UUID),
		gracePeriod: gracePeriod,
	}, nil
}

// Register registra um agendamento no scheduler interno.
// Se o agendamento já estiver registrado, ele é removido e re-registrado.
func (runner *CronRunner) Register(schedule *domain.Schedule) error {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	// Remove o job anterior se já existir
	if _, exists := runner.jobsByID[schedule.ID]; exists {
		runner.unregisterLocked(schedule.ID)
	}

	uid := uuid.New()

	var job gocron.Job
	var err error

	switch schedule.ScheduleType {
	case domain.ScheduleTypeCron:
		job, err = runner.scheduler.NewJob(
			gocron.CronJob(schedule.CronExpr, false),
			gocron.NewTask(func(id string) { runner.service.ExecuteSchedule(id) }, schedule.ID),
			gocron.WithIdentifier(uid),
		)
	case domain.ScheduleTypeOneShot:
		if schedule.RunAt == nil {
			return domain.ErrInvalidSchedule
		}
		job, err = runner.scheduler.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(*schedule.RunAt)),
			gocron.NewTask(func(id string) { runner.service.ExecuteSchedule(id) }, schedule.ID),
			gocron.WithIdentifier(uid),
		)
	default:
		return domain.ErrInvalidSchedule
	}

	if err != nil {
		return err
	}

	runner.jobsByID[schedule.ID] = job.ID()
	log.Printf("[CronRunner] Schedule %s (%s) registrado com job ID %s", schedule.ID, schedule.ScheduleType, job.ID())
	return nil
}

// Unregister remove um agendamento do scheduler interno.
func (runner *CronRunner) Unregister(scheduleID string) {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	runner.unregisterLocked(scheduleID)
}

// unregisterLocked remove um job sem adquirir o mutex (deve ser chamado com o mutex já adquirido).
func (runner *CronRunner) unregisterLocked(scheduleID string) {
	jobID, exists := runner.jobsByID[scheduleID]
	if !exists {
		return
	}
	if err := runner.scheduler.RemoveJob(jobID); err != nil {
		log.Printf("[CronRunner] Erro ao remover job do schedule %s: %v", scheduleID, err)
	}
	delete(runner.jobsByID, scheduleID)
	log.Printf("[CronRunner] Schedule %s removido do scheduler", scheduleID)
}

// LoadAndStart carrega todos os agendamentos habilitados, realiza recovery de execuções
// perdidas e inicia o scheduler.
func (runner *CronRunner) LoadAndStart(ctx context.Context) error {
	schedules, err := runner.service.LoadAllEnabled(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, s := range schedules {
		if err := runner.Register(s); err != nil {
			log.Printf("[CronRunner] Falha ao registrar schedule %s: %v", s.ID, err)
			continue
		}

		// Recovery de one-shot: executa se RunAt está dentro da grace period passada
		if s.ScheduleType == domain.ScheduleTypeOneShot &&
			s.RunAt != nil &&
			s.RunAt.Before(now) &&
			s.RunAt.After(now.Add(-runner.gracePeriod)) {
			log.Printf("[CronRunner] Recuperando execução perdida do schedule one-shot %s (RunAt: %s)", s.ID, s.RunAt)
			go runner.service.ExecuteSchedule(s.ID)
		}
	}

	runner.scheduler.Start()
	log.Printf("[CronRunner] Scheduler iniciado com %d agendamentos registrados", len(runner.jobsByID))
	return nil
}

// Stop encerra o scheduler de forma ordenada.
func (runner *CronRunner) Stop() {
	if err := runner.scheduler.Shutdown(); err != nil {
		log.Printf("[CronRunner] Erro ao encerrar scheduler: %v", err)
	}
	log.Println("[CronRunner] Scheduler encerrado")
}
