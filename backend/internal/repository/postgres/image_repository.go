package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type ImageRepository struct {
	pool *pgxpool.Pool
}

func NewImageRepository(pool *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{
		pool: pool,
	}
}
