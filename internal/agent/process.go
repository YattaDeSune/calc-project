package agent

import (
	"context"
	"strconv"
	"time"

	"github.com/YattaDeSune/calc-project/internal/logger"
	pb "github.com/YattaDeSune/calc-project/internal/proto"
	"go.uber.org/zap"
)

// Вычисление таски
func (a *Agent) processTask(ctx context.Context, task *pb.GetTaskResponse) *pb.SubmitResultRequest {
	logger := logger.FromContext(ctx)

	logger.Info("Processed task",
		zap.String("task id", task.Id),
		zap.String("arg1", task.Arg1),
		zap.String("arg2", task.Arg2),
		zap.String("operation", task.Operation),
	)

	// Нет смысла обрабатывать err потому что такого рода ошибки сюда не дойдут
	arg1, err := strconv.ParseFloat(task.Arg1, 64)
	if err != nil {
		return &pb.SubmitResultRequest{Id: task.Id, Error: ErrInvalidOperator.Error()}
	}
	arg2, err := strconv.ParseFloat(task.Arg2, 64)
	if err != nil && task.Operation != "~" {
		return &pb.SubmitResultRequest{Id: task.Id, Error: ErrInvalidOperator.Error()}
	}

	switch task.Operation {
	case "+":
		time.Sleep(time.Duration(a.cfg.TimeAdditionMs) * time.Millisecond)
		return &pb.SubmitResultRequest{Id: task.Id, Result: arg1 + arg2}
	case "-":
		time.Sleep(time.Duration(a.cfg.TimeSubtractionMs) * time.Millisecond)
		return &pb.SubmitResultRequest{Id: task.Id, Result: arg1 - arg2}
	case "*":
		time.Sleep(time.Duration(a.cfg.TimeMultiplicationMs) * time.Millisecond)
		return &pb.SubmitResultRequest{Id: task.Id, Result: arg1 * arg2}
	case "/":
		if arg2 == 0 {
			return &pb.SubmitResultRequest{Id: task.Id, Error: ErrDevisionByZero.Error()}
		}
		time.Sleep(time.Duration(a.cfg.TimeDivisionMs) * time.Millisecond)
		return &pb.SubmitResultRequest{Id: task.Id, Result: arg1 / arg2}
	case "~":
		time.Sleep(time.Duration(a.cfg.TimeSubtractionMs) * time.Millisecond)
		return &pb.SubmitResultRequest{Id: task.Id, Result: -arg1}
	default:
		logger.Error("Invalid operation", zap.String("task id", task.Id), zap.String("operation", task.Operation))
		return &pb.SubmitResultRequest{Id: task.Id, Error: ErrInvalidOperation.Error()}
	}
}
