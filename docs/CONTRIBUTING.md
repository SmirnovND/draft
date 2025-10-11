# 🤝 Руководство по разработке

Это руководство поможет вам эффективно работать с шаблоном проекта.

## 📋 Содержание

- [Стандарты кода](#стандарты-кода)
- [Структура коммитов](#структура-коммитов)
- [Процесс разработки](#процесс-разработки)
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

## 📝 Структура коммитов

Используйте [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Типы коммитов

- `feat`: новая функциональность
- `fix`: исправление бага
- `docs`: изменения в документации
- `style`: форматирование кода
- `refactor`: рефакторинг
- `test`: добавление тестов
- `chore`: обновление зависимостей, конфигурации

### Примеры

```bash
feat(user): add user registration endpoint

- Add UserRepository.Create method
- Add UserUsecase.Register method
- Add UserController.Register handler
- Add POST /api/users route

Closes #123
```

```bash
fix(auth): handle expired tokens correctly

Previously expired tokens were not properly validated,
causing security issues.
```

## 🔄 Процесс разработки

### 1. Создание новой функции

```bash
# 1. Создайте ветку
git checkout -b feat/user-registration

# 2. Создайте миграцию (если нужно)
make migrate-create name=add_users_table

# 3. Разработайте функцию (см. QUICKSTART.md)

# 4. Запустите статический анализ
make lint

# 5. Запустите тесты
make test

# 6. Проверьте код
go vet ./...
go fmt ./...

# 7. Закоммитьте изменения
git add .
git commit -m "feat(user): add user registration"

# 8. Отправьте в репозиторий
git push origin feat/user-registration
```

### 2. Порядок разработки слоев

Следуйте принципу "изнутри наружу":

1. **Domain** → Создайте модель
2. **Repository** → Реализуйте работу с БД
3. **Usecase** → Добавьте бизнес-логику
4. **Controller** → Создайте HTTP-обработчик
5. **Router** → Зарегистрируйте маршрут
6. **Container** → Добавьте в DI

### 3. Работа с миграциями

```bash
# Создать миграцию
make migrate-create name=add_email_to_users

# Применить миграции
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Проверить статус
make migrate-status
```

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

### Версионирование

Используем [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): Новая функциональность (обратно совместимая)
- **PATCH** (0.0.1): Исправления багов

### Создание релиза

```bash
# 1. Обновите CHANGELOG.md
# 2. Создайте тег
git tag -a v1.0.0 -m "Release v1.0.0"

# 3. Отправьте тег
git push origin v1.0.0
```

## 📚 Полезные ссылки

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

## 💡 Советы

1. **Держите функции маленькими** - одна функция = одна ответственность
2. **Используйте интерфейсы** - это упрощает тестирование
3. **Не игнорируйте ошибки** - всегда обрабатывайте их
4. **Пишите тесты** - они экономят время в будущем
5. **Документируйте публичные API** - используйте godoc комментарии
6. **Логируйте важные события** - это помогает при отладке

## ❓ Вопросы?

Если у вас есть вопросы или предложения:

1. Проверьте документацию в `/docs`
2. Посмотрите примеры в README файлах слоев
3. Создайте issue в репозитории