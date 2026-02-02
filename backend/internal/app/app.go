package app

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/broker/kafka"
	"imageprocessor/backend/internal/config"
	httpserver "imageprocessor/backend/internal/http-server"
	"imageprocessor/backend/internal/http-server/handler"
	"imageprocessor/backend/internal/repository/cloud/s3"
	"imageprocessor/backend/internal/repository/postgres"
	imageservice "imageprocessor/backend/internal/service/image_service"
	statsservice "imageprocessor/backend/internal/service/stats_service"

	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type App struct {
	cfg    *config.ServiceConfig
	log    *zap.Logger
	server *httpserver.Server
}

func NewApp(ctx context.Context, cfg *config.ServiceConfig, log *zap.Logger) (*App, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	storage, err := postgres.NewDatabase(ctx, cfg.DbConfig.DBConn)
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Connected to database", zap.String("dsn", cfg.DbConfig.DBConn))

	dbPool := storage.GetPool()
	defer func() {
		log.Info("Closing database connection...")
		dbPool.Close()
		log.Info("Database connection closed")
	}()

	s3Client, err := s3.NewS3CloudStorage(context.Background(), cfg.CloudStorageConfig, log)
	if err != nil {
		log.Fatal("Failed to initialize S3 client", zap.Error(err))
	}
	log.Info("S3 client initialized")

	// Инициализация Kafka producer
	kafkaProducer := kafka.NewProducer(cfg.BrokerConfig, log)
	defer kafkaProducer.Close()

	// Проверка и создание топика Kafka
	if err := kafka.EnsureTopicExists(cfg.BrokerConfig, log); err != nil {
		log.Warn("Failed to ensure Kafka topic exists", zap.Error(err))
	}

	// Инициализация репозиториев
	imageRepo := postgres.NewImageRepository(dbPool, log)
	statsRepo := postgres.NewStatisticsRepository(dbPool, log)

	// Инициализация сервисов
	imageService := imageservice.NewImageService(
		imageRepo,
		s3Client,
		kafkaProducer,
		log,
		cfg.CloudStorageConfig.Bucket,
	)

	statsService := statsservice.NewStatsService(statsRepo, log)

	// Инициализация хэндлеров
	handlers := handler.NewHandler(log, imageService, statsService)

	server := httpserver.NewServer(log, cfg, handlers)
	return &App{
		cfg:    cfg,
		log:    log,
		server: server,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serverDone := make(chan error, 1)
	go func() {
		a.log.Info("Starting HTTP server...")
		if err := a.server.Run(); err != nil {
			serverDone <- err
		}
		close(serverDone)
	}()

	select {
	case sig := <-sigChan:
		a.log.Info("Received shutdown signal", zap.String("signal", sig.String()))

		cancel()

		if err := a.server.Shutdown(ctx); err != nil {
			a.log.Error("Failed to shutdown HTTP server", zap.Error(err))
		}

		a.log.Info("Waiting for goroutines to finish...")

		a.log.Info("Application gracefully shut down")
		return nil

	case err := <-serverDone:

		cancel()

		if err != nil {
			a.log.Error("HTTP server stopped with error", zap.Error(err))
			return err
		}
		return nil
	}
}
