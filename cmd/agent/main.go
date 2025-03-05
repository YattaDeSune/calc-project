package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YattaDeSune/calc-project/internal/agent"
	"github.com/YattaDeSune/calc-project/internal/logger"
	"go.uber.org/zap"
)

func main() {
	zapLogger := logger.NewLogger()
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctxWithLogger := logger.WithLogger(ctxWithCancel, zapLogger)

	agent := agent.New(ctxWithLogger)
	// Запуск агента в отдельной горутине чтобы не блокировать дальнейший код
	go func() {
		if err := agent.RunAgent(ctxWithLogger, cancel); err != nil {
			zapLogger.Fatal("Failed to run agent", zap.Error(err))
		}
	}()
	zapLogger.Info("Agent launched")

	// Обработка системных сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		zapLogger.Info("Graceful shutdown")
		time.Sleep(time.Second * 3) // можно сделать функцию которая отправляет невыполненные задачи обратно оркестратору
		return
	case <-ctxWithLogger.Done():
		zapLogger.Info("Stopped by context")
		return
	}
}
