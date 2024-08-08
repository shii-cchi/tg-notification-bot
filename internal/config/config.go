package config

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken    string
	RabbitMQURL string
	QueueTTLs   []int
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load(".env")

	if err != nil {
		return nil, err
	}

	botToken := os.Getenv("BOT_TOKEN")

	if botToken == "" {
		return nil, errors.New("BOT_TOKEN parameter is not defined")
	}

	rabbitMQURL := os.Getenv("RABBIT_MQ_URL")

	if rabbitMQURL == "" {
		return nil, errors.New("RABBIT_MQ_URL parameter is not defined")
	}

	queueTTLsStr := os.Getenv("QUEUE_TTLS")

	if queueTTLsStr == "" {
		return nil, errors.New("QUEUE_TTLS parameter is not defined")
	}

	ttlsStr := strings.Split(queueTTLsStr, ",")
	var queueTTLs []int
	for _, ttlStr := range ttlsStr {
		ttl, err := strconv.Atoi(ttlStr)
		if err != nil {
			return nil, errors.New("error parsing QUEUE_TTLS params")
		}
		queueTTLs = append(queueTTLs, ttl)
	}

	return &Config{
		BotToken:    botToken,
		RabbitMQURL: rabbitMQURL,
		QueueTTLs:   queueTTLs,
	}, nil
}
