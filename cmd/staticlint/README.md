# 🔍 Staticlint - Кастомный Multichecker

Кастомный инструмент статического анализа кода на основе `golang.org/x/tools/go/analysis/multichecker`.

## 🎯 Назначение

Staticlint объединяет множество анализаторов для комплексной проверки качества кода:
- Обнаружение потенциальных багов
- Проверка стиля кода
- Поиск проблем с производительностью
- Выявление небезопасных конструкций

## 🚀 Использование

### Через Makefile (рекомендуется)

```bash
make lint
```

### Напрямую

```bash
go run ./cmd/staticlint/main.go ./...
```

### Для конкретного пакета

```bash
go run ./cmd/staticlint/main.go ./internal/repositories
```

## 📋 Включенные анализаторы

### 1. Стандартные анализаторы (golang.org/x/tools)

| Анализатор | Описание |
|------------|----------|
| `printf` | Проверка корректности форматирования строк в функциях типа Printf |
| `shadow` | Обнаружение затенения переменных |
| `structtag` | Проверка корректности тегов структур (json, db, xml и т.д.) |
| `unreachable` | Поиск недостижимого кода |

### 2. Staticcheck SA (проверки на баги)

Все анализаторы категории **SA** из [staticcheck.io](https://staticcheck.io/docs/checks):

- `SA1000` - `SA1030`: Проверки на некорректное использование стандартной библиотеки
- `SA2000` - `SA2003`: Проверки на проблемы с конкурентностью
- `SA3000` - `SA3001`: Проверки на проблемы с тестированием
- `SA4000` - `SA4031`: Проверки на логические ошибки
- `SA5000` - `SA5012`: Проверки на некорректное использование языковых конструкций
- `SA6000` - `SA6005`: Проверки на проблемы с производительностью
- `SA9000` - `SA9008`: Проверки на дублирование кода и другие проблемы

**Примеры:**
- `SA1019` - использование deprecated функций
- `SA4006` - присваивание значения переменной, которая никогда не используется
- `SA5007` - бесконечные рекурсии

### 3. Staticcheck ST1000

- `ST1000` - проверка именования пакетов (должны быть lowercase без подчеркиваний)

### 4. Публичные анализаторы

| Анализатор | Описание | Репозиторий |
|------------|----------|-------------|
| `nilerr` | Обнаружение игнорирования ошибок (возврат nil вместо err) | [gostaticanalysis/nilerr](https://github.com/gostaticanalysis/nilerr) |
| `bodyclose` | Проверка закрытия `http.Response.Body` | [timakin/bodyclose](https://github.com/timakin/bodyclose) |

### 5. Кастомный анализатор

#### exitchecker

Запрещает прямые вызовы `os.Exit()` в функции `main` пакета `main`.

**Почему это важно:**
- `os.Exit()` немедленно завершает программу, минуя defer-функции
- Это может привести к утечкам ресурсов и некорректному завершению
- Лучше использовать `return` с кодом ошибки

**Плохо:**
```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
        os.Exit(1) // ❌ Будет обнаружено
    }
}
```

**Хорошо:**
```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err) // ✅ log.Fatal сам вызывает os.Exit
        return
    }
}
```

## 📊 Пример вывода

```
Enabled analyzers:
printf
shadow
structtag
unreachable
SA1000
SA1001
...
ST1000
nilerr
bodyclose
exitchecker

internal/repositories/user_repository.go:45:2: 
  this value of err is never used (SA4006)
internal/controllers/user_controller.go:78:15: 
  response body must be closed (bodyclose)
cmd/server/main.go:25:5: 
  direct call to os.Exit in main is not allowed (exitchecker)
```

## 🔧 Настройка

### Добавление нового анализатора

Отредактируйте `cmd/staticlint/main.go`:

```go
import "your/analyzer/package"

func main() {
    // ...
    analyzers = append(analyzers, youranalyzer.Analyzer)
    // ...
}
```

### Исключение анализатора

Закомментируйте или удалите соответствующую строку в `main.go`.

### Создание собственного анализатора

1. Создайте пакет в `cmd/staticlint/youranalyzer/`
2. Реализуйте `analysis.Analyzer`
3. Добавьте в список анализаторов в `main.go`

**Пример структуры:**

```go
package youranalyzer

import "golang.org/x/tools/go/analysis"

var Analyzer = &analysis.Analyzer{
    Name: "youranalyzer",
    Doc:  "description of your analyzer",
    Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
    // Ваша логика анализа
    return nil, nil
}
```

## 🎓 Полезные ссылки

- [Writing Go analysis passes](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- [Staticcheck documentation](https://staticcheck.io/docs/)
- [Go AST Viewer](https://yuroyoro.github.io/goast-viewer/) - для изучения AST

## 💡 Советы

1. **Запускайте регулярно**: Добавьте `make lint` в CI/CD pipeline
2. **Исправляйте постепенно**: Начните с критичных ошибок (SA категория)
3. **Изучайте предупреждения**: Каждое предупреждение - возможность улучшить код
4. **Настраивайте под проект**: Отключайте неактуальные анализаторы

## 🐛 Известные ограничения

- Анализ может занимать время на больших проектах
- Некоторые анализаторы могут давать ложные срабатывания
- Не все проблемы могут быть обнаружены статическим анализом

## 📝 Интеграция с IDE

### GoLand / IntelliJ IDEA

1. Settings → Tools → File Watchers
2. Добавьте новый watcher для `.go` файлов
3. Program: `make`
4. Arguments: `lint`

### VS Code

Добавьте в `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "lint",
      "type": "shell",
      "command": "make lint",
      "group": "test"
    }
  ]
}
```

---

**Создано для обеспечения высокого качества кода в Go проектах** 🚀