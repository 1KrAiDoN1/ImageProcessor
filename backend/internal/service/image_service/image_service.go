package imageservice

import "go.uber.org/zap"

type ImageService struct {
	imageRepo ImageRepositoryInterface
	log       *zap.Logger
}

func NewImageService(imageRepo ImageRepositoryInterface, log *zap.Logger) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		log:       log,
	}
}
