# Use Cases

Слой бизнес-логики приложения. Орхестрирует работу сервисов.

**Ключевое правило:** Usecase получает сервисы через DI, **НИКОГДА** репозитории напрямую!

## Архитектура потока данных

```
Usecase → Services → Repositories → Database
```

## Пример use case

```go
package usecases

import (
	"context"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/SmirnovND/gobase/internal/services"
)

type UserUsecase struct {
	userProfileService *services.UserProfileService
	emailService       *services.EmailService
}

func NewUserUsecase(
	userProfileService *services.UserProfileService,
	emailService *services.EmailService,
) *UserUsecase {
	return &UserUsecase{
		userProfileService: userProfileService,
		emailService:       emailService,
	}
}

// GetUser - получить пользователя
func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return uc.userProfileService.GetUser(ctx, id)
}

// CreateUser - создать пользователя с отправкой приветственного письма
func (uc *UserUsecase) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	// Бизнес-логика: создание пользователя через сервис
	user, err := uc.userProfileService.CreateUser(ctx, name, email)
	if err != nil {
		return nil, err
	}
	
	// Отправка приветственного письма
	_ = uc.emailService.SendEmail(
		ctx,
		email,
		"Welcome!",
		"Thank you for registering!",
	)
	
	return user, nil
}
```

## Регистрация в DI контейнере

В файле `internal/container/container.go`:

```go
func (c *Container) provideUsecase() {
	// Usecase получает сервисы, не репозитории!
	c.container.Provide(usecases.NewUserUsecase)
}
```

## Лучшие практики

✅ **Правильно:**
- Usecase работает с Services
- Service инкапсулирует работу с Repository
- Каждый Service отвечает за одну область (UserProfileService, EmailService)

❌ **Неправильно:**
- Usecase напрямую использует Repository
- Usecase напрямую работает с DB
- Service вызывает другие Usecases