package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	client *kgo.Client
}

func NewConsumer(brokers []string, groupID string, topics ...string) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(), // Commit manually after successful processing
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{client: client}, nil
}

func (c *Consumer) Run(ctx context.Context, handler func(ctx context.Context, record *kgo.Record) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}

		errs := fetches.Errors()
		for _, err := range errs {
			log.Printf("Consumer poll error: %v", err.Err)
		}

		fetches.EachRecord(func(record *kgo.Record) {
			if err := handler(ctx, record); err != nil {
				log.Printf("Failed to handle record topic=%s partition=%d offset=%d: %v", record.Topic, record.Partition, record.Offset, err)
				return // Do not commit, so it can be re-processed if logic is updated
			}

			if err := c.client.CommitRecords(ctx, record); err != nil {
				log.Printf("Failed to commit record: %v", err)
			}
		})
	}
}

func (c *Consumer) Unmarshal(record *kgo.Record, v any) error {
	return json.Unmarshal(record.Value, v)
}

func (c *Consumer) Close() {
	c.client.Close()
}
