package main

import (
	"context"
	"fmt"
	"github.com/SmirnovND/gobase/internal/container"
	"go.uber.org/zap"
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(nil, "cron failed: %v", err)
		panic(err)
	}
}

func Run() error {
	diContainer := container.NewContainer()

	var logger *zap.Logger
	if err := diContainer.Invoke(func(l *zap.Logger) {
		logger = l
	}); err != nil {
		return err
	}

	logger.Info("Starting example cron job")

	ctx := context.Background()

	// Здесь ваша логика крон скрипта
	if err := ExampleJob(ctx, logger); err != nil {
		logger.Error("Cron job failed", zap.Error(err))
		return err
	}

	logger.Info("Cron job completed successfully")
	return nil
}

// ExampleJob — пример функции для крон задачи
func ExampleJob(ctx context.Context, log *zap.Logger) error {
	log.Info("Executing example job")
	// TODO: реализовать логику
	return nil
}
