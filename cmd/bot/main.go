package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/lavralex/fairy_tale_bot/internal/bot"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()
	conf, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Config loaded successfully")
	store, err := storage.New(ctx, conf.DatabaseURL)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("DB connected successfully")
	err = bot.Run(ctx, conf, store)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Bot stopped")
}
