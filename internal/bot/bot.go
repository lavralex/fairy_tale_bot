package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

type Bot struct {
	conf  *config.Config
	store *storage.Storage
	api   *tgbotapi.BotAPI
}

func New(conf *config.Config, store *storage.Storage) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(conf.BotToken)
	if err != nil {
		return nil, err
	}
	return &Bot{
		conf:  conf,
		store: store,
		api:   api,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	b.api.Debug = b.conf.Debug

	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			message := update.Message
			if message != nil {
				log.Printf("[%s] %s", message.From.UserName, message.Text)
				msg := tgbotapi.NewMessage(message.Chat.ID, "")
				if message.IsCommand() {
					switch message.Command() {
					case "start":
						b.handleStart(ctx, message)
					default:
						msg.Text = "Команда не найдена"
						b.api.Send(msg)
					}
				}
			}
		}
	}
}
