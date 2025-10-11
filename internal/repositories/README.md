# Repositories

Слой для работы с базой данных.

## Пример репозитория

```go
package repositories

import (
	"context"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (name, email) 
		VALUES ($1, $2) 
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.Name, user.Email).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
```

## Регистрация в DI контейнере

В файле `internal/container/container.go`:

```go
func (c *Container) provideRepo() {
	c.container.Provide(repositories.NewUserRepository)
}
```