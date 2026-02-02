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
)

func main() {
	ctx := context.Background()
	log := zaplogger.SetupLoggerWithLevel(zapcore.DebugLevel)
	log.Info("WORKER Service started")
	config, err := config.LoadServiceConfig(log, configPath, dbPasswordPath)
	if err != nil {
		log.Error("Failed to load WORKER service config", zaplogger.Err(err))
		os.Exit(1)
	}

	worker, err := worker.NewWorker(ctx, config, log)
	if err != nil {
		log.Error("building WORKER SERVICE", zaplogger.Err(err))
	}
	if err := worker.Run(); err != nil {
		log.Error("starting WORKER SERVICE", zaplogger.Err(err))

	}
	log.Info("Worker service stopped")
}
