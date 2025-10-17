# Services

Слой для работы с бизнес-логикой и внешними системами.

**Назначение:** Services предоставляют методы для work с данными и сервисами. Они инкапсулируют:
- Работу с репозиториями (получение данных из БД)
- Бизнес-логику обработки данных
- Интеграцию с внешними сервисами (email, SMS, API)

**Важно:** Usecases работают **только** через Services, никогда напрямую через Repositories.

## Пример 1: Domain Service (работа с БД)

```go
package services

import (
	"context"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/SmirnovND/gobase/internal/repositories"
)

type UserProfileService struct {
	userRepo *repositories.UserRepository
}

func NewUserProfileService(userRepo *repositories.UserRepository) *UserProfileService {
	return &UserProfileService{
		userRepo: userRepo,
	}
}

// GetUser - получить пользователя по ID (обертка над репозиторием)
func (s *UserProfileService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// CreateUser - создать пользователя с дополнительной бизнес-логикой
func (s *UserProfileService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
	// Можно добавить бизнес-логику перед созданием
	user := &domain.User{
		Name:  name,
		Email: email,
	}
	
	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	
	// Можно добавить логику после создания
	return user, nil
}
```

**Когда использовать:** Когда нужна работа с данными из БД

## Пример 2: External Service (интеграция с внешними системами)

```go
package services

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailService struct {
	smtpHost string
	smtpPort string
	from     string
	password string
}

func NewEmailService(host, port, from, password string) *EmailService {
	return &EmailService{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		password: password,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	auth := smtp.PlainAuth("", s.from, s.password, s.smtpHost)
	
	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))
	
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
```

**Когда использовать:** Когда нужна интеграция с внешними системами

## Регистрация в DI контейнере

В файле `internal/container/container.go`:

```go
func (c *Container) provideService() {
	// Domain сервисы (работают с репозиториями)
	c.container.Provide(services.NewUserProfileService)
	
	// External сервисы
	c.container.Provide(func() *services.EmailService {
		return services.NewEmailService(
			"smtp.gmail.com",
			"587",
			"your-email@gmail.com",
			"your-password",
		)
	})
}
```

## Как работает в Usecase

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

// Usecase работает ТОЛЬКО через сервисы!
func (uc *UserUsecase) RegisterUser(ctx context.Context, name, email string) (*domain.User, error) {
	// Получаем данные через сервис, не через репозиторий
	user, err := uc.userProfileService.CreateUser(ctx, name, email)
	if err != nil {
		return nil, err
	}
	
	// Используем другой сервис
	_ = uc.emailService.SendEmail(ctx, email, "Welcome", "Thanks for registering!")
	
	return user, nil
}
```