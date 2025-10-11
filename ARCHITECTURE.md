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
└────────────┬───────────────────────┬────────────────────┘
             │                       │
┌────────────▼────────────┐ ┌───────▼────────────────────┐
│      Repositories       │ │       Services             │
│  (Работа с БД)          │ │ (Email, SMS, External API) │
└────────────┬────────────┘ └────────────────────────────┘
             │
┌────────────▼────────────────────────────────────────────┐
│                      Database                           │
│                   (PostgreSQL)                          │
└─────────────────────────────────────────────────────────┘
```

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

**Назначение:** Вспомогательные сервисы.

**Примеры:**
- Email сервис
- SMS сервис
- Работа с внешними API
- Кэширование
- Файловое хранилище

**Характеристики:**
- Независимы от бизнес-логики
- Переиспользуемые компоненты
- Могут использоваться в разных use cases

### 4. Use Case Layer (`internal/usecases/`)

**Назначение:** Бизнес-логика приложения.

**Характеристики:**
- Оркестрирует работу репозиториев и сервисов
- Содержит бизнес-правила
- Не зависит от деталей реализации (HTTP, gRPC и т.д.)

**Пример:**
```go
func (uc *UserUsecase) RegisterUser(ctx context.Context, name, email string) (*domain.User, error) {
    // 1. Проверка бизнес-правил
    if !uc.emailService.IsValid(ctx, email) {
        return nil, errors.New("invalid email")
    }
    
    // 2. Создание пользователя
    user := &domain.User{Name: name, Email: email}
    err := uc.userRepo.Create(ctx, user)
    if err != nil {
        return nil, err
    }
    
    // 3. Отправка приветственного письма
    uc.emailService.SendWelcome(ctx, email)
    
    return user, nil
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
    c.container.Provide(config.NewConfig())
    c.container.Provide(db.NewDB)
}

func (c *Container) provideRepo() {
    c.container.Provide(repositories.NewUserRepository)
}

func (c *Container) provideUsecase() {
    c.container.Provide(usecases.NewUserUsecase)
}

func (c *Container) provideController() {
    c.container.Provide(controllers.NewUserController)
}
```

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
   - Вызывает UserRepository.Create
   - Может вызвать EmailService.SendWelcome
   
5. Repository
   - Выполняет SQL INSERT
   - Возвращает созданного пользователя
   
6. Controller
   - Формирует JSON ответ
   - Устанавливает статус 201 Created
   
7. HTTP Response
   {"id": 1, "name": "John", "email": "john@example.com", ...}
```

## Middleware

Проект использует middleware для:
- Логирования запросов
- Обработки паник
- CORS
- Аутентификации (при необходимости)

Middleware применяется в `cmd/server/main.go`:

```go
http.ListenAndServe(addr, middleware.ChainMiddleware(
    router.Handler(diContainer),
    logger.WithLogging,
))
```

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

1. **Не смешивайте слои** - каждый слой должен знать только о нижележащих
2. **Используйте интерфейсы** - для лучшей тестируемости
3. **Держите контроллеры тонкими** - вся логика в use cases
4. **Один use case = одна бизнес-операция**
5. **Репозитории работают только с БД** - никакой бизнес-логики
6. **Используйте транзакции** - для атомарных операций

## Расширение проекта

### Добавление нового функционала:

1. Создайте доменную модель в `domain/`
2. Создайте миграцию
3. Создайте репозиторий в `repositories/`
4. Создайте use case в `usecases/`
5. Создайте контроллер в `controllers/`
6. Зарегистрируйте в `container/`
7. Добавьте маршруты в `router/`

### Добавление нового транспорта (gRPC, WebSocket):

1. Создайте новый слой транспорта
2. Используйте те же use cases
3. Не дублируйте бизнес-логику

Это позволяет легко поддерживать несколько транспортных протоколов одновременно.