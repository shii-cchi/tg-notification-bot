package service

import (
	"encoding/json"
	"fmt"
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

func (rs *RabbitService) Publish(task model.Task) error {
	msg, err := json.Marshal(task)

	if err != nil {
		return fmt.Errorf("error decoding message: %s\n", err)
	}

	queues := rs.rabbit.GetQueueList()

	for _, queue := range queues {
		if task.TaskTimeMs <= queue.TTL {
			err = rs.rabbit.Publish(queue.Queue, msg)

			if err != nil {
				return fmt.Errorf("error publishing message: %s\n", err)
			}

			return nil
		}
	}

	err = rs.rabbit.Publish(queues[len(queues)-1].Queue, msg)

	if err != nil {
		return fmt.Errorf("error publishing message: %s\n", err)
	}

	return nil
}

func (rs *RabbitService) Consume() (model.Task, error) {
	msgByte, err := rs.rabbit.Consume()

	if err != nil {
		return model.Task{}, err
	}

	msg := model.Task{}

	err = json.Unmarshal(msgByte, &msg)

	if err != nil {
		return model.Task{}, err
	}

	const toleranceMs = 59 * 1000

	if msg.TaskTimeMs-int(time.Now().Sub(msg.CreatedAt).Milliseconds()) <= toleranceMs {
		return msg, nil
	} else {
		msg.TaskTimeMs -= int(time.Now().Sub(msg.CreatedAt).Milliseconds())
		msg.CreatedAt = time.Now()

		err = rs.Publish(msg)

		if err != nil {
			return model.Task{}, err
		}

		return model.Task{}, nil
	}
}