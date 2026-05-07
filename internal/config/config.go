package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)


type Config struct{
	BotToken string
	DatabaseURL string
	ServerPort string
}

func Load() (*Config, error) {
	godotenv.Load() // на проде файла нет - игнорируем ошибку намеренно

	conf := &Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServerPort: os.Getenv("SERVER_PORT"),
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