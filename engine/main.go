package main

import (
	"context"
	"log"

	"github.com/samirkhairati/backtesting/engine/data/store/postgres"
	"github.com/samirkhairati/backtesting/engine/db"
	"github.com/samirkhairati/backtesting/engine/simulator"
	"github.com/samirkhairati/backtesting/engine/strategy"
)

func main() {
	log.Println("Starting Backtesting Engine...")
	ctx := context.Background()

	pool := db.Setup(ctx)
	defer pool.Close()

	store := postgres.NewStore(pool)
	symbol := "AAPL"
	log.Printf("Fetching historical data for %s...", symbol)

	iter, err := store.GetBars(ctx, symbol)
	if err != nil {
		log.Fatalf("Failed to get bars for %s: %v", symbol, err)
	}

	strat := strategy.NewSMACrossStrategy(50, 200)
	simEngine := simulator.NewEngine(iter, strat)

	log.Println("Starting Market Replay Simulation...")
	if err := simEngine.Run(); err != nil {
		log.Fatalf("Simulation encountered an error: %v", err)
	}

	log.Printf("Simulation finished! Final clock: %v", simEngine.CurrentTime.Format("2006-01-02"))
	log.Printf("Final Portfolio Value: $%.2f", simEngine.Portfolio.TotalValue())
}
