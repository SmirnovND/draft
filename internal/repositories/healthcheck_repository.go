package repositories

import (
	"context"
	"github.com/SmirnovND/gobase/internal/interfaces"
	"github.com/jmoiron/sqlx"
)

type healthcheckRepository struct {
	db *sqlx.DB
}

func NewHealthcheckRepository(db *sqlx.DB) interfaces.HealthcheckRepository {
	return &healthcheckRepository{
		db: db,
	}
}

// Ping проверяет подключение к БД
func (r *healthcheckRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
