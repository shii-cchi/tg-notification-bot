package config

import (
	"errors"
	"os"
)

type Config struct {
	BotToken     string
	RabbitMQURL  string
	QueueConfigs map[string]int
}

func LoadConfig() (*Config, error) {
	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		return nil, errors.New("BOT_TOKEN parameter is not defined")
	}

	rabbitMQURL := os.Getenv("RABBIT_MQ_URL")

	if rabbitMQURL == "" {
		return nil, errors.New("RABBIT_MQ_URL parameter is not defined")
	}

	queueConfigs := map[string]int{
		"queue_1min":  1 * 60 * 1000,
		"queue_5min":  5 * 60 * 1000,
		"queue_10min": 10 * 60 * 1000,
		"queue_30min": 30 * 60 * 1000,
		"queue_1hour": 60 * 60 * 1000,
		"queue_2hour": 120 * 60 * 1000,
		"queue_3hour": 180 * 60 * 1000,
	}

	return &Config{
		BotToken:     botToken,
		RabbitMQURL:  rabbitMQURL,
		QueueConfigs: queueConfigs,
	}, nil
}
