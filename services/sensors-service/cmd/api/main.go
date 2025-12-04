package main

import (
	"log"
	"os"
	"time"

	"aurora/services/sensors-service/internal/application"
	"aurora/services/sensors-service/internal/infrastructure/http"
	"aurora/services/sensors-service/internal/infrastructure/messaging"
	"aurora/services/sensors-service/internal/infrastructure/repository"
	"aurora/services/sensors-service/internal/infrastructure/security"
)

func main() {
	// Configurações
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-dev-secret-key"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8081"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	// Aguardar NATS estar disponível
	var natsConn *messaging.NATSConnection
	var err error
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

	// Infraestrutura
	sensorRepo := repository.NewMemorySensorRepository()
	eventPublisher := messaging.NewNATSPublisher(natsConn)
	jwtValidator := security.NewJWTValidator(jwtSecret)

	// Application Layer
	motionService := application.NewMotionService(sensorRepo, eventPublisher)

	// HTTP Server
	server := http.NewServer(motionService, jwtValidator, serverPort)

	log.Printf("Sensors Service starting on port %s...", serverPort)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
