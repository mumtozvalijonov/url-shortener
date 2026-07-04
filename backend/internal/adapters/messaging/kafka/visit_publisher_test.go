package kafka_test

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

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

		publisher := kafkaadapter.NewShortURLVisitedPublisher(producer, "test")

		err = publisher.Publish(ctx, domain.ShortURLVisited{})
		require.NoError(t, err)
	})

}
