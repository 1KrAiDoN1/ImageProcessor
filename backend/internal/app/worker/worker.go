package worker

import (
	"imageprocessor/backend/internal/broker"
	"imageprocessor/backend/internal/config"
	imageprocessor "imageprocessor/backend/internal/domain/image_processor"
	"imageprocessor/backend/internal/repository/cloud"
	imageservice "imageprocessor/backend/internal/service/image_service"
	"sync"

	"go.uber.org/zap"
)

type Worker struct {
	cfg          *config.ServiceConfig
	log          *zap.Logger
	consumer     broker.ConsumerMessageBrokerInterface
	processor    imageprocessor.ImageProcessorInterface
	imageRepo    imageservice.ImageRepositoryInterface
	cloudStorage cloud.CloudStorageInterface
	numWorkers   int
	wg           *sync.WaitGroup
	stopChan     chan struct{}
}

func NewWorker(cfg config.ServiceConfig) *Worker {
	return &Worker{}
}
