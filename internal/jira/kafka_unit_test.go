package connector

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/config"
)

func testProducerConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true // required by sarama's mock producer
	return cfg
}

func TestPublishProjectUpdated_SendsEvent(t *testing.T) {
	prod := mocks.NewSyncProducer(t, testProducerConfig())
	prod.ExpectSendMessageWithCheckerFunctionAndSucceed(func(val []byte) error {
		var m map[string]interface{}
		if err := json.Unmarshal(val, &m); err != nil {
			return err
		}
		if m["event"] != "project_updated" {
			return fmt.Errorf("event = %v, want project_updated", m["event"])
		}
		if m["project"] != "ABC" {
			return fmt.Errorf("project = %v, want ABC", m["project"])
		}
		if m["status"] != "success" {
			return fmt.Errorf("status = %v, want success", m["status"])
		}
		if _, ok := m["timestamp"].(string); !ok {
			return fmt.Errorf("timestamp missing or not a string")
		}
		return nil
	})

	s := NewGRPCServer(nil, nil, config.ProgramSettings{}, discardLogger(), prod, "events")
	require.NoError(t, s.publishProjectUpdated(discardLogger(), "ABC"))
	require.NoError(t, prod.Close())
}

func TestPublishProjectUpdated_ProducerError(t *testing.T) {
	prod := mocks.NewSyncProducer(t, testProducerConfig())
	prod.ExpectSendMessageAndFail(sarama.ErrBrokerNotAvailable)

	s := NewGRPCServer(nil, nil, config.ProgramSettings{}, discardLogger(), prod, "events")
	err := s.publishProjectUpdated(discardLogger(), "ABC")
	require.Error(t, err)
	require.NoError(t, prod.Close())
}
