# 🚀 Go Project Template

> Готовый шаблон для создания Go веб-приложений с Clean Architecture, Dependency Injection и PostgreSQL

[![Go Version](https://img.shields.io/badge/Go-1.24.1-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ⚡ Быстрый старт (3 команды)

```bash
cp cmd/server/config.example.yaml cmd/server/config.yaml
make up-docker && sleep 3 && make migrate-up
make up-server
```

Проверьте: `curl http://localhost:8080/ping` → `{"status":"ok"}` ✅

## ✨ Что включено?

- 🏗️ **Clean Architecture** - 5 слоев (domain, repositories, usecases, services, controllers)
- 💉 **Dependency Injection** - Uber Dig для управления зависимостями
- 🌐 **Chi Router v5** - быстрый HTTP роутер с middleware
- 🗄️ **PostgreSQL 17** - современная БД + sqlx + миграции
- 🔄 **Транзакции в БД** - поддержка ACID с TransactionManager и примерами
- 📝 **Структурированное логирование** - Uber Zap
- 🐰 **RabbitMQ интеграция** - Producer/Consumer с поддержкой отложенных сообщений
- 🔍 **Статический анализ** - кастомный multichecker с 20+ анализаторами
- 🐳 **Docker Ready** - готовый docker-compose для разработки
- 📚 **Примеры кода** - для каждого слоя архитектуры

## 📋 Требования

- Go 1.24.1+
- Docker и Docker Compose
- golang-migrate (для миграций)

## 🎯 Что дальше?

### Вариант 1: Изучить шаблон

```bash
# Запустите проект
make up-docker && make migrate-up
make up-server

# Изучите структуру
tree internal/

# Читайте примеры в каждом слое
cat internal/repositories/README.md
cat internal/usecases/README.md
cat internal/controllers/README.md
```

📖 **Подробнее:** [QUICKSTART.md](QUICKSTART.md)

### Вариант 2: Создать новый проект

```bash
# 1. Скопируйте шаблон
cp -r go-base my-awesome-project
cd my-awesome-project

# 2. Переименуйте модуль
# Замените в go.mod: github.com/SmirnovND/draft → github.com/my-awesome-project

# 3. Замените импорты во всех файлах
find . -type f -name "*.go" -exec sed -i '' 's|github.com/SmirnovND/draft|github.com/my-awesome-project|g' {} +

# 4. Начинайте разработку!
```

📖 **Подробнее:** [docs/SETUP_NEW_PROJECT.md](docs/SETUP_NEW_PROJECT.md)

## 📁 Структура проекта

```
.
├── cmd/
│   ├── server/             # Точка входа приложения
│   ├── crons/              # Cron-скрипты и фоновые задачи
│   └── staticlint/         # Кастомный multichecker для анализа кода
├── internal/
│   ├── config/             # Конфигурация
│   ├── container/          # DI-контейнер (Uber Dig)
│   ├── controllers/        # HTTP-контроллеры (+ примеры)
│   ├── domain/             # Доменные модели
│   ├── interfaces/         # Интерфейсы для зависимостей
│   ├── repositories/       # Работа с БД (+ примеры)
│   ├── router/             # Маршрутизация
│   ├── services/           # Вспомогательные сервисы (+ примеры)
│   └── usecases/           # Бизнес-логика (+ примеры)
├── migrations/             # SQL миграции
├── docker/                 # Docker конфигурация
├── docs/                   # Документация
├── config.example.yaml     # Пример конфигурации
├── docker-compose.yml      # PostgreSQL для разработки
└── Makefile               # Команды для разработки
```

## 🔧 Основные команды

```bash
make help              # Показать все команды
make up-docker         # Запустить PostgreSQL
make down-docker       # Остановить PostgreSQL
make up-server         # Запустить сервер
make migrate-up        # Применить миграции
make migrate-down      # Откатить последнюю миграцию
make migrate-create    # Создать новую миграцию
make lint              # Запустить статический анализ кода
make test              # Запустить тесты
make deps              # Установить зависимости
make clean             # Очистить артефакты
```

## 🏗️ Архитектура

Проект следует принципам **Clean Architecture**:

```
HTTP Request
    ↓
[Controller] ← зависит от interfaces.Service
    ↓ (использует interfaces.*)
[Usecase] ← зависит от interfaces.Service
    ↓ (использует interfaces.*)
[Services] ← реализуют interfaces.Service, зависят от interfaces.Repository
    ↓ (используют interfaces.*)
[Repository] ← реализуют interfaces.Repository
    ↓
[Database]
```

**Ключевые правила:**
1. 🔗 **Только интерфейсы** - все слои зависят от интерфейсов, не от конкретных типов
2. 💉 **Dependency Injection** - Uber Dig автоматически разрешает зависимости через интерфейсы
3. 🧪 **Тестируемость** - легко подменить зависимости мокированными интерфейсами
4. 🎯 **Slabaya связанность** - каждый слой независим от деталей реализации нижних слоёв

**Слои:**
1. **Domain** - доменные модели (User, Product, etc.)
2. **Repository** - работа с БД (реализуют `interfaces.Repository`)
3. **Service** - инкапсуляция логики (реализуют `interfaces.Service`, зависят от `interfaces.Repository`)
4. **Usecase** - бизнес-логика (зависят от `interfaces.Service`)
5. **Controller** - HTTP обработчики (зависят от `interfaces.Service`)

**Пример сигнатуры:**
```go
// Repository → Service → Controller (через интерфейсы!)
func NewService(repo interfaces.Repository) interfaces.Service { ... }
func NewController(svc interfaces.Service) interfaces.Controller { ... }
```

📖 **Подробнее:** [ARCHITECTURE.md](ARCHITECTURE.md)

## 🔄 Cron-скрипты и фоновые задачи

Проект поддерживает cron-скрипты через отдельные точки входа в `cmd/crons/`.

**Структура:**
```
cmd/crons/
├── example/               # Пример cron-скрипта
│   └── main.go           # Пример использования DI контейнера
```

**Особенности:**
- Используют тот же **DI контейнер**, что и основной сервер
- Имеют доступ ко всем **Services, Repositories, Logger**
- Могут быть запущены через cron, systemd или другие планировщики

📖 **Подробнее:** [ARCHITECTURE.md](ARCHITECTURE.md#cron-скрипты)

## 📝 Добавление нового функционала

Пример: добавим управление продуктами

### 1️⃣ Добавьте интерфейсы в `internal/interfaces/`

```go
// internal/interfaces/repository.go (добавьте)
type ProductRepository interface {
    Create(ctx context.Context, product *domain.Product) error
    GetByID(ctx context.Context, id int64) (*domain.Product, error)
}

// internal/interfaces/service.go (добавьте)
type ProductService interface {
    GetProduct(ctx context.Context, id int64) (*domain.Product, error)
    CreateProduct(ctx context.Context, name string, price float64) (*domain.Product, error)
}

// internal/interfaces/controller.go (добавьте)
type ProductController interface {
    GetProduct(w http.ResponseWriter, r *http.Request)
    CreateProduct(w http.ResponseWriter, r *http.Request)
}
```

### 2️⃣ Создайте миграцию

```bash
make migrate-create name=create_products_table
```

### 3️⃣ Создайте модель в `domain/`

```go
// internal/domain/product.go
type Product struct {
    ID        int64   `db:"id"`
    Name      string  `db:"name"`
    Price     float64 `db:"price"`
    CreatedAt string  `db:"created_at"`
}
```

### 4️⃣ Создайте репозиторий (реализует интерфейс!)

```go
// internal/repositories/product_repository.go
type productRepository struct {
    db *sqlx.DB
}

// ✅ Возвращаем интерфейс!
func NewProductRepository(db *sqlx.DB) interfaces.ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
    // ... реализация
}
```

### 5️⃣ Создайте сервис (зависит от интерфейса, реализует интерфейс!)

```go
// internal/services/product_service.go
type productService struct {
    repo interfaces.ProductRepository  // ← интерфейс!
}

// ✅ Зависит от интерфейса, возвращает интерфейс!
func NewProductService(repo interfaces.ProductRepository) interfaces.ProductService {
    return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, name string, price float64) (*domain.Product, error) {
    product := &domain.Product{Name: name, Price: price}
    return product, s.repo.Create(ctx, product)
}
```

### 6️⃣ Создайте контроллер (зависит от интерфейса, реализует интерфейс!)

```go
// internal/controllers/product_controller.go
type productController struct {
    productService interfaces.ProductService  // ← интерфейс!
}

// ✅ Зависит от интерфейса, возвращает интерфейс!
func NewProductController(productService interfaces.ProductService) interfaces.ProductController {
    return &productController{productService: productService}
}

func (c *productController) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string  `json:"name"`
        Price float64 `json:"price"`
    }
    
    json.NewDecoder(r.Body).Decode(&req)
    product, err := c.productService.CreateProduct(r.Context(), req.Name, req.Price)
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(product)
}
```

### 7️⃣ Зарегистрируйте в DI контейнере

```go
// internal/container/container.go
func (c *Container) provideRepo() {
    c.container.Provide(repositories.NewProductRepository)  // ← добавьте
}

func (c *Container) provideService() {
    c.container.Provide(services.NewProductService)  // ← добавьте
}

func (c *Container) provideController() {
    c.container.Provide(controllers.NewProductController)  // ← добавьте
}
```

### 8️⃣ Добавьте маршрут

```go
// internal/router/router.go
var productController interfaces.ProductController
diContainer.Invoke(func(ctrl interfaces.ProductController) {
    productController = ctrl
})

r.Get("/api/products/{id}", productController.GetProduct)
r.Post("/api/products", productController.CreateProduct)
```

**Результат:** Dig автоматически построит цепочку:
```
*sqlx.DB → NewProductRepository → interfaces.ProductRepository
         → NewProductService → interfaces.ProductService
         → NewProductController → interfaces.ProductController
```

📖 **Больше примеров:** [QUICKSTART.md](QUICKSTART.md)

## 📚 Документация

| Документ | Описание |
|----------|----------|
| [QUICKSTART.md](QUICKSTART.md) | Быстрый старт и примеры использования |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Подробное описание архитектуры |
| [docs/TRANSACTIONS.md](docs/TRANSACTIONS.md) | Использование ACID-транзакций в проекте |
| [docs/RABBITMQ.md](docs/RABBITMQ.md) | Работа с RabbitMQ: Publisher/Consumer примеры |
| [docs/SETUP_NEW_PROJECT.md](docs/SETUP_NEW_PROJECT.md) | Создание нового проекта из шаблона |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | Стандарты разработки и code review |

## 🔍 Статический анализ кода

Проект включает кастомный multichecker для проверки качества кода:

```bash
make lint
```

**Включенные анализаторы:**

1. **Стандартные анализаторы** (golang.org/x/tools):
   - `printf` - проверка форматирования строк
   - `shadow` - обнаружение затенения переменных
   - `structtag` - проверка тегов структур
   - `unreachable` - поиск недостижимого кода

2. **Staticcheck SA** - все анализаторы категории SA (проверки на баги)

3. **Staticcheck ST1000** - проверка именования пакетов

4. **Публичные анализаторы:**
   - `nilerr` - обнаружение игнорирования ошибок
   - `bodyclose` - проверка закрытия HTTP Response Body

5. **Кастомный анализатор:**
   - `exitchecker` - запрет прямых вызовов `os.Exit()` в функции `main`

## 🔌 Зависимости

- [Chi](https://github.com/go-chi/chi) - HTTP router
- [sqlx](https://github.com/jmoiron/sqlx) - расширение для database/sql
- [Uber Dig](https://github.com/uber-go/dig) - dependency injection
- [Uber Zap](https://github.com/uber-go/zap) - логирование
- [AMQP](https://github.com/streadway/amqp) - клиент RabbitMQ
- [golang-migrate](https://github.com/golang-migrate/migrate) - миграции
- [staticcheck](https://staticcheck.io/) - статический анализ кода

## ❓ Частые вопросы

**Q: Как добавить новую таблицу?**  
A: `make migrate-create name=add_products_table`

**Q: Где примеры кода?**  
A: В каждом слое есть `README.md` с примерами

**Q: Нужно ли знать Clean Architecture?**  
A: Нет, документация объясняет все концепции

**Q: Как переименовать проект?**  
A: Смотрите [docs/SETUP_NEW_PROJECT.md](docs/SETUP_NEW_PROJECT.md)

## 🤝 Вклад в развитие

Идеи и предложения приветствуются! Смотрите [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)