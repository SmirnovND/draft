# 🤝 Руководство по разработке

Это руководство поможет вам эффективно работать с шаблоном проекта.

## 📋 Содержание

- [Стандарты кода](#стандарты-кода)
- [Тестирование](#тестирование)

## 🎯 Стандарты кода

### Именование

- **Файлы**: snake_case (`user_repository.go`)
- **Пакеты**: lowercase, короткие (`user`, `auth`, `http`)
- **Интерфейсы**: PascalCase с суффиксом (например, `UserRepository`, `ConfigServer`)
- **Структуры**: PascalCase (`User`, `Product`)
- **Методы**: PascalCase для экспортируемых, camelCase для приватных

### Организация кода

```go
// 1. Package declaration
package user

// 2. Imports (стандартная библиотека, затем внешние пакеты)
import (
    "context"
    "time"
    
    "github.com/jmoiron/sqlx"
)

// 3. Constants
const (
    DefaultTimeout = 30 * time.Second
)

// 4. Types
type User struct {
    ID    int64  `db:"id"`
    Name  string `db:"name"`
}

// 5. Constructor
func NewUser(name string) *User {
    return &User{Name: name}
}

// 6. Methods
func (u *User) Validate() error {
    // ...
}
```

### Обработка ошибок

```go
// ✅ Хорошо
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// ❌ Плохо
if err != nil {
    return err
}
```

### Контекст

Всегда передавайте `context.Context` первым параметром:

```go
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    // ...
}
```

### Статический анализ

Перед коммитом всегда запускайте линтер:

```bash
make lint
```

Проект использует кастомный multichecker с 20+ анализаторами:
- Стандартные анализаторы Go (printf, shadow, structtag, unreachable)
- Все анализаторы SA из staticcheck
- Публичные анализаторы (nilerr, bodyclose)
- Кастомный анализатор exitchecker

📖 **Подробнее:** [cmd/staticlint/README.md](../cmd/staticlint/README.md)

**Важно:** Исправляйте все найденные проблемы перед созданием PR.

## 🧪 Тестирование

### Структура тестов

```
internal/
├── repositories/
│   ├── user_repository.go
│   └── user_repository_test.go
├── usecases/
│   ├── user_usecase.go
│   └── user_usecase_test.go
```

### Пример теста

```go
package repositories_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    user := &User{Name: "John Doe"}
    
    // Act
    err := repo.Create(context.Background(), user)
    
    // Assert
    require.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

### Запуск тестов

```bash
# Все тесты
make test

# Конкретный пакет
go test ./internal/repositories/...

# С покрытием
go test -cover ./...

# Verbose
go test -v ./...
```

## 🔍 Code Review

### Чеклист для ревьюера

- [ ] Код следует архитектуре проекта
- [ ] Все ошибки обрабатываются корректно
- [ ] Добавлены необходимые тесты
- [ ] Документация обновлена
- [ ] Нет хардкода (используется конфигурация)
- [ ] Логирование добавлено в критических местах
- [ ] SQL-запросы защищены от инъекций
- [ ] Контекст передается корректно

### Чеклист для автора

- [ ] Код прошел `go vet` и `go fmt`
- [ ] Все тесты проходят
- [ ] Миграции применяются и откатываются корректно
- [ ] Конфигурация обновлена (если нужно)
- [ ] README обновлен (если нужно)
- [ ] Коммит следует Conventional Commits

## 🚀 Релизы

## 💡 Советы

1. **Держите функции маленькими** - одна функция = одна ответственность
2. **Используйте интерфейсы** - это упрощает тестирование
3. **Не игнорируйте ошибки** - всегда обрабатывайте их
4. **Пишите тесты** - они экономят время в будущем
5. **Документируйте публичные API** - используйте godoc комментарии
6. **Логируйте важные события** - это помогает при отладке