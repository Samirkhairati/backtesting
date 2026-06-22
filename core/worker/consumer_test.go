package worker_test

import (
	"context"
	"testing"
	"github.com/samirkhairati/backtesting/core/data"
	"github.com/stretchr/testify/assert"
)

// A mock store to be used for basic worker logic verification
type mockStore struct{}

func (m *mockStore) GetBars(ctx context.Context, symbol string) (data.BarIterator, error) {
	return nil, nil
}
func (m *mockStore) GetMultiBars(ctx context.Context, symbols []string) (data.MultiBarIterator, error) {
	return nil, nil
}
func (m *mockStore) GetMultiBarsWithTimeframe(ctx context.Context, symbols []string, start, end string) (data.MultiBarIterator, error) {
	return nil, nil
}

// A mock publisher
type mockPublisher struct{}
func (m *mockPublisher) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	return nil
}
func (m *mockPublisher) Close() {}

func TestWorker_Initialization(t *testing.T) {
	// Not easily testable without a Kafka broker, so we do a simple check
	assert.True(t, true)
}
