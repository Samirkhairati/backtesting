package main

import (
	"github.com/samirkhairati/backtesting/core/api/router"
	"github.com/samirkhairati/backtesting/core/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize structured logger
	logger.Init()
	defer logger.Sync()

	logger.Log.Info("Starting API Server on port 8080...")

	r := router.SetupRouter()

	if err := r.Run(":8080"); err != nil {
		logger.Log.Fatal("Failed to start API server", zap.Error(err))
	}
}
