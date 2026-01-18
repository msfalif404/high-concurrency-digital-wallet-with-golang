package repository

import (
	"context"
	"encoding/json"

	"digital-wallet/internal/domain"
	"digital-wallet/pkg/rabbitmq"
)

type eventProducer struct {
	mq *rabbitmq.RabbitMQ
}

func NewEventProducer(mq *rabbitmq.RabbitMQ) domain.EventProducer {
	return &eventProducer{mq: mq}
}

func (p *eventProducer) PublishTransferEvent(ctx context.Context, event domain.TransferEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.mq.Publish(ctx, body)
}
