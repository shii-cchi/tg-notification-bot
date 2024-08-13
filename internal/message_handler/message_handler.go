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
	var task string
	state := StateIdle

	for update := range mh.UpdatesChannel {
		if update.Message == nil {
			continue
		}

		switch update.Message.Text {
		case "/start":
			mh.handleStart(update.Message.Chat.ID)
			state = StateIdle

		case "/add":
			mh.handleAddTaskCommand(update.Message.Chat.ID)
			state = StateAddingTask

		case "/list":
			mh.handleList(update.Message.Chat.ID)
			state = StateIdle

		case "/cancel":
			state = StateIdle
			task = ""

		default:
			mh.HandleDefault(update.Message.Chat.ID, update.Message.Text, &task, &state)
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

			mh.sendMessage(notification.ChatID, notificationMessage+notification.Task)
		}
	}
}

func (mh *MessageHandler) handleStart(chatID int64) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	mh.sendMessage(chatID, startMessages[rng.Intn(len(startMessages))])
	mh.sendSticker(chatID, startSticker)
}

func (mh *MessageHandler) handleAddTaskCommand(chatID int64) {
	mh.sendMessage(chatID, addTaskMessage)
}

func (mh *MessageHandler) handleAddTimeCommand(chatID int64) {
	mh.sendMessage(chatID, addTimeMessage)
}

func (mh *MessageHandler) handleTaskAddition(chatID int64, task, time string) {
	log.Printf("starting adding task - %s in queue\n", task)

	err := mh.MessageService.AddTask(task, time, chatID)

	if err != nil {
		mh.sendMessage(chatID, addingErrMessage)
		return
	}

	log.Printf("task - %s has been added in queue\n", task)
	mh.sendMessage(chatID, successAddMessage+task)
	mh.sendSticker(chatID, successAddSticker)

}

func (mh *MessageHandler) handleUnknownCommand(chatID int64) {
	mh.sendMessage(chatID, unknownCommandMessage)
	mh.sendSticker(chatID, unknownCommandSticker)
}

func (mh *MessageHandler) handleList(chatID int64) {
	taskList, err := mh.MessageService.GetTaskList(chatID)

	if err != nil {
		mh.sendMessage(chatID, gettingListErrMessage)
		return
	}

	if taskList != "" {
		mh.sendMessage(chatID, getListMessage+taskList)
	} else {
		mh.sendMessage(chatID, noTasksMessage)
	}
}

func (mh *MessageHandler) HandleDefault(chatID int64, msg string, task *string, state *State) {
	switch *state {
	case StateAddingTask:
		mh.handleAddTimeCommand(chatID)
		*task = msg
		*state = StateAddingTime

	case StateAddingTime:
		mh.handleTaskAddition(chatID, *task, msg)
		*state = StateIdle

	default:
		mh.handleUnknownCommand(chatID)
	}
}

func (mh *MessageHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := mh.Bot.Send(msg); err != nil {
		log.Printf("error sending message: %v", err)
	}
}

func (mh *MessageHandler) sendSticker(chatID int64, stickerID string) {
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(stickerID))
	if _, err := mh.Bot.Send(sticker); err != nil {
		log.Printf("error sending sticker: %v", err)
	}
}
