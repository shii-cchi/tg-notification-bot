package message_handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"tg-notification-bot/internal/model"
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
	var lastMessageID int
	var taskInfoList []model.TaskInfo

	for update := range mh.UpdatesChannel {
		if update.Message != nil {
			mh.handleMessage(update, &task, &state, &lastMessageID, &taskInfoList)
		} else if update.CallbackQuery != nil {
			if state == StateAddingTime {
				taskTime := mh.handlePushTimeButton(update.CallbackQuery, &state)

				if state == StateAddingInQueue {
					mh.handleTaskAddition(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, task, taskTime)

					state = StateIdle
				}
			}

			if state == StateShowList {
				mh.handlePushTaskButton(update.CallbackQuery, &state, taskInfoList)
			}

			if state == StateEditTask {
				mh.handlePushEditTaskButton(update.CallbackQuery, &state, taskInfoList)
			}
		}
	}
}

func (mh *MessageHandler) handleMessage(update tgbotapi.Update, task *string, state *State, lastMessageID *int, taskInfoList *[]model.TaskInfo) {
	if *state == StateAddingTime || *state == StateShowList || *state == StateEditTask {
		deleteConfig := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, *lastMessageID)

		if _, err := mh.Bot.Request(deleteConfig); err != nil {
			log.Printf("error delete message: %v", err)
		}

		*state = StateIdle
	}

	switch update.Message.Text {
	case "/start":
		mh.handleStart(update.Message.Chat.ID)
		*state = StateIdle

	case "/add":
		mh.handleAddTaskCommand(update.Message.Chat.ID)
		*state = StateAddingTask

	case "/list":
		*lastMessageID, *taskInfoList = mh.handleList(update.Message.Chat.ID, 0)

		if *taskInfoList == nil {
			*state = StateIdle
		} else {
			*state = StateShowList
		}

	case "/cancel":
		*state = StateIdle
		*task = ""

	default:
		switch *state {
		case StateAddingTask:
			*lastMessageID = mh.handleAddTimeCommand(update.Message.Chat.ID)
			*task = update.Message.Text
			*state = StateAddingTime

		default:
			mh.handleUnknownCommand(update.Message.Chat.ID)
		}
	}
}

func (mh *MessageHandler) handlePushTimeButton(callbackQuery *tgbotapi.CallbackQuery, state *State) string {
	parts := strings.Split(callbackQuery.Data, ":")

	if len(parts) != 4 {
		log.Printf("unexpected callback data: %v", callbackQuery.Data)
		return ""
	}

	action := parts[0]
	hours, _ := strconv.Atoi(parts[1])
	minutes, _ := strconv.Atoi(parts[2])
	seconds, _ := strconv.Atoi(parts[3])

	switch action {
	case "increase_hours":
		hours = (hours + 1) % 24
	case "increase_min":
		minutes = (minutes + 1) % 60
	case "increase_sec":
		seconds = (seconds + 1) % 60
	case "decrease_hours":
		hours = (hours - 1 + 24) % 24
	case "decrease_min":
		minutes = (minutes - 1 + 60) % 60
	case "decrease_sec":
		seconds = (seconds - 1 + 60) % 60
	case "confirm_time":
		*state = StateAddingInQueue
		return strings.Join(parts[1:], ":")
	}

	newKeyboard := mh.createTimeKeyboard(hours, minutes, seconds)

	editMsg := tgbotapi.NewEditMessageReplyMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, newKeyboard)
	if _, err := mh.Bot.Send(editMsg); err != nil {
		log.Printf("error updating message with new keyboard: %v", err)
	}

	callbackConfig := tgbotapi.NewCallback(callbackQuery.ID, "")
	if _, err := mh.Bot.Request(callbackConfig); err != nil {
		log.Printf("error sending callback confirmation: %v", err)
	}

	return ""
}

func (mh *MessageHandler) handlePushTaskButton(callbackQuery *tgbotapi.CallbackQuery, state *State, taskInfoList []model.TaskInfo) {
	parts := strings.Split(callbackQuery.Data, "_")

	if len(parts) < 2 {
		log.Printf("unexpected callback data: %v", callbackQuery.Data)
		return
	}

	action := parts[0]
	value := parts[1]

	switch action {
	case "task":
		id, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("error converting id number: %v", err)
			return
		}

		*state = StateEditTask

		keyboard := mh.createTaskKeyboard(taskInfoList[id-1].TaskID, id)

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, taskInfoList[id-1].TaskWithTime, keyboard)

		if _, err := mh.Bot.Send(editMsg); err != nil {
			log.Printf("error sending edited message: %v", err)
		}

		callbackConfig := tgbotapi.NewCallback(callbackQuery.ID, "")
		if _, err := mh.Bot.Request(callbackConfig); err != nil {
			log.Printf("error sending callback confirmation: %v", err)
		}

	case "page":
		page, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("error converting page number: %v", err)
			return
		}

		var builder strings.Builder

		startIndex := page * maxTasksButton
		endIndex := startIndex + maxTasksButton

		if endIndex > len(taskInfoList) {
			endIndex = len(taskInfoList)
		}

		for i := startIndex; i < endIndex; i++ {
			builder.WriteString(taskInfoList[i].TaskWithTime)
		}

		keyboard := mh.createTaskListKeyboard(len(taskInfoList), page)

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, getListMessage+builder.String(), keyboard)

		if _, err := mh.Bot.Send(editMsg); err != nil {
			log.Printf("error sending edited message: %v", err)
		}

		callbackConfig := tgbotapi.NewCallback(callbackQuery.ID, "")
		if _, err := mh.Bot.Request(callbackConfig); err != nil {
			log.Printf("error sending callback confirmation: %v", err)
		}
	}
}

func (mh *MessageHandler) handlePushEditTaskButton(callbackQuery *tgbotapi.CallbackQuery, state *State, taskInfoList []model.TaskInfo) {
	parts := strings.Split(callbackQuery.Data, "_")

	if len(parts) < 2 {
		log.Printf("unexpected callback data: %v", callbackQuery.Data)
		return
	}

	action := parts[0]
	value := parts[1]

	switch action {
	case "delete":
		id, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("error converting id number: %v", err)
			return
		}

		deleteConfig := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)

		if _, err = mh.Bot.Request(deleteConfig); err != nil {
			log.Printf("error delete message: %v", err)
		}

		err = mh.MessageService.DeleteTask(int64(id))

		if err != nil {
			mh.sendMessage(callbackQuery.Message.Chat.ID, deletingTaskFailedMessage)
		} else {
			mh.sendMessage(callbackQuery.Message.Chat.ID, deletingTaskSuccessMessage)
		}

		*state = StateIdle

		callbackConfig := tgbotapi.NewCallback(callbackQuery.ID, "")
		if _, err := mh.Bot.Request(callbackConfig); err != nil {
			log.Printf("error sending callback confirmation: %v", err)
		}

	case "back":
		taskNumber, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("error converting id number: %v", err)
			return
		}

		page := (taskNumber - 1) / maxTasksButton

		var builder strings.Builder

		startIndex := page * maxTasksButton
		endIndex := startIndex + maxTasksButton

		if endIndex > len(taskInfoList) {
			endIndex = len(taskInfoList)
		}

		for i := startIndex; i < endIndex; i++ {
			builder.WriteString(taskInfoList[i].TaskWithTime)
		}

		keyboard := mh.createTaskListKeyboard(len(taskInfoList), page)

		editMsg := tgbotapi.NewEditMessageTextAndMarkup(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, getListMessage+builder.String(), keyboard)

		_, err = mh.Bot.Send(editMsg)
		if err != nil {
			log.Printf("error sending message: %v", err)
		}

		*state = StateShowList

		callbackConfig := tgbotapi.NewCallback(callbackQuery.ID, "")
		if _, err := mh.Bot.Request(callbackConfig); err != nil {
			log.Printf("error sending callback confirmation: %v", err)
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

func (mh *MessageHandler) handleAddTimeCommand(chatID int64) int {
	hours, minutes, seconds := 0, 0, 0
	keyboard := mh.createTimeKeyboard(hours, minutes, seconds)

	msg := tgbotapi.NewMessage(chatID, addTimeMessage)
	msg.ReplyMarkup = keyboard

	message, err := mh.Bot.Send(msg)
	if err != nil {
		log.Printf("error sending message: %v", err)
	}

	return message.MessageID
}

func (mh *MessageHandler) handleTaskAddition(chatID int64, messageID int, task, time string) {
	log.Printf("starting adding task - %s in queue\n", task)

	err := mh.MessageService.AddTask(task, time, chatID)

	if err != nil {
		mh.sendMessage(chatID, addingErrMessage)
		return
	}

	log.Printf("task - %s has been added in queue\n", task)
	mh.sendMessage(chatID, successAddMessage+task)
	mh.sendSticker(chatID, successAddSticker)

	deleteConfig := tgbotapi.NewDeleteMessage(chatID, messageID)

	if _, err = mh.Bot.Request(deleteConfig); err != nil {
		log.Printf("error delete message: %v", err)
	}
}

func (mh *MessageHandler) handleUnknownCommand(chatID int64) {
	mh.sendMessage(chatID, unknownCommandMessage)
	mh.sendSticker(chatID, unknownCommandSticker)
}

func (mh *MessageHandler) handleList(chatID int64, page int) (int, []model.TaskInfo) {
	taskInfoList, err := mh.MessageService.GetTaskList(chatID)

	if err != nil {
		mh.sendMessage(chatID, gettingListErrMessage)
		return 0, nil
	}

	if len(taskInfoList) != 0 {
		var builder strings.Builder

		startIndex := page * maxTasksButton
		endIndex := startIndex + maxTasksButton

		if endIndex > len(taskInfoList) {
			endIndex = len(taskInfoList)
		}

		for i := startIndex; i < endIndex; i++ {
			builder.WriteString(taskInfoList[i].TaskWithTime)
		}

		keyboard := mh.createTaskListKeyboard(len(taskInfoList), page)
		msg := tgbotapi.NewMessage(chatID, getListMessage+builder.String())
		msg.ReplyMarkup = keyboard

		message, err := mh.Bot.Send(msg)
		if err != nil {
			log.Printf("error sending message: %v", err)
		}

		return message.MessageID, taskInfoList
	} else {
		mh.sendMessage(chatID, noTasksMessage)
		return 0, nil
	}
}

func (mh *MessageHandler) createTimeKeyboard(hours, minutes, seconds int) tgbotapi.InlineKeyboardMarkup {
	h := fmt.Sprintf("%02d", hours)
	m := fmt.Sprintf("%02d", minutes)
	s := fmt.Sprintf("%02d", seconds)

	topRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬆️", fmt.Sprintf("increase_hours:%s:%s:%s", h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData("⬆️", fmt.Sprintf("increase_min:%s:%s:%s", h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData("⬆️", fmt.Sprintf("increase_sec:%s:%s:%s", h, m, s)),
	)

	middleRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h, "ignore_hours"),
		tgbotapi.NewInlineKeyboardButtonData(m, "ignore_min"),
		tgbotapi.NewInlineKeyboardButtonData(s, "ignore_sec"),
	)

	bottomRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬇️", fmt.Sprintf("decrease_hours:%s:%s:%s", h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData("⬇️", fmt.Sprintf("decrease_min:%s:%s:%s", h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData("⬇️", fmt.Sprintf("decrease_sec:%s:%s:%s", h, m, s)),
	)

	confirmRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("подтвердить", fmt.Sprintf("confirm_time:%s:%s:%s", h, m, s)))

	return tgbotapi.NewInlineKeyboardMarkup(topRow, middleRow, bottomRow, confirmRow)
}

func (mh *MessageHandler) createTaskListKeyboard(lenList, page int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	startIndex := page * maxTasksButton
	endIndex := startIndex + maxTasksButton
	lastPage := lenList / maxTasksButton

	if endIndex > lenList {
		endIndex = lenList
	}

	for i := startIndex; i < endIndex; i++ {
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", i+1), fmt.Sprintf("task_%d", i+1))
		row = append(row, button)
	}

	rows = append(rows, row)

	if lenList > endIndex {
		if page != lastPage {
			nextPageButton := tgbotapi.NewInlineKeyboardButtonData("далее", fmt.Sprintf("page_%d", page+1))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(nextPageButton))
		}
	}

	if page != 0 {
		prevPageButton := tgbotapi.NewInlineKeyboardButtonData("назад", fmt.Sprintf("page_%d", page-1))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(prevPageButton))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (mh *MessageHandler) createTaskKeyboard(id int64, taskNumber int) tgbotapi.InlineKeyboardMarkup {
	row := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("удалить", fmt.Sprintf("delete_%d", id)),
		tgbotapi.NewInlineKeyboardButtonData("назад", fmt.Sprintf("back_%d", taskNumber)),
	)

	return tgbotapi.NewInlineKeyboardMarkup(row)
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
