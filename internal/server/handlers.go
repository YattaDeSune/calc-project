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
	//
	log.Printf("Add expression, id: %v", resp)
}

type GetExpressionsResponce struct {
	Expressions []Expression `json:"expressions"`
}

// /expressions GET
func (s *Server) GetExpressions(w http.ResponseWriter, r *http.Request) {
	resp := &GetExpressionsResponce{
		Expressions: s.storage.GetExpressions(),
	}

	w.WriteHeader(http.StatusAccepted) // 200
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (GetExpressions)", http.StatusInternalServerError) // 500
		return
	}
	//
	log.Println("Get expressions:", resp)
}

type GetExpressionResponce struct {
	Expression Expression `json:"expression"`
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

	resp := &GetExpressionResponce{
		Expression: s.storage.GetExpressionByID(id - 1),
	}

	w.WriteHeader(http.StatusAccepted) // 200
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (GetExpressionByID)", http.StatusInternalServerError) // 500
		return
	}

	log.Printf("Get expression by ID: %v", resp)
}

// /task GET
func (s *Server) GetTask(w http.ResponseWriter, r *http.Request) {

}

func (s *Server) SubmitResult(w http.ResponseWriter, r *http.Request) {}
