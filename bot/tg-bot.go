package tgbot

import (
	"context"
	"fmt"

	tgbotapi "github.com/Syfaro/telegram-bot-api"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

type SessionProvider interface {
	RegisterSession(ctx context.Context, userName string, chatID int64) (uuid.UUID, error)
	RegisterPinCode(ctx context.Context, sessionID uuid.UUID, pin string) error
}

type TgBot struct {
	logger   *logrus.Logger
	api      *tgbotapi.BotAPI
	config   tgbotapi.UpdateConfig
	provider SessionProvider
}

func NewTgBot(log *logrus.Logger, token string, provider SessionProvider) (*TgBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Telegram bot API: %v", err)
	}
	config := tgbotapi.NewUpdate(0)
	config.Timeout = 60
	return &TgBot{
		logger:   log,
		api:      bot,
		config:   config,
		provider: provider,
	}, nil
}

func (bot *TgBot) Serve(ctx context.Context) error {
	updates, err := bot.api.GetUpdatesChan(bot.config)
	if err != nil {
		return fmt.Errorf("unable to get Telegram API updates channel: %v", err)
	}
	for upd := range updates {
		sessionID, err := bot.provider.RegisterSession(ctx, upd.Message.From.UserName, upd.Message.Chat.ID)
		if err != nil {
			return fmt.Errorf("failed chat session registration: %v", err)
		}
		bot.logger.Info("Register session, ID: ", sessionID)

		if err := bot.provider.RegisterPinCode(ctx, sessionID, upd.Message.Text); err != nil {
			return fmt.Errorf("failed pin code registration: %v", err)
		}
	}
	return nil
}
