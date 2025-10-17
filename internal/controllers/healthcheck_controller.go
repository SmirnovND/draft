package controllers

import (
	"encoding/json"
	"github.com/SmirnovND/gobase/internal/services"
	"net/http"
)

type HealthcheckController struct {
	healthcheckService *services.HealthcheckService
}

func NewHealthcheckController(healthcheckService *services.HealthcheckService) *HealthcheckController {
	return &HealthcheckController{
		healthcheckService: healthcheckService,
	}
}

// HandlePing godoc
// @Summary      Проверка здоровья сервиса
// @Description  Проверяет доступность сервиса и подключение к базе данных
// @Tags         healthcheck
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "OK"
// @Failure      500  {object}  map[string]interface{}  "Service unhealthy"
// @Router       /ping [get]
func (hc *HealthcheckController) HandlePing(w http.ResponseWriter, r *http.Request) {
	status, err := hc.healthcheckService.Check(r.Context())

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}
