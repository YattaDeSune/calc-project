package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"go.uber.org/zap"
)

type GetTaskResponse struct {
	ID        string `json:"id"`
	Arg1      string `json:"arg1"`
	Arg2      string `json:"arg2"`
	Operation string `json:"operation"`
}

// Получение задачи
func (a *Agent) GetTask(ctx context.Context) (*GetTaskResponse, error) {
	logger := logger.FromContext(ctx)

	resp, err := http.Get("http://localhost:" + a.cfg.RequestPort + "/api/v1/task")
	if err != nil {
		logger.Error("Failed to get request from orchestrator", zap.String("server port", a.cfg.RequestPort))
		return nil, ErrFailedToConnect
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Info("No free tasks", zap.Int("status code", resp.StatusCode))
		return nil, nil
	}

	body, _ := io.ReadAll(resp.Body)

	var task GetTaskResponse
	err = json.Unmarshal(body, &task)
	if err != nil {
		logger.Error("Got incorrect task")
		return nil, nil
	}
	logger.Info("Task recieved", zap.String("task id", task.ID))

	return &task, nil
}

type SendResultResponce struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
	Error  string  `json:"error"`
}

// Отправление задачи
func (a *Agent) SendResult(ctx context.Context, readyTask *SendResultResponce) error {
	logger := logger.FromContext(ctx)

	jsonTask, _ := json.Marshal(readyTask)
	req, err := http.NewRequest("POST", "http://localhost:"+a.cfg.RequestPort+"/api/v1/task", bytes.NewBuffer(jsonTask))
	if err != nil {
		logger.Error("Failed to send request to orchestrator", zap.String("server port", a.cfg.RequestPort))
		return ErrFailedToConnect
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send request to orchestrator", zap.String("server port", a.cfg.RequestPort))
		return ErrFailedToConnect
	}
	logger.Info("Task sended to orchestrator", zap.String("task id", readyTask.ID))
	defer resp.Body.Close()

	return nil
}
