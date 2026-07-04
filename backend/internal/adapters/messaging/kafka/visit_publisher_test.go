package kafka_test

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaadapter "github.com/mumtozvalijonov/url-shortener/internal/adapters/messaging/kafka"
	"github.com/mumtozvalijonov/url-shortener/internal/domain"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	kafkacontainer "github.com/testcontainers/testcontainers-go/modules/kafka"
)

type kafkaTestCluster struct {
	brokers []string
}

func startKafka(ctx context.Context, t *testing.T) *kafkaTestCluster {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	t.Cleanup(cancel)

	container, err := kafkacontainer.Run(
		ctx,
		"confluentinc/confluent-local:8.2.2",
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, container)

	brokers, err := container.Brokers(ctx)
	require.NoError(t, err)

	return &kafkaTestCluster{
		brokers: brokers,
	}
}

func TestVisitPublisher(t *testing.T) {
	cluster := startKafka(context.Background(), t)

	t.Run("Publish", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		producer, err := ckafka.NewProducer(&ckafka.ConfigMap{
			"bootstrap.servers": strings.Join(cluster.brokers, ","),
			"client.id":         "myProducer",
			"acks":              "all"})
		if err != nil {
			log.Fatalf("create Kafka producer: %v", err)
		}
		defer func() {
			producer.Flush(5_000)
			producer.Close()
		}()

		publisher := kafkaadapter.NewShortURLVisitedPublisher(producer, "test-topic")

		visitedAt := time.Date(2026, 1, 1, 15, 0, 0, 0, time.UTC)
		err = publisher.Publish(ctx, domain.ShortURLVisited{
			ShortCode: "abcde",
			VisitedAt: visitedAt,
		})
		require.NoError(t, err)

		event := <-producer.Events()
		switch ev := event.(type) {
		case *kafka.Message:
			require.NoError(t, ev.TopicPartition.Error)
			require.Equal(t, "test-topic", *ev.TopicPartition.Topic)
			var visitedMessage struct {
				ShortCode string    `json:"short_code"`
				VisitedAt time.Time `json:"visited_at"`
			}
			err := json.Unmarshal(ev.Value, &visitedMessage)
			require.NoError(t, err)
			require.Equal(t, "abcde", visitedMessage.ShortCode)
			require.Equal(t, visitedAt, visitedMessage.VisitedAt)
		}
	})
}
