package main

import (
	"log"
	"os"

	"aurora/pkg/database"
	"aurora/services/auth-service/internal/application"
	"aurora/services/auth-service/internal/infrastructure/http"
	"aurora/services/auth-service/internal/infrastructure/migrations"
	"aurora/services/auth-service/internal/infrastructure/repository"
	"aurora/services/auth-service/internal/infrastructure/security"
)

func main() {
	// Configurações
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-dev-secret-key"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	// Configurações do PostgreSQL
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "aurora"),
		Password: getEnv("DB_PASSWORD", "aurora_secret"),
		DBName:   getEnv("DB_NAME", "aurora_home"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Conectar ao PostgreSQL
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Executar migrations
	if err := migrations.RunAuthMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Infraestrutura
	userRepo := repository.NewPostgresUserRepository(db)
	jwtManager := security.NewJWTManager(jwtSecret)

	// Application Layer
	authService := application.NewAuthService(userRepo, jwtManager)

	// HTTP Server
	server := http.NewServer(authService, serverPort)

	log.Printf("Auth Service starting on port %s...", serverPort)
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
