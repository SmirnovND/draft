package container

import (
	config "github.com/SmirnovND/gobase/internal/config/server"
	"github.com/SmirnovND/gobase/internal/controllers"
	"github.com/SmirnovND/gobase/internal/interfaces"
	"github.com/SmirnovND/toolbox/pkg/db"
	"github.com/SmirnovND/toolbox/pkg/http"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/dig"
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
	c.container.Provide(config.NewConfig())
	c.container.Provide(func(configServer interfaces.ConfigServer) *sqlx.DB {
		return db.NewDB(configServer.GetDBDsn())
	})
	c.container.Provide(db.NewTransactionManager)
	c.container.Provide(http.NewAPIClient)
}

// provideUsecase - регистрация use case слоя
func (c *Container) provideUsecase() {
	// Пример: c.container.Provide(usecase.NewUserUsecase)
}

// provideRepo - регистрация репозиториев
func (c *Container) provideRepo() {
	// Пример: c.container.Provide(repository.NewUserRepository)
}

// provideService - регистрация сервисов
func (c *Container) provideService() {
	// Пример: c.container.Provide(service.NewEmailService)
}

// provideController - регистрация контроллеров
func (c *Container) provideController() {
	c.container.Provide(controllers.NewHealthcheckController)
	// Пример: c.container.Provide(controllers.NewUserController)
}

// Invoke - функция для вызова и инжекта зависимостей
func (c *Container) Invoke(function interface{}) error {
	return c.container.Invoke(function)
}
