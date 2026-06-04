package main

import (
	"context"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/data/store/postgres"
	"github.com/samirkhairati/backtesting/engine/simulator"
)

func main() {
	log.Println("Starting Backtesting Engine...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Run migrations
	log.Println("Running database migrations...")
	m, err := migrate.New(
		"file://db/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}
	
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations applied successfully!")

	ctx := context.Background()

	// Connect to database using pgxpool for connection pooling
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Successfully connected to the PostgreSQL database!")

	// Initialize the PostgreSQL data store
	store := postgres.NewStore(pool)

	// In a real scenario, the symbol would come from config/request
	symbol := "AAPL"
	log.Printf("Fetching historical data for %s...", symbol)

	iter, err := store.GetBars(ctx, symbol)
	if err != nil {
		log.Fatalf("Failed to get bars for %s: %v", symbol, err)
	}

	// Instantiate our dummy strategy
	strat := &PrintingStrategy{}

	// Initialize the Market Replay Engine
	simEngine := simulator.NewEngine(iter, strat)

	log.Println("Starting Market Replay Simulation...")
	if err := simEngine.Run(); err != nil {
		log.Fatalf("Simulation encountered an error: %v", err)
	}

	log.Printf("Simulation finished! Processed %d bars. Final simulated clock: %v", strat.count, simEngine.CurrentTime.Format("2006-01-02"))
}

// PrintingStrategy is a dummy strategy to verify the replay engine works.
type PrintingStrategy struct {
	count int
}

func (s *PrintingStrategy) OnBar(bar data.Bar) {
	s.count++
	
	// Print the first few bars as a sanity check
	if s.count <= 3 {
		log.Printf("[Strategy] Received Bar %d: %s [%v] O:%.2f H:%.2f L:%.2f C:%.2f V:%d", 
			s.count, bar.Symbol, bar.Timestamp.Format("2006-01-02"), 
			bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	}
}
