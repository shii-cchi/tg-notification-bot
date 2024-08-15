package message_handler

type State int

const (
	StateIdle State = iota
	StateAddingTask
	StateAddingTime
	StateAddingInQueue
	StateShowList
	StateEditTask
)

var startMessages = []string{
	"пиветики^-^",
	"привет^^",
	"привет-привет~",
}

const addTaskMessage = "о чем тебе напомнить?"
const addTimeMessage = "введи время"
const addingErrMessage = "ошибка добавления("
const unknownCommandMessage = "такой команды нет("
const successAddMessage = "успешно добавлено: "
const notificationMessage = "пора "
const getListMessage = "вот все твои дела:\n\n"
const gettingListErrMessage = "ошибка получения списка дел("
const noTasksMessage = "у тебя нет добавленных дел"
const deletingTaskSuccessMessage = "задача удалена"
const deletingTaskFailedMessage = "задача не была удалена("

const startSticker = "CAACAgIAAxkBAAEMoF9mtkhyWknPycFAHoFr_r3jjIdOCgACIRMAAqr1AAFIRBQx6LiUPhQ1BA"
const successAddSticker = "CAACAgIAAxkBAAEMoGFmtkjyTquTfcvcjFpkbbb3WUBssQAC3z0AAospUUsQBB-1YCaT3zUE"
const unknownCommandSticker = "CAACAgIAAxkBAAEMoGVmtknUWSZ1ezd6JnxeNQ5OltZkYwACpEgAAtbBUUthLNyzsLGbRjUE"

const maxTasksButton = 7
