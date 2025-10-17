# Services

Слой для работы с бизнес-логикой и внешними системами.

**Назначение:** Services предоставляют методы для работы с данными и внешними системами. Они инкапсулируют:
- Работу с репозиториями (получение данных из БД)
- Бизнес-логику обработки данных
- Интеграцию с внешними сервисами (email, SMS, API)

Все зависимости передаются через **интерфейсы**, не конкретные типы

- ✅ `func New(repo interfaces.Repository) interfaces.Service`
- ❌ `func New(repo *repositories.UserRepository) *UserProfileService`

## Пример 1: Domain Service (работа с БД через интерфейсы)

```go
package services

import (
	"context"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/SmirnovND/gobase/internal/interfaces"
)

// ✅ Приватная структура!
type userProfileService struct {
	userRepo interfaces.UserRepository  // ← интерфейс, не конкретный тип!
}

// ✅ Конструктор принимает интерфейс, возвращает интерфейс!
func NewUserProfileService(userRepo interfaces.UserRepository) interfaces.UserService {
	return &userProfileService{
		userRepo: userRepo,
	}
}

// GetUser - получить пользователя по ID (обертка над репозиторием)
func (s *userProfileService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// CreateUser - создать пользователя с дополнительной бизнес-логикой
func (s *userProfileService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
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
	"github.com/SmirnovND/gobase/internal/interfaces"
	"net/smtp"
)

// ✅ Приватная структура!
type emailService struct {
	smtpHost string
	smtpPort string
	from     string
	password string
}

// ✅ Возвращаем интерфейс!
func NewEmailService(host, port, from, password string) interfaces.EmailService {
	return &emailService{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		password: password,
	}
}

func (s *emailService) SendEmail(ctx context.Context, to, subject, body string) error {
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
	
	// External сервисы (с кастомной логикой инициализации)
	c.container.Provide(func(cfg interfaces.ConfigServer) interfaces.EmailService {
		return services.NewEmailService(
			cfg.GetSmtpHost(),
			cfg.GetSmtpPort(),
			cfg.GetSmtpUser(),
			cfg.GetSmtpPassword(),
		)
	})
}
```

**Важно:** Dig **автоматически разрешит зависимости** (видит `interfaces.ConfigServer` в параметрах)

## Как работает в Usecase

```go
package usecases

import (
	"context"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/SmirnovND/gobase/internal/interfaces"
)

// ✅ Приватная структура!
type userUsecase struct {
	userService  interfaces.UserService   // ← интерфейс!
	emailService interfaces.EmailService  // ← интерфейс!
}

// ✅ Конструктор принимает интерфейсы, возвращает интерфейс!
func NewUserUsecase(
	userService interfaces.UserService,
	emailService interfaces.EmailService,
) interfaces.UserUsecase {
	return &userUsecase{
		userService:  userService,
		emailService: emailService,
	}
}

// Usecase работает ТОЛЬКО через сервисы (через интерфейсы)!
func (uc *userUsecase) RegisterUser(ctx context.Context, name, email string) (*domain.User, error) {
	// Получаем данные через сервис, не через репозиторий
	user, err := uc.userService.CreateUser(ctx, name, email)
	if err != nil {
		return nil, err
	}
	
	// Используем другой сервис
	_ = uc.emailService.SendEmail(ctx, email, "Welcome", "Thanks for registering!")
	
	return user, nil
}
```

## Тестирование Service (с использованием моков)

```go
package services

import (
	"context"
	"testing"
	"github.com/SmirnovND/gobase/internal/domain"
	"github.com/SmirnovND/gobase/internal/interfaces"
)

func TestUserProfileService_CreateUser(t *testing.T) {
	// ✅ Используем встроенный мок!
	mockRepo := &interfaces.MockUserRepository{
		CreateFunc: func(ctx context.Context, user *domain.User) error {
			user.ID = 1  // Имитируем присвоение ID
			return nil
		},
	}
	
	service := NewUserProfileService(mockRepo)
	user, err := service.CreateUser(context.Background(), "John", "john@example.com")
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if user.Name != "John" {
		t.Errorf("expected name 'John', got %s", user.Name)
	}
}
```