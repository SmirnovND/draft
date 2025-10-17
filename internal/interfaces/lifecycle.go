package interfaces

import "context"

// Closer интерфейс для компонентов, которые нужно закрыть при shutdown
type Closer interface {
	Close() error
}

// Shutdowner интерфейс для компонентов с graceful shutdown
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// LoggerCloser интерфейс для логгеров (zap использует Sync вместо Close)
type LoggerCloser interface {
	Sync() error
}