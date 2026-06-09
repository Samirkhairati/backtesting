package db

import (
	"context"
	"github.com/samirkhairati/backtesting/core/logger"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(ctx context.Context) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Log.Fatal("DATABASE_URL environment variable is not set")
	}

	logger.Log.Info("Running database migrations...")
	m, err := migrate.New("file://core/db/migrations", dbURL)
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to initialize migrations: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log.Sugar().Fatalf("Failed to run migrations: %v", err)
	}
	logger.Log.Info("Database migrations applied successfully!")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Log.Sugar().Fatalf("Unable to connect to database: %v", err)
	}

	logger.Log.Info("Successfully connected to the PostgreSQL database!")
	return pool
}
