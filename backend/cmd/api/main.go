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
	configPath     = "backend/internal/config/config.yaml"
	dbPasswordPath = "DB_PASSWORD"
	cloudAccessKey = "CLOUD_ACCESS_KEY"
	cloudSecretKey = "CLOUD_SECRET_KEY"
)

func main() {
	ctx := context.Background()

	log := zaplogger.SetupLoggerWithLevel(zapcore.DebugLevel)
	log.Info("API Service started")

	config, err := config.LoadServiceConfig(log, configPath, dbPasswordPath, cloudAccessKey, cloudSecretKey)
	if err != nil {
		log.Error("Failed to load API service config", zaplogger.Err(err))
		os.Exit(1)
	}

	app, err := app.NewApp(ctx, config, log)
	if err != nil {
		log.Error("Failed to create application service", zaplogger.Err(err))
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		log.Error("Failed to Run application service", zaplogger.Err(err))
		os.Exit(1)
	}

}
