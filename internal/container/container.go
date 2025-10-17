package container

import (
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
)

// Container - структура контейнера, обертывающая dig-контейнер
type Container struct {
	container *dig.Container
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
		return logger
	})

	// Регистрируем RabbitMQ компоненты
	c.container.Provide(func(configServer interfaces.ConfigServer) *rabbitmq.RabbitMQConnection {
		return rabbitmq.NewRabbitMQConnection(configServer.GetRabbitMQURL())
	})

	c.container.Provide(func(conn *rabbitmq.RabbitMQConnection) *rabbitmq.RabbitMQProducer {
		return rabbitmq.NewRabbitMQProducer(conn.Conn)
	})

	c.container.Provide(func(conn *rabbitmq.RabbitMQConnection) *rabbitmq.RabbitMQConsumer {
		// Используется стандартное имя очереди "default_queue"
		// Для специфичных очередей создавайте отдельные consumer в сервисах
		return rabbitmq.NewRabbitMQConsumer(conn.Conn, "default_queue")
	})
}

// provideUsecase - регистрация use case слоя
func (c *Container) provideUsecase() {
}

// provideRepo - регистрация репозиториев
func (c *Container) provideRepo() {
	c.container.Provide(repositories.NewHealthcheckRepository)
}

// provideService - регистрация сервисов
func (c *Container) provideService() {
	c.container.Provide(services.NewHealthcheckService)
}

// provideController - регистрация контроллеров
func (c *Container) provideController() {
	c.container.Provide(controllers.NewHealthcheckController)
}

// Invoke - функция для вызова и инжекта зависимостей
func (c *Container) Invoke(function interface{}) error {
	return c.container.Invoke(function)
}
