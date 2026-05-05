package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	_ "time/tzdata"

	"aurora/pkg/database"
	"aurora/services/schedule-service/internal/application"
	"aurora/services/schedule-service/internal/infrastructure/dispatcher"
	httpinfra "aurora/services/schedule-service/internal/infrastructure/http"
	"aurora/services/schedule-service/internal/infrastructure/messaging"
	"aurora/services/schedule-service/internal/infrastructure/migrations"
	"aurora/services/schedule-service/internal/infrastructure/repository"
	"aurora/services/schedule-service/internal/infrastructure/scheduler"
	"aurora/services/schedule-service/internal/infrastructure/security"

	_ "aurora/services/schedule-service/docs"
)

// @title           Schedule Service API
// @version         1.0
// @description     Serviço de agendamentos da plataforma Aurora Home.

// @host      localhost:8086
// @BasePath  /

func main() {
	jwtSecret := getEnv("JWT_SECRET", "")
	serverPort := getEnv("SERVER_PORT", "8086")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	lightingURL := getEnv("LIGHTING_SERVICE_URL", "http://lighting-service:8082")
	securityURL := getEnv("SECURITY_SERVICE_URL", "http://security-service:8083")
	deviceAPIKey := getEnv("DEVICE_API_KEY", "")
	gracePeriodStr := getEnv("MISSED_EXECUTION_GRACE_SECONDS", "300")
	dispatchTimeoutStr := getEnv("DISPATCH_TIMEOUT_SECONDS", "30")

	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "aurora"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "aurora_home"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Conecta ao PostgreSQL
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Executa migrations
	if err := migrations.RunScheduleMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Repositórios
	scheduleRepo := repository.NewPostgresScheduleRepository(db)
	executionRepo := repository.NewPostgresScheduleExecutionRepository(db)

	// NATS Publisher
	natsPublisher, err := messaging.NewNATSPublisher(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsPublisher.Close()

	// HTTP Action Client e Composite Dispatcher
	httpClient := httpinfra.NewHTTPActionClient(lightingURL, securityURL, deviceAPIKey)
	compositeDispatcher := dispatcher.NewCompositeDispatcher(natsPublisher, httpClient)

	// Parse de timeouts
	dispatchTimeoutSecs, err := strconv.Atoi(dispatchTimeoutStr)
	if err != nil {
		dispatchTimeoutSecs = 30
	}
	dispatchTimeout := time.Duration(dispatchTimeoutSecs) * time.Second

	gracePeriodSecs, err := strconv.Atoi(gracePeriodStr)
	if err != nil {
		gracePeriodSecs = 300
	}
	gracePeriod := time.Duration(gracePeriodSecs) * time.Second

	// Cria ScheduleService com registry nil — será definido após criar o CronRunner
	scheduleService := application.NewScheduleService(scheduleRepo, executionRepo, compositeDispatcher, nil, dispatchTimeout)

	// Cria CronRunner com referência ao service
	cronRunner, err := scheduler.NewCronRunner(scheduleService, gracePeriod)
	if err != nil {
		log.Fatalf("Failed to create cron runner: %v", err)
	}

	// Resolve dependência circular: define o registry no service
	scheduleService.SetRegistry(cronRunner)

	// Carrega agendamentos existentes e inicia o scheduler
	ctx := context.Background()
	if err := cronRunner.LoadAndStart(ctx); err != nil {
		log.Fatalf("Failed to load and start cron runner: %v", err)
	}

	// Configuração do JWT validator
	jwtValidator := security.NewJWTValidator(jwtSecret)

	// Tratamento de sinais para shutdown graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-quit
		log.Printf("Schedule Service received signal %s, shutting down...", sig)
		cronRunner.Stop()
		db.Close()
		natsPublisher.Close()
		os.Exit(0)
	}()

	debug := os.Getenv("DEBUG") == "true"

	// Inicia o servidor HTTP
	server := httpinfra.NewServer(scheduleService, jwtValidator, serverPort, debug)
	log.Printf("Schedule Service starting on port %s...", serverPort)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
