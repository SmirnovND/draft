package container

import (
	"context"
	"github.com/SmirnovND/gobase/internal/adapter"
	config "github.com/SmirnovND/gobase/internal/config/server"
	"github.com/SmirnovND/gobase/internal/controllers"
	"github.com/SmirnovND/gobase/internal/interfaces"
	"github.com/SmirnovND/gobase/internal/repositories"
	"github.com/SmirnovND/gobase/internal/services"
	"github.com/SmirnovND/toolbox/pkg/db"
	"github.com/SmirnovND/toolbox/pkg/http"
	"github.com/SmirnovND/toolbox/pkg/rabbitmq"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"log"
)

// Container - структура контейнера, обертывающая dig-контейнер
type Container struct {
	container *dig.Container
	closers   []interfaces.Closer
	loggers   []interfaces.LoggerCloser
}

func NewContainer() *Container {
	c := &Container{container: dig.New()}
	c.provideDependencies()
	c.provideRepo()
	c.provideService()
	c.provideUsecase()
	c.provideController()
	return c
}

// provideDependencies - функция, регистрирующая зависимости
func (c *Container) provideDependencies() {
	// Регистрируем конфигурацию
	c.container.Provide(config.NewConfig)
	c.container.Provide(func(configServer interfaces.ConfigServer) *sqlx.DB {
		return db.NewDB(configServer.GetDBDsn())
	})
	c.container.Provide(db.NewTransactionManager)
	c.container.Provide(http.NewAPIClient)
	c.container.Provide(func() *zap.Logger {
		logger, _ := zap.NewProduction()
		c.loggers = append(c.loggers, logger)
		return logger
	})

	// Регистрируем RabbitMQ компоненты
	c.container.Provide(func(configServer interfaces.ConfigServer) *rabbitmq.RabbitMQConnection {
		conn := rabbitmq.NewRabbitMQConnection(configServer.GetRabbitMQURL())
		c.closers = append(c.closers, adapter.NewRabbitMQConnectionCloser(conn))
		return conn
	})

	c.container.Provide(func(conn *rabbitmq.RabbitMQConnection) *rabbitmq.RabbitMQProducer {
		producer := rabbitmq.NewRabbitMQProducer(conn.Conn)
		c.closers = append(c.closers, adapter.NewRabbitMQProducerCloser(producer))
		return producer
	})

	c.container.Provide(func(conn *rabbitmq.RabbitMQConnection) *rabbitmq.RabbitMQConsumer {
		// Используется стандартное имя очереди "default_queue"
		// Для специфичных очередей создавайте отдельные consumer в сервисах
		consumer := rabbitmq.NewRabbitMQConsumer(conn.Conn, "default_queue")
		c.closers = append(c.closers, adapter.NewRabbitMQConsumerCloser(consumer))
		return consumer
	})
}

// provideUsecase - регистрация use case слоя
func (c *Container) provideUsecase() {
}

// provideRepo - регистрация репозиториев
// dig смотрит на сигнатуру NewHealthcheckRepository(db *sqlx.DB) interfaces.HealthcheckRepository
// и автоматически регистрирует interfaces.HealthcheckRepository
func (c *Container) provideRepo() {
	c.container.Provide(repositories.NewHealthcheckRepository)
}

// provideService - регистрация сервисов
// dig смотрит на сигнатуру NewHealthcheckService(repo interfaces.HealthcheckRepository) interfaces.HealthcheckService
// находит interfaces.HealthcheckRepository (уже зарегистрирован в provideRepo)
// и автоматически вызывает конструктор с правильными аргументами
func (c *Container) provideService() {
	c.container.Provide(services.NewHealthcheckService)
}

// provideController - регистрация контроллеров
// dig смотрит на сигнатуру NewHealthcheckController(service interfaces.HealthcheckService) interfaces.HealthcheckController
// находит interfaces.HealthcheckService (уже зарегистрирован в provideService)
// и автоматически вызывает конструктор с правильными аргументами
func (c *Container) provideController() {
	c.container.Provide(controllers.NewHealthcheckController)
}

// Invoke - функция для вызова и инжекта зависимостей
func (c *Container) Invoke(function interface{}) error {
	return c.container.Invoke(function)
}

// Shutdown - graceful shutdown контейнера и всех зависимостей
func (c *Container) Shutdown(ctx context.Context) error {
	// Сначала синхронизируем логгеры (flush buffers)
	for _, logger := range c.loggers {
		if err := logger.Sync(); err != nil {
			log.Printf("error syncing logger: %v", err)
		}
	}

	// Затем закрываем остальные компоненты в обратном порядке
	for i := len(c.closers) - 1; i >= 0; i-- {
		if err := c.closers[i].Close(); err != nil {
			log.Printf("error closing component: %v", err)
		}
	}

	return nil
}

// Close - закрытие контейнера без контекста (для defer)
func (c *Container) Close() error {
	return c.Shutdown(context.Background())
}
