package update_handler

func (mh *MessageHandler) handleMessageUpdate(msg string, chatID int64) {
	mh.initializeUserState(chatID)

	mh.resetStateIfNecessary(chatID)

	switch msg {
	case "/start":
		mh.handleStartCommand(chatID)

	case "/add":
		mh.handleAddCommand(chatID)

	case "/list":
		mh.handleListCommand(chatID, 0)

	case "/cancel":
		mh.handleCancelCommand(chatID)

	default:
		mh.handleDefault(chatID, msg)
	}
}

func (mh *MessageHandler) resetStateIfNecessary(chatID int64) {
	if mh.userStates[chatID].state == StateAddingTime || mh.userStates[chatID].state == StateShowingList || mh.userStates[chatID].state == StateEditingTask {
		mh.sendDeleteMessage(chatID, mh.userStates[chatID].lastMessageID)

		mh.userStates[chatID].state = StateIdle
	}
}

func (mh *MessageHandler) handleStartCommand(chatID int64) {
	mh.sendMessage(chatID, msgStart)
	mh.sendSticker(chatID, stickerStart)

	mh.userStates[chatID].state = StateIdle
}

func (mh *MessageHandler) handleAddCommand(chatID int64) {
	mh.sendMessage(chatID, msgAddTask)

	mh.userStates[chatID].state = StateAddingTask
}

func (mh *MessageHandler) handleListCommand(chatID int64, page int) {
	taskInfoList, err := mh.messageService.GetTaskList(chatID)

	if err != nil {
		mh.sendMessage(chatID, msgGettingListErr)
		mh.userStates[chatID].state = StateIdle
		return
	}

	mh.userStates[chatID].taskInfoList = taskInfoList

	if len(taskInfoList) == 0 {
		mh.sendMessage(chatID, msgNoTasks)
		mh.userStates[chatID].state = StateIdle
		return
	}

	tasksForPage := getTasksForPage(page, taskInfoList)

	keyboard := mh.createTaskListKeyboard(len(taskInfoList), page)
	mh.userStates[chatID].lastMessageID = mh.sendMessageAndKeyboard(chatID, msgGetList+tasksForPage, keyboard)

	mh.userStates[chatID].state = StateShowingList
}

func (mh *MessageHandler) handleCancelCommand(chatID int64) {
	mh.userStates[chatID].state = StateIdle
	mh.userStates[chatID].task = ""
}

func (mh *MessageHandler) handleDefault(chatID int64, msg string) {
	if mh.userStates[chatID].state == StateAddingTask {
		mh.userStates[chatID].lastMessageID = mh.handleAddTimeCommand(chatID)
		mh.userStates[chatID].task = msg
		mh.userStates[chatID].state = StateAddingTime
	} else {
		mh.handleUnknownCommand(chatID)
	}
}

func (mh *MessageHandler) handleAddTimeCommand(chatID int64) int {
	hours, minutes, seconds := 0, 0, 0
	keyboard := mh.createTimeKeyboard(hours, minutes, seconds)

	messageID := mh.sendMessageAndKeyboard(chatID, msgAddTime, keyboard)

	return messageID
}

func (mh *MessageHandler) handleUnknownCommand(chatID int64) {
	mh.sendMessage(chatID, msgUnknownCommand)
	mh.sendSticker(chatID, stickerUnknownCommand)
}
