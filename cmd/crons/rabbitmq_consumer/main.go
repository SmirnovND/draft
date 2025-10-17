package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SmirnovND/gobase/internal/container"
	"github.com/SmirnovND/toolbox/pkg/rabbitmq"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "consumer failed: %v\n", err)
		os.Exit(1)
	}
}

func Run() error {
	diContainer := container.NewContainer()
	defer diContainer.Close()

	var logger *zap.Logger
	if err := diContainer.Invoke(func(l *zap.Logger) {
		logger = l
	}); err != nil {
		return err
	}

	var conn *rabbitmq.RabbitMQConnection
	if err := diContainer.Invoke(func(c *rabbitmq.RabbitMQConnection) {
		conn = c
	}); err != nil {
		return err
	}

	if conn == nil {
		return fmt.Errorf("failed to create RabbitMQ connection")
	}

	// Создаем consumer для конкретной очереди
	// Здесь используется очередь "tasks_queue", замените на вашу
	consumer := rabbitmq.NewRabbitMQConsumer(conn.Conn, "tasks_queue")
	defer func() {
		consumer.Close()
		logger.Info("Consumer closed")
	}()

	logger.Info("RabbitMQ consumer started")

	// Создаем контекст для отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработчик сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received signal", zap.String("signal", sig.String()))
		cancel()
	}()

	// Запускаем потребление сообщений
	return consumeMessages(ctx, consumer, logger)
}

// consumeMessages осуществляет потребление и обработку сообщений
func consumeMessages(ctx context.Context, consumer *rabbitmq.RabbitMQConsumer, logger *zap.Logger) error {
	messages, err := consumer.Consume()
	if err != nil {
		logger.Error("Failed to start consuming messages", zap.Error(err))
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("Consumer context cancelled")
			return nil

		case msg := <-messages:
			if msg.DeliveryTag == 0 {
				logger.Info("Consumer closed")
				return nil
			}

			startTime := time.Now()

			// Обработка сообщения
			if err := handleMessage(msg.Body, logger); err != nil {
				logger.Error("Failed to process message", zap.Error(err))
				// Отклоняем сообщение и возвращаем в очередь для повторной обработки
				msg.Nack(false, true)
				continue
			}

			// Подтверждаем успешную обработку
			msg.Ack(false)
			logger.Info("Message processed successfully",
				zap.Duration("processingTime", time.Since(startTime)),
			)
		}
	}
}

// handleMessage обрабатывает полученное сообщение
// Рекомендуемый паттерн использования:
// 1. Распарсьте сообщение в нужную вам структуру (task, event, и т.д.)
// 2. Вызовите соответствующий UseCase или Service через DI контейнер
// 3. НЕ размещайте бизнес-логику прямо здесь - это просто транспортный слой
//
// Пример правильного использования:
//
//	func handleMessage(ctx context.Context, body []byte, taskUseCase interfaces.TaskUseCase, logger *zap.Logger) error {
//	    var task domain.Task
//	    if err := json.Unmarshal(body, &task); err != nil {
//	        return err
//	    }
//	    // Вызываем use case - вся бизнес-логика там
//	    return taskUseCase.ProcessTask(ctx, task)
//	}
func handleMessage(body []byte, logger *zap.Logger) error {
	var message map[string]interface{}
	if err := json.Unmarshal(body, &message); err != nil {
		logger.Error("Failed to unmarshal message", zap.Error(err))
		return err
	}

	logger.Info("Processing message", zap.Any("message", message))

	// TODO: Получите нужный UseCase/Service из DI контейнера
	// и вызовите его метод для обработки сообщения.
	// Не добавляйте бизнес-логику прямо сюда!

	return nil
}
