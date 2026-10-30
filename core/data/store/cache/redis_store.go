package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samirkhairati/backtesting/core/data"
	"github.com/samirkhairati/backtesting/core/data/store"
	"github.com/samirkhairati/backtesting/core/logger"
	"go.uber.org/zap"
)

var _ store.Store = (*RedisStore)(nil)

type RedisStore struct {
	client   *redis.Client
	fallback store.Store
}

func NewRedisStore(client *redis.Client, fallback store.Store) store.Store {
	return &RedisStore{
		client:   client,
		fallback: fallback,
	}
}

func (r *RedisStore) fetchAndCacheSymbol(ctx context.Context, symbol string) ([]data.Bar, error) {
	key := fmt.Sprintf("bars:%s", symbol)

	// 1. Try Redis Cache
	val, err := r.client.Get(ctx, key).Bytes()
	if err == nil {
		var bars []data.Bar
		if err := json.Unmarshal(val, &bars); err == nil {
			logger.Log.Debug("Cache hit for symbol", zap.String("symbol", symbol), zap.Int("bars", len(bars)))
			return bars, nil
		}
		logger.Log.Warn("Failed to unmarshal cached bars, falling back to DB", zap.Error(err), zap.String("symbol", symbol))
	}

	logger.Log.Info("Cache miss for symbol, fetching full history from DB", zap.String("symbol", symbol))

	// 2. Cache Miss: Fetch ALL bars from Postgres
	iter, err := r.fallback.GetBars(ctx, symbol)
	if err != nil {
		return nil, err
	}

	var bars []data.Bar
	for iter.Next() {
		bars = append(bars, iter.Bar())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	// 3. Cache to Redis. Never cache an empty result: it would hide data that arrives later.
	// Historical bars only change when ingestion adds them, and ingestion deletes the key
	// when it does, so a long TTL is safe.
	if len(bars) == 0 {
		return bars, nil
	}
	if dataBytes, err := json.Marshal(bars); err == nil {
		if setErr := r.client.Set(ctx, key, dataBytes, 7*24*time.Hour).Err(); setErr != nil {
			logger.Log.Error("Failed to set cache in Redis", zap.Error(setErr), zap.String("symbol", symbol))
		}
	} else {
		logger.Log.Error("Failed to marshal bars for Redis", zap.Error(err), zap.String("symbol", symbol))
	}

	return bars, nil
}

func (r *RedisStore) GetBars(ctx context.Context, symbol string) (data.BarIterator, error) {
	allBars, err := r.fetchAndCacheSymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return data.NewMemoryBarIterator(allBars), nil
}

func (r *RedisStore) GetMultiBars(ctx context.Context, symbols []string) (data.MultiBarIterator, error) {
	barsPerSymbol := make(map[string][]data.Bar)
	for _, sym := range symbols {
		allBars, err := r.fetchAndCacheSymbol(ctx, sym)
		if err != nil {
			return nil, err
		}
		barsPerSymbol[sym] = allBars
	}
	return data.NewMemoryMultiBarIterator(symbols, barsPerSymbol), nil
}

func parseTime(tStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, tStr)
	if err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", tStr)
}

func (r *RedisStore) GetMultiBarsWithTimeframe(ctx context.Context, symbols []string, start, end string) (data.MultiBarIterator, error) {
	startTime, err := parseTime(start)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	endTime, err := parseTime(end)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	barsPerSymbol := make(map[string][]data.Bar)
	for _, sym := range symbols {
		allBars, err := r.fetchAndCacheSymbol(ctx, sym)
		if err != nil {
			return nil, err
		}

		// In-memory filtering
		var filtered []data.Bar
		for _, b := range allBars {
			if !b.Timestamp.Before(startTime) && !b.Timestamp.After(endTime) {
				filtered = append(filtered, b)
			}
		}
		barsPerSymbol[sym] = filtered
	}

	return data.NewMemoryMultiBarIterator(symbols, barsPerSymbol), nil
}
