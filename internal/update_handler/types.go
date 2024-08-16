package update_handler

type State int

const (
	StateIdle State = iota
	StateAddingTask
	StateAddingTime
	StateAddingInQueue
	StateShowingList
	StateEditingTask
)

const (
	msgStart              = "приветики^-^"
	msgAddTask            = "о чем тебе напомнить?"
	msgAddTime            = "введи время"
	msgAddingErr          = "ошибка добавления("
	msgUnknownCommand     = "такой команды нет("
	msgSuccessAdd         = "успешно добавлено: "
	msgNotification       = "пора "
	msgGetList            = "вот все твои дела:\n\n"
	msgGettingListErr     = "ошибка получения списка дел("
	msgNoTasks            = "у тебя нет добавленных дел"
	msgTaskDeletedSuccess = "задача удалена"
	msgTaskDeletedFailed  = "задача не была удалена("
)

const (
	stickerStart          = "CAACAgIAAxkBAAEMoF9mtkhyWknPycFAHoFr_r3jjIdOCgACIRMAAqr1AAFIRBQx6LiUPhQ1BA"
	stickerSuccessAdd     = "CAACAgIAAxkBAAEMoGFmtkjyTquTfcvcjFpkbbb3WUBssQAC3z0AAospUUsQBB-1YCaT3zUE"
	stickerUnknownCommand = "CAACAgIAAxkBAAEMoGVmtknUWSZ1ezd6JnxeNQ5OltZkYwACpEgAAtbBUUthLNyzsLGbRjUE"
)

const (
	actionIncreaseHours = "increase_hours"
	actionIncreaseMin   = "increase_min"
	actionIncreaseSec   = "increase_sec"
	actionDecreaseHours = "decrease_hours"
	actionDecreaseMin   = "decrease_min"
	actionDecreaseSec   = "decrease_sec"
	actionIgnoreHours   = "ignore_hours"
	actionIgnoreMin     = "ignore_min"
	actionIgnoreSec     = "ignore_sec"
	actionConfirmTime   = "confirm_time"
	actionTask          = "task"
	actionPage          = "page"
	actionDelete        = "delete"
	actionBack          = "back"
)

const (
	iconIncrease = "⬆️"
	iconDecrease = "⬇️"
	textConfirm  = "подтвердить"
	textNextPage = "далее"
	textPrevPage = "назад"
	textDelete   = "удалить"
	textBack     = "назад"
)

const (
	maxTasksButton = 7
)
