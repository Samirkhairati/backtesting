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
