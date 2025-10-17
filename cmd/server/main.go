package main

import (
	"context"
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
	"os"
	"os/signal"
	"syscall"
	"time"

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
	defer diContainer.Close()

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

	// Инициализация RabbitMQ компонентов через контейнер (управление жизненным циклом)
	if err := diContainer.Invoke(func(conn *rabbitmq.RabbitMQConnection) {}); err != nil {
		return err
	}

	if err := diContainer.Invoke(func(producer *rabbitmq.RabbitMQProducer) {}); err != nil {
		return err
	}

	if err := diContainer.Invoke(func(consumer *rabbitmq.RabbitMQConsumer) {}); err != nil {
		return err
	}

	log.Println("RabbitMQ initialized successfully")

	// Создание HTTP сервера
	server := &http.Server{
		Addr: cf.GetRunAddr(),
		Handler: middleware.ChainMiddleware(
			router.Handler(diContainer),
			logger.WithLogging,
		),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Канал для ошибок при запуске сервера
	serverErrors := make(chan error, 1)

	// Запуск сервера в горутине
	go func() {
		log.Printf("Starting server on %s", cf.GetRunAddr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Обработка OS сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return err
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)

		// Graceful shutdown с timeout в 30 секунд
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Println("Shutting down server gracefully...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
			return err
		}

		log.Println("Server shut down successfully")

		// Контейнер закроется автоматически через defer
		return nil
	}
}
