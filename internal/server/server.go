package server

import (
	"log"
	"net/http"

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

func (s *Server) RunServer() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/calculate", s.AddExpression)
	mux.HandleFunc("/api/v1/expressions", s.GetExpressions)
	mux.HandleFunc("/api/v1/expressions/", s.GetExpressionByID)
	mux.HandleFunc("/internal/task", s.GetTask)
	mux.HandleFunc("/internal/task", s.SubmitResult)

	if err := http.ListenAndServe(":"+s.cfg.Addr, mux); err != nil {
		return err
	}

	return nil
}
