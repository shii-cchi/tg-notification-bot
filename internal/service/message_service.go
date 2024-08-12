package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"tg-notification-bot/internal/database"
	"tg-notification-bot/internal/model"
	"tg-notification-bot/internal/rabbitmq"
	"time"
)

type MessageService struct {
	rabbitService *RabbitService
	queries       *database.Queries
}

func NewMessageService(r *rabbitmq.Rabbit, q *database.Queries) *MessageService {
	return &MessageService{
		rabbitService: NewRabbitService(r),
		queries:       q,
	}
}

func (ms *MessageService) AddTask(msg, msgTime string, chatID int64) error {
	taskTimeMs, err := parseMsgTime(msgTime)

	if err != nil {
		log.Printf("invalid message format: %s\n", err)
		return err
	}

	createdAt := time.Now()

	err = ms.rabbitService.Publish(model.Task{
		Task:       msg,
		ChatID:     chatID,
		TaskTimeMs: taskTimeMs,
		CreatedAt:  createdAt,
	})

	if err != nil {
		log.Printf("error publishing message: %s\n", err)
		return err
	}

	_, err = ms.queries.CreateTask(context.Background(), database.CreateTaskParams{
		Task:      msg,
		TaskTime:  msgTime,
		ChatID:    chatID,
		CreatedAt: createdAt.UTC(),
	})

	if err != nil {
		log.Printf("error adding task into db: %s\n", err)
		return err
	}

	log.Printf("task %s has been successfully added into db\n", msg)

	return nil
}

func (ms *MessageService) GetNotification() model.Task {
	msg, err := ms.rabbitService.Consume()

	if err != nil {
		log.Printf("error consuming message: %s\n", err)
		return model.Task{}
	}

	if msg.Task != "" {
		err = ms.updateStatus(msg)

		if err != nil {
			log.Printf("error updating task status: %s\n", err)
		}
	}

	return msg
}

func parseMsgTime(msgTime string) (int, error) {
	timeParts := strings.Split(msgTime, ":")
	if len(timeParts) != 3 {
		return 0, fmt.Errorf("invalid time format, expected 'hh:mm:ss'")
	}

	hours, err := strconv.Atoi(timeParts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours value")
	}

	minutes, err := strconv.Atoi(timeParts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes value")
	}

	seconds, err := strconv.Atoi(timeParts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds value")
	}

	totalMilliseconds := (hours*3600 + minutes*60 + seconds) * 1000

	return totalMilliseconds, nil
}

func (ms *MessageService) updateStatus(msg model.Task) error {
	tasks, err := ms.queries.GetTaskId(context.Background(), database.GetTaskIdParams{
		Task:   msg.Task,
		ChatID: msg.ChatID,
	})

	if err != nil {
		return err
	}

	if len(tasks) == 1 {
		err = ms.queries.UpdateTaskStatus(context.Background(), tasks[0].ID)

		if err != nil {
			return err
		}
	} else {
		for _, task := range tasks {
			taskTimeMs, _ := parseMsgTime(task.TaskTime)

			elapsedTime := int(time.Now().Sub(task.CreatedAt).Milliseconds())

			if taskTimeMs <= elapsedTime {
				err = ms.queries.UpdateTaskStatus(context.Background(), task.ID)

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ms *MessageService) GetTaskList(chatID int64) (string, error) {
	list, err := ms.queries.GetAllTasks(context.Background(), chatID)

	if err != nil {
		return "", err
	}

	var res string

	for _, task := range list {
		taskTimeMs, _ := parseMsgTime(task.TaskTime)

		elapsedTime := int(time.Now().UTC().Sub(task.CreatedAt).Milliseconds())

		remainingTime := (taskTimeMs - elapsedTime) / 60000

		if remainingTime < 0 {
			remainingTime = 0
		}

		res += fmt.Sprintf("%s - %s (через ~%d минут)\n", task.Task, task.TaskTime, remainingTime)
	}

	return res, nil
}
