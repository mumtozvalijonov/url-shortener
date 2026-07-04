package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
)

type shortURLVisitedMessage struct {
	ShortCode string    `json:"short_code"`
	VisitedAt time.Time `json:"visited_at"`
}

type ShortURLVisitedPublisher struct {
	producer *kafka.Producer
	topic    string
}

func NewShortURLVisitedPublisher(
	producer *kafka.Producer,
	topic string,
) *ShortURLVisitedPublisher {
	return &ShortURLVisitedPublisher{
		producer: producer,
		topic:    topic,
	}
}

func (p *ShortURLVisitedPublisher) Publish(
	ctx context.Context,
	event domain.ShortURLVisited,
) error {
	message := shortURLVisitedMessage{
		ShortCode: event.ShortCode,
		VisitedAt: event.VisitedAt,
	}

	value, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal short URL visited event: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.ShortCode),
		Value: value,
	}

	return p.producer.Produce(msg, nil)
}
