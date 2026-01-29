package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type StatisticsRepository struct {
	pool *pgxpool.Pool
}

func NewStatisticsRepository(pool *pgxpool.Pool) *StatisticsRepository {
	return &StatisticsRepository{
		pool: pool,
	}
}
