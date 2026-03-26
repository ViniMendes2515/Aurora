package main

import (
	"log"
	"os"

	"aurora/pkg/database"
	"aurora/services/notifications-service/internal/application"
	httpserver "aurora/services/notifications-service/internal/infrastructure/http"
	"aurora/services/notifications-service/internal/infrastructure/messaging"
	"aurora/services/notifications-service/internal/infrastructure/migrations"
	"aurora/services/notifications-service/internal/infrastructure/repository"
	"aurora/services/notifications-service/internal/infrastructure/security"
)

func main() {
	jwtSecret := getEnv("JWT_SECRET", "")
	serverPort := getEnv("SERVER_PORT", "8085")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

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
		log.Fatalf("Falha ao conectar ao banco: %v", err)
	}
	defer db.Close()

	if err := migrations.RunNotificationMigrations(db); err != nil {
		log.Fatalf("Falha ao executar migrations: %v", err)
	}

	notifRepo := repository.NewPostgresNotificationRepository(db)
	notifService := application.NewNotificationService(notifRepo)
	jwtValidator := security.NewJWTValidator(jwtSecret)

	subscriber, err := messaging.NewNATSSubscriber(natsURL, notifService)
	if err != nil {
		log.Fatalf("Falha ao conectar ao NATS: %v", err)
	}
	defer subscriber.Close()

	if err := subscriber.Subscribe(); err != nil {
		log.Fatalf("Falha ao registrar subscricoes NATS: %v", err)
	}

	server := httpserver.NewServer(notifService, jwtValidator, serverPort)
	log.Printf("Notifications Service iniciando na porta %s...", serverPort)
	if err := server.Start(); err != nil {
		log.Fatalf("Falha ao iniciar servidor: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
