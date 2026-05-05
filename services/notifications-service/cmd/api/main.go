package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"aurora/pkg/database"
	"aurora/services/notifications-service/internal/application"
	httpserver "aurora/services/notifications-service/internal/infrastructure/http"
	"aurora/services/notifications-service/internal/infrastructure/messaging"
	"aurora/services/notifications-service/internal/infrastructure/migrations"
	"aurora/services/notifications-service/internal/infrastructure/repository"
	"aurora/services/notifications-service/internal/infrastructure/security"
	"aurora/services/notifications-service/internal/infrastructure/telegram"

	_ "aurora/services/notifications-service/docs"
)

// @title           Notifications Service API
// @version         1.0
// @description     Serviço de notificações da plataforma Aurora Home.

// @host      localhost:8085
// @BasePath  /

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	jwtSecret := getEnv("JWT_SECRET", "")
	serverPort := getEnv("SERVER_PORT", "8085")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	telegramToken := getEnv("TELEGRAM_BOT_TOKEN", "")

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
	telegramRepo := repository.NewPostgresTelegramPreferenceRepository(db)
	jwtValidator := security.NewJWTValidator(jwtSecret)

	// Inicializa TelegramService primeiro, depois injeta o bot (quebra dependencia circular)
	telegramSvc := application.NewTelegramService(telegramRepo)

	if telegramToken != "" {
		bot, err := telegram.NewBot(telegramToken, telegramSvc)
		if err != nil {
			log.Printf("Aviso: falha ao iniciar Telegram bot: %v", err)
		} else {
			telegramSvc.SetBot(bot)
			go bot.Start(ctx)
			log.Println("Telegram bot iniciado com sucesso")
		}
	} else {
		log.Println("TELEGRAM_BOT_TOKEN nao configurado — notificacoes Telegram desabilitadas")
	}

	notifService := application.NewNotificationService(notifRepo, telegramSvc)

	subscriber, err := messaging.NewNATSSubscriber(natsURL, notifService)
	if err != nil {
		log.Fatalf("Falha ao conectar ao NATS: %v", err)
	}
	defer subscriber.Close()

	if err := subscriber.Subscribe(); err != nil {
		log.Fatalf("Falha ao registrar subscricoes NATS: %v", err)
	}

	debug := os.Getenv("DEBUG") == "true"
	server := httpserver.NewServer(notifService, telegramSvc, jwtValidator, serverPort, debug)
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
