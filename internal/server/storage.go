package server

import "sync"

type Expression struct {
	ID         int    `json:"id"`
	Expression string `json:"expression"`
	Status     string `json:"status"`
	Result     any    `json:"result"`
}

type Storage struct {
	mu   *sync.Mutex
	data []Expression
}

func NewStorage() *Storage {
	return &Storage{
		mu:   &sync.Mutex{},
		data: make([]Expression, 0),
	}
}

func (s *Storage) GetLen() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}

func (s *Storage) AddExpression(expr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, Expression{
		ID:         len(s.data) + 1,
		Expression: expr,
		Status:     "in progress",
		Result:     nil,
	})
}

func (s *Storage) GetExpressions() []Expression {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data
}

func (s *Storage) GetExpressionByID(id int) Expression {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[id]
}
