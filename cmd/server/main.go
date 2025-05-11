package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/YattaDeSune/calc-project/internal/server"
	"go.uber.org/zap"
)

func main() {
	zapLogger := logger.NewLogger()
	ctxWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctxWithLogger := logger.WithLogger(ctxWithCancel, zapLogger)

	server := server.New(ctxWithLogger)

	// Запуск оркестратора в отдельной горутине чтобы не блокировать дальнейший код
	go func() {
		if err := server.RunServer(); err != nil {
			zapLogger.Fatal("Failed to run server (HTTP)", zap.Error(err))
		}
	}()
	zapLogger.Info("Orchestrator launched (HTTP)")

	go func() {
		if err := server.RunGRPCServer(); err != nil {
			zapLogger.Fatal("Failed to run server (gRPC)", zap.Error(err))
		}
	}()
	zapLogger.Info("Orchestrator launched (gRPC)")

	// Обработка системных сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		zapLogger.Info("Graceful shutdown")
		return
	case <-ctxWithLogger.Done():
		zapLogger.Info("Stopped by context")
		return
	}
}
