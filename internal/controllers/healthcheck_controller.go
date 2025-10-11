package controllers

import (
	"github.com/jmoiron/sqlx"
	"net/http"
)

type HealthcheckController struct {
	DB *sqlx.DB
}

func NewHealthcheckController(DB *sqlx.DB) *HealthcheckController {
	return &HealthcheckController{
		DB: DB,
	}
}

// HandlePing godoc
// @Summary      Проверка здоровья сервиса
// @Description  Проверяет доступность сервиса и подключение к базе данных
// @Tags         healthcheck
// @Produce      plain
// @Success      200  {string}  string  "OK"
// @Failure      500  {string}  string  "Failed to connect DB"
// @Router       /ping [get]
func (hc *HealthcheckController) HandlePing(w http.ResponseWriter, r *http.Request) {
	err := hc.DB.PingContext(r.Context())
	if err != nil {
		http.Error(w, "Failed to connect DB", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
