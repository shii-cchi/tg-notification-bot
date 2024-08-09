package service

import (
	"encoding/json"
	"fmt"
	"log"
	"tg-notification-bot/internal/model"
	"tg-notification-bot/internal/rabbitmq"
	"time"
)

type RabbitService struct {
	rabbit *rabbitmq.Rabbit
}

func NewRabbitService(r *rabbitmq.Rabbit) *RabbitService {
	return &RabbitService{
		rabbit: r,
	}
}

const errorMarginMs = 1000

func (rs *RabbitService) Publish(task model.Task) error {
	msg, err := json.Marshal(task)

	if err != nil {
		return fmt.Errorf("error decoding message: %s\n", err)
	}

	queues := rs.rabbit.GetQueueList()

	for i := len(queues) - 1; i >= 0; i-- {
		if task.TaskTimeMs >= queues[i].TTL-errorMarginMs {
			return rs.rabbit.Publish(queues[i].Queue, msg)
		}
	}

	return rs.rabbit.Publish(queues[0].Queue, msg)
}

func (rs *RabbitService) Consume() (model.Task, error) {
	msgByte, err := rs.rabbit.Consume()

	if err != nil {
		return model.Task{}, err
	}

	var msg model.Task

	if err = json.Unmarshal(msgByte, &msg); err != nil {
		return model.Task{}, err
	}

	queues := rs.rabbit.GetQueueList()

	if rs.isTaskDue(msg) || msg.TaskTimeMs < queues[1].TTL {
		log.Printf("notify about a task %s\n", msg.Task)
		return msg, nil
	}

	return rs.requeueTask(msg)
}

func (rs *RabbitService) isTaskDue(msg model.Task) bool {
	elapsedTime := int(time.Now().Sub(msg.CreatedAt).Milliseconds())
	return msg.TaskTimeMs <= elapsedTime
}

func (rs *RabbitService) requeueTask(msg model.Task) (model.Task, error) {
	log.Printf("task %s should not be completed yet, requeued\n", msg.Task)

	elapsedTime := int(time.Now().Sub(msg.CreatedAt).Milliseconds())
	msg.TaskTimeMs -= elapsedTime
	msg.CreatedAt = time.Now()

	if err := rs.Publish(msg); err != nil {
		return model.Task{}, err
	}

	return model.Task{}, nil
}
