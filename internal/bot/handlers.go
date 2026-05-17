package bot

import (
	"context"
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lavralex/fairy_tale_bot/internal/models"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

func (b *Bot) handleStart(ctx context.Context, message *tgbotapi.Message) {
	telegramUser := message.From
	err := b.store.CreateUser(
		ctx,
		models.User{
			TelegramID: telegramUser.ID,
			Username:   &telegramUser.UserName,
		},
	)
	response := tgbotapi.NewMessage(message.Chat.ID, "")
	if err != nil {
		response.Text = "Произошла ошибка регистрации"
		b.api.Send(response)
		log.Printf("CreateUser error: %v", err)
		return
	}
	dbUser, err := b.store.GetUserByTelegramID(ctx, telegramUser.ID)
	if errors.Is(err, storage.ErrUserNotFound) {
		response.Text = "Пользователь не найден"
		b.api.Send(response)
		log.Printf("GetUserByTelegramID error: %v", err)
		return
	}
	if err != nil {
		response.Text = "Произошла ошибка поиска пользователя"
		b.api.Send(response)
		log.Printf("GetUserByTelegramID error: %v", err)
		return
	}
	response.Text = fmt.Sprintf("Здравствуйте, %s", *dbUser.Username)
	webApp := tgbotapi.WebAppInfo{URL: "https://example.com"}
	response.ReplyMarkup = getStartKeyboard("Вход в магазин", webApp)
	b.api.Send(response)
}
