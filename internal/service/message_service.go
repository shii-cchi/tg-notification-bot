package service

import (
	"context"
	"database/sql"
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
		log.Printf("error getting task id: %s\n", err)
		return err
	}

	if len(tasks) == 1 {
		return ms.queries.UpdateTaskStatus(context.Background(), tasks[0].ID)
	}

	for _, task := range tasks {
		taskTimeMs, _ := parseMsgTime(task.TaskTime)

		elapsedTime := int(time.Since(task.CreatedAt).Milliseconds())

		if taskTimeMs <= elapsedTime {
			if err = ms.queries.UpdateTaskStatus(context.Background(), task.ID); err != nil {
				log.Printf("error updating status for task ID '%d': %v", task.ID, err)
				return err
			}
		}
	}

	return nil
}

func (ms *MessageService) GetTaskList(chatID int64) ([]model.TaskInfo, error) {
	taskList, err := ms.queries.GetAllTasks(context.Background(), chatID)

	if err != nil {
		log.Printf("error getting task list: %s\n", err)
		return nil, err
	}

	taskInfoList := make([]model.TaskInfo, len(taskList))

	for i, task := range taskList {
		taskTimeMs, _ := parseMsgTime(task.TaskTime)

		elapsedTime := int(time.Since(task.CreatedAt).Milliseconds())

		remainingTime := (taskTimeMs - elapsedTime) / 60000

		remainingTimeMsg := fmt.Sprintf("через ~%d минут", remainingTime)

		if remainingTime <= 0 {
			remainingTimeMsg = "уже истекло"
		}

		taskInfoList[i] = model.TaskInfo{
			TaskID:       task.ID,
			TaskWithTime: fmt.Sprintf("%d) %s - %s (%s)\n", i+1, task.Task, task.TaskTime, remainingTimeMsg),
		}
	}

	return taskInfoList, nil
}

func (ms *MessageService) DeleteTask(id int64) error {
	deletedAt := sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	return ms.queries.DeleteTask(context.Background(), database.DeleteTaskParams{ID: id, DeletedAt: deletedAt})
}
