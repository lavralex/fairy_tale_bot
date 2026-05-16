package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/lavralex/fairy_tale_bot/internal/bot"
	"github.com/lavralex/fairy_tale_bot/internal/config"
	"github.com/lavralex/fairy_tale_bot/internal/storage"
	"golang.org/x/sync/errgroup"
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
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return bot.Run(ctx, conf, store) })
	if err := g.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Bot stopped")
		} else {
			log.Fatalln(err)
		}
	}
}
