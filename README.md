# Backtesting Engine

## Commands

**Start API Development Server (with Auto-Reload)**
```bash
docker compose up api-dev
```

**Run All Tests (with DB & Migrations)**
```bash
docker compose --profile test run --rm test
```

**Generate a New Database Migration**
```bash
migrate create -ext sql -dir core/db/migrations -seq your_migration_name
```

**Reset the Database Completely**
```bash
docker compose down -v
```

**Start Platform with Scaled Workers**
```bash
# By default, Docker Compose runs 1 instance of each service.
# You can scale backtesting workers (worker-dev) and ingestion workers (ingestion-dev) like this:
docker compose --profile dev up -d --scale worker-dev=5 --scale ingestion-dev=3
```
