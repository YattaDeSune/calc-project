package agent

import (
	"context"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"go.uber.org/zap"
)

// Воркер принимает задачу из канала и возвращает результат в другой канал
func (a *Agent) worker(ctx context.Context, cancel context.CancelFunc, num int) {
	logger := logger.FromContext(ctx)

	logger.Info("Worker started", zap.Int("worker number", num))
	for task := range a.taskChan {

		logger.Info("Worker starts to process task",
			zap.Int("worker number", num),
			zap.String("task id", task.ID),
		)
		a.readyTaskChan <- a.processTask(ctx, task)

		// Ожидание результата

		readyTask := <-a.readyTaskChan
		logger.Info("Worker finished to process task",
			zap.Int("worker number", num),
			zap.String("task id", readyTask.ID),
			zap.Float64("task result", readyTask.Result),
		)

		// Если нет подключения к оркестратору - кладем агента
		if err := a.SendResult(ctx, readyTask); err == ErrFailedToConnect {
			cancel()
		}
	}
}
