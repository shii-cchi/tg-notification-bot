package update_handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (mh *MessageHandler) createTimeKeyboard(hours, minutes, seconds int) tgbotapi.InlineKeyboardMarkup {
	h := fmt.Sprintf("%02d", hours)
	m := fmt.Sprintf("%02d", minutes)
	s := fmt.Sprintf("%02d", seconds)

	topRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(iconIncrease, fmt.Sprintf("%s:%s:%s:%s", actionIncreaseHours, h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData(iconIncrease, fmt.Sprintf("%s:%s:%s:%s", actionIncreaseMin, h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData(iconIncrease, fmt.Sprintf("%s:%s:%s:%s", actionIncreaseSec, h, m, s)),
	)

	middleRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(h, actionIgnoreHours),
		tgbotapi.NewInlineKeyboardButtonData(m, actionIgnoreMin),
		tgbotapi.NewInlineKeyboardButtonData(s, actionIgnoreSec),
	)

	bottomRow := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(iconDecrease, fmt.Sprintf("%s:%s:%s:%s", actionDecreaseHours, h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData(iconDecrease, fmt.Sprintf("%s:%s:%s:%s", actionDecreaseMin, h, m, s)),
		tgbotapi.NewInlineKeyboardButtonData(iconDecrease, fmt.Sprintf("%s:%s:%s:%s", actionDecreaseSec, h, m, s)),
	)

	confirmRow := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(textConfirm, fmt.Sprintf("%s:%s:%s:%s", actionConfirmTime, h, m, s)))

	return tgbotapi.NewInlineKeyboardMarkup(topRow, middleRow, bottomRow, confirmRow)
}

func (mh *MessageHandler) createTaskListKeyboard(lenList, page int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var taskRow []tgbotapi.InlineKeyboardButton
	var pageSwitchRow []tgbotapi.InlineKeyboardButton

	startIndex := page * maxTasksButton
	endIndex := startIndex + maxTasksButton

	if endIndex > lenList {
		endIndex = lenList
	}

	for i := startIndex; i < endIndex; i++ {
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", i+1), fmt.Sprintf("%s_%d", actionTask, i+1))
		taskRow = append(taskRow, button)
	}

	rows = append(rows, taskRow)

	if page > 0 {
		prevPageButton := tgbotapi.NewInlineKeyboardButtonData(textPrevPage, fmt.Sprintf("%s_%d", actionPage, page-1))
		pageSwitchRow = append(pageSwitchRow, prevPageButton)
	}

	if endIndex < lenList {
		nextPageButton := tgbotapi.NewInlineKeyboardButtonData(textNextPage, fmt.Sprintf("%s_%d", actionPage, page+1))
		pageSwitchRow = append(pageSwitchRow, nextPageButton)
	}

	if len(pageSwitchRow) != 0 {
		rows = append(rows, pageSwitchRow)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (mh *MessageHandler) createTaskKeyboard(id int64, taskNumber int) tgbotapi.InlineKeyboardMarkup {
	row := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(textDelete, fmt.Sprintf("%s_%d", actionDelete, id)),
		tgbotapi.NewInlineKeyboardButtonData(textBack, fmt.Sprintf("%s_%d", actionBack, taskNumber)),
	)

	return tgbotapi.NewInlineKeyboardMarkup(row)
}

func (mh *MessageHandler) initializeUserState(chatID int64) {
	_, exists := mh.userStates[chatID]

	if !exists {
		mh.userStates[chatID] = &HandlerState{
			state:         StateIdle,
			task:          "",
			taskTime:      "",
			taskInfoList:  nil,
			lastMessageID: 0,
		}
	}
}

func (mh *MessageHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := mh.bot.Send(msg); err != nil {
		log.Printf("error sending message: %v", err)
	}
}

func (mh *MessageHandler) sendMessageAndKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) int {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	message, err := mh.bot.Send(msg)
	if err != nil {
		log.Printf("error sending message: %v", err)
	}

	return message.MessageID
}

func (mh *MessageHandler) sendEditMessageAndKeyboard(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, keyboard)
	if _, err := mh.bot.Send(editMsg); err != nil {
		log.Printf("error updating message with new keyboard: %v", err)
	}
}

func (mh *MessageHandler) sendCallbackResponse(callbackID string) {
	callbackConfig := tgbotapi.NewCallback(callbackID, "")
	if _, err := mh.bot.Request(callbackConfig); err != nil {
		log.Printf("error sending callback confirmation: %v", err)
	}
}

func (mh *MessageHandler) sendDeleteMessage(chatID int64, messageID int) {
	deleteConfig := tgbotapi.NewDeleteMessage(chatID, messageID)

	if _, err := mh.bot.Request(deleteConfig); err != nil {
		log.Printf("error delete message: %v", err)
	}
}

func (mh *MessageHandler) sendSticker(chatID int64, stickerID string) {
	sticker := tgbotapi.NewSticker(chatID, tgbotapi.FileID(stickerID))
	if _, err := mh.bot.Send(sticker); err != nil {
		log.Printf("error sending sticker: %v", err)
	}
}
