package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"github.com/lavralex/fairy_tale_bot/internal/api"
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
	srv := api.New(conf, store)
	b, err := bot.New(conf, store)
	if err != nil {
		log.Fatalln(err)
	}
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return b.Run(ctx) })
	g.Go(func() error { return srv.Run(ctx) })
	if err := g.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Shutting down")
		} else {
			log.Fatalln(err)
		}
	}
}
