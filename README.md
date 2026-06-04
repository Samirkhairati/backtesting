# Backtesting Engine

A highly accurate, event-driven market replay engine built in Go.

## Overview

This project is designed with a strict "no lookahead bias" architecture. Market data is fed sequentially into a `Strategy` interface to guarantee that strategies only ever have access to currently available historical data, perfectly simulating a live market environment.

## Development Setup

The local environment is built entirely on `docker-compose`. Ensure Docker Desktop is installed and running.

### 1. Start the Environment

To start the engine and run the main entrypoint:

```bash
docker-compose run --rm engine-dev sh -c "go mod tidy && go run main.go"
```

This will automatically:
1. Spin up the PostgreSQL database in the background.
2. Run all database schema migrations and seed the data.
3. Start the market replay simulation.

### 2. Run the Unit Tests

To execute all unit tests across the Go workspace (using a mock dataset without needing to connect to the database), run:

```bash
docker-compose run --rm engine-dev go test ./... -v
```

## Architecture Notes

* **Data Consistency:** All historical data is rigidly validated against duplicates, negative prices, and negative volume upon ingestion.
* **Database:** `golang-migrate` is used to enforce schema consistency, and `pgxpool` manages PostgreSQL connections efficiently.
* **Simulated Clock:** The `simulator.Engine` is responsible for ticking the internal clock forward one bar at a time.
