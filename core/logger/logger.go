package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func init() {
	// Initialize a default no-op logger so tests don't panic if they forget to call Init()
	Log = zap.NewNop()
}

// Init initializes the global logger based on the environment (dev vs prod)
func Init() {
	var config zap.Config

	env := os.Getenv("GIN_MODE")
	if env == "release" {
		// Production: structured JSON logging
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Development: colorful console logging
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	Log = logger
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
