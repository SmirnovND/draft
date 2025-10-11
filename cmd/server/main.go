package main

import (
	"github.com/SmirnovND/gobase/internal/container"
	"github.com/SmirnovND/gobase/internal/interfaces"
	"github.com/SmirnovND/gobase/internal/router"
	"github.com/SmirnovND/toolbox/pkg/logger"
	"github.com/SmirnovND/toolbox/pkg/middleware"
	"github.com/SmirnovND/toolbox/pkg/migrations"
	"github.com/jmoiron/sqlx"
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
	diContainer.Invoke(func(c interfaces.ConfigServer) {
		cf = c
	})

	var dbx *sqlx.DB
	diContainer.Invoke(func(db *sqlx.DB) {
		dbx = db
	})

	dbBase := dbx.DB
	migrations.StartMigrations(dbBase)

	return http.ListenAndServe(cf.GetRunAddr(), middleware.ChainMiddleware(
		router.Handler(diContainer),
		logger.WithLogging,
	))
}
