package interfaces

import (
	"context"
	"net/http"
)

// ==================== Примеры моков для тестирования ====================

// MockHealthcheckRepository - мок репозитория для тестирования
type MockHealthcheckRepository struct {
	PingFunc func(ctx context.Context) error
}

func (m *MockHealthcheckRepository) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

// MockHealthcheckService - мок сервиса для тестирования
type MockHealthcheckService struct {
	CheckFunc func(ctx context.Context) (map[string]interface{}, error)
}

func (m *MockHealthcheckService) Check(ctx context.Context) (map[string]interface{}, error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx)
	}
	return map[string]interface{}{"status": "ok"}, nil
}

// MockHealthcheckController - мок контроллера для тестирования
type MockHealthcheckController struct {
	HandlePingFunc func(w http.ResponseWriter, r *http.Request)
}

func (m *MockHealthcheckController) HandlePing(w http.ResponseWriter, r *http.Request) {
	if m.HandlePingFunc != nil {
		m.HandlePingFunc(w, r)
	}
}

// ==================== Примеры использования в тестах ====================

/*
Пример использования в unit тестах:

func TestHealthcheckService(t *testing.T) {
	// Мок репозитория с успешным Ping
	mockRepo := &interfaces.MockHealthcheckRepository{
		PingFunc: func(ctx context.Context) error {
			return nil
		},
	}

	service := NewHealthcheckService(mockRepo)
	status, err := service.Check(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", status["status"])
	}
}

// Мок репозитория с ошибкой
func TestHealthcheckServiceWithError(t *testing.T) {
	mockRepo := &interfaces.MockHealthcheckRepository{
		PingFunc: func(ctx context.Context) error {
			return errors.New("connection failed")
		},
	}

	service := NewHealthcheckService(mockRepo)
	_, err := service.Check(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
*/