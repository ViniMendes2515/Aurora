package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"aurora/services/notifications-service/internal/domain"
)

var notificationTemplates = map[string]string{
	string(domain.TypeMotionDetected): "🚨 <b>Movimento detectado</b> em %s",
	string(domain.TypeLightOn):        "💡 Luz <b>ligada</b> em %s",
	string(domain.TypeLightOff):       "💡 Luz <b>desligada</b> em %s",
	string(domain.TypeAlarmTriggered): "🔔 <b>Alarme acionado</b> em %s",
}

// TelegramService e a interface que o bot usa para manipular preferencias
type TelegramService interface {
	LinkAccount(token string, chatID int64) error
	UnlinkByChatID(chatID int64) error
}

// Bot encapsula o bot Telegram do Aurora
type Bot struct {
	bot     *tgbot.Bot
	service TelegramService
}

// NewBot cria e configura o bot Telegram
func NewBot(token string, service TelegramService) (*Bot, error) {
	b := &Bot{service: service}

	opts := []tgbot.Option{
		tgbot.WithDefaultHandler(b.defaultHandler),
	}

	tb, err := tgbot.New(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar bot Telegram: %w", err)
	}
	b.bot = tb

	tb.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypePrefix, b.handleStart)
	tb.RegisterHandler(tgbot.HandlerTypeMessageText, "/stop", tgbot.MatchTypeExact, b.handleStop)
	tb.RegisterHandler(tgbot.HandlerTypeMessageText, "/ajuda", tgbot.MatchTypeExact, b.handleHelp)

	return b, nil
}

// Start inicia o bot
func (b *Bot) Start(ctx context.Context) {
	log.Println("Telegram bot iniciado")
	b.bot.Start(ctx)
}

// SendNotification envia uma mensagem formatada ao chat especificado
func (b *Bot) SendNotification(ctx context.Context, chatID int64, notifType, location string) {
	tmpl, ok := notificationTemplates[notifType]
	if !ok {
		return
	}
	text := fmt.Sprintf(tmpl, location)
	_, err := b.bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Telegram: erro ao enviar mensagem para chat %d: %v", chatID, err)
	}
}

func (b *Bot) handleStart(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	parts := strings.Fields(update.Message.Text)

	if len(parts) < 2 {
		b.reply(ctx, tb, update, "👋 Ola! Envie <code>/start TOKEN</code> com o codigo gerado no app Aurora para vincular sua conta.")
		return
	}

	token := parts[1]
	chatID := update.Message.Chat.ID

	if err := b.service.LinkAccount(token, chatID); err != nil {
		switch err {
		case domain.ErrTelegramTokenExpired:
			b.reply(ctx, tb, update, "❌ Token invalido ou expirado. Gere um novo token no app Aurora.")
		default:
			log.Printf("Telegram: erro ao vincular conta: %v", err)
			b.reply(ctx, tb, update, "❌ Erro ao vincular conta. Tente novamente.")
		}
		return
	}

	b.reply(ctx, tb, update, "✅ <b>Conta vinculada com sucesso.</b>\n\nVoce receberá notificações da Aurora aqui.\nUse /ajuda para ver os comandos disponíveis.")
}

func (b *Bot) handleStop(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID
	if err := b.service.UnlinkByChatID(chatID); err != nil {
		b.reply(ctx, tb, update, "❌ Nenhuma conta vinculada encontrada.")
		return
	}
	b.reply(ctx, tb, update, "🔕 Notificacoes desativadas. Use /start com um novo token para reativar.")
}

func (b *Bot) handleHelp(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	msg := "🏠 <b>Aurora Home Bot</b>\n\n" +
		"Comandos disponíveis:\n" +
		"• <code>/start TOKEN</code> — vincular conta Aurora\n" +
		"• <code>/stop</code> — desativar notificacoes\n" +
		"• <code>/ajuda</code> — mostrar esta mensagem\n\n" +
		"Para gerenciar alertas, acesse o app Aurora."
	b.reply(ctx, tb, update, msg)
}

func (b *Bot) defaultHandler(ctx context.Context, tb *tgbot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	b.reply(ctx, tb, update, "Nao entendi. Use /ajuda para ver os comandos disponíveis.")
}

func (b *Bot) reply(ctx context.Context, tb *tgbot.Bot, update *models.Update, text string) {
	_, err := tb.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Telegram: erro ao responder: %v", err)
	}
}
