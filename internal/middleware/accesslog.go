package middleware

import (
	"context"
	"net/http"

	"time"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"go.uber.org/zap"
)

func AccessLog(ctx context.Context, next http.Handler) http.Handler {
	logger := logger.FromContext(ctx)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("New request",
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("url", r.URL.Path),
			zap.Duration("time", time.Since(start)),
		)
	})
}
