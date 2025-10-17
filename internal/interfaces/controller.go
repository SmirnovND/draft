package interfaces

import "net/http"

// HealthcheckController интерфейс контроллера
type HealthcheckController interface {
	HandlePing(w http.ResponseWriter, r *http.Request)
}
