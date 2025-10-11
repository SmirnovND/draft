# Use Cases

Слой бизнес-логики приложения.

## Пример use case

```go
package usecases

import (
	"context"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/SmirnovND/gobase/internal/repositories"
)

type UserUsecase struct {
	userRepo *repositories.UserRepository
}

func NewUserUsecase(userRepo *repositories.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *UserUsecase) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	user := &domain.User{
		Name:  name,
		Email: email,
	}
	
	err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}
```

## Регистрация в DI контейнере

В файле `internal/container/container.go`:

```go
func (c *Container) provideUsecase() {
	c.container.Provide(usecases.NewUserUsecase)
}
```