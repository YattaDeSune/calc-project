package middleware

import (
	"context"
	"net/http"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"go.uber.org/zap"
)

func PanicRecover(ctx context.Context, next http.Handler) http.Handler {
	logger := logger.FromContext(ctx)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered", zap.Any("Error", err), zap.String("URL", r.URL.Path))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
