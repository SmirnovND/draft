# Архитектура проекта

## Обзор

Проект построен на принципах **Clean Architecture** с использованием **Dependency Injection** через Uber Dig.

**Ключевой принцип:** Каждый слой зависит от **интерфейсов** более низких слоёв, а не от конкретных реализаций. Это обеспечивает слабую связанность и высокую тестируемость.

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
                     │ (использует interfaces.*)
┌────────────────────▼────────────────────────────────────┐
│                   Use Cases                             │
│              (Бизнес-логика приложения)                 │
│          (Орхестрирует работу сервисов)                 │
└────────────────────┬────────────────────────────────────┘
                     │ (использует interfaces.*)
┌────────────────────▼────────────────────────────────────┐
│                   Services                              │
│  (Domain Services + External Services + Some Services)  │
│  - Инкапсулируют работу с Repositories                  │
│  - Интеграция с внешними системами                      │
└────────────┬───────────────────────┬────────────────────┘
             │ (использует interfaces.*)
┌────────────▼────────────┐ ┌───────▼────────────────────┐
│      Repositories       │ │  Other Service         │
│ реализуют interfaces.*  │ │  (Email, SMS, API)         │
│  (Работа с БД)          │ │  (Email, SMS, API)         │
└────────────┬────────────┘ └────────────────────────────┘
             │
┌────────────▼────────────────────────────────────────────┐
│                      Database                           │
│                   (PostgreSQL)                          │
└─────────────────────────────────────────────────────────┘
```

**Ключевые правила:**
1. 🔗 Use Cases работают **только** через Services (не напрямую через Repositories!)
2. 🎯 Controllers зависят от **интерфейсов** Services, не от конкретных реализаций
3. 💉 Services зависят от **интерфейсов** Repositories, не от конкретных реализаций
4. 🧪 Благодаря интерфейсам легко создавать моки для тестирования

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
- **Реализует интерфейсы из `internal/interfaces/repository.go`**

**Ответственность:**
- CRUD операции
- Сложные запросы
- Транзакции

**Пример структуры:**
```go
// internal/interfaces/repository.go
type HealthcheckRepository interface {
    Ping(ctx context.Context) error
}

// internal/repositories/healthcheck_repository.go
type healthcheckRepository struct {
    db *sqlx.DB
}

// Конструктор возвращает интерфейс, не конкретный тип
func NewHealthcheckRepository(db *sqlx.DB) interfaces.HealthcheckRepository {
    return &healthcheckRepository{db: db}
}

func (r *healthcheckRepository) Ping(ctx context.Context) error {
    return r.db.PingContext(ctx)
}
```

### 3. Service Layer (`internal/services/`)

**Назначение:** Слой сервисов - инкапсулирует работу с данными и внешними системами.

**Характеристики:**
- Узко специализированы (одна ответственность)
- Инкапсулируют Repositories
- Переиспользуются в разных Use Cases
- Могут содержать простую бизнес-логику
- Сервис может загружать в себя другой сервис
- **Зависят от интерфейсов Repositories (`interfaces.HealthcheckRepository`)**
- **Реализуют интерфейсы из `internal/interfaces/service.go`**

**Пример Service:**
```go
// internal/interfaces/service.go
type HealthcheckService interface {
    Check(ctx context.Context) (map[string]interface{}, error)
}

// internal/services/healthcheck_service.go
type healthcheckService struct {
    healthRepo interfaces.HealthcheckRepository  // ← зависит от интерфейса!
}

// Конструктор возвращает интерфейс, не конкретный тип
func NewHealthcheckService(healthRepo interfaces.HealthcheckRepository) interfaces.HealthcheckService {
    return &healthcheckService{
        healthRepo: healthRepo,
    }
}

func (s *healthcheckService) Check(ctx context.Context) (map[string]interface{}, error) {
    err := s.healthRepo.Ping(ctx)
    if err != nil {
        return nil, err
    }
    return map[string]interface{}{"status": "ok"}, nil
}
```

**Преимущества использования интерфейсов:**
- ✅ Можно подменить репозиторий на мок при тестировании
- ✅ Можно подменить на другую реализацию без изменения сервиса
- ✅ Легко понять зависимости (просто смотрим параметры конструктора)

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
- Вызов use cases/сервисов
- Формирование HTTP ответов
- **Зависят от интерфейсов Services (`interfaces.HealthcheckService`)**
- **Реализуют интерфейсы из `internal/interfaces/controller.go`**

**Ответственность:**
- HTTP статус коды
- Сериализация/десериализация JSON
- Обработка ошибок на уровне HTTP

**Пример Controller:**
```go
// internal/interfaces/controller.go
type HealthcheckController interface {
    HandlePing(w http.ResponseWriter, r *http.Request)
}

// internal/controllers/healthcheck_controller.go
type healthcheckController struct {
    healthcheckService interfaces.HealthcheckService  // ← зависит от интерфейса!
}

// Конструктор возвращает интерфейс, не конкретный тип
func NewHealthcheckController(healthcheckService interfaces.HealthcheckService) interfaces.HealthcheckController {
    return &healthcheckController{
        healthcheckService: healthcheckService,
    }
}

func (hc *healthcheckController) HandlePing(w http.ResponseWriter, r *http.Request) {
    status, err := hc.healthcheckService.Check(r.Context())
    w.Header().Set("Content-Type", "application/json")
    
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "error",
            "error":  err.Error(),
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(status)
}
```

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

## Dependency Injection через Uber Dig

Проект использует **Uber Dig** для управления зависимостями с акцентом на **интерфейсы**.

### Ключевые принципы:

1. **Инверсия зависимостей** - зависимости передаются через интерфейсы, а не конкретные типы
2. **Слабая связанность** - каждый компонент зависит от интерфейсов, а не реализаций
3. **Тестируемость** - легко подменить зависимости мокированными интерфейсами
4. **Автоматическое разрешение** - Dig анализирует сигнатуру функций и находит нужные зависимости

### Как работает Dig с интерфейсами:

```
┌─ Конструктор функции ────────────────────────────────────┐
│                                                           │
│  func NewHealthcheckService(                              │
│      repo interfaces.HealthcheckRepository  ← нужен       │
│  ) interfaces.HealthcheckService {           ← возвращает │
│      return &healthcheckService{repo: repo}              │
│  }                                                        │
│                                                           │
│  Dig анализирует параметры ↓                              │
│  "Мне нужен interfaces.HealthcheckRepository"             │
│                                                           │
│  Dig ищет в контейнере ↓                                 │
│  "Кто создаёт interfaces.HealthcheckRepository?"          │
│  → repositories.NewHealthcheckRepository                  │
│                                                           │
│  Dig вызывает нужные конструкторы по порядку:             │
│  1. NewHealthcheckRepository(db)                          │
│  2. NewHealthcheckService(repo)                           │
│                                                           │
│  Результат: interfaces.HealthcheckService зарегистрирован │
└───────────────────────────────────────────────────────────┘
```

### Регистрация зависимостей:

```go
// internal/container/container.go
func (c *Container) provideDependencies() {
    // 1. Базовые зависимости (конкретные типы)
    c.container.Provide(config.NewConfig)           // → interfaces.ConfigServer
    c.container.Provide(db.NewDB)                   // → *sqlx.DB
}

func (c *Container) provideRepo() {
    // 2. Репозитории (зарегистрируются как интерфейсы!)
    // NewHealthcheckRepository(db *sqlx.DB) → interfaces.HealthcheckRepository
    c.container.Provide(repositories.NewHealthcheckRepository)
}

func (c *Container) provideService() {
    // 3. Сервисы (работают через интерфейсы репозиториев!)
    // NewHealthcheckService(repo interfaces.HealthcheckRepository) → interfaces.HealthcheckService
    c.container.Provide(services.NewHealthcheckService)
}

func (c *Container) provideController() {
    // 4. Контроллеры (работают через интерфейсы сервисов!)
    // NewHealthcheckController(svc interfaces.HealthcheckService) → interfaces.HealthcheckController
    c.container.Provide(controllers.NewHealthcheckController)
}
```

**Порядок вызова при Invoke:**

Когда вызываем `c.Invoke()` для получения контроллера:
```
1. Dig ищет interfaces.HealthcheckController
2. Находит NewHealthcheckController
3. NewHealthcheckController нужен interfaces.HealthcheckService
4. Находит NewHealthcheckService
5. NewHealthcheckService нужен interfaces.HealthcheckRepository
6. Находит NewHealthcheckRepository
7. NewHealthcheckRepository нужен *sqlx.DB
8. Находит провайдер БД
9. Рекурсивно вызывает все конструкторы в правильном порядке
10. Возвращает полностью инициализированный interfaces.HealthcheckController
```

### Использование:

```go
// Router получает контроллер из DI контейнера
var healthcheckController interfaces.HealthcheckController
err := diContainer.Invoke(func(
    ctrl interfaces.HealthcheckController,  // ← Dig автоматически разрешит все зависимости
) {
    healthcheckController = ctrl
})

// Dig сам построит цепочку:
// *sqlx.DB → interfaces.HealthcheckRepository → interfaces.HealthcheckService → interfaces.HealthcheckController
```

### Для сложных сервисов с кастомной логикой:

```go
// Если нужна специальная логика инициализации
func (c *Container) provideService() {
    c.container.Provide(func(
        minioCfg interfaces.ConfigServer,  // ← Dig разрешит зависимости
        repo interfaces.HealthcheckRepository,
    ) interfaces.CloudService {
        // Кастомная инициализация
        return service.NewCloud(minioCfg, repo)
    })
}
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

**Ключевое преимущество архитектуры с интерфейсами:** вы легко можете подменять зависимости на моки!

### Unit тесты сервисов:

```go
// Пример 1: Простой мок с функциями (самый простой способ)
func TestHealthcheckService(t *testing.T) {
    // Мокируем репозиторий
    mockRepo := &interfaces.MockHealthcheckRepository{
        PingFunc: func(ctx context.Context) error {
            return nil  // Успешно
        },
    }
    
    // Создаём сервис с мокированным репозиторием
    service := services.NewHealthcheckService(mockRepo)
    
    // Тестируем
    status, err := service.Check(context.Background())
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if status["status"] != "ok" {
        t.Errorf("expected status 'ok', got %v", status["status"])
    }
}

// Пример 2: Тест с ошибкой
func TestHealthcheckService_WithError(t *testing.T) {
    mockRepo := &interfaces.MockHealthcheckRepository{
        PingFunc: func(ctx context.Context) error {
            return errors.New("connection failed")
        },
    }
    
    service := services.NewHealthcheckService(mockRepo)
    _, err := service.Check(context.Background())
    
    if err == nil {
        t.Fatal("expected error, got nil")
    }
}
```

**Заметьте:** никаких фреймворков для моков не нужно! Просто используйте встроенные моки из `interfaces/mocks.go`

### Unit тесты контроллеров:

```go
func TestHealthcheckController(t *testing.T) {
    // Мокируем сервис
    mockService := &interfaces.MockHealthcheckService{
        CheckFunc: func(ctx context.Context) (map[string]interface{}, error) {
            return map[string]interface{}{"status": "ok"}, nil
        },
    }
    
    // Создаём контроллер с мокированным сервисом
    controller := controllers.NewHealthcheckController(mockService)
    
    // Тестируем HTTP обработчик
    req := httptest.NewRequest("GET", "/ping", nil)
    w := httptest.NewRecorder()
    
    controller.HandlePing(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
    
    var result map[string]interface{}
    json.NewDecoder(w.Body).Decode(&result)
    
    if result["status"] != "ok" {
        t.Errorf("expected status 'ok', got %v", result["status"])
    }
}
```

### Интеграционные тесты с реальным DI контейнером:

```go
func TestIntegration_HealthcheckFlow(t *testing.T) {
    // Используем реальный контейнер с реальной БД
    container := container.NewContainer()
    
    var controller interfaces.HealthcheckController
    err := container.Invoke(func(ctrl interfaces.HealthcheckController) {
        controller = ctrl
    })
    
    if err != nil {
        t.Fatalf("failed to invoke controller: %v", err)
    }
    
    // Тестируем полный цепочку
    req := httptest.NewRequest("GET", "/ping", nil)
    w := httptest.NewRecorder()
    
    controller.HandlePing(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

### Лучшие практики для тестирования:

1. **Unit тесты** - используйте моки для быстрого тестирования логики
2. **Интеграционные тесты** - используйте реальный контейнер для тестирования цепочки
3. **Мокируйте только то, что нужно** - если можно использовать реальный объект, используйте его
4. **Тесты находятся рядом с кодом** - `service_test.go` рядом с `service.go`

## Лучшие практики

### ✅ **ПРАВИЛЬНО:**

1. **Зависимости через интерфейсы** - всегда передавайте интерфейсы, не конкретные типы
   ```go
   // ✅ ПРАВИЛЬНО
   func NewService(repo interfaces.Repository) interfaces.Service {
       return &service{repo: repo}
   }
   
   // ❌ НЕПРАВИЛЬНО
   func NewService(repo *repositories.UserRepository) *userService {
       return &userService{repo: repo}
   }
   ```

2. **Конструкторы возвращают интерфейсы**
   ```go
   // ✅ ПРАВИЛЬНО - возвращаем интерфейс
   func NewHealthcheckService(repo interfaces.HealthcheckRepository) interfaces.HealthcheckService
   
   // ❌ НЕПРАВИЛЬНО - возвращаем конкретный тип
   func NewHealthcheckService(repo interfaces.HealthcheckRepository) *healthcheckService
   ```

3. **Структуры приватные** - только интерфейсы публичные
   ```go
   // ✅ ПРАВИЛЬНО
   type healthcheckService struct { ... }  // приватная
   func NewHealthcheckService(...) interfaces.HealthcheckService  // публичный конструктор
   
   // ❌ НЕПРАВИЛЬНО
   type HealthcheckService struct { ... }  // публичная структура
   ```

4. **Слои знают только о нижележащих** через интерфейсы
   ```
   Controllers → interfaces.Services
   Services → interfaces.Repositories
   Repositories → Database
   ```

5. **Usecases работают ТОЛЬКО с Services** - никогда напрямую с Repositories
6. **Services инкапсулируют Repositories** - используйте методы-обертки
7. **Узко специализированные Services** - HealthcheckService, UserProfileService, и т.д.
8. **Одна ответственность** - каждый сервис делает одно
9. **Используйте транзакции** - для атомарных операций

### ❌ **НЕПРАВИЛЬНО:**

- ❌ Controller напрямую работает с Repository (нарушает слои)
- ❌ Usecase напрямую использует Repository (должен через Service)
- ❌ Зависимости от конкретных типов вместо интерфейсов
- ❌ Публичные структуры вместо интерфейсов
- ❌ Repository содержит бизнес-логику

## Расширение проекта

### Пошаговое добавление нового функционала (например, UserService):

#### 1. Добавьте интерфейсы в `internal/interfaces/`

```go
// internal/interfaces/repository.go (добавьте)
type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*domain.User, error)
    Create(ctx context.Context, user *domain.User) error
}

// internal/interfaces/service.go (добавьте)
type UserService interface {
    GetUser(ctx context.Context, id int64) (*domain.User, error)
    CreateUser(ctx context.Context, name string) (*domain.User, error)
}

// internal/interfaces/controller.go (добавьте)
type UserController interface {
    GetUser(w http.ResponseWriter, r *http.Request)
    CreateUser(w http.ResponseWriter, r *http.Request)
}
```

#### 2. Создайте доменную модель в `domain/`

```go
// internal/domain/user.go
package domain

type User struct {
    ID        int64  `db:"id"`
    Name      string `db:"name"`
    CreatedAt string `db:"created_at"`
}
```

#### 3. Создайте миграцию

```bash
make migrate-create name=create_users_table
```

#### 4. Создайте репозиторий в `repositories/`

```go
// internal/repositories/user_repository.go
package repositories

type userRepository struct {
    db *sqlx.DB
}

// ✅ Возвращаем интерфейс!
func NewUserRepository(db *sqlx.DB) interfaces.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
    var user domain.User
    // ... SQL запрос
    return &user, nil
}
```

#### 5. Создайте сервис в `services/`

```go
// internal/services/user_service.go
package services

type userService struct {
    userRepo interfaces.UserRepository  // ← интерфейс!
}

// ✅ Зависит от интерфейса репозитория, возвращаем интерфейс!
func NewUserService(userRepo interfaces.UserRepository) interfaces.UserService {
    return &userService{userRepo: userRepo}
}

func (s *userService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
    return s.userRepo.GetByID(ctx, id)
}
```

#### 6. Создайте контроллер в `controllers/`

```go
// internal/controllers/user_controller.go
package controllers

type userController struct {
    userService interfaces.UserService  // ← интерфейс!
}

// ✅ Зависит от интерфейса сервиса, возвращаем интерфейс!
func NewUserController(userService interfaces.UserService) interfaces.UserController {
    return &userController{userService: userService}
}

func (c *userController) GetUser(w http.ResponseWriter, r *http.Request) {
    // ... обработка
}
```

#### 7. Зарегистрируйте в DI контейнере

```go
// internal/container/container.go
func (c *Container) provideRepo() {
    c.container.Provide(repositories.NewHealthcheckRepository)
    c.container.Provide(repositories.NewUserRepository)  // ← добавьте
}

func (c *Container) provideService() {
    c.container.Provide(services.NewHealthcheckService)
    c.container.Provide(services.NewUserService)  // ← добавьте
}

func (c *Container) provideController() {
    c.container.Provide(controllers.NewHealthcheckController)
    c.container.Provide(controllers.NewUserController)  // ← добавьте
}
```

#### 8. Добавьте маршруты

```go
// internal/router/router.go
func Handler(diContainer *container.Container) http.Handler {
    var userController interfaces.UserController
    err := diContainer.Invoke(func(
        ctrl interfaces.UserController,  // ← Dig автоматически разрешит цепочку!
    ) {
        userController = ctrl
    })
    
    // ...
    r.Get("/users/{id}", userController.GetUser)
    r.Post("/users", userController.CreateUser)
    
    return r
}
```

**Заметьте:** Dig **автоматически разрешит всю цепочку**:
```
*sqlx.DB 
  ↓
NewUserRepository → interfaces.UserRepository
  ↓
NewUserService → interfaces.UserService
  ↓
NewUserController → interfaces.UserController
```

Никаких ручных подключений - только регистрируем конструкторы, Dig сам найдёт зависимости!