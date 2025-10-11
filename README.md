# 🚀 Go Project Template

> Готовый шаблон для создания Go веб-приложений с Clean Architecture, Dependency Injection и PostgreSQL

[![Go Version](https://img.shields.io/badge/Go-1.24.1-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ⚡ Быстрый старт (3 команды)

```bash
cp config.example.yaml config.yaml
make up-docker && sleep 3 && make migrate-up
make up-server
```

Проверьте: `curl http://localhost:8080/ping` → `{"status":"ok"}` ✅

## ✨ Что включено?

- 🏗️ **Clean Architecture** - 5 слоев (domain, repositories, usecases, services, controllers)
- 💉 **Dependency Injection** - Uber Dig для управления зависимостями
- 🌐 **Chi Router v5** - быстрый HTTP роутер с middleware
- 🗄️ **PostgreSQL 17** - современная БД + sqlx + миграции
- 📝 **Структурированное логирование** - zerolog
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
cp -r thinker my-awesome-project
cd my-awesome-project

# 2. Переименуйте модуль
# Замените в go.mod: github.com/SmirnovND/gobase → github.com/yourname/my-awesome-project

# 3. Замените импорты во всех файлах
find . -type f -name "*.go" -exec sed -i '' 's|github.com/SmirnovND/gobase|github.com/yourname/my-awesome-project|g' {} +

# 4. Начинайте разработку!
```

📖 **Подробнее:** [docs/SETUP_NEW_PROJECT.md](docs/SETUP_NEW_PROJECT.md)

## 📁 Структура проекта

```
.
├── cmd/
│   ├── server/             # Точка входа приложения
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
[Controller] ← обработка HTTP, валидация
    ↓
[Usecase] ← бизнес-логика
    ↓
[Repository] ← работа с БД
    ↓
[Database]
```

**Слои:**
1. **Domain** - доменные модели (User, Product, etc.)
2. **Repository** - работа с базой данных
3. **Usecase** - бизнес-логика приложения
4. **Service** - вспомогательные сервисы
5. **Controller** - HTTP обработчики

📖 **Подробнее:** [ARCHITECTURE.md](ARCHITECTURE.md)

## 📝 Добавление нового функционала

Пример: добавим управление продуктами

### 1. Создайте миграцию

```bash
make migrate-create name=create_products_table
```

```sql
-- migrations/000002_create_products_table.up.sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. Создайте модель

```go
// internal/domain/product.go
package domain

type Product struct {
    ID        int64   `db:"id"`
    Name      string  `db:"name"`
    Price     float64 `db:"price"`
    CreatedAt string  `db:"created_at"`
}
```

### 3. Создайте репозиторий

```go
// internal/repositories/product_repository.go
package repositories

type ProductRepository interface {
    Create(ctx context.Context, product *domain.Product) error
    GetByID(ctx context.Context, id int64) (*domain.Product, error)
}

type productRepository struct {
    db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
    return &productRepository{db: db}
}
```

### 4. Создайте usecase

```go
// internal/usecases/product_usecase.go
package usecases

type ProductUsecase interface {
    CreateProduct(ctx context.Context, name string, price float64) error
}

type productUsecase struct {
    repo repositories.ProductRepository
}

func NewProductUsecase(repo repositories.ProductRepository) ProductUsecase {
    return &productUsecase{repo: repo}
}
```

### 5. Создайте контроллер

```go
// internal/controllers/product_controller.go
package controllers

type ProductController struct {
    usecase usecases.ProductUsecase
}

func NewProductController(usecase usecases.ProductUsecase) *ProductController {
    return &ProductController{usecase: usecase}
}

func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
    // Обработка запроса
}
```

### 6. Зарегистрируйте в DI

```go
// internal/container/container.go
func (c *Container) RegisterRepositories() {
    c.container.Provide(repositories.NewProductRepository)
}

func (c *Container) RegisterUsecases() {
    c.container.Provide(usecases.NewProductUsecase)
}

func (c *Container) RegisterControllers() {
    c.container.Provide(controllers.NewProductController)
}
```

### 7. Добавьте маршрут

```go
// internal/router/router.go
r.Post("/api/products", productController.Create)
```

📖 **Больше примеров:** [QUICKSTART.md](QUICKSTART.md)

## 📚 Документация

| Документ | Описание |
|----------|----------|
| [QUICKSTART.md](QUICKSTART.md) | Быстрый старт и примеры использования |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Подробное описание архитектуры |
| [docs/SETUP_NEW_PROJECT.md](docs/SETUP_NEW_PROJECT.md) | Создание нового проекта из шаблона |
| [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) | Стандарты разработки и code review |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | История изменений |

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
- [zerolog](https://github.com/rs/zerolog) - логирование
- [golang-migrate](https://github.com/golang-migrate/migrate) - миграции
- [staticcheck](https://staticcheck.io/) - статический анализ кода

## ❓ Частые вопросы

**Q: Можно использовать без Docker?**  
A: Да, запустите PostgreSQL локально и обновите `config.yaml`

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

## 📄 Лицензия

MIT - используйте свободно в любых проектах

---

**Создано с ❤️ для Go сообщества**