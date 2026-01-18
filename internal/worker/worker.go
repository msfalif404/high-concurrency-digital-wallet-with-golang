package worker

import (
	"encoding/json"
	"log"
	"time"

	"digital-wallet/internal/domain"
	"digital-wallet/pkg/rabbitmq"
)

type Worker struct {
	mq *rabbitmq.RabbitMQ
}

func NewWorker(mq *rabbitmq.RabbitMQ) *Worker {
	return &Worker{mq: mq}
}

func (w *Worker) Start() {
	msgs, err := w.mq.Consume()
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	go func() {
		for d := range msgs {
			var event domain.TransferEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error decoding event: %v", err)
				d.Nack(false, false) // discard
				continue
			}

			w.processTransfer(event)
			d.Ack(false)
		}
	}()
	log.Println("Worker started consuming events")
}

func (w *Worker) processTransfer(event domain.TransferEvent) {
	// Simulate Email Sending
	log.Printf("Processing transfer event: %s. Sending email...", event.TransactionID)
	time.Sleep(2 * time.Second)
	log.Printf("Email Sent for transaction %s", event.TransactionID)
}
