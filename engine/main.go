package main

import (
	"context"
	"log"

	"github.com/samirkhairati/backtesting/engine/data/store/postgres"
	"github.com/samirkhairati/backtesting/engine/db"
	"github.com/samirkhairati/backtesting/engine/metrics"
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

	strategies := map[string]strategy.Strategy{
		"Buy and Hold":          strategy.NewBuyAndHoldStrategy(),
		"SMA Cross (50, 200)":   strategy.NewSMACrossStrategy(50, 200),
		"MACD (12, 26, 9)":      strategy.NewMACDStrategy(),
		"RSI Mean Reversion":    strategy.NewMeanReversionStrategy(),
		"Bollinger Bands":       strategy.NewBollingerBandsStrategy(),
	}

	results := make(map[string]*metrics.PerformanceReport)

	for name, strat := range strategies {
		log.Printf("\n--- Running Strategy: %s ---", name)
		
		iter, err := store.GetBars(ctx, symbol)
		if err != nil {
			log.Fatalf("Failed to get bars for %s: %v", symbol, err)
		}

		simEngine := simulator.NewEngine(iter, strat)
		
		if err := simEngine.Run(); err != nil {
			log.Fatalf("Simulation encountered an error on %s: %v", name, err)
		}
		
		results[name] = metrics.GenerateReport(simEngine.Portfolio)
	}

	log.Println("\n\n======================================")
	log.Println("       FINAL SIMULATION RESULTS       ")
	log.Println("======================================")
	for name, report := range results {
		log.Printf("\n>>> STRATEGY: %s <<<\n%s", name, report.String())
	}
}
