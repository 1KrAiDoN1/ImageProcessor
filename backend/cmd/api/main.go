package main

import (
	"context"
	"imageprocessor/backend/internal/app"
	"imageprocessor/backend/internal/config"
	"imageprocessor/backend/pkg/lib/logger/zaplogger"
	"os"

	"go.uber.org/zap/zapcore"
)

const (
	configPath     = "internal/config/config.yaml"
	dbPasswordPath = "DB_PASSWORD"
)

func main() {
	ctx := context.Background()
	log := zaplogger.SetupLoggerWithLevel(zapcore.DebugLevel)
	log.Info("Service started")
	config, err := config.LoadServiceConfig(log, configPath, dbPasswordPath)
	if err != nil {
		log.Error("Failed to load service config", zaplogger.Err(err))
		os.Exit(1)
	}

	if err := app.Run(ctx, log, config); err != nil {
		log.Error("Failed to Run application service", zaplogger.Err(err))
		os.Exit(1)
	}

}
