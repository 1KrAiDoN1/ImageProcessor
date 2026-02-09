package main

import (
	"context"
	"imageprocessor/backend/internal/app/worker"
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
	log.Info("WORKER SERVICE started")

	config, err := config.LoadServiceConfig(log, configPath, dbPasswordPath, cloudAccessKey, cloudSecretKey)
	if err != nil {
		log.Error("Failed to load WORKER SERVICE config", zaplogger.Err(err))
		os.Exit(1)
	}

	worker, err := worker.NewWorker(ctx, config, log)
	if err != nil {
		log.Error("building WORKER SERVICE", zaplogger.Err(err))
	}

	if err := worker.Run(ctx); err != nil {
		log.Error("starting WORKER SERVICE", zaplogger.Err(err))

	}
	log.Info("Worker SERVICE stopped")
}
