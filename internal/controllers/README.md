# Controllers

Слой HTTP-контроллеров для обработки запросов.

## Пример контроллера

```go
package controllers

import (
	"encoding/json"
	"github.com/SmirnovND/gobase/internal/usecases"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type UserController struct {
	userUsecase *usecases.UserUsecase
}

func NewUserController(userUsecase *usecases.UserUsecase) *UserController {
	return &UserController{
		userUsecase: userUsecase,
	}
}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := c.userUsecase.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := c.userUsecase.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
```

## Регистрация в DI контейнере

В файле `internal/container/container.go`:

```go
func (c *Container) provideController() {
	c.container.Provide(controllers.NewHealthcheckController)
	c.container.Provide(controllers.NewUserController)
}
```

## Добавление маршрутов

В файле `internal/router/router.go`:

```go
func Handler(diContainer *container.Container) http.Handler {
	var (
		HealthcheckController *controllers.HealthcheckController
		UserController        *controllers.UserController
	)
	
	err := diContainer.Invoke(func(
		d *sqlx.DB,
		c interfaces.ConfigServer,
		healthcheckControl *controllers.HealthcheckController,
		userControl *controllers.UserController,
	) {
		HealthcheckController = healthcheckControl
		UserController = userControl
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)

	r.Get("/ping", HealthcheckController.HandlePing)
	
	// API маршруты
	r.Route("/api/users", func(r chi.Router) {
		r.Get("/{id}", UserController.GetUser)
		r.Post("/", UserController.CreateUser)
		r.Put("/{id}", UserController.UpdateUser)
		r.Delete("/{id}", UserController.DeleteUser)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Route not found", http.StatusNotFound)
	})

	return r
}
```