package message_handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand"
	"tg-notification-bot/internal/service"
	"time"
)

type MessageHandler struct {
	Bot            *tgbotapi.BotAPI
	UpdatesChannel tgbotapi.UpdatesChannel
	MessageService *service.MessageService
}

func NewMessageHandler(token string, messageService *service.MessageService) (*MessageHandler, error) {
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start using notification message_handler"},
		{Command: "add_task", Description: "Add task in queue"},
	}

	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)

	_, err = bot.Request(setCommandsConfig)

	if err != nil {
		return nil, err
	}

	updatesChan := bot.GetUpdatesChan(u)

	return &MessageHandler{
		Bot:            bot,
		UpdatesChannel: updatesChan,
		MessageService: messageService,
	}, nil
}

func (mh *MessageHandler) HandleMessage() {
	isAddingTask := false

	for update := range mh.UpdatesChannel {
		if update.Message == nil {
			continue
		}

		rng := rand.New(rand.NewSource(time.Now().UnixNano()))

		var msg tgbotapi.MessageConfig

		switch update.Message.Text {
		case "/start":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, startMessages[rng.Intn(len(startMessages))])
			sticker := tgbotapi.NewSticker(update.Message.Chat.ID, tgbotapi.FileID(startSticker))

			mh.Bot.Send(msg)
			mh.Bot.Send(sticker)

		case "/add_task":
			isAddingTask = true
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, addTaskMessage)
			mh.Bot.Send(msg)

		default:
			if isAddingTask {
				log.Printf("starting adding task - %s in queue\n", update.Message.Text)
				err := mh.MessageService.AddTask(update.Message.Text, update.Message.Chat.ID)

				if err != nil {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, errMessage)
					mh.Bot.Send(msg)
				} else {
					log.Printf("task - %s has been added in queue\n", update.Message.Text)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, successAddMessage+update.Message.Text)
					sticker := tgbotapi.NewSticker(update.Message.Chat.ID, tgbotapi.FileID(successAddSticker))
					isAddingTask = false
					mh.Bot.Send(msg)
					mh.Bot.Send(sticker)
				}

			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, noCommandMessage)
				sticker := tgbotapi.NewSticker(update.Message.Chat.ID, tgbotapi.FileID(noCommandSticker))
				mh.Bot.Send(msg)
				mh.Bot.Send(sticker)
			}
		}
	}
}

func (mh *MessageHandler) Notify() {
	for {
		log.Println("waiting notification")

		notification := mh.MessageService.GetNotification()

		log.Printf("getting notification %s\n", notification.Task)

		if notification.Task != "" {
			log.Printf("sending notification %s\n", notification.Task)

			msg := tgbotapi.NewMessage(notification.ChatID, notificationMessage+notification.Task)
			mh.Bot.Send(msg)
		}
	}
}
