# Настройка нового проекта на основе шаблона

Пошаговая инструкция для создания нового проекта на основе этого шаблона.

## Шаг 1: Клонирование

```bash
# Клонируйте репозиторий
git clone <template-repo-url> my-new-project
cd my-new-project

# Удалите историю git шаблона
rm -rf .git

# Инициализируйте новый репозиторий
git init
```

## Шаг 2: Переименование модуля

### 2.1 Определите новое имя модуля

Например: `github.com/yourusername/my-new-project`

### 2.2 Обновите go.mod

Откройте `go.mod` и измените первую строку:

```go
// Было:
module github.com/SmirnovND/gobase

// Стало:
module github.com/yourusername/my-new-project
```

### 2.3 Замените импорты во всех файлах

**Вариант 1: Используя find и sed (Linux/Mac):**

```bash
# Замените yourusername/my-new-project на ваш путь
find . -type f -name "*.go" -exec sed -i '' 's|github.com/SmirnovND/gobase|github.com/yourusername/my-new-project|g' {} +
```

**Вариант 2: Используя IDE:**

В GoLand/VSCode:
1. Нажмите `Cmd+Shift+R` (Mac) или `Ctrl+Shift+R` (Windows/Linux)
2. Найдите: `github.com/SmirnovND/gobase`
3. Замените на: `github.com/yourusername/my-new-project`
4. Замените во всех файлах

**Вариант 3: Вручную:**

Обновите импорты в следующих файлах:
- `cmd/server/main.go`
- `internal/container/container.go`
- `internal/router/router.go`
- Все файлы в `internal/controllers/`

### 2.4 Обновите зависимости

```bash
go mod tidy
```

## Шаг 3: Настройка конфигурации

```bash
# Создайте конфигурационный файл
cp config.example.yaml config.yaml

# Отредактируйте config.yaml под ваши нужды
# Измените порт, DSN базы данных и т.д.
```

## Шаг 4: Настройка базы данных

### 4.1 Обновите docker-compose.yml

Измените имя базы данных:

```yaml
environment:
  - POSTGRES_DB=my_new_project  # Было: gobase
  - POSTGRES_USER=developer
  - POSTGRES_PASSWORD=developer
```

### 4.2 Обновите Makefile

Измените переменную `DB_DSN`:

```makefile
DB_DSN=postgresql://developer:developer@localhost:5432/my_new_project?sslmode=disable
```

### 4.3 Обновите config.yaml

```yaml
db:
  dsn: "postgresql://developer:developer@localhost:5432/my_new_project?sslmode=disable"
```

## Шаг 5: Запуск проекта

```bash
# Установите зависимости
make deps

# Запустите PostgreSQL
make up-docker

# Подождите несколько секунд, затем примените миграции
make migrate-up

# Запустите сервер
make up-server
```

## Шаг 6: Проверка

```bash
# Проверьте healthcheck
curl http://localhost:8080/ping

# Должен вернуть статус 200 OK
```

## Шаг 7: Первый коммит

```bash
git add .
git commit -m "Initial commit from template"

# Добавьте remote репозиторий
git remote add origin <your-repo-url>
git push -u origin main
```

## Шаг 8: Начало разработки

Теперь вы готовы к разработке! Следуйте инструкциям в:

- `QUICKSTART.md` - для создания нового функционала
- `ARCHITECTURE.md` - для понимания архитектуры
- README файлы в каждом слое - для примеров кода

## Опциональные шаги

### Переименование проекта в файлах

Если вы хотите изменить название проекта в документации:

```bash
# Замените "thinker" на название вашего проекта
find . -type f \( -name "*.md" -o -name "*.yaml" \) -exec sed -i '' 's/thinker/my-new-project/g' {} +
```

### Настройка CI/CD

Добавьте `.github/workflows/` для GitHub Actions или `.gitlab-ci.yml` для GitLab CI.

### Добавление дополнительных зависимостей

```bash
# Пример: добавление Redis
go get github.com/go-redis/redis/v8

# Пример: добавление валидатора
go get github.com/go-playground/validator/v10

# Обновите зависимости
go mod tidy
```

## Проверочный список

- [ ] Клонирован репозиторий
- [ ] Удалена история git шаблона
- [ ] Обновлен `go.mod`
- [ ] Заменены все импорты
- [ ] Создан `config.yaml`
- [ ] Обновлен `docker-compose.yml`
- [ ] Обновлен `Makefile`
- [ ] Запущен PostgreSQL
- [ ] Применены миграции
- [ ] Сервер запускается
- [ ] Healthcheck работает
- [ ] Сделан первый коммит

## Возможные проблемы

### Ошибка "module not found"

```bash
# Убедитесь, что заменили все импорты
grep -r "github.com/SmirnovND/gobase" .

# Обновите зависимости
go mod tidy
```

### Ошибка подключения к БД

```bash
# Проверьте, что PostgreSQL запущен
docker ps

# Проверьте логи
docker-compose logs postgres

# Проверьте DSN в config.yaml
```

### Порт уже занят

Измените порт в `config.yaml`:

```yaml
app:
  run_addr: "localhost:8081"  # Вместо 8080
```

## Полезные команды

```bash
# Просмотр всех импортов
go list -f '{{.ImportPath}}: {{.Imports}}' ./...

# Проверка на ошибки
go vet ./...

# Форматирование кода
go fmt ./...

# Запуск тестов
go test ./...
```

Удачи с новым проектом! 🎉