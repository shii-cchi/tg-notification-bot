package rabbitmq

import (
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Rabbit struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	queues          []Queue
	deliveryChannel <-chan amqp.Delivery
}

type Queue struct {
	Queue amqp.Queue
	TTL   int
}

func InitRabbit(url string, queueTTLs []int) (*Rabbit, error) {
	conn, err := amqp.Dial(url)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ %s\n", err)
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, errors.New("failed to open a channel")
	}

	queues := make([]Queue, 0)

	outputQueue, err := ch.QueueDeclare(
		"output_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, errors.New("failed to create output_queue")
	}

	queues = append(queues, Queue{Queue: outputQueue, TTL: 0})

	for _, ttl := range queueTTLs {
		args := amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "output_queue",
			"x-message-ttl":             int32(ttl) * 1000,
		}

		queue, err := ch.QueueDeclare(
			fmt.Sprintf("queue_%dmin", ttl/60),
			true,
			false,
			false,
			false,
			args,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create queue_%dmin\n", ttl/60)
		}

		queues = append(queues, Queue{Queue: queue, TTL: ttl * 1000})
	}

	deliveryChan, err := ch.Consume(
		queues[0].Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &Rabbit{
		conn:            conn,
		channel:         ch,
		queues:          queues,
		deliveryChannel: deliveryChan,
	}, nil
}

func (r *Rabbit) Close() {
	r.channel.Close()
	r.conn.Close()
}

func (r *Rabbit) Publish(queue amqp.Queue, msg []byte) error {
	err := r.channel.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msg,
		})

	if err != nil {
		log.Printf("error publishing message %s\n", err)
		return err
	}

	log.Printf("sent msg: %s\n into queue %s\n", string(msg), queue.Name)

	return nil
}

func (r *Rabbit) Consume() ([]byte, error) {
	log.Println("waiting message from output queue")

	delivery := <-r.deliveryChannel

	log.Printf("received a message: %s\n", delivery.Body)

	return delivery.Body, nil
}

func (r *Rabbit) GetQueueList() []Queue {
	return r.queues
}
