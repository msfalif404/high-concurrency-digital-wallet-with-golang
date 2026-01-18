package rabbitmq

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

const QueueName = "wallet_transfers"

func NewConnection(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		QueueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	log.Println("Connected to RabbitMQ")
	return &RabbitMQ{conn: conn, channel: ch, queue: q}, nil
}

func (r *RabbitMQ) Publish(ctx context.Context, body []byte) error {
	return r.channel.PublishWithContext(ctx,
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (r *RabbitMQ) Consume() (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack (We want manual Ack)
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
