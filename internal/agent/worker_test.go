package agent

import (
	"context"
	"testing"
	"time"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWorker_ProcessTask(t *testing.T) {
	cfg := &Config{
		TimeAdditionMs:       1,
		TimeSubtractionMs:    1,
		TimeMultiplicationMs: 1,
		TimeDivisionMs:       1,
	}

	agent := &Agent{
		cfg:           cfg,
		taskChan:      make(chan *GetTaskResponse, 1),
		readyTaskChan: make(chan *SendResultResponce, 1),
	}

	ctx := logger.WithLogger(context.Background(), zap.NewNop())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go agent.worker(ctx, cancel, 1)

	task := &GetTaskResponse{
		ID:        "123",
		Arg1:      "10",
		Arg2:      "5",
		Operation: "+",
	}

	agent.taskChan <- task

	select {
	case result := <-agent.readyTaskChan:
		assert.Equal(t, "123", result.ID)
		assert.Equal(t, 15.0, result.Result)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

func TestWorker_InvalidTask(t *testing.T) {
	cfg := &Config{
		TimeAdditionMs:       1,
		TimeSubtractionMs:    1,
		TimeMultiplicationMs: 1,
		TimeDivisionMs:       1,
	}

	agent := &Agent{
		cfg:           cfg,
		taskChan:      make(chan *GetTaskResponse, 1),
		readyTaskChan: make(chan *SendResultResponce, 1),
	}

	ctx := logger.WithLogger(context.Background(), zap.NewNop())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go agent.worker(ctx, cancel, 1)

	task := &GetTaskResponse{
		ID:        "123",
		Arg1:      "10",
		Arg2:      "5",
		Operation: "invalid",
	}

	agent.taskChan <- task

	select {
	case result := <-agent.readyTaskChan:
		assert.Equal(t, "123", result.ID)
		assert.Equal(t, ErrInvalidOperation.Error(), result.Error)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

func TestWorker_DivisionByZero(t *testing.T) {
	cfg := &Config{
		TimeAdditionMs:       1,
		TimeSubtractionMs:    1,
		TimeMultiplicationMs: 1,
		TimeDivisionMs:       1,
	}

	agent := &Agent{
		cfg:           cfg,
		taskChan:      make(chan *GetTaskResponse, 1),
		readyTaskChan: make(chan *SendResultResponce, 1),
	}

	ctx := logger.WithLogger(context.Background(), zap.NewNop())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go agent.worker(ctx, cancel, 1)

	task := &GetTaskResponse{
		ID:        "123",
		Arg1:      "10",
		Arg2:      "0",
		Operation: "/",
	}

	agent.taskChan <- task

	select {
	case result := <-agent.readyTaskChan:
		assert.Equal(t, "123", result.ID)
		assert.Equal(t, ErrDevisionByZero.Error(), result.Error)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}
