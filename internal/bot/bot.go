package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/models"
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

	for update := range updates {
		message := update.Message
		if message != nil {
			log.Printf("[%s] %s", message.From.UserName, message.Text)
			msg := tgbotapi.NewMessage(message.Chat.ID, "")
			if message.IsCommand() {
				switch message.Command() {
				case "start":
					telegramUser := message.From
					err = store.CreateUser(
						ctx,
						models.User{
							TelegramID: telegramUser.ID,
							Username:   &telegramUser.UserName,
						},
					)
					if err != nil {
						msg.Text = "Произошла ошибка регистрации"
						bot.Send(msg)
						log.Printf("CreateUser error: %v", err)
						continue
					}
					dbUser, err := store.GetUserByTelegramID(ctx, telegramUser.ID)
					if errors.Is(err, storage.ErrUserNotFound) {
						msg.Text = "Пользователь не найден"
						bot.Send(msg)
						log.Printf("GetUserByTelegramID error: %v", err)
						continue
					}
					if err != nil {
						msg.Text = "Произошла ошибка поиска пользователя"
						bot.Send(msg)
						log.Printf("GetUserByTelegramID error: %v", err)
						continue
					}
					msg.Text = fmt.Sprintf("Здравствуйте, %s", *dbUser.Username)
					webApp := tgbotapi.WebAppInfo{URL: "https://example.com"}
					msg.ReplyMarkup = getStartKeyboard("Вход в магазин", webApp)
				default:
					msg.Text = "Команда не найдена"
				}
			} else {
				msg.Text = message.Text
				msg.ReplyToMessageID = message.MessageID
			}
			bot.Send(msg)
		}
	}
	return nil
}
