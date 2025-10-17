package services

import (
	"context"
	"github.com/SmirnovND/gobase/internal/interfaces"
)

type healthcheckService struct {
	healthRepo interfaces.HealthcheckRepository
}

func NewHealthcheckService(healthRepo interfaces.HealthcheckRepository) interfaces.HealthcheckService {
	return &healthcheckService{
		healthRepo: healthRepo,
	}
}

// Check проверяет здоровье сервиса и базы данных
func (s *healthcheckService) Check(ctx context.Context) (map[string]interface{}, error) {
	err := s.healthRepo.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": "ok",
	}, nil
}
