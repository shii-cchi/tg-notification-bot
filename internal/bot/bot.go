package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand"
	"time"
)

func NewBot(token string) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel, error) {
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10

	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Start using notification bot"},
		{Command: "add_task", Description: "Add task in queue"},
	}

	setCommandsConfig := tgbotapi.NewSetMyCommands(commands...)

	_, err = bot.Request(setCommandsConfig)

	if err != nil {
		return nil, nil, err
	}

	updatesChan := bot.GetUpdatesChan(u)

	return bot, updatesChan, nil
}

func HandleMessage(bot *tgbotapi.BotAPI, updatesChan tgbotapi.UpdatesChannel) {
	isAddingTask := false

	for update := range updatesChan {
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
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "задача добавлена: "+update.Message.Text)
				isAddingTask = false
				log.Printf("A task has been received - %s\n", update.Message.Text)
			} else {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "такой команды нет(")
			}
		}

		bot.Send(msg)
	}
}
