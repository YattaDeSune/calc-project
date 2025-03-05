package server

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/YattaDeSune/calc-project/internal/entities"
	"github.com/YattaDeSune/calc-project/internal/logger"
	"github.com/YattaDeSune/calc-project/pkg/calculation"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Содержит в себе выражения, для каждого выражения - слайс "тасок"
type Storage struct {
	mu   *sync.Mutex
	data []*entities.Expression
	ctx  context.Context
}

func NewStorage(ctx context.Context) *Storage {
	return &Storage{
		mu:   &sync.Mutex{},
		data: make([]*entities.Expression, 0),
		ctx:  ctx,
	}
}

// для проверки id (соответствуют положению в слайсе)
func (s *Storage) GetLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

// EXPRESSIONS
func (s *Storage) GetExpressions() []*entities.Expression {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (s *Storage) GetExpressionByID(id int) *entities.Expression {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[id]
}

func (s *Storage) AddExpression(expr string) {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()
	id := len(s.data) + 1

	// Вычисление ОПН для выражения
	RPN, err := calculation.ToRPN(calculation.Tokenize(expr))
	// Если при создании ОПН найдена ошибка - не проводим вычисления и ставим результатом ошибку
	if err != nil {
		s.data = append(s.data, &entities.Expression{
			ID:         id,
			Expression: expr,
			Status:     entities.CompletedWithError, // завершено с ошибкой
			Result:     err.Error(),
		})
		logger.Info("End with RPN error", zap.Error(err))
		return
	}

	// Создание стека для хранения состояния вычислений
	stack := make([]string, 0)
	// Вычисление первой таски, и сохранение состояния (новая ОПН и новый стек ДО вычисление самой таски)
	arg1, arg2, operation, newRPN, newStack, err := calculation.NextTask(RPN, stack)
	if err != nil {
		s.data = append(s.data, &entities.Expression{
			ID:         id,
			Expression: expr,
			Status:     entities.CompletedWithError, // завершено с ошибкой
			Result:     err.Error(),
		})
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
	s.data = append(s.data, task)
	logger.Info("Add first task", zap.Any("task", task))
}

// TASKS

// Меняем результат таски и запускаем следующую таску, либо добавляем результат выражения
func (s *Storage) SubmitTaskResult(result *SubmitResultRequest) {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	exprIDstr := strings.Split(result.ID, "_")[0]
	exprID, _ := strconv.ParseInt(exprIDstr, 10, 64)
	expression := s.data[int(exprID)-1]

	// Если таска не "в прогрессе", значит либо она уже посчиталась, либо вернулась и посчитается позже
	if expression.Tasks[len(expression.Tasks)-1].Status != entities.InProgress {
		logger.Info("Task is not in progress, ignoring result", zap.String("id", result.ID))
		return
	}

	// Если таска пришла с ошибкой, добавляем результат выражения
	if result.Error != "" {
		expression.Result = result.Error
		expression.Tasks = nil                          // удаляем все таски
		expression.Status = entities.CompletedWithError // завершено с ошибкой
		logger.Info("Task error, expression completed with error", zap.Int("expression id", expression.ID))
		return
	}

	// Если стек и ОПН пусты, добавляем результат выражения
	if len(expression.Stack) == 0 && len(expression.RPN) == 0 {
		expression.Result = result.Result
		expression.Tasks = nil                 // удаляем все таски
		expression.Status = entities.Completed // завершено
		logger.Info("Tasks completed, expression completed", zap.Int("expression id", expression.ID))
		return
	}

	// Если стек не пуст, продолжаем вычислять выражение
	s.data[int(exprID)-1].Stack = append(s.data[int(exprID)-1].Stack, fmt.Sprint(result.Result))

	log.Println(expression.RPN)
	arg1, arg2, operation, newRPN, newStack, err := calculation.NextTask(expression.RPN, expression.Stack)
	if err != nil {
		expression.Result = err.Error()
		expression.Tasks = nil                          // удаляем все таски
		expression.Status = entities.CompletedWithError // завершено с ошибкой
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
func (s *Storage) GetTaskForAgent() *entities.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, expr := range s.data {
		for _, task := range expr.Tasks {
			if task.Status == entities.Accepted {
				task.Status = entities.InProgress // таска принята в работу
				task.LastUpdated = time.Now()
				expr.Status = entities.InProgress // выражение принято в работу
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
