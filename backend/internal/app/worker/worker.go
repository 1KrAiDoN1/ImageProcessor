package worker

import (
	"context"
	"fmt"
	"imageprocessor/backend/internal/broker"
	"imageprocessor/backend/internal/broker/kafka"
	"imageprocessor/backend/internal/config"
	"imageprocessor/backend/internal/domain/entity"
	"imageprocessor/backend/internal/repository/cloud"
	"imageprocessor/backend/internal/repository/cloud/s3"
	"imageprocessor/backend/internal/repository/postgres"
	"imageprocessor/backend/internal/service/image_processor/processor"
	imageservice "imageprocessor/backend/internal/service/image_service"
	statsservice "imageprocessor/backend/internal/service/stats_service"
	workerservice "imageprocessor/backend/internal/service/worker"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Worker struct {
	cfg          *config.ServiceConfig
	log          *zap.Logger
	consumer     broker.ConsumerMessageBrokerInterface
	processor    processor.ImageProcessorImpl
	imageRepo    imageservice.ImageRepositoryInterface
	statsRepo    statsservice.StatsRepositoryInterface
	cloudStorage cloud.CloudStorageInterface
	workService  workerServiceInterface
	numWorkers   int
	wg           *sync.WaitGroup
	stopChan     chan struct{}
}

func NewWorker(ctx context.Context, cfg *config.ServiceConfig, log *zap.Logger) (*Worker, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := &sync.WaitGroup{}

	stopChan := make(chan struct{})

	storage, err := postgres.NewDatabase(ctx, cfg.DbConfig.DBConn)
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	log.Info("Connected to database", zap.String("dsn", cfg.DbConfig.DBConn))

	dbpool := storage.GetPool()
	defer func() {
		log.Info("Closing database connection...")
		dbpool.Close()
		log.Info("Database connection closed")
	}()

	// Инициализация S3 клиента
	s3Client, err := s3.NewS3CloudStorage(context.Background(), cfg.CloudStorageConfig, log)
	if err != nil {
		log.Fatal("Failed to initialize S3 client", zap.Error(err))
	}
	log.Info("S3 client initialized")

	// Инициализация Kafka consumer
	kafkaConsumer := kafka.NewConsumer(cfg.BrokerConfig, log)
	defer kafkaConsumer.Close()

	// Проверка и создание топика Kafka
	if err := kafka.EnsureTopicExists(cfg.BrokerConfig, log); err != nil {
		log.Warn("Failed to ensure Kafka topic exists", zap.Error(err))
	}
	imageRepo := postgres.NewImageRepository(dbpool, log)
	statsRepo := postgres.NewStatisticsRepository(dbpool, log)

	statsService := statsservice.NewStatsService(statsRepo, log)

	imageProcessor := processor.NewImageProcessor(log)

	workerService := workerservice.NewWorkerService(
		imageProcessor,
		s3Client,
		imageRepo,
		statsService,
		log,
		cfg.CloudStorageConfig.Bucket,
	)

	return &Worker{
		cfg:          cfg,
		log:          log,
		consumer:     kafkaConsumer,
		processor:    *imageProcessor,
		imageRepo:    imageRepo,
		statsRepo:    statsRepo,
		cloudStorage: s3Client,
		workService:  workerService,
		numWorkers:   cfg.WorkerConfig.NumWorkers,
		wg:           wg,
		stopChan:     stopChan,
	}, nil

}

func (w *Worker) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskHandler := func(ctx context.Context, task *entity.ProcessingTask) error {
		w.log.Info("Received task from Kafka",
			zap.String("taskId", task.ID),
			zap.String("imageId", task.ImageID),
		)

		// Обрабатываем задачу с повторными попытками
		err := w.workService.ProcessTaskWithRetry(ctx, task, w.cfg.WorkerConfig.RetryAttempts)
		if err != nil {
			w.log.Error("Failed to process task",
				zap.Error(err),
				zap.String("taskId", task.ID),
			)
			return err
		}

		return nil
	}
	for i := 0; i < w.numWorkers; i++ {
		w.wg.Add(1)
		workerID := i + 1

		go func(id int) {
			defer w.wg.Done()
			w.log.Info("Worker started", zap.Int("workerId", id))

			// Каждый воркер читает сообщения
			err := w.consumer.Start(ctx, taskHandler)
			if err != nil && err != context.Canceled {
				w.log.Error("Worker stopped with error",
					zap.Int("workerId", id),
					zap.Error(err),
				)
			} else {
				w.log.Info("Worker stopped gracefully", zap.Int("workerId", id))
			}
		}(workerID)
	}

	w.log.Info("All workers started successfully")

	// Мониторинг статистики Kafka consumer
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msg, byt := w.consumer.Stats()

				w.log.Info("Kafka consumer stats",
					zap.Int64("messages", msg),
					zap.Int64("bytes", byt),
				)
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	w.log.Info("Shutting down workers...")
	cancel()

	// Ждем завершения всех воркеров
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	// Таймаут для graceful shutdown
	select {
	case <-done:
		w.log.Info("All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		w.log.Warn("Shutdown timeout exceeded, forcing exit")
	}
	return nil
}

type workerServiceInterface interface {
	ProcessTask(ctx context.Context, task *entity.ProcessingTask) error
	ProcessTaskWithRetry(ctx context.Context, task *entity.ProcessingTask, maxRetries int) error
}
