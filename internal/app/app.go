package app

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"tg-notification-bot/internal/config"
	"tg-notification-bot/internal/database"
	"tg-notification-bot/internal/message_handler"
	"tg-notification-bot/internal/rabbitmq"
	"tg-notification-bot/internal/service"
)

func RunBot() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("error loading config: %s\n", err)
	}

	log.Println("config has been loaded successfully")

	rabbit, err := rabbitmq.InitRabbit(cfg.RabbitMQURL, cfg.QueueTTLs)

	if err != nil {
		log.Fatalf("error init rabbitmq: %s\n", err)
	}

	defer rabbit.Close()

	conn, err := sql.Open("postgres", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", cfg.DbUser, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbName))

	if err != nil {
		log.Fatalf("error connecting to db: %s\n", err)
	}

	queries := database.New(conn)

	messageService := service.NewMessageService(rabbit, queries)

	messageHandler, err := message_handler.NewMessageHandler(cfg.BotToken, messageService)

	if err != nil {
		log.Fatal("error creating bot")
	}

	log.Println("initialization was a complete success")
	log.Println("starting to handle message")

	go messageHandler.Notify()

	messageHandler.HandleMessage()
}
