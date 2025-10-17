# Архитектура проекта

## Обзор

Проект построен на принципах **Clean Architecture** с использованием **Dependency Injection** через Uber Dig.

```
┌─────────────────────────────────────────────────────────┐
│                     HTTP Layer                          │
│                  (Chi Router + Middleware)              │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   Controllers                           │
│          (Обработка HTTP запросов/ответов)             │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   Use Cases                             │
│              (Бизнес-логика приложения)                 │
│          (Орхестрирует работу сервисов)                 │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   Services                              │
│  (Domain Services + External Services + Some Services)                  │
│  - Инкапсулируют работу с Repositories                  │
│  - Интеграция с внешними системами                      │
└────────────┬───────────────────────┬────────────────────┘
             │                       │
┌────────────▼────────────┐ ┌───────▼────────────────────┐
│      Repositories       │ │  Other Service         │
│  (Работа с БД)          │ │  (Email, SMS, API)         │
└────────────┬────────────┘ └────────────────────────────┘
             │
┌────────────▼────────────────────────────────────────────┐
│                      Database                           │
│                   (PostgreSQL)                          │
└─────────────────────────────────────────────────────────┘
```

**Ключевое правило:** Use Cases работают **только** через Services, никогда напрямую через Repositories!

## Слои приложения

### 1. Domain Layer (`internal/domain/`)

**Назначение:** Доменные модели и бизнес-сущности.

**Характеристики:**
- Не зависит от других слоёв
- Содержит только структуры данных
- Используется всеми остальными слоями

**Пример:**
```go
type User struct {
    ID        int       `db:"id" json:"id"`
    Name      string    `db:"name" json:"name"`
    Email     string    `db:"email" json:"email"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}
```

### 2. Repository Layer (`internal/repositories/`)

**Назначение:** Работа с базой данных.

**Характеристики:**
- Инкапсулирует SQL-запросы
- Возвращает доменные модели
- Использует sqlx для работы с БД

**Ответственность:**
- CRUD операции
- Сложные запросы
- Транзакции

### 3. Service Layer (`internal/services/`)

**Назначение:** Слой сервисов - инкапсулирует работу с данными и внешними системами.

**Характеристики:**
- Узко специализированы (одна ответственность)
- Инкапсулируют Repositories
- Переиспользуются в разных Use Cases
- Могут содержать простую бизнес-логику
- Сервис может загружать в себя другой сервис

**Пример Service:**
```go
type UserProfileService struct {
    userRepo *repositories.UserRepository
}

func (s *UserProfileService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    return s.userRepo.GetByID(ctx, id)
}

func (s *UserProfileService) CreateUser(ctx context.Context, name, email string) (*domain.User, error) {
    user := &domain.User{Name: name, Email: email}
    return user, s.userRepo.Create(ctx, user)
}
```

### 4. Use Case Layer (`internal/usecases/`)

**Назначение:** Бизнес-логика приложения.

Usecase получает **Services через DI**, никогда не работает с Repositories напрямую!

**Характеристики:**
- Оркестрирует работу сервисов
- Содержит бизнес-правила
- Не зависит от деталей реализации (HTTP, gRPC и т.д.)
- Не имеет прямого доступа к Repositories

**Пример:**
```go
type UserUsecase struct {
    userProfileService *services.UserProfileService  // ✅ Сервис
    emailService       *services.EmailService         // ✅ Сервис
}

func (uc *UserUsecase) RegisterUser(ctx context.Context, name, email string) (*domain.User, error) {
    // 1. Создание пользователя через сервис
    user, err := uc.userProfileService.CreateUser(ctx, name, email)
    if err != nil {
        return nil, err
    }
    
    // 2. Отправка приветственного письма
    _ = uc.emailService.SendEmail(ctx, email, "Welcome", "Thanks!")
    
    return user, nil
}
```

**Неправильный пример (анти-паттерн):**
```go
type UserUsecase struct {
    userRepo *repositories.UserRepository  // ❌ НЕПРАВИЛЬНО!
}
```

### 5. Controller Layer (`internal/controllers/`)

**Назначение:** Обработка HTTP запросов.

**Характеристики:**
- Парсинг входных данных
- Валидация запросов
- Вызов use cases
- Формирование HTTP ответов

**Ответственность:**
- HTTP статус коды
- Сериализация/десериализация JSON
- Обработка ошибок на уровне HTTP

### 6. Router Layer (`internal/router/`)

**Назначение:** Маршрутизация HTTP запросов.

**Характеристики:**
- Определение endpoints
- Группировка маршрутов
- Применение middleware

### 7. Configuration Layer (`internal/config/`)

**Назначение:** Конфигурация приложения.

**Характеристики:**
- Чтение YAML конфигурации
- Предоставление настроек через интерфейсы
- Валидация конфигурации

### 8. Container Layer (`internal/container/`)

**Назначение:** Dependency Injection контейнер.

**Характеристики:**
- Регистрация всех зависимостей
- Автоматическое разрешение зависимостей
- Управление жизненным циклом объектов

## Dependency Injection

Проект использует **Uber Dig** для управления зависимостями.

### Принципы:

1. **Инверсия зависимостей** - зависимости передаются извне
2. **Единственная ответственность** - каждый компонент делает одну вещь
3. **Тестируемость** - легко подменить зависимости в тестах

### Регистрация зависимостей:

```go
// В container.go
func (c *Container) provideDependencies() {
    // 1. Базовые зависимости
    c.container.Provide(config.NewConfig())
    c.container.Provide(db.NewDB)
}

func (c *Container) provideRepositories() {
    // 2. Репозитории (работают с БД)
    c.container.Provide(repositories.NewUserRepository)
    c.container.Provide(repositories.NewHealthcheckRepository)
}

func (c *Container) provideServices() {
    // 3. Сервисы (работают с репозиториями)
    c.container.Provide(services.NewUserProfileService)
    c.container.Provide(services.NewHealthcheckService)
    c.container.Provide(services.NewEmailService)
}

func (c *Container) provideUsecases() {
    // 4. Usecases (работают с сервисами, не с репозиториями!)
    c.container.Provide(usecases.NewUserUsecase)
}

func (c *Container) provideControllers() {
    // 5. Контроллеры (работают с usecases)
    c.container.Provide(controllers.NewUserController)
    c.container.Provide(controllers.NewHealthcheckController)
}
```

**Зависимости:**
- Controllers → Usecases
- Usecases → Services
- Services → Repositories
- Repositories → Database

### Использование:

```go
// Dig автоматически разрешит все зависимости
var userController *controllers.UserController
diContainer.Invoke(func(uc *controllers.UserController) {
    userController = uc
})
```

## Поток данных

### Пример: Создание пользователя

```
1. HTTP Request
   POST /api/users
   {"name": "John", "email": "john@example.com"}
   
2. Router
   Направляет запрос в UserController.CreateUser
   
3. Controller
   - Парсит JSON
   - Валидирует данные
   - Вызывает UserUsecase.CreateUser
   
4. Use Case
   - Применяет бизнес-правила
   - Вызывает UserProfileService.CreateUser (сервис!)
   - Может вызвать EmailService.SendEmail (сервис!)
   
5. Services
   - UserProfileService.CreateUser:
     * Вызывает UserRepository.Create
     * Возвращает созданного пользователя
   - EmailService.SendEmail:
     * Отправляет письмо через SMTP
   
6. Repository
   - Выполняет SQL INSERT
   - Возвращает созданного пользователя
   
7. Controller
   - Формирует JSON ответ
   - Устанавливает статус 201 Created
   
8. HTTP Response
   {"id": 1, "name": "John", "email": "john@example.com", ...}
```

## Middleware

Middleware применяется в `cmd/server/main.go`:

```go
http.ListenAndServe(addr, middleware.ChainMiddleware(
    router.Handler(diContainer),
    logger.WithLogging,
))
```

## Cron-скрипты

### Назначение

Cron-скрипты — это отдельные приложения для фоновых задач и периодических операций.

**Примеры:**
- Очистка старых данных
- Отправка отложенных писем
- Синхронизация с внешними системами
- Переиндексация данных
- Мониторинг и оповещения

### Структура

```
cmd/crons/
├── example/              # Пример cron-скрипта
│   └── main.go          # Точка входа
├── cleanup-records/     # Задача очистки
│   └── main.go
└── sync-external/       # Задача синхронизации
    └── main.go
```

### Принципы

1. **Переиспользование DI контейнера** - cron получает доступ к Services, Repositories, Logger
2. **Независимость от HTTP** - работают как консольные приложения
3. **Логирование** - используют тот же Logger, что и сервер
4. **Обработка ошибок** - логируют ошибки и возвращают статус выхода

## Миграции

Используется **golang-migrate** для управления схемой БД.

**Структура миграции:**
- `000001_name.up.sql` - применение изменений
- `000001_name.down.sql` - откат изменений

**Применение:**
```bash
make migrate-up    # Применить все миграции
make migrate-down  # Откатить последнюю
```

## Обработка ошибок

### Уровни обработки:

1. **Repository** - возвращает ошибки БД
2. **Use Case** - оборачивает ошибки с контекстом
3. **Controller** - преобразует в HTTP статусы

### Пример:

```go
// Repository
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
    var user domain.User
    err := r.db.GetContext(ctx, &user, query, id)
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    return &user, err
}

// Use Case
func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    user, err := uc.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %d: %w", id, err)
    }
    return user, nil
}

// Controller
func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }
    
    user, err := c.userUsecase.GetUser(r.Context(), id)
    if errors.Is(err, repositories.ErrUserNotFound) {
        http.Error(w, "User not found", http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(user)
}
```

## Тестирование

### Unit тесты:

```go
// Мокируем зависимости
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*domain.User), args.Error(1)
}

// Тестируем use case
func TestUserUsecase_GetUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    usecase := usecases.NewUserUsecase(mockRepo)
    
    ctx := context.Background()
    mockRepo.On("GetByID", ctx, int64(1)).Return(&domain.User{ID: 1}, nil)
    
    user, err := usecase.GetUser(ctx, 1)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), user.ID)
}
```

## Лучшие практики

✅ **ПРАВИЛЬНО:**
1. **Не смешивайте слои** - каждый слой должен знать только о нижележащих
2. **Используйте интерфейсы** - для лучшей тестируемости
3. **Держите контроллеры тонкими** - вся логика в use cases
4. **Один use case = одна бизнес-операция**
5. **Репозитории работают только с БД** - никакой бизнес-логики
6. **Используйте транзакции** - для атомарных операций
7. **Usecases работают ТОЛЬКО с Services** - никогда напрямую с Repositories
8. **Services инкапсулируют Repositories** - используйте методы-обертки
9. **Узко специализированные Services** - UserProfileService, OrderService и т.д.

❌ **НЕПРАВИЛЬНО:**
- Usecase напрямую использует Repository
- Controller напрямую использует Repository
- Repository содержит бизнес-логику

## Расширение проекта

### Добавление нового функционала:

1. Создайте доменную модель в `domain/`
2. Создайте миграцию
3. Создайте репозиторий в `repositories/`
4. **Создайте сервис в `services/`** 
5. Создайте use case в `usecases/`
6. Создайте контроллер в `controllers/`
7. Зарегистрируйте в `container/`:
8. Добавьте маршруты в `router/`