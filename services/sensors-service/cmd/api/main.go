package main

import (
	"log"
	"os"
	"time"

	"aurora/pkg/database"
	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/infrastructure/http"
	"aurora/services/sensors-service/internal/infrastructure/messaging"
	"aurora/services/sensors-service/internal/infrastructure/migrations"
	"aurora/services/sensors-service/internal/infrastructure/repository"
	"aurora/services/sensors-service/internal/infrastructure/security"
	"aurora/services/sensors-service/internal/infrastructure/ws"
)

func main() {
	jwtSecret := getEnv("JWT_SECRET", "")
	serverPort := getEnv("SERVER_PORT", "8081")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	deviceAPIKey := getEnv("DEVICE_API_KEY", "")

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

	if err := migrations.RunSensorMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := migrations.SeedDemoSensors(db); err != nil {
		log.Fatalf("Failed to seed demo sensors: %v", err)
	}

	var natsConn *messaging.NATSConnection
	for i := 0; i < 10; i++ {
		natsConn, err = messaging.NewNATSConnection(natsURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for NATS... attempt %d/10", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	hub := ws.NewHub()
	sensorRepo := repository.NewPostgresSensorRepository(db)
	eventPublisher := messaging.NewNATSPublisher(natsConn, hub)
	jwtValidator := security.NewJWTValidator(jwtSecret)

	motionService := application.NewMotionService(sensorRepo, eventPublisher)
	lightService := application.NewLightService(sensorRepo, eventPublisher)

	server := http.NewServer(motionService, lightService, sensorRepo, jwtValidator, deviceAPIKey, hub, serverPort)

	log.Printf("Sensors Service starting on port %s...", serverPort)
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
