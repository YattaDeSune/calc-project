package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type Config struct {
	Addr string `env:"SERVER_PORT"`
}

func GetCfgFromEnv(ctx context.Context) *Config {
	logger := logger.FromContext(ctx)

	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)

	var Addr string = "8081"
	if err != nil {

		logger.Error("Error loading config, loaded default values",
			zap.Error(err),
			zap.String("Addr", Addr),
		)
		return &Config{
			Addr: Addr,
		}
	}

	if cfg.Addr == "" {
		logger.Error("Empty address, using default config values",
			zap.Error(err),
			zap.String("Addr", Addr),
		)
		return &Config{
			Addr: "8081",
		}
	}

	logger.Info("Config loaded", zap.String("Addr", cfg.Addr))
	return &cfg
}

type Server struct {
	cfg     *Config
	storage *Storage
	ctx     context.Context
}

func New(ctx context.Context) *Server {
	return &Server{
		cfg:     GetCfgFromEnv(ctx),
		storage: NewStorage(ctx),
		ctx:     ctx,
	}
}

// Проверка тасок на "живучесть", костыльная защита от падения агента
func (s *Server) StartRecover() {
	ctx := s.ctx
	func() {
		for {
			// Каждую минуту проверяем таски
			time.Sleep(time.Minute)
			s.storage.CheckAndRecoverTasks(ctx)
		}
	}()
}

// Middleware для обработки CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) RunServer() error {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	// Фоновая проверка раз в минуту
	go s.StartRecover()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", s.AddExpression)
	mux.HandleFunc("/api/v1/expressions", s.GetExpressions)
	mux.HandleFunc("/api/v1/expressions/", s.GetExpressionByID)

	mux.HandleFunc("/api/v1/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.GetTask(w, r)
		case http.MethodPost:
			s.SubmitResult(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	go func() error {
		if err := http.ListenAndServe(":"+s.cfg.Addr, enableCORS(mux)); err != nil {
			logger.Error("Failed to launch server", zap.String("port", s.cfg.Addr))
			return err
		}

		return nil
	}()

	logger.Info("Server launched", zap.String("port", s.cfg.Addr))
	return nil
}
