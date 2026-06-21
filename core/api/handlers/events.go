package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samirkhairati/backtesting/core/kafka"
	"github.com/samirkhairati/backtesting/core/logger"
	"github.com/samirkhairati/backtesting/core/models"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan models.ExperimentEvent
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan models.ExperimentEvent),
	}
}

func (b *EventBus) Subscribe(experimentID string) chan models.ExperimentEvent {
	ch := make(chan models.ExperimentEvent, 16)
	b.mu.Lock()
	b.subscribers[experimentID] = append(b.subscribers[experimentID], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBus) Unsubscribe(experimentID string, ch chan models.ExperimentEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[experimentID]
	for i, sub := range subs {
		if sub == ch {
			b.subscribers[experimentID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(b.subscribers[experimentID]) == 0 {
		delete(b.subscribers, experimentID)
	}
}

func (b *EventBus) Dispatch(event models.ExperimentEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[event.ExperimentID] {
		select {
		case ch <- event:
		default:
			// Drop if full to avoid blocking
		}
	}
}

type EventStreamConsumer struct {
	Bus      *EventBus
	consumer *kafka.Consumer
}

func NewEventStreamConsumer(brokers []string, bus *EventBus) (*EventStreamConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "api-group", "event-stream")
	if err != nil {
		return nil, err
	}
	return &EventStreamConsumer{
		Bus:      bus,
		consumer: consumer,
	}, nil
}

func (ec *EventStreamConsumer) Run(ctx context.Context) {
	ec.consumer.Run(ctx, ec.handle)
}

func (ec *EventStreamConsumer) handle(ctx context.Context, record *kgo.Record) error {
	var event models.ExperimentEvent
	if err := ec.consumer.Unmarshal(record, &event); err != nil {
		logger.Log.Error("Failed to unmarshal event", zap.Error(err))
		return nil
	}
	ec.Bus.Dispatch(event)
	return nil
}

func (ec *EventStreamConsumer) Close() {
	ec.consumer.Close()
}

type EventHandler struct {
	Bus *EventBus
}

func (h *EventHandler) StreamEvents(c *gin.Context) {
	experimentID := c.Param("id")
	ch := h.Bus.Subscribe(experimentID)
	defer h.Bus.Unsubscribe(experimentID, ch)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	clientGone := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case <-ticker.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			return true
		case event, ok := <-ch:
			if !ok {
				return false
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			return true
		}
	})
}
