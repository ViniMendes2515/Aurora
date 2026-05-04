package main

import (
	"log"
	"os"

	"aurora/pkg/database"
	"aurora/services/lighting-service/internal/application"
	"aurora/services/lighting-service/internal/infrastructure/device"
	httpserver "aurora/services/lighting-service/internal/infrastructure/http"
	"aurora/services/lighting-service/internal/infrastructure/migrations"
	"aurora/services/lighting-service/internal/infrastructure/repository"
	"aurora/services/lighting-service/internal/infrastructure/security"
	"aurora/services/lighting-service/internal/infrastructure/ws"

	_ "aurora/services/lighting-service/docs"
)

// @title           Lighting Service API
// @version         1.0
// @description     Serviço de controle de iluminação da plataforma Aurora Home.

// @host      localhost:8082
// @BasePath  /

func main() {
	jwtSecret := getEnv("JWT_SECRET", "")
	serverPort := getEnv("SERVER_PORT", "8082")
	deviceAPIKey := getEnv("DEVICE_API_KEY", "")
	esp32IP := getEnv("ESP32_IP", "")

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

	if err := migrations.RunLightMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := migrations.SeedDemoLights(db); err != nil {
		log.Fatalf("Failed to seed demo lights: %v", err)
	}

	lightRepo := repository.NewPostgresLightRepository(db)
	esp32Client := device.NewESP32Client(esp32IP, deviceAPIKey)
	jwtValidator := security.NewJWTValidator(jwtSecret)

	hub := ws.NewHub()
	lightService := application.NewLightService(lightRepo, esp32Client, hub)

	debug := os.Getenv("DEBUG") == "true"
	server := httpserver.NewServer(lightService, esp32Client, jwtValidator, deviceAPIKey, hub, serverPort, debug)

	log.Printf("Lighting Service starting on port %s...", serverPort)
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
