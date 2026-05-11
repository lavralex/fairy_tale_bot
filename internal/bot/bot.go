package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lavralex/fairy_tale_bot/internal/config"
)

func Run(conf *config.Config) error {
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
					msg.Text = "Приветствуем вас!"
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
