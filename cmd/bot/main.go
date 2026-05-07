package main

import (
	"log"

	"github.com/lavralex/fairy_tale_bot/internal/config"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Config loaded successfully")
	_ = conf // временная заглушка
}
