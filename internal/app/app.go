package app

import (
	"log"
	"tg-notification-bot/internal/bot"
	"tg-notification-bot/internal/config"
)

func RunBot() {
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("error loading config: %s\n", err)
	}

	log.Println("config has been loaded successfully")

	notificationBot, updatesChan, err := bot.NewBot(cfg.BotToken)

	if err != nil {
		log.Fatal("error creating bot")
	}

	log.Println("bot has been created successfully")

	bot.HandleMessage(notificationBot, updatesChan)
}
