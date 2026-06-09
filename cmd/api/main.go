package main

import (
	"context"
	"os"

	"github.com/samirkhairati/backtesting/core/api/router"
	"github.com/samirkhairati/backtesting/core/db"
	"github.com/samirkhairati/backtesting/core/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize structured logger
	logger.Init()
	defer logger.Sync()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Log.Sugar().Infof("Starting API Server on port %s...", port)

	ctx := context.Background()
	dbPool := db.Setup(ctx)
	defer dbPool.Close()

	r := router.SetupRouter()

	if err := r.Run(":" + port); err != nil {
		logger.Log.Fatal("Failed to start API server", zap.Error(err))
	}
}
