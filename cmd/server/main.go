package main

import (
	"errors"
	"github.com/SmirnovND/gobase/internal/container"
	"github.com/SmirnovND/gobase/internal/interfaces"
	"github.com/SmirnovND/gobase/internal/router"
	"github.com/SmirnovND/toolbox/pkg/logger"
	"github.com/SmirnovND/toolbox/pkg/middleware"
	"github.com/SmirnovND/toolbox/pkg/migrations"
	"github.com/SmirnovND/toolbox/pkg/rabbitmq"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"

	_ "github.com/SmirnovND/gobase/docs" // Swagger docs
)

// @title           GoBase API
// @version         1.0
// @description     REST API для GoBase проекта
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @schemes http https
func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}

func Run() error {
	diContainer := container.NewContainer()

	var cf interfaces.ConfigServer
	if err := diContainer.Invoke(func(c interfaces.ConfigServer) {
		cf = c
	}); err != nil {
		return err
	}

	var dbx *sqlx.DB
	if err := diContainer.Invoke(func(db *sqlx.DB) {
		dbx = db
	}); err != nil {
		return err
	}

	if dbx == nil {
		return errors.New("database connection is nil")
	}

	dbBase := dbx.DB
	migrations.StartMigrations(dbBase)

	// Инициализация RabbitMQ компонентов
	var rmqConn *rabbitmq.RabbitMQConnection
	var rmqProducer *rabbitmq.RabbitMQProducer
	var rmqConsumer *rabbitmq.RabbitMQConsumer

	if err := diContainer.Invoke(func(conn *rabbitmq.RabbitMQConnection) {
		rmqConn = conn
	}); err != nil {
		return err
	}

	if err := diContainer.Invoke(func(producer *rabbitmq.RabbitMQProducer) {
		rmqProducer = producer
	}); err != nil {
		return err
	}

	if err := diContainer.Invoke(func(consumer *rabbitmq.RabbitMQConsumer) {
		rmqConsumer = consumer
	}); err != nil {
		return err
	}

	log.Println("RabbitMQ initialized successfully")
	defer func() {
		if rmqProducer != nil {
			rmqProducer.Close()
			log.Println("RabbitMQ producer closed")
		}
		if rmqConsumer != nil {
			rmqConsumer.Close()
			log.Println("RabbitMQ consumer closed")
		}
		if rmqConn != nil {
			rmqConn.Close()
			log.Println("RabbitMQ connection closed")
		}
	}()

	return http.ListenAndServe(cf.GetRunAddr(), middleware.ChainMiddleware(
		router.Handler(diContainer),
		httplog.WithLogging,
	))
}
