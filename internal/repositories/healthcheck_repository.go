package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type HealthcheckRepository interface {
	Ping(ctx context.Context) error
}

type healthcheckRepository struct {
	db *sqlx.DB
}

func NewHealthcheckRepository(db *sqlx.DB) HealthcheckRepository {
	return &healthcheckRepository{
		db: db,
	}
}

// Ping проверяет подключение к БД
func (r *healthcheckRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}