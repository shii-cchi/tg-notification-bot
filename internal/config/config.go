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
	RabbitMQDefaultUser string
	RabbitMQDefaultPass string
	RabbitMQHost string
	RabbitMQPort string
	QueueTTLs   []int
	DbUser      string
	DbPassword  string
	DbHost      string
	DbPort      string
	DbName      string
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

	rabbitMQDefaultUser := os.Getenv("RABBITMQ_DEFAULT_USER")

	if rabbitMQDefaultUser == "" {
		return nil, errors.New("RABBITMQ_DEFAULT_USER parameter is not defined")
	}

	rabbitMQDefaultPass := os.Getenv("RABBITMQ_DEFAULT_PASS")

	if rabbitMQDefaultPass == "" {
		return nil, errors.New("RABBITMQ_DEFAULT_PASS parameter is not defined")
	}

	rabbitMQHost := os.Getenv("RABBITMQ_HOST")

	if rabbitMQHost == "" {
		return nil, errors.New("RABBITMQ_HOST parameter is not defined")
	}

	rabbitMQPort := os.Getenv("RABBITMQ_PORT")

	if rabbitMQPort == "" {
		return nil, errors.New("RABBITMQ_PORT parameter is not defined")
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

	dbUser := os.Getenv("DB_USER")

	if dbUser == "" {
		return nil, errors.New("DB_USER parameter is not defined")
	}

	dbPassword := os.Getenv("DB_PASSWORD")

	if dbPassword == "" {
		return nil, errors.New("DB_PASSWORD parameter is not defined")
	}

	dbHost := os.Getenv("DB_HOST")

	if dbHost == "" {
		return nil, errors.New("DB_HOST parameter is not defined")
	}

	dbPort := os.Getenv("DB_PORT")

	if dbPort == "" {
		return nil, errors.New("DB_PORT parameter is not defined")
	}

	dbName := os.Getenv("DB_NAME")

	if dbName == "" {
		return nil, errors.New("DB_NAME parameter is not defined")
	}

	return &Config{
		BotToken:    botToken,
		RabbitMQDefaultUser: rabbitMQDefaultUser,
		RabbitMQDefaultPass: rabbitMQDefaultPass,
		RabbitMQHost: rabbitMQHost,
		RabbitMQPort: rabbitMQPort,
		QueueTTLs:   queueTTLs,
		DbUser:      dbUser,
		DbPassword:  dbPassword,
		DbHost:      dbHost,
		DbPort:      dbPort,
		DbName:      dbName,
	}, nil
}
