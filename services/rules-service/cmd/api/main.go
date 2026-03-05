package main

import (
	"log"
	"os"
	"time"

	"aurora/pkg/database"
	"aurora/services/rules-service/internal/application"
	httpserver "aurora/services/rules-service/internal/infrastructure/http"
	"aurora/services/rules-service/internal/infrastructure/messaging"
	"aurora/services/rules-service/internal/infrastructure/migrations"
	"aurora/services/rules-service/internal/infrastructure/repository"
	"aurora/services/rules-service/internal/infrastructure/security"
)

func main() {
	jwtSecret := getEnv("JWT_SECRET", "default-dev-secret-key")
	serverPort := getEnv("SERVER_PORT", "8084")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	lightingURL := getEnv("LIGHTING_SERVICE_URL", "http://lighting-service:8082")
	securityURL := getEnv("SECURITY_SERVICE_URL", "http://security-service:8083")
	deviceAPIKey := getEnv("DEVICE_API_KEY", "aurora-device-key-2024")
	motionTimeoutSeconds := getEnv("MOTION_OFF_TIMEOUT_SECONDS", "300")

	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "aurora"),
		Password: getEnv("DB_PASSWORD", "aurora_secret"),
		DBName:   getEnv("DB_NAME", "aurora_home"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunRuleMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := migrations.SeedDemoRules(db); err != nil {
		log.Fatalf("Failed to seed demo rules: %v", err)
	}

	ruleRepo := repository.NewPostgresRuleRepository(db)
	jwtValidator := security.NewJWTValidator(jwtSecret)
	timeout, err := time.ParseDuration(motionTimeoutSeconds + "s")
	if err != nil {
		timeout = 5 * time.Minute
	}
	rulesEngine := application.NewRulesEngine(ruleRepo, lightingURL, securityURL, deviceAPIKey, timeout)

	subscriber, err := messaging.NewNATSSubscriber(natsURL, rulesEngine)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer subscriber.Close()

	if err := subscriber.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe to NATS: %v", err)
	}

	server := httpserver.NewServer(rulesEngine, jwtValidator, serverPort)

	log.Printf("Rules Service starting on port %s...", serverPort)
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
