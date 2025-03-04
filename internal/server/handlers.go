package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type AddExpressionRequest struct {
	Expression string `json:"expression"`
}

type AddExpressionResponce struct {
	ID int `json:"id"`
}

// /calculate POST
func (s *Server) AddExpression(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read expression", http.StatusBadRequest) // 400
		return
	}
	defer r.Body.Close()

	var req AddExpressionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusUnprocessableEntity) // 422
		return
	}

	if req.Expression == "" {
		http.Error(w, "Expression cannot be empty", http.StatusUnprocessableEntity) // 422
		return
	}

	s.storage.AddExpression(req.Expression)

	resp := &AddExpressionResponce{
		ID: s.storage.GetExpressionByID(len(s.storage.data) - 1).ID,
	}

	w.WriteHeader(http.StatusCreated) // 201
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (AddExpression)", http.StatusInternalServerError) // 500
		return
	}
	log.Printf("Add expression, id: %v", resp)
}

type localExpression struct {
	ID         int    `json:"id"`
	Expression string `json:"expression"`
	Status     string `json:"status"`
	Result     any    `json:"result"`
}

type GetExpressionsResponce struct {
	Expressions []localExpression `json:"expressions"`
}

// /expressions GET
func (s *Server) GetExpressions(w http.ResponseWriter, r *http.Request) {
	exprs := s.storage.GetExpressions()

	var resp GetExpressionsResponce
	for _, expr := range exprs {
		resp.Expressions = append(resp.Expressions, localExpression{
			ID:         expr.ID,
			Expression: expr.Expression,
			Status:     expr.Status,
			Result:     expr.Result,
		})
	}

	w.WriteHeader(http.StatusOK) // 200
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (GetExpressions)", http.StatusInternalServerError) // 500
		return
	}
	log.Println("Get expressions:", resp)
}

type GetExpressionResponce struct {
	Expression localExpression `json:"expression"`
}

// /expression/:id GET
func (s *Server) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest) // 400
		return
	}

	id, err := strconv.Atoi(parts[4])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound) // 404
		return
	}

	if id > s.storage.GetLen() {
		http.Error(w, "Invalid ID", http.StatusNotFound) // 404
		log.Printf("Invalid ID: %v", id)
		return
	}

	expr := s.storage.GetExpressionByID(id - 1)

	resp := &localExpression{
		ID:         expr.ID,
		Expression: expr.Expression,
		Status:     expr.Status,
		Result:     expr.Result,
	}

	w.WriteHeader(http.StatusOK) // 200
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (GetExpressionByID)", http.StatusInternalServerError) // 500
		return
	}
	log.Println("Get expression by ID:", resp)
}

type GetTaskResponce struct {
	ID        string `json:"id"`
	Arg1      string `json:"arg1"`
	Arg2      string `json:"arg2"`
	Operation string `json:"operation"`
}

// /task GET
func (s *Server) GetTask(w http.ResponseWriter, r *http.Request) {
	task := s.storage.GetTaskForAgent()

	if task == nil {
		http.Error(w, "No tasks for agent", http.StatusNotFound) // 404
		return
	}

	resp := &GetTaskResponce{
		ID:        task.ID,
		Arg1:      task.Arg1,
		Arg2:      task.Arg2,
		Operation: task.Operation,
	}

	w.WriteHeader(http.StatusOK) // 200
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (GetTask)", http.StatusInternalServerError) // 500
		return
	}
	log.Printf("Get task for agent, id: %v", task.ID)
}

type SubmitResultRequest struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
	Error  string  `json:"error"`
}

// /task POST
func (s *Server) SubmitResult(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read result", http.StatusBadRequest) // 400
		return
	}

	var result SubmitResultRequest
	err = json.Unmarshal(body, &result)
	if err != nil {
		http.Error(w, "Failed to unmarshal JSON", http.StatusUnprocessableEntity) // 422
		return
	}

	s.storage.SubmitTaskResult(&result)

	w.WriteHeader(http.StatusOK) // 200
	log.Printf("")
}
