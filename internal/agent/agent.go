package agent

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	time_addition_ms       int `env:"TIME_ADDITION_MS"`
	time_subtraction_ms    int `env:"TIME_SUBTRACTION_MS"`
	time_multiplication_ms int `env:"TIME_MULTIPLICATIONS_MS"`
	time_division_ms       int `env:"TIME_DIVISIONS_MS"`
	computing_power        int `env:"COMPUTING_POWER"`
}

func GetCfgFromEnv() *Config {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Printf("Error loading config, loaded default values: %v", err)
		return &Config{
			time_addition_ms:       2000,
			time_subtraction_ms:    2000,
			time_multiplication_ms: 5000,
			time_division_ms:       5000,
			computing_power:        4,
		}
	}

	return &cfg
}

type Agent struct {
	cfg *Config
}

func New() *Agent {
	return &Agent{cfg: GetCfgFromEnv()}
}

func (a *Agent) GetTask() {
	// получение таски
}

func (a *Agent) worker() {
	for {
		// получение тасок
	}
}

func (a *Agent) RunAgent() error {
	// запуск агента с несколькими горутинами
	for i := 1; i <= a.cfg.computing_power; i++ {
		go a.worker()
	}

	return nil
}
