package interfaces

import "context"

// HealthcheckRepository интерфейс для репозитория
type HealthcheckRepository interface {
	Ping(ctx context.Context) error
}
