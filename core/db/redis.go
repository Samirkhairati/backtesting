package db

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/samirkhairati/backtesting/core/logger"
	"go.uber.org/zap"
)

// SetupRedis connects to Redis and returns the client
func SetupRedis(ctx context.Context) *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	// Ping to ensure connection
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Log.Fatal("Failed to connect to Redis", zap.Error(err), zap.String("url", redisURL))
	}

	logger.Log.Info("Successfully connected to Redis!", zap.String("url", redisURL))
	return client
}
