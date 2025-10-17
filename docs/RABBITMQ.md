# RabbitMQ в Go-Base

Этот документ описывает использование RabbitMQ компонента из go-toolbox в проекте go-base.

## Оглавление
- [Обзор](#обзор)
- [Настройка](#настройка)
- [Использование](#использование)
- [Примеры кода](#примеры-кода)
- [Лучшие практики](#лучшие-практики)

## Обзор

Компонент RabbitMQ в go-toolbox предоставляет следующие возможности:

- **Управление подключением** - создание и управление подключением к RabbitMQ
- **Производитель сообщений (Producer)** - публикация сообщений с поддержкой задержки доставки
- **Потребитель сообщений (Consumer)** - получение и обработка сообщений из очередей

## Настройка

### 1. Конфигурация

В файле `config.yaml` добавьте секцию для RabbitMQ:

```yaml
db:
  dsn: "postgresql://developer:developer@localhost:5432/gobase?sslmode=disable"

app:
  run_addr: "localhost:8080"

rabbitmq:
  url: "amqp://guest:guest@localhost:5672/"
```

Доступные параметры:
- `url` - строка подключения в формате `amqp://username:password@host:port/vhost`

## Использование

### Инициализация в DI контейнере

Компонент автоматически инициализируется в контейнере зависимостей и доступен для инжекта:

```go
type MyService struct {
    producer *rabbitmq.RabbitMQProducer
    consumer *rabbitmq.RabbitMQConsumer
}

func NewMyService(producer *rabbitmq.RabbitMQProducer, consumer *rabbitmq.RabbitMQConsumer) *MyService {
    return &MyService{
        producer: producer,
        consumer: consumer,
    }
}
```

### Producer (Производитель)

#### Публикация простого сообщения без задержки

```go
import (
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
)

// Публикация сообщения в exchange
err := producer.Publish(
    []byte("Hello, World!"),
    0,                    // Без задержки
    "my_exchange",        // Имя exchange
    "routing.key",        // Routing key
)
if err != nil {
    log.Printf("Failed to publish message: %v", err)
}
```

#### Публикация сообщения с задержкой

```go
import (
    "time"
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
)

// Публикация сообщения с задержкой 5 секунд
err := producer.Publish(
    []byte("Delayed message"),
    5 * time.Second,      // Задержка
    "delayed_exchange",   // Имя exchange для отложенных сообщений
    "task.process",       // Routing key
)
if err != nil {
    log.Printf("Failed to publish delayed message: %v", err)
}
```

### Consumer (Потребитель)

#### Потребление сообщений с обработкой

```go
import (
    "context"
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
)

// Начать потребление сообщений
messages, err := consumer.Consume()
if err != nil {
    log.Printf("Failed to start consuming: %v", err)
    return
}

// Обработка сообщений
for msg := range messages {
    // Обработка тела сообщения
    body := msg.Body
    log.Printf("Received message: %s", string(body))
    
    // Подтверждение доставки
    msg.Ack(false)
    
    // Или отклонение с повторной доставкой
    // msg.Nack(false, true)
}
```

#### Потребление с контекстом отмены

```go
import (
    "context"
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
)

func ConsumeMessages(ctx context.Context, consumer *rabbitmq.RabbitMQConsumer) error {
    messages, err := consumer.Consume()
    if err != nil {
        return err
    }

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case msg := <-messages:
            if msg.DeliveryTag == 0 {
                return nil // Consumer закрыт
            }
            
            // Обработка сообщения
            log.Printf("Processing: %s", string(msg.Body))
            msg.Ack(false)
        }
    }
}
```

## Примеры кода

### Пример 1: Использование в контроллере

```go
package controllers

import (
    "encoding/json"
    "time"
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
    "go.uber.org/zap"
)

type TaskController struct {
    producer *rabbitmq.RabbitMQProducer
    logger   *zap.Logger
}

func NewTaskController(producer *rabbitmq.RabbitMQProducer, logger *zap.Logger) *TaskController {
    return &TaskController{
        producer: producer,
        logger:   logger,
    }
}

// PublishTask публикует задачу на обработку
func (tc *TaskController) PublishTask(taskID string, data map[string]interface{}) error {
    payload, err := json.Marshal(map[string]interface{}{
        "id":   taskID,
        "data": data,
    })
    if err != nil {
        return err
    }

    err = tc.producer.Publish(
        payload,
        5 * time.Second,      // Задержка перед обработкой
        "tasks_exchange",
        "task.create",
    )
    
    if err != nil {
        tc.logger.Error("Failed to publish task", zap.Error(err))
        return err
    }

    tc.logger.Info("Task published", zap.String("taskID", taskID))
    return nil
}
```

### Пример 2: Использование в фоновом воркере

```go
package workers

import (
    "context"
    "encoding/json"
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
    "go.uber.org/zap"
)

type TaskWorker struct {
    consumer *rabbitmq.RabbitMQConsumer
    logger   *zap.Logger
}

func NewTaskWorker(consumer *rabbitmq.RabbitMQConsumer, logger *zap.Logger) *TaskWorker {
    return &TaskWorker{
        consumer: consumer,
        logger:   logger,
    }
}

// Start начинает обработку задач
func (w *TaskWorker) Start(ctx context.Context) error {
    messages, err := w.consumer.Consume()
    if err != nil {
        return err
    }

    for {
        select {
        case <-ctx.Done():
            w.logger.Info("Task worker stopped")
            return ctx.Err()
        case msg := <-messages:
            if msg.DeliveryTag == 0 {
                w.logger.Info("Consumer closed")
                return nil
            }

            if err := w.processMessage(msg.Body); err != nil {
                w.logger.Error("Failed to process message", zap.Error(err))
                msg.Nack(false, true) // Повторить доставку
                continue
            }

            msg.Ack(false) // Подтверждение успешной обработки
        }
    }
}

func (w *TaskWorker) processMessage(body []byte) error {
    var task map[string]interface{}
    if err := json.Unmarshal(body, &task); err != nil {
        return err
    }

    w.logger.Info("Processing task", zap.Any("task", task))
    // Логика обработки задачи
    return nil
}
```

### Пример 3: Использование в сервисе

```go
package services

import (
    "encoding/json"
    "time"
    "github.com/SmirnovND/toolbox/pkg/rabbitmq"
    "go.uber.org/zap"
)

type NotificationService struct {
    producer *rabbitmq.RabbitMQProducer
    logger   *zap.Logger
}

func NewNotificationService(producer *rabbitmq.RabbitMQProducer, logger *zap.Logger) *NotificationService {
    return &NotificationService{
        producer: producer,
        logger:   logger,
    }
}

// SendNotification отправляет уведомление
func (ns *NotificationService) SendNotification(userID string, notification map[string]interface{}) error {
    payload, err := json.Marshal(map[string]interface{}{
        "user_id": userID,
        "data":    notification,
    })
    if err != nil {
        return err
    }

    // Публикация с задержкой - например, для батчинга уведомлений
    err = ns.producer.Publish(
        payload,
        2 * time.Second,
        "notifications_exchange",
        "notification.send",
    )
    
    if err != nil {
        ns.logger.Error("Failed to send notification", zap.Error(err))
        return err
    }

    ns.logger.Info("Notification queued", zap.String("userID", userID))
    return nil
}
```

## Лучшие практики

### 1. Управление жизненным циклом

```go
// Consumer и Producer должны быть закрыты при завершении приложения
defer func() {
    if producer != nil {
        producer.Close()
    }
    if consumer != nil {
        consumer.Close()
    }
    if conn != nil {
        conn.Close()
    }
}()
```

### 2. Обработка ошибок

- Всегда проверяйте ошибки при публикации сообщений
- Используйте `Nack()` с флагом `requeue=true` для повторной обработки при ошибках
- Логируйте ошибки для диагностики проблем

```go
if err := producer.Publish(payload, delay, exchange, key); err != nil {
    logger.Error("Publish failed", zap.Error(err))
    // Обработайте ошибку (retry, fallback, etc.)
}
```

### 3. Структурирование сообщений

Используйте структурированные форматы для сообщений:

```go
type Message struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

payload, _ := json.Marshal(Message{
    ID:        uuid.New().String(),
    Type:      "task.created",
    Timestamp: time.Now(),
    Data:      taskData,
})
```

### 4. Масштабирование потребления

Для обработки больших объемов сообщений используйте несколько worker горутин:

```go
for i := 0; i < numWorkers; i++ {
    go func(workerID int) {
        if err := worker.Start(ctx); err != nil {
            logger.Error("Worker failed", 
                zap.Int("workerID", workerID),
                zap.Error(err),
            )
        }
    }(i)
}
```

### 5. Мониторинг и логирование

```go
logger.Info("Message published",
    zap.String("exchange", exchange),
    zap.String("routingKey", key),
    zap.Duration("delay", delay),
)

logger.Info("Message processed",
    zap.String("messageID", msg.ID),
    zap.Duration("processingTime", time.Since(startTime)),
)
```