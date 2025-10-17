package services

import (
	"context"
	"github.com/SmirnovND/gobase/internal/repositories"
)

type HealthcheckService struct {
	healthRepo repositories.HealthcheckRepository
}

func NewHealthcheckService(healthRepo repositories.HealthcheckRepository) *HealthcheckService {
	return &HealthcheckService{
		healthRepo: healthRepo,
	}
}

// Check проверяет здоровье сервиса и базы данных
func (s *HealthcheckService) Check(ctx context.Context) (map[string]interface{}, error) {
	err := s.healthRepo.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": "ok",
	}, nil
}
