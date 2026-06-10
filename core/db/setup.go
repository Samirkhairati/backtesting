package db

import (
	"context"
	"github.com/samirkhairati/backtesting/core/logger"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(ctx context.Context) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Log.Fatal("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Log.Sugar().Fatalf("Unable to connect to database: %v", err)
	}

	logger.Log.Info("Successfully connected to the PostgreSQL database!")
	return pool
}
