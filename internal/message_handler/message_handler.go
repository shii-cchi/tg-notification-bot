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

		startMessages := []string{
			"пиветики,будут новые задачи?",
			"жду новые задачи^-^",
			"привет-привет,давай задачи",
		}

		var msg tgbotapi.MessageConfig

		switch update.Message.Text {
		case "/start":
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, startMessages[rng.Intn(len(startMessages))])

		case "/add_task":
			isAddingTask = true
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "введи задачу в формате: задача - время, через которое напомнить о ней")

		default:
			if isAddingTask {
				err := mh.MessageService.AddTask(update.Message.Text, update.Message.Chat.ID)

				if err != nil {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ошибка добавления задачи(")
				} else {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "задача добавлена: "+update.Message.Text)
					isAddingTask = false
					log.Printf("A task has been received - %s\n", update.Message.Text)
				}
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "такой команды нет(")
			}
		}

		mh.Bot.Send(msg)
	}
}

func (mh *MessageHandler) Notify() {
	for {
		notification := mh.MessageService.GetNotification()

		if notification.Task != "" {
			msg := tgbotapi.NewMessage(notification.ChatID, "пора "+notification.Task)

			mh.Bot.Send(msg)
		}
	}
}
