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
		{Command: "start", Description: "Start using notification"},
		{Command: "add", Description: "Add task in queue"},
		{Command: "list", Description: "Show a list of your tasks"},
		{Command: "cancel", Description: "Cancel adding task or time"},
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
	isAddingTime := false
	var task string

	for update := range mh.UpdatesChannel {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/start":
			mh.handleStart(update.Message.Chat.ID)

		case "/add":
			isAddingTask = true
			mh.handleAddTaskCommand(update.Message.Chat.ID)

		case "/list":
			mh.handleList(update.Message.Chat.ID)

		case "/cancel":
			isAddingTask = false
			isAddingTime = false
			task = ""

		default:
			if isAddingTask || isAddingTime {
				if isAddingTask {
					mh.handleAddTimeCommand(update.Message.Chat.ID)
					task = update.Message.Text
					isAddingTask = false
					isAddingTime = true
					continue
				}

				if isAddingTime {
					mh.handleTaskAddition(update.Message.Chat.ID, task, update.Message.Text)
					isAddingTime = false
				}
			} else {
				mh.handleUnknownCommand(update.Message.Chat.ID)
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

			if _, err := mh.Bot.Send(msg); err != nil {
				log.Printf("error sending notification message: %v", err)
			}
		}
	}
}

func (mh *MessageHandler) handleStart(chatID int64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	msg := tgbotapi.NewMessage(chatID, startMessages[rng.Intn(len(startMessages))])
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(startSticker))

	if _, err := mh.Bot.Send(msg); err != nil {
		log.Printf("error sending start message: %v", err)
	}

	if _, err := mh.Bot.Send(sticker); err != nil {
		log.Printf("error sending start sticker: %v", err)
	}
}

func (mh *MessageHandler) handleAddTaskCommand(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, addTaskMessage)

	if _, err := mh.Bot.Send(msg); err != nil {
		log.Printf("error sending add task message: %v", err)
	}
}

func (mh *MessageHandler) handleAddTimeCommand(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, addTimeMessage)

	if _, err := mh.Bot.Send(msg); err != nil {
		log.Printf("error sending add time message: %v", err)
	}
}

func (mh *MessageHandler) handleTaskAddition(chatID int64, task, time string) {
	log.Printf("starting adding task - %s in queue\n", task)

	err := mh.MessageService.AddTask(task, time, chatID)

	if err != nil {
		msg := tgbotapi.NewMessage(chatID, addingErrMessage)

		if _, err = mh.Bot.Send(msg); err != nil {
			log.Printf("error sending error message: %v", err)
		}

	} else {
		log.Printf("task - %s has been added in queue\n", task)
		msg := tgbotapi.NewMessage(chatID, successAddMessage+task)
		sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(successAddSticker))

		if _, err = mh.Bot.Send(msg); err != nil {
			log.Printf("error sending success add task message: %v", err)
		}

		if _, err = mh.Bot.Send(sticker); err != nil {
			log.Printf("error sending success add task sticker: %v", err)
		}
	}
}

func (mh *MessageHandler) handleUnknownCommand(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, unknownCommandMessage)
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(unknownCommandSticker))

	if _, err := mh.Bot.Send(msg); err != nil {
		log.Printf("error sending unknown command message: %v", err)
	}

	if _, err := mh.Bot.Send(sticker); err != nil {
		log.Printf("error sending unknown command sticker: %v", err)
	}
}

func (mh *MessageHandler) handleList(chatID int64) {
	list, err := mh.MessageService.GetTaskList(chatID)

	if err != nil {
		msg := tgbotapi.NewMessage(chatID, gettingListErrMessage)

		if _, err = mh.Bot.Send(msg); err != nil {
			log.Printf("error sending error message: %v", err)
		}
	} else if list != "" {
		msg := tgbotapi.NewMessage(chatID, getListMessage+list)

		if _, err := mh.Bot.Send(msg); err != nil {
			log.Printf("error sending list of tasks: %v", err)
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, noTasksMessage)

		if _, err := mh.Bot.Send(msg); err != nil {
			log.Printf("error sending list of tasks: %v", err)
		}
	}
}
