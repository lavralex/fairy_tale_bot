package bot

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

func Run(ctx context.Context, conf *config.Config, store *storage.Storage) error {
	bot, err := tgbotapi.NewBotAPI(conf.BotToken)
	if err != nil {
		return err
	}

	bot.Debug = conf.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	hs := newHandler(bot, store, conf)

	for update := range updates {
		message := update.Message
		if message != nil {
			log.Printf("[%s] %s", message.From.UserName, message.Text)
			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			if message.IsCommand() {
				switch message.Command() {
				case "start":
					hs.handleStart(ctx, message)
				default:
					msg.Text = "Команда не найдена"
					bot.Send(msg)
				}
			}
		}
	}
	return nil
}
