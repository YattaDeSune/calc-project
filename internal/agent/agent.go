package agent

import (
	"context"

	"github.com/YattaDeSune/calc-project/internal/logger"
	pb "github.com/YattaDeSune/calc-project/internal/proto"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Config struct {
	HTTPPort string `env:"HTTP_SERVER_PORT"`
	GRPCPort string `env:"GRPC_SERVER_PORT"`

	TimeAdditionMs       int `env:"TIME_ADDITION_MS"`
	TimeSubtractionMs    int `env:"TIME_SUBTRACTION_MS"`
	TimeMultiplicationMs int `env:"TIME_MULTIPLICATIONS_MS"`
	TimeDivisionMs       int `env:"TIME_DIVISIONS_MS"`
	ComputingPower       int `env:"COMPUTING_POWER"`
}

func GetCfgFromEnv(ctx context.Context) *Config {
	logger := logger.FromContext(ctx)

	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		var (
			httpPort             = "8081"
			grpcPort             = "9090"
			TimeAdditionMs       = 2000
			TimeSubtractionMs    = 2000
			TimeMultiplicationMs = 5000
			TimeDivisionMs       = 5000
			ComputingPower       = 4
		)

		logger.Error("Error loading config, loaded default values",
			zap.Error(err),
			zap.String("httpPort", httpPort),
			zap.String("grpcPort", grpcPort),
			zap.Int("TimeAdditionMs", TimeAdditionMs),
			zap.Int("TimeSubtractionMs", TimeSubtractionMs),
			zap.Int("TimeMultiplicationMs", TimeMultiplicationMs),
			zap.Int("TimeDivisionMs", TimeDivisionMs),
			zap.Int("ComputingPower", ComputingPower),
		)
		return &Config{
			HTTPPort:             httpPort,
			GRPCPort:             grpcPort,
			TimeAdditionMs:       TimeAdditionMs,
			TimeSubtractionMs:    TimeSubtractionMs,
			TimeMultiplicationMs: TimeMultiplicationMs,
			TimeDivisionMs:       TimeDivisionMs,
			ComputingPower:       ComputingPower,
		}
	}

	logger.Info("Loaded config",
		zap.String("httpPort", cfg.HTTPPort),
		zap.String("grpcPort", cfg.GRPCPort),
		zap.Int("TimeAdditionMs", cfg.TimeAdditionMs),
		zap.Int("TimeSubtractionMs", cfg.TimeSubtractionMs),
		zap.Int("TimeMultiplicationMs", cfg.TimeMultiplicationMs),
		zap.Int("TimeDivisionMs", cfg.TimeDivisionMs),
		zap.Int("ComputingPower", cfg.ComputingPower),
	)

	return &cfg
}

type Agent struct {
	client        pb.TaskServiceClient
	cfg           *Config
	taskChan      chan *pb.GetTaskResponse
	readyTaskChan chan *pb.SubmitResultRequest
}

func New(ctx context.Context) *Agent {
	logger := logger.FromContext(ctx)

	agent := &Agent{
		cfg:           GetCfgFromEnv(ctx),
		taskChan:      make(chan *pb.GetTaskResponse, 100),     // для получения задач
		readyTaskChan: make(chan *pb.SubmitResultRequest, 100), // для результатов
	}

	conn, err := grpc.Dial("localhost:"+agent.cfg.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect gRPC server", zap.Error(err))
	}
	agent.client = pb.NewTaskServiceClient(conn)

	return agent
}

func (a *Agent) RunAgent(ctx context.Context, cancel context.CancelFunc) error {
	logger := logger.FromContext(ctx)

	// Запуск воркеров
	for i := 1; i <= a.cfg.ComputingPower; i++ {
		go a.worker(ctx, cancel, i)
	}

	// Бесконечный цикл для запроса задач
	func() {
		for {
			task, err := a.client.GetTask(ctx, &pb.GetTaskRequest{})
			// Если нет подключения к оркестратору - кладем агента
			if status.Code(err) == codes.Unavailable {
				logger.Warn("Failed to connect gRPC server", zap.Error(err))
				cancel()
			}

			if task != nil {
				a.taskChan <- task
			}
		}
	}()

	return nil
}
