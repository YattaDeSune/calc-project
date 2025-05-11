package middleware

import (
	"context"
	"net/http"

	"strings"

	"github.com/YattaDeSune/calc-project/internal/auth"
	"github.com/YattaDeSune/calc-project/internal/entities"
	"github.com/YattaDeSune/calc-project/internal/logger"
	"go.uber.org/zap"
)

func AuthMiddleware(ctx context.Context, jwtManager auth.JWTManager, next http.Handler) http.Handler {
	logger := logger.FromContext(ctx)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/register" || r.URL.Path == "/api/v1/login" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "empty Authorization header", http.StatusUnauthorized)
			logger.Error("empty Authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid Authorization format", http.StatusUnauthorized)
			logger.Error("invalid Authorization format")
			return
		}

		token := parts[1]

		claims, err := jwtManager.Verify(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			logger.Error("invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), entities.UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, entities.UserLoginKey, claims.Login)
		logger.Info("User authorized", zap.Int("user_id", claims.UserID), zap.String("user_login", claims.Login))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
