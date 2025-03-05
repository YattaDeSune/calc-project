package agent

import (
	"context"
	"time"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type Config struct {
	RequestPort          string `env:"SERVER_PORT"`
	TimeAdditionMs       int    `env:"TIME_ADDITION_MS"`
	TimeSubtractionMs    int    `env:"TIME_SUBTRACTION_MS"`
	TimeMultiplicationMs int    `env:"TIME_MULTIPLICATIONS_MS"`
	TimeDivisionMs       int    `env:"TIME_DIVISIONS_MS"`
	ComputingPower       int    `env:"COMPUTING_POWER"`
}

func GetCfgFromEnv(ctx context.Context) *Config {
	logger := logger.FromContext(ctx)

	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		var (
			TimeAdditionMs       = 2000
			TimeSubtractionMs    = 2000
			TimeMultiplicationMs = 5000
			TimeDivisionMs       = 5000
			ComputingPower       = 4
		)

		logger.Error("Error loading config, loaded default values",
			zap.Error(err),
			zap.Int("TimeAdditionMs", TimeAdditionMs),
			zap.Int("TimeSubtractionMs", TimeSubtractionMs),
			zap.Int("TimeMultiplicationMs", TimeMultiplicationMs),
			zap.Int("TimeDivisionMs", TimeDivisionMs),
			zap.Int("ComputingPower", ComputingPower),
		)
		return &Config{
			TimeAdditionMs:       TimeAdditionMs,
			TimeSubtractionMs:    TimeSubtractionMs,
			TimeMultiplicationMs: TimeMultiplicationMs,
			TimeDivisionMs:       TimeDivisionMs,
			ComputingPower:       ComputingPower,
		}
	}

	logger.Info("Loaded config",
		zap.Int("TimeAdditionMs", cfg.TimeAdditionMs),
		zap.Int("TimeSubtractionMs", cfg.TimeSubtractionMs),
		zap.Int("TimeMultiplicationMs", cfg.TimeMultiplicationMs),
		zap.Int("TimeDivisionMs", cfg.TimeDivisionMs),
		zap.Int("ComputingPower", cfg.ComputingPower),
	)

	return &cfg
}

type Agent struct {
	cfg           *Config
	taskChan      chan *GetTaskResponse
	readyTaskChan chan *SendResultResponce
}

func New(ctx context.Context) *Agent {
	return &Agent{
		cfg:           GetCfgFromEnv(ctx),
		taskChan:      make(chan *GetTaskResponse, 100),    // для получения задач
		readyTaskChan: make(chan *SendResultResponce, 100), // для результатов
	}
}

func (a *Agent) RunAgent(ctx context.Context, cancel context.CancelFunc) error {
	// Запуск воркеров
	for i := 1; i <= a.cfg.ComputingPower; i++ {
		go a.worker(ctx, cancel, i)
	}

	// Бесконечный цикл для запроса задач
	func() {
		for {
			task, err := a.GetTask(ctx)
			// Если нет подключения к оркестратору - кладем агента
			if err == ErrFailedToConnect {
				cancel()
			}

			if task != nil {
				a.taskChan <- task
			}

			time.Sleep(1 * time.Second) // Задержка между запросами задач
		}
	}()

	return nil
}
