package main

import (
	"context"
	"log"

	"github.com/lavralex/fairy_tale_bot/internal/bot"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
)

func main() {
	ctx := context.Background()
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
}
