package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Логгер - zap
type Logger = *zap.Logger

// Кастомный формат времени
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	formated := t.Format("2006-01-02 15:04.05.000 MST")
	enc.AppendString(formated)
}

func NewLogger() Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = CustomTimeEncoder
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	logger, err := cfg.Build()
	if err != nil {
		logger.Error("Failed to build zap logger, usuing fallback logger:", zap.Error(err))
		// Запуск фейк-логгера
		return zap.NewNop()
	}
	defer logger.Sync()

	return logger
}

type key string

// Логгер в контексте
func WithLogger(ctx context.Context, logger Logger) context.Context {
	var logKey key = "logger"
	return context.WithValue(ctx, logKey, logger)
}

// Извлечение логгера из контекста
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value("logger").(Logger); ok {
		return logger
	}

	return NewLogger()
}
