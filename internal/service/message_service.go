package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"tg-notification-bot/internal/model"
	"tg-notification-bot/internal/rabbitmq"
	"time"
)

type MessageService struct {
	rabbitService *RabbitService
}

func NewMessageService(r *rabbitmq.Rabbit) *MessageService {
	return &MessageService{
		rabbitService: NewRabbitService(r),
	}
}

func (ms *MessageService) AddTask(msg string, chatID int64) error {
	task, taskTimeMs, err := parseMsg(msg)

	if err != nil {
		log.Printf("invalid message format: %s\n", err)
		return err
	}

	err = ms.rabbitService.Publish(model.Task{
		Task:       task,
		ChatID:     chatID,
		TaskTimeMs: taskTimeMs,
		CreatedAt:  time.Now(),
	})

	if err != nil {
		log.Printf("error publishing message: %s\n", err)
		return err
	}

	return nil
}

func (ms *MessageService) GetNotification() model.Task {
	msg, err := ms.rabbitService.Consume()

	if err != nil {
		log.Printf("error consuming message: %s\n", err)
		return model.Task{}
	}

	return msg
}

func parseMsg(msg string) (string, int, error) {
	parts := strings.SplitN(msg, " - ", 2)

	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid input format, expected 'TASK - TIME'")
	}

	task := parts[0]
	timeStr := parts[1]

	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 3 {
		return "", 0, fmt.Errorf("invalid time format, expected 'hh:mm:ss'")
	}

	hours, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return "", 0, fmt.Errorf("invalid hours value")
	}

	minutes, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid minutes value")
	}

	seconds, err := strconv.Atoi(timeParts[2])
	if err != nil {
		return "", 0, fmt.Errorf("invalid seconds value")
	}

	totalMilliseconds := (hours*3600 + minutes*60 + seconds) * 1000

	return task, totalMilliseconds, nil
}
