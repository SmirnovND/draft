# 🚀 Быстрый старт

Пошаговая инструкция для начала работы с шаблоном.

## 📋 Содержание

- [Первоначальная настройка](#первоначальная-настройка)
- [Проверка работы](#проверка-работы)
- [Добавление нового функционала](#добавление-нового-функционала)
- [Примеры API запросов](#примеры-api-запросов)
- [Полезные команды](#полезные-команды)

## Первоначальная настройка

### 1. Установка и запуск

```bash
# Создайте конфигурацию
cp cmd/server/config.example.yaml cmd/server/config.yaml

# Запустите PostgreSQL
make up-docker

# Подождите 3 секунды и примените миграции
sleep 3 && make migrate-up

# Запустите сервер
make up-server
```

### 2. Проверка работы

```bash
# Проверьте healthcheck
curl http://localhost:8080/ping

# Ожидаемый ответ:
# {"status":"ok"}
```

✅ Готово! Сервер работает на `http://localhost:8080`

## Добавление нового функционала

Пример: создадим API для управления продуктами (CRUD операции)

### Шаг 1: Создайте миграцию

```bash
make migrate-create name=create_products_table
```

Отредактируйте файлы в `migrations/`:

**000002_create_products_table.up.sql:**
```sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**000002_create_products_table.down.sql:**
```sql
DROP TABLE IF EXISTS products;
```

Примените миграцию:
```bash
make migrate-up
```

### Шаг 2: Создайте доменную модель

**Файл:** `internal/domain/product.go`

```go
package domain

import "time"

type Product struct {
    ID          int64     `db:"id" json:"id"`
    Name        string    `db:"name" json:"name"`
    Description string    `db:"description" json:"description"`
    Price       float64   `db:"price" json:"price"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
```

### Шаг 3: Создайте репозиторий

**Файл:** `internal/repositories/product_repository.go`

```go
package repositories

import (
    "context"
    "github.com/SmirnovND/gobase/internal/domain"
    "github.com/jmoiron/sqlx"
)

type ProductRepository interface {
    Create(ctx context.Context, product *domain.Product) error
    GetByID(ctx context.Context, id int64) (*domain.Product, error)
    GetAll(ctx context.Context) ([]domain.Product, error)
    Update(ctx context.Context, product *domain.Product) error
    Delete(ctx context.Context, id int64) error
}

type productRepository struct {
    db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
    query := `
        INSERT INTO products (name, description, price)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `
    return r.db.QueryRowContext(ctx, query, product.Name, product.Description, product.Price).
        Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
    var product domain.Product
    query := `SELECT * FROM products WHERE id = $1`
    err := r.db.GetContext(ctx, &product, query, id)
    return &product, err
}

func (r *productRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
    var products []domain.Product
    query := `SELECT * FROM products ORDER BY created_at DESC`
    err := r.db.SelectContext(ctx, &products, query)
    return products, err
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
    query := `
        UPDATE products 
        SET name = $1, description = $2, price = $3, updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
        RETURNING updated_at
    `
    return r.db.QueryRowContext(ctx, query, product.Name, product.Description, product.Price, product.ID).
        Scan(&product.UpdatedAt)
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
    query := `DELETE FROM products WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}
```

### Шаг 4: Создайте usecase

**Файл:** `internal/usecases/product_usecase.go`

```go
package usecases

import (
    "context"
    "github.com/SmirnovND/gobase/internal/domain"
    "github.com/SmirnovND/gobase/internal/repositories"
)

type ProductUsecase interface {
    CreateProduct(ctx context.Context, name, description string, price float64) (*domain.Product, error)
    GetProduct(ctx context.Context, id int64) (*domain.Product, error)
    GetAllProducts(ctx context.Context) ([]domain.Product, error)
    UpdateProduct(ctx context.Context, id int64, name, description string, price float64) (*domain.Product, error)
    DeleteProduct(ctx context.Context, id int64) error
}

type productUsecase struct {
    repo repositories.ProductRepository
}

func NewProductUsecase(repo repositories.ProductRepository) ProductUsecase {
    return &productUsecase{repo: repo}
}

func (uc *productUsecase) CreateProduct(ctx context.Context, name, description string, price float64) (*domain.Product, error) {
    product := &domain.Product{
        Name:        name,
        Description: description,
        Price:       price,
    }
    err := uc.repo.Create(ctx, product)
    return product, err
}

func (uc *productUsecase) GetProduct(ctx context.Context, id int64) (*domain.Product, error) {
    return uc.repo.GetByID(ctx, id)
}

func (uc *productUsecase) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
    return uc.repo.GetAll(ctx)
}

func (uc *productUsecase) UpdateProduct(ctx context.Context, id int64, name, description string, price float64) (*domain.Product, error) {
    product := &domain.Product{
        ID:          id,
        Name:        name,
        Description: description,
        Price:       price,
    }
    err := uc.repo.Update(ctx, product)
    return product, err
}

func (uc *productUsecase) DeleteProduct(ctx context.Context, id int64) error {
    return uc.repo.Delete(ctx, id)
}
```

### Шаг 5: Создайте контроллер

**Файл:** `internal/controllers/product_controller.go`

```go
package controllers

import (
    "encoding/json"
    "github.com/SmirnovND/gobase/internal/usecases"
    "github.com/go-chi/chi/v5"
    "net/http"
    "strconv"
)

type ProductController struct {
    usecase usecases.ProductUsecase
}

func NewProductController(usecase usecases.ProductUsecase) *ProductController {
    return &ProductController{usecase: usecase}
}

type CreateProductRequest struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}

func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    product, err := c.usecase.CreateProduct(r.Context(), req.Name, req.Description, req.Price)
    if err != nil {
        http.Error(w, "Failed to create product", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(product)
}

func (c *ProductController) GetByID(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    product, err := c.usecase.GetProduct(r.Context(), id)
    if err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

func (c *ProductController) GetAll(w http.ResponseWriter, r *http.Request) {
    products, err := c.usecase.GetAllProducts(r.Context())
    if err != nil {
        http.Error(w, "Failed to get products", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

func (c *ProductController) Update(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var req CreateProductRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    product, err := c.usecase.UpdateProduct(r.Context(), id, req.Name, req.Description, req.Price)
    if err != nil {
        http.Error(w, "Failed to update product", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

func (c *ProductController) Delete(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    if err := c.usecase.DeleteProduct(r.Context(), id); err != nil {
        http.Error(w, "Failed to delete product", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
```

### Шаг 6: Зарегистрируйте в DI контейнере

**Файл:** `internal/container/container.go`

```go
func (c *Container) RegisterRepositories() {
    c.container.Provide(repositories.NewProductRepository)
}

func (c *Container) RegisterUsecases() {
    c.container.Provide(usecases.NewProductUsecase)
}

func (c *Container) RegisterControllers() {
    c.container.Provide(controllers.NewHealthcheckController)
    c.container.Provide(controllers.NewProductController) // добавить
}
```

### Шаг 7: Добавьте маршруты

**Файл:** `internal/router/router.go`

```go
func Handler(diContainer *container.Container) http.Handler {
    var (
        HealthcheckController *controllers.HealthcheckController
        ProductController     *controllers.ProductController
    )
    
    err := diContainer.Invoke(func(
        d *sqlx.DB,
        c interfaces.ConfigServer,
        healthcheckControl *controllers.HealthcheckController,
        productControl *controllers.ProductController,
    ) {
        HealthcheckController = healthcheckControl
        ProductController = productControl
    })
    if err != nil {
        fmt.Println(err)
        return nil
    }

    r := chi.NewRouter()
    r.Use(middleware.StripSlashes)

    // Healthcheck
    r.Get("/ping", HealthcheckController.HandlePing)

    // Products API
    r.Route("/api/products", func(r chi.Router) {
        r.Get("/", ProductController.GetAll)
        r.Post("/", ProductController.Create)
        r.Get("/{id}", ProductController.GetByID)
        r.Put("/{id}", ProductController.Update)
        r.Delete("/{id}", ProductController.Delete)
    })

    return r
}
```

### Шаг 8: Перезапустите сервер

```bash
# Остановите сервер (Ctrl+C)
# Запустите снова
make up-server
```

## Примеры API запросов

### Healthcheck

```bash
curl http://localhost:8080/ping
```

### Создать продукт

```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "description": "High-performance laptop",
    "price": 1299.99
  }'
```

### Получить все продукты

```bash
curl http://localhost:8080/api/products
```

### Получить продукт по ID

```bash
curl http://localhost:8080/api/products/1
```

### Обновить продукт

```bash
curl -X PUT http://localhost:8080/api/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gaming Laptop",
    "description": "High-end gaming laptop",
    "price": 1599.99
  }'
```

### Удалить продукт

```bash
curl -X DELETE http://localhost:8080/api/products/1
```

## Полезные команды

### Работа с Docker

```bash
# Просмотр логов PostgreSQL
docker-compose logs -f postgres

# Подключение к PostgreSQL
docker exec -it go-base-postgres-1 psql -U developer -d gobase

# Остановка всех сервисов
make down-docker
```

### Работа с миграциями

```bash
# Создать новую миграцию
make migrate-create name=add_categories_table

# Применить все миграции
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Проверить статус миграций
make migrate-status
```

### Разработка

```bash
# Запустить статический анализ кода
make lint

# Запустить тесты
make test

# Установить зависимости
make deps

# Очистить артефакты сборки
make clean

# Показать все команды
make help
```

## Следующие шаги

1. 📖 Изучите [ARCHITECTURE.md](ARCHITECTURE.md) для понимания архитектуры
2. 🔧 Добавьте валидацию входных данных
3. 🧪 Напишите тесты для вашего кода
4. 📝 Настройте обработку ошибок
5. 🚀 Настройте CI/CD

**Удачи в разработке!** 🎉