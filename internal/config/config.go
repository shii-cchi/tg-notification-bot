package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken string
}

func LoadConfig() (*Config, error) {
	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		return nil, errors.New("BOT_TOKEN parameter is not defined")
	}

	return &Config{
		BotToken: botToken,
	}, nil
}
