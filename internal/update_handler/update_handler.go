package update_handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"tg-notification-bot/internal/model"
	"tg-notification-bot/internal/service"
)

type MessageHandler struct {
	bot            *tgbotapi.BotAPI
	updatesChannel tgbotapi.UpdatesChannel
	messageService *service.MessageService
	userStates     map[int64]*HandlerState
}

type HandlerState struct {
	state         State
	task          string
	taskTime      string
	taskInfoList  []model.TaskInfo
	lastMessageID int
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
		{Command: "cancel", Description: "Cancel any operation"},
	}

	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)

	_, err = bot.Request(setCommandsConfig)

	if err != nil {
		return nil, err
	}

	updatesChan := bot.GetUpdatesChan(u)

	return &MessageHandler{
		bot:            bot,
		updatesChannel: updatesChan,
		messageService: messageService,
		userStates:     make(map[int64]*HandlerState),
	}, nil
}

func (mh *MessageHandler) HandleUpdate() {
	for update := range mh.updatesChannel {
		if update.Message != nil {
			mh.handleMessageUpdate(update.Message.Text, update.Message.Chat.ID)
		} else if update.CallbackQuery != nil {
			mh.handleCallbackQueryUpdate(update.CallbackQuery, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
		}
	}
}

func (mh *MessageHandler) Notify() {
	for {
		log.Println("waiting notification")

		notification := mh.messageService.GetNotification()

		log.Printf("getting notification %s\n", notification.Task)

		if notification.Task != "" {
			log.Printf("sending notification %s\n", notification.Task)

			mh.sendMessage(notification.ChatID, msgNotification+notification.Task)
		}
	}
}
