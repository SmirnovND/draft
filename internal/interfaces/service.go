package interfaces

import "context"

// HealthcheckService интерфейс сервиса проверки здоровья
type HealthcheckService interface {
	Check(ctx context.Context) (map[string]interface{}, error)
}