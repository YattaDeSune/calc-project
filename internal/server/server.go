package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/YattaDeSune/calc-project/internal/auth"
	"github.com/YattaDeSune/calc-project/internal/db"
	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/YattaDeSune/calc-project/internal/middleware"
	pb "github.com/YattaDeSune/calc-project/internal/proto"
	"github.com/gorilla/mux"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	HTTPPort string `env:"HTTP_SERVER_PORT"`
	GRPCPort string `env:"GRPC_SERVER_PORT"`
}

func GetCfgFromEnv(ctx context.Context) *Config {
	logger := logger.FromContext(ctx)

	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)

	httpPort := "8081"
	grpcPort := "9090"
	if err != nil {
		logger.Error("Error loading config, loaded default values",
			zap.Error(err),
			zap.String("httpPort", httpPort),
			zap.String("grpcPort", grpcPort),
		)
		return &Config{
			HTTPPort: httpPort,
			GRPCPort: grpcPort,
		}
	}

	if cfg.HTTPPort == "" || cfg.GRPCPort == "" {
		logger.Error("Empty ports, using default config values",
			zap.Error(err),
			zap.String("httpPort", httpPort),
			zap.String("grpcPort", grpcPort),
		)
		return &Config{
			HTTPPort: "8081",
			GRPCPort: "9090",
		}
	}

	logger.Info("Config loaded", zap.String("httpPort", cfg.HTTPPort), zap.String("grpcPort", cfg.GRPCPort))
	return &cfg
}

type Server struct {
	pb.TaskServiceServer
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

	mux := middleware.AccessLog(ctx, r)
	mux = middleware.AuthMiddleware(ctx, *s.jwt, mux)
	mux = middleware.PanicRecover(ctx, mux)
	mux = middleware.EnableCORS(mux)

	go func() error {
		if err := http.ListenAndServe(":"+s.cfg.HTTPPort, mux); err != nil {
			logger.Error("Failed to launch server", zap.String("http port", s.cfg.HTTPPort))
			return err
		}

		return nil
	}()

	logger.Info("HTTP server listening", zap.String("http port", s.cfg.HTTPPort))
	return nil
}

func (s *Server) RunGRPCServer() error {
	logger := logger.FromContext(s.ctx)

	lis, err := net.Listen("tcp", ":"+s.cfg.GRPCPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterTaskServiceServer(grpcServer, s)
	logger.Info("gRPC server listening", zap.String("grpc port", s.cfg.GRPCPort))

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
		return err
	}

	return nil
}
