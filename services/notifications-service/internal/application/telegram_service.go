package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"aurora/services/notifications-service/internal/domain"
)

// TelegramBot e a interface que o servico usa para enviar mensagens
type TelegramBot interface {
	SendNotification(ctx context.Context, chatID int64, notifType, location string)
}

// TelegramService contem a logica de negocio das preferencias Telegram
type TelegramService struct {
	repo domain.TelegramPreferenceRepository
	bot  TelegramBot
}

// NewTelegramService cria uma nova instancia do servico
func NewTelegramService(repo domain.TelegramPreferenceRepository) *TelegramService {
	return &TelegramService{repo: repo}
}

// SetBot define o bot apos a inicializacao (necessario para quebrar dependencia circular)
func (s *TelegramService) SetBot(bot TelegramBot) {
	s.bot = bot
}

// GenerateLinkToken gera um token temporario (TTL 15min) para o usuario vincular o Telegram
func (s *TelegramService) GenerateLinkToken(userID string) (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("erro ao gerar token: %w", err)
	}
	token := "AURORA-" + strings.ToUpper(hex.EncodeToString(b))

	linkToken := &domain.TelegramLinkToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.repo.SaveLinkToken(linkToken); err != nil {
		return "", err
	}
	return token, nil
}

// LinkAccount valida o token e vincula o chat_id ao usuario
func (s *TelegramService) LinkAccount(token string, chatID int64) error {
	linkToken, err := s.repo.FindLinkToken(token)
	if err != nil {
		return err
	}

	pref := &domain.TelegramPreference{
		UserID:       linkToken.UserID,
		ChatID:       chatID,
		EnabledTypes: domain.DefaultEnabledTypes(),
		Active:       true,
		LinkedAt:     time.Now(),
	}

	if err := s.repo.Save(pref); err != nil {
		return err
	}

	return s.repo.DeleteLinkToken(token)
}

// UnlinkByChatID remove a vinculacao pelo chat_id (chamado pelo /stop do bot)
func (s *TelegramService) UnlinkByChatID(chatID int64) error {
	pref, err := s.repo.FindByChatID(chatID)
	if err != nil {
		return err
	}
	return s.repo.Delete(pref.UserID)
}

// UnlinkByUserID remove a vinculacao pelo user_id (chamado pela API REST)
func (s *TelegramService) UnlinkByUserID(userID string) error {
	return s.repo.Delete(userID)
}

// GetPreferences retorna as preferencias Telegram do usuario
func (s *TelegramService) GetPreferences(userID string) (*domain.TelegramPreference, error) {
	return s.repo.FindByUserID(userID)
}

// UpdatePreferences atualiza os tipos de notificacao habilitados para o usuario
func (s *TelegramService) UpdatePreferences(userID string, types []string) error {
	return s.repo.UpdateTypes(userID, types)
}

// SendIfEnabled envia notificacao Telegram se o usuario tiver o tipo habilitado
func (s *TelegramService) SendIfEnabled(ctx context.Context, userID, notifType, location string) {
	if userID == "*" {
		return
	}

	pref, err := s.repo.FindByUserID(userID)
	if err != nil {
		fmt.Printf("[Telegram] FindByUserID erro userID=%s: %v\n", userID, err)
		return
	}
	if !pref.Active {
		fmt.Printf("[Telegram] preferencia inativa userID=%s\n", userID)
		return
	}

	fmt.Printf("[Telegram] enabled_types=%v notifType=%s enabled=%v\n", pref.EnabledTypes, notifType, pref.IsTypeEnabled(notifType))
	if pref.IsTypeEnabled(notifType) {
		s.bot.SendNotification(ctx, pref.ChatID, notifType, location)
	}
}
