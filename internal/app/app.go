package app

import (
	"log"
	"tg-notification-bot/internal/config"
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

	rabbit, err := rabbitmq.InitRabbit(cfg.RabbitMQURL, cfg.QueueConfigs)

	if err != nil {
		log.Fatalf("error init rabbitmq: %s\n", err)
	}

	defer rabbit.Close()

	messageService := service.NewMessageService(rabbit)

	messageHandler, err := message_handler.NewMessageHandler(cfg.BotToken, messageService)

	if err != nil {
		log.Fatal("error creating bot")
	}

	log.Println("initialization was a complete success")
	log.Println("starting to handle message")

	go messageHandler.Notify()

	messageHandler.HandleMessage()
}
