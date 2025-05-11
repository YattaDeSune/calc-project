package server

import (
	"context"
	"net/http"
	"time"

	"github.com/YattaDeSune/calc-project/internal/auth"
	"github.com/YattaDeSune/calc-project/internal/db"
	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/YattaDeSune/calc-project/internal/middleware"
	"github.com/gorilla/mux"
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

	Addr := "8081"
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
	db      *db.Database
	ctx     context.Context
	jwt     *auth.JWTManager
}

func New(ctx context.Context) *Server {
	logger := logger.FromContext(ctx)

	db, err := db.New(ctx)
	if err != nil {
		logger.Fatal("Failed to create db", zap.Error(err))
	}

	return &Server{
		cfg:     GetCfgFromEnv(ctx),
		storage: NewStorage(ctx),
		ctx:     ctx,
		db:      db,

		// безопасность придумают завтра)
		jwt: auth.NewJWTManager("smeshariki2005", 24*time.Hour),
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

func (s *Server) RunServer() error {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	// Фоновая проверка раз в минуту
	go s.StartRecover()

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/register", s.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", s.Login).Methods("POST")

	r.HandleFunc("/api/v1/calculate", s.AddExpression).Methods("POST")
	r.HandleFunc("/api/v1/expressions", s.GetExpressions).Methods("GET")
	r.HandleFunc("/api/v1/expressions/{id}", s.GetExpressionByID).Methods("GET")

	r.HandleFunc("/api/v1/task", s.GetTask).Methods("GET")
	r.HandleFunc("/api/v1/task", s.SubmitResult).Methods("POST")

	mux := middleware.AccessLog(ctx, r)
	mux = middleware.AuthMiddleware(ctx, *s.jwt, mux)
	mux = middleware.PanicRecover(ctx, mux)
	mux = middleware.EnableCORS(mux)

	go func() error {
		if err := http.ListenAndServe(":"+s.cfg.Addr, mux); err != nil {
			logger.Error("Failed to launch server", zap.String("port", s.cfg.Addr))
			return err
		}

		return nil
	}()

	logger.Info("Server launched", zap.String("port", s.cfg.Addr))
	return nil
}
