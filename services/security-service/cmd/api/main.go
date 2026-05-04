package main

import (
	"log"
	"os"
	"strconv"

	"aurora/pkg/database"
	"aurora/services/security-service/internal/application"
	"aurora/services/security-service/internal/infrastructure/device"
	httpserver "aurora/services/security-service/internal/infrastructure/http"
	"aurora/services/security-service/internal/infrastructure/messaging"
	"aurora/services/security-service/internal/infrastructure/migrations"
	"aurora/services/security-service/internal/infrastructure/repository"
	"aurora/services/security-service/internal/infrastructure/security"

	_ "aurora/services/security-service/docs"
)

// @title           Security Service API
// @version         1.0
// @description     Serviço de segurança e alarmes da plataforma Aurora Home.

// @host      localhost:8083
// @BasePath  /

func main() {
	jwtSecret := getEnv("JWT_SECRET", "")
	serverPort := getEnv("SERVER_PORT", "8083")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	deviceAPIKey := getEnv("DEVICE_API_KEY", "")
	esp32IP := getEnv("ESP32_IP", "")
	deviceID := getEnv("DEVICE_ID", "esp32-main")
	buzzerDurationMs, _ := strconv.Atoi(getEnv("BUZZER_DURATION_MS", "3000"))
	autoAlarmOnMotion := getEnv("AUTO_ALARM_ON_MOTION", "false") == "true"

	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "aurora"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "aurora_home"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunAlarmMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	alarmRepo := repository.NewPostgresAlarmRepository(db)
	buzzerClient := device.NewBuzzerClient(esp32IP, deviceAPIKey)
	jwtValidator := security.NewJWTValidator(jwtSecret)

	alarmService := application.NewAlarmService(alarmRepo, buzzerClient, buzzerDurationMs, deviceID)

	// Subscriber NATS (recebe eventos de movimento e dispara alarme)
	subscriber, err := messaging.NewNATSSubscriber(natsURL, alarmService, autoAlarmOnMotion)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer subscriber.Close()

	if err := subscriber.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}

	debug := os.Getenv("DEBUG") == "true"
	server := httpserver.NewServer(alarmService, jwtValidator, deviceAPIKey, serverPort, debug)

	log.Printf("Security Service starting on port %s...", serverPort)
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
