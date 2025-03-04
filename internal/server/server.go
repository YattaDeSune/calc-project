package server

import (
	"log"
	"net/http"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Addr string `env:"SERVER_PORT"`
}

func GetCfgFromEnv() *Config {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Printf("Error loading config, loaded default values: %v", err)
		return &Config{
			Addr: "8081",
		}
	}

	if cfg.Addr == "" {
		log.Println("Empty address, using default config values")
		return &Config{
			Addr: "8081",
		}
	}

	log.Printf("Loaded config, port %v:", cfg)
	return &cfg
}

type Server struct {
	cfg     *Config
	storage *Storage
}

func New() *Server {
	return &Server{
		cfg:     GetCfgFromEnv(),
		storage: NewStorage(),
	}
}

// Фоновая проверка тасок на "живучесть", костыльная защита от падения агента
func (s *Server) StartRecover() {
	go func() {
		for {
			// Каждую минуту проверяем таски
			time.Sleep(time.Minute)
			s.storage.CheckAndRecoverTasks()
		}
	}()
}

func (s *Server) RunServer() error {
	s.StartRecover()

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

	if err := http.ListenAndServe(":"+s.cfg.Addr, mux); err != nil {
		return err
	}

	return nil
}
