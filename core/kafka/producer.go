package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Publisher interface {
	Publish(ctx context.Context, topic, key string, v any) error
}

type Producer struct {
	client *kgo.Client
}

func NewProducer(brokers []string) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ProducerBatchCompression(kgo.NoCompression()),
	)
	if err != nil {
		return nil, err
	}
	return &Producer{client: client}, nil
}

func (p *Producer) Publish(ctx context.Context, topic, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}

	results := p.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
