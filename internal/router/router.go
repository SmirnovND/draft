package router

import (
	"fmt"
	"github.com/SmirnovND/gobase/internal/container"
	"github.com/SmirnovND/gobase/internal/controllers"
	"github.com/SmirnovND/gobase/internal/interfaces"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

func Handler(diContainer *container.Container) http.Handler {
	var HealthcheckController *controllers.HealthcheckController
	err := diContainer.Invoke(func(
		d *sqlx.DB,
		c interfaces.ConfigServer,
		healthcheckControl *controllers.HealthcheckController,
	) {
		HealthcheckController = healthcheckControl
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	r.Get("/ping", HealthcheckController.HandlePing)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Обработчик для неподходящего метода (405 Method Not Allowed)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Обработчик для несуществующих маршрутов (404 Not Found)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Route not found", http.StatusNotFound)
	})

	return r
}
