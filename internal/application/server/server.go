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

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Printf("Error loading config, loaded default values: %v", err)
		return &Config{
			Addr: "8082",
		}
	}

	return &cfg
}

type Server struct {
	cfg *Config
}

func New() *Server {
	return &Server{cfg: GetCfgFromEnv()}
}

// func (s)

func (s *Server) RunServer() error {
	mux := http.NewServeMux()

	http.ListenAndServe(s.cfg.Addr, mux)

	return nil
}
