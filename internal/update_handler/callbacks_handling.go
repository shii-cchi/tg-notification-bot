package update_handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"tg-notification-bot/internal/model"
)

func (mh *MessageHandler) handleCallbackQueryUpdate(callbackQuery *tgbotapi.CallbackQuery, chatID int64, messageID int) {
	mh.initializeUserState(chatID)

	switch mh.userStates[chatID].state {
	case StateAddingTime:
		mh.handlePushTimeButton(callbackQuery)

		if mh.userStates[chatID].state == StateAddingInQueue {
			mh.handleTaskAddition(chatID, messageID)

			mh.userStates[chatID].state = StateIdle
		}

	case StateShowingList:
		mh.handlePushTaskButton(callbackQuery)

	case StateEditingTask:
		mh.handlePushEditTaskButton(callbackQuery)

	default:
		log.Printf("unexpected state: %v", mh.userStates[chatID].state)
	}
}

func (mh *MessageHandler) handlePushTimeButton(callbackQuery *tgbotapi.CallbackQuery) {
	defer mh.sendCallbackResponse(callbackQuery.ID)

	parts := strings.Split(callbackQuery.Data, ":")

	if len(parts) != 4 {
		if callbackQuery.Data != actionIgnoreHours && callbackQuery.Data != actionIgnoreMin && callbackQuery.Data != actionIgnoreSec {
			log.Printf("unexpected callback data: %v", callbackQuery.Data)
			return
		}

		mh.userStates[callbackQuery.Message.Chat.ID].state = StateIdle
		return
	}

	action := parts[0]
	hours, _ := strconv.Atoi(parts[1])
	minutes, _ := strconv.Atoi(parts[2])
	seconds, _ := strconv.Atoi(parts[3])

	switch action {
	case actionIncreaseHours:
		hours = (hours + 1) % 24
	case actionIncreaseMin:
		minutes = (minutes + 1) % 60
	case actionIncreaseSec:
		seconds = (seconds + 1) % 60
	case actionDecreaseHours:
		hours = (hours - 1 + 24) % 24
	case actionDecreaseMin:
		minutes = (minutes - 1 + 60) % 60
	case actionDecreaseSec:
		seconds = (seconds - 1 + 60) % 60
	case actionConfirmTime:
		mh.userStates[callbackQuery.Message.Chat.ID].state = StateAddingInQueue
		mh.userStates[callbackQuery.Message.Chat.ID].taskTime = strings.Join(parts[1:], ":")
		return
	}

	newKeyboard := mh.createTimeKeyboard(hours, minutes, seconds)

	mh.sendEditMessageAndKeyboard(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, msgAddTime, newKeyboard)

	return
}

func (mh *MessageHandler) handleTaskAddition(chatID int64, messageID int) {
	log.Printf("starting adding task - %s in queue\n", mh.userStates[chatID].task)

	if err := mh.messageService.AddTask(mh.userStates[chatID].task, mh.userStates[chatID].taskTime, chatID); err != nil {
		mh.sendMessage(chatID, msgAddingErr)
		mh.sendDeleteMessage(chatID, messageID)
		return
	}

	log.Printf("task - %s has been added in queue\n", mh.userStates[chatID].task)

	mh.sendMessage(chatID, msgSuccessAdd+mh.userStates[chatID].task)
	mh.sendSticker(chatID, stickerSuccessAdd)
	mh.sendDeleteMessage(chatID, messageID)
}

func (mh *MessageHandler) handlePushTaskButton(callbackQuery *tgbotapi.CallbackQuery) {
	defer mh.sendCallbackResponse(callbackQuery.ID)

	parts := strings.Split(callbackQuery.Data, "_")

	if len(parts) < 2 {
		log.Printf("unexpected callback data: %v", callbackQuery.Data)
		mh.userStates[callbackQuery.Message.Chat.ID].state = StateIdle
		return
	}

	action, value := parts[0], parts[1]

	switch action {
	case actionTask:
		id, _ := strconv.Atoi(value)

		newKeyboard := mh.createTaskKeyboard(mh.userStates[callbackQuery.Message.Chat.ID].taskInfoList[id-1].TaskID, id)
		mh.sendEditMessageAndKeyboard(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, mh.userStates[callbackQuery.Message.Chat.ID].taskInfoList[id-1].TaskWithTime, newKeyboard)

		mh.userStates[callbackQuery.Message.Chat.ID].state = StateEditingTask

	case actionPage:
		page, _ := strconv.Atoi(value)

		tasksForPage := getTasksForPage(page, mh.userStates[callbackQuery.Message.Chat.ID].taskInfoList)

		newKeyboard := mh.createTaskListKeyboard(len(mh.userStates[callbackQuery.Message.Chat.ID].taskInfoList), page)
		mh.sendEditMessageAndKeyboard(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, msgGetList+tasksForPage, newKeyboard)
	}
}

func (mh *MessageHandler) handlePushEditTaskButton(callbackQuery *tgbotapi.CallbackQuery) {
	defer mh.sendCallbackResponse(callbackQuery.ID)

	parts := strings.Split(callbackQuery.Data, "_")

	if len(parts) < 2 {
		log.Printf("unexpected callback data: %v", callbackQuery.Data)
		mh.userStates[callbackQuery.Message.Chat.ID].state = StateIdle
		return
	}

	action, value := parts[0], parts[1]

	switch action {
	case actionDelete:
		id, _ := strconv.Atoi(value)

		mh.sendDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)

		if err := mh.messageService.DeleteTask(int64(id)); err != nil {
			log.Printf("error deleting task: %s\n", err)
			mh.sendMessage(callbackQuery.Message.Chat.ID, msgTaskDeletedFailed)
		} else {
			mh.sendMessage(callbackQuery.Message.Chat.ID, msgTaskDeletedSuccess)
		}

		mh.userStates[callbackQuery.Message.Chat.ID].state = StateIdle

	case actionBack:
		taskNumber, _ := strconv.Atoi(value)

		page := (taskNumber - 1) / maxTasksButton

		tasksForPage := getTasksForPage(page, mh.userStates[callbackQuery.Message.Chat.ID].taskInfoList)

		newKeyboard := mh.createTaskListKeyboard(len(mh.userStates[callbackQuery.Message.Chat.ID].taskInfoList), page)
		mh.sendEditMessageAndKeyboard(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, msgGetList+tasksForPage, newKeyboard)

		mh.userStates[callbackQuery.Message.Chat.ID].state = StateShowingList
	}
}

func getTasksForPage(page int, taskInfoList []model.TaskInfo) string {
	var builder strings.Builder

	startIndex := page * maxTasksButton
	endIndex := startIndex + maxTasksButton

	if endIndex > len(taskInfoList) {
		endIndex = len(taskInfoList)
	}

	for i := startIndex; i < endIndex; i++ {
		builder.WriteString(taskInfoList[i].TaskWithTime)
	}

	return builder.String()
}
