package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/YattaDeSune/calc-project/internal/db"
	"github.com/YattaDeSune/calc-project/internal/entities"
	"github.com/YattaDeSune/calc-project/internal/logger"
	pb "github.com/YattaDeSune/calc-project/internal/proto"
	"github.com/YattaDeSune/calc-project/pkg/calculation"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Содержит в себе выражения для вычислений
type Storage struct {
	mu   *sync.Mutex
	data map[int]*entities.Expression
	ctx  context.Context
}

func NewStorage(ctx context.Context) *Storage {
	return &Storage{
		mu:   &sync.Mutex{},
		data: make(map[int]*entities.Expression),
		ctx:  ctx,
	}
}

// EXPRESSIONS
func (s *Storage) AddExpression(db *db.Database, id int, expr string) {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Вычисление ОПН для выражения
	RPN, err := calculation.ToRPN(calculation.Tokenize(expr))
	// Если при создании ОПН найдена ошибка - не проводим вычисления и ставим результатом ошибку
	if err != nil {
		// меняем результат в бд
		if errdb := db.UpdateExpressionResult(ctx, id, err.Error(), entities.CompletedWithError); errdb != nil {
			logger.Error("Failed to update expression result", zap.Error(errdb), zap.Int("id", id))
			return
		}
		logger.Info("End with RPN error", zap.Error(err))
		return
	}

	// Создание стека для хранения состояния вычислений
	stack := make([]string, 0)
	// Вычисление первой таски, и сохранение состояния (новая ОПН и новый стек ДО вычисление самой таски)
	arg1, arg2, operation, newRPN, newStack, err := calculation.NextTask(RPN, stack)
	if err != nil {
		// меняем результат в бд
		if errdb := db.UpdateExpressionResult(ctx, id, err.Error(), entities.CompletedWithError); errdb != nil {
			logger.Error("Failed to update expression result", zap.Error(errdb), zap.Int("id", id))
			return
		}
		logger.Info("End with RPN error", zap.Error(err))
		return
	}

	// Добавление в хранилище выражения с первой таской
	tasks := make([]*entities.Task, 0)
	taskID := fmt.Sprintf("%d_%s", id, uuid.New().String())
	task := &entities.Expression{
		ID:         id,
		Expression: expr,
		Status:     entities.Accepted, // Выражение принято
		RPN:        newRPN,
		Stack:      newStack,
		Tasks: append(tasks, &entities.Task{
			ID:        taskID,
			Arg1:      arg1,
			Arg2:      arg2,
			Operation: operation,
			Status:    entities.Accepted, // Таска принята
		}),
	}
	s.data[id] = task
	logger.Info("Add first task", zap.Any("task", task))
}

// TASKS

// Меняем результат таски и запускаем следующую таску, либо добавляем результат выражения
func (s *Storage) SubmitTaskResult(db *db.Database, result *pb.SubmitResultRequest) {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	exprIDstr := strings.Split(result.Id, "_")[0]
	exprID, _ := strconv.Atoi(exprIDstr)
	expression, ok := s.data[exprID]
	if !ok {
		logger.Error("Invalid expression id")
		return
	}

	// Если таска не "в прогрессе", значит либо она уже посчиталась, либо вернулась и посчитается позже
	if expression.Tasks[len(expression.Tasks)-1].Status != entities.InProgress {
		logger.Info("Task is not in progress, ignoring result", zap.String("id", result.Id))
		return
	}

	// Если таска пришла с ошибкой, добавляем результат выражения
	if result.Error != "" {
		// меняем результат в бд
		if errdb := db.UpdateExpressionResult(ctx, exprID, result.Error, entities.CompletedWithError); errdb != nil {
			logger.Error("Failed to update expression result", zap.Error(errdb), zap.Int("id", exprID))
			return
		}
		// сносим выражение локально
		delete(s.data, exprID)

		logger.Info("Task error, expression completed with error", zap.Int("expression id", expression.ID))
		return
	}

	// Если стек и ОПН пусты, добавляем результат выражения
	if len(expression.Stack) == 0 && len(expression.RPN) == 0 {
		// меняем результат в бд
		if errdb := db.UpdateExpressionResult(ctx, exprID, result.Result, entities.Completed); errdb != nil {
			logger.Error("Failed to update expression result", zap.Error(errdb), zap.Int("id", exprID))
			return
		}
		// сносим выражение локально
		delete(s.data, exprID)

		logger.Info("Tasks completed, expression completed", zap.Int("expression id", expression.ID))
		return
	}

	// Если стек не пуст, продолжаем вычислять выражение
	s.data[exprID].Stack = append(expression.Stack, fmt.Sprint(result.Result))

	arg1, arg2, operation, newRPN, newStack, err := calculation.NextTask(expression.RPN, expression.Stack)
	if err != nil {
		// меняем результат в бд
		if errdb := db.UpdateExpressionResult(ctx, exprID, err.Error(), entities.CompletedWithError); errdb != nil {
			logger.Error("Failed to update expression result", zap.Error(errdb), zap.Int("id", exprID))
			return
		}
		// сносим выражение локально
		delete(s.data, exprID)

		logger.Info("End with RPN error", zap.Error(err))
		return
	}

	// Добавляем новую таску
	taskID := fmt.Sprintf("%d_%s", expression.ID, uuid.New().String())
	expression.Tasks = append(expression.Tasks, &entities.Task{
		ID:        taskID,
		Arg1:      arg1,
		Arg2:      arg2,
		Operation: operation,
		Status:    entities.Accepted, // Таска принята
	})
	expression.RPN = newRPN
	expression.Stack = newStack
	logger.Info("Add task", zap.Any("task", expression.Tasks[len(expression.Tasks)-1]))
}

// Ищем таску для агента
func (s *Storage) GetTaskForAgent(db *db.Database) *entities.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := logger.FromContext(s.ctx)

	for _, expr := range s.data {
		for _, task := range expr.Tasks {
			if task.Status == entities.Accepted {
				task.Status = entities.InProgress // таска принята в работу
				task.LastUpdated = time.Now()
				expr.Status = entities.InProgress // выражение принято в работу
				// меняем статус в бд
				if errdb := db.UpdateExpressionStatus(s.ctx, expr.ID, entities.InProgress); errdb != nil {
					logger.Error("Failed to update expression result", zap.Error(errdb), zap.Int("id", expr.ID))
					return nil
				}

				return task
			}
		}
	}

	return nil
}

// Проверка тасок на время исполнения
func (s *Storage) CheckAndRecoverTasks(ctx context.Context) {
	logger := logger.FromContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, expr := range s.data {
		for _, task := range expr.Tasks {
			if task.Status == entities.InProgress && time.Since(task.LastUpdated) > 2*time.Minute {
				// Возвращаем задачу в статус accepted через 2 минуты
				task.Status = entities.Accepted
				task.LastUpdated = time.Now()
				logger.Info("Task recovered to 'accepted' status", zap.String("task id", task.ID))
			}
		}
	}
}
