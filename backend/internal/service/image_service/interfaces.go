package imageservice

import "context"

type ImageRepositoryInterface interface {
	Create(ctx context.Context, image *Image) error
	GetByID(ctx context.Context, id string) (*Image, error)
	Update(ctx context.Context, image *Image) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status ImageStatus) error
	UpdateVersions(ctx context.Context, id string, versions map[VersionType]string) error
}
