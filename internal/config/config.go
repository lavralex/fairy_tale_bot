package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken    string
	DatabaseURL string
	ServerPort  string
	Debug       bool
}

func Load() (*Config, error) {
	_ = godotenv.Load() // на проде файла нет - игнорируем ошибку намеренно
	isDebugText := os.Getenv("DEBUG")
	isDebug, err := strconv.ParseBool(isDebugText)
	if isDebugText == "" {
		isDebug = false
	} else if err != nil {
		return nil, errors.New("DEBUG value incorrect")
	}
	conf := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		Debug:       isDebug,
	}

	if conf.BotToken == "" {
		return nil, errors.New("Bot Token required")
	}
	if conf.DatabaseURL == "" {
		return nil, errors.New("Database URL required")
	}
	if conf.ServerPort == "" {
		return nil, errors.New("Server Port required")
	}

	return conf, nil
}
