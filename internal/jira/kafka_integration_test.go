//go:build integration

// Kafka integration test. Run with: go test -tags=integration ./internal/jira/
// Requires a working Docker daemon (testcontainers starts a Kafka broker).
package connector

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"

	"hse-2026-golang-project/internal/config"
)

func TestPublishProjectUpdated_Integration(t *testing.T) {
	ctx := context.Background()

	ctr, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ctr)

	brokers, err := ctr.Brokers(ctx)
	require.NoError(t, err)

	const topic = "etl-events"

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_1_0
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	// Pre-create the topic so the consumer can read from partition 0.
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	require.NoError(t, err)
	require.NoError(t, admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false))
	require.NoError(t, admin.Close())

	// Publish through the production code path (same producer setup as main.go).
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	require.NoError(t, err)
	defer producer.Close()

	s := NewGRPCServer(nil, nil, config.ProgramSettings{}, discardLogger(), producer, topic)
	require.NoError(t, s.publishProjectUpdated("ABC"))

	// Consume it back and verify the event contract.
	consumer, err := sarama.NewConsumer(brokers, cfg)
	require.NoError(t, err)
	defer consumer.Close()

	pc, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	require.NoError(t, err)
	defer pc.Close()

	select {
	case msg := <-pc.Messages():
		var m map[string]interface{}
		require.NoError(t, json.Unmarshal(msg.Value, &m))
		assert.Equal(t, "project_updated", m["event"])
		assert.Equal(t, "ABC", m["project"])
		assert.Equal(t, "success", m["status"])
		assert.NotEmpty(t, m["timestamp"])
	case <-time.After(20 * time.Second):
		t.Fatal("timed out waiting for kafka message")
	}
}
