package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAgent_GetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		task := GetTaskResponse{
			ID:        "123",
			Arg1:      "10",
			Arg2:      "5",
			Operation: "+",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)
	}))
	defer server.Close()

	cfg := &Config{
		RequestPort: server.URL[len("http://localhost:"):],
	}

	agent := &Agent{
		cfg: cfg,
	}

	ctx := logger.WithLogger(context.Background(), zap.NewNop())

	task, err := agent.GetTask(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "123", task.ID)
	assert.Equal(t, "10", task.Arg1)
	assert.Equal(t, "5", task.Arg2)
	assert.Equal(t, "+", task.Operation)
}

func TestAgent_SendResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &Config{
		RequestPort: server.URL[len("http://localhost:"):],
	}

	agent := &Agent{
		cfg: cfg,
	}

	ctx := logger.WithLogger(context.Background(), zap.NewNop())

	result := &SendResultResponce{
		ID:     "123",
		Result: 15.0,
	}
	err := agent.SendResult(ctx, result)
	assert.NoError(t, err)
}

func TestAgent_processTask(t *testing.T) {
	agent := &Agent{
		cfg: &Config{
			TimeAdditionMs:       100,
			TimeSubtractionMs:    100,
			TimeMultiplicationMs: 100,
			TimeDivisionMs:       100,
		},
	}

	ctx := logger.WithLogger(context.Background(), zap.NewNop())

	tests := []struct {
		name     string
		task     *GetTaskResponse
		expected *SendResultResponce
	}{
		{
			name: "addition",
			task: &GetTaskResponse{
				ID:        "1",
				Arg1:      "10",
				Arg2:      "5",
				Operation: "+",
			},
			expected: &SendResultResponce{
				ID:     "1",
				Result: 15.0,
			},
		},
		{
			name: "subtraction",
			task: &GetTaskResponse{
				ID:        "2",
				Arg1:      "10",
				Arg2:      "5",
				Operation: "-",
			},
			expected: &SendResultResponce{
				ID:     "2",
				Result: 5.0,
			},
		},
		{
			name: "multiplication",
			task: &GetTaskResponse{
				ID:        "3",
				Arg1:      "10",
				Arg2:      "5",
				Operation: "*",
			},
			expected: &SendResultResponce{
				ID:     "3",
				Result: 50.0,
			},
		},
		{
			name: "division",
			task: &GetTaskResponse{
				ID:        "4",
				Arg1:      "10",
				Arg2:      "5",
				Operation: "/",
			},
			expected: &SendResultResponce{
				ID:     "4",
				Result: 2.0,
			},
		},
		{
			name: "division by zero",
			task: &GetTaskResponse{
				ID:        "5",
				Arg1:      "10",
				Arg2:      "0",
				Operation: "/",
			},
			expected: &SendResultResponce{
				ID:    "5",
				Error: ErrDevisionByZero.Error(),
			},
		},
		{
			name: "invalid operation",
			task: &GetTaskResponse{
				ID:        "6",
				Arg1:      "10",
				Arg2:      "5",
				Operation: "lol",
			},
			expected: &SendResultResponce{
				ID:    "6",
				Error: ErrInvalidOperation.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := agent.processTask(ctx, tt.task)
			assert.Equal(t, tt.expected, result)
		})
	}
}
