package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/YattaDeSune/calc-project/internal/auth"
	"github.com/YattaDeSune/calc-project/internal/entities"
	"github.com/YattaDeSune/calc-project/internal/errors"
	"github.com/YattaDeSune/calc-project/internal/logger"
	pb "github.com/YattaDeSune/calc-project/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterResponce struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponce struct {
	Token string `json:"token"`
}

// /register POST
func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	logger := logger.FromContext(s.ctx)

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	userID, err := s.db.CreateUser(s.ctx, req.Login, hash)
	if err != nil {
		if err == errors.ErrUserExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Register error", http.StatusInternalServerError)
		return
	}
	logger.Info("User created", zap.Int("ID", userID))

	token, err := s.jwt.Generate(userID, req.Login)
	if err != nil {
		http.Error(w, "JWT generate error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RegisterResponce{Token: token})
}

// /login POST
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	logger := logger.FromContext(s.ctx)

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUserByLogin(s.ctx, req.Login)
	if err != nil {
		if err == errors.ErrWrongLogin {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("User logged in", zap.Int("ID", user.ID), zap.String("Login", user.Login))

	if !auth.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, errors.ErrWrongPassword.Error(), http.StatusUnauthorized)
		return
	}

	token, err := s.jwt.Generate(user.ID, req.Login)
	if err != nil {
		http.Error(w, "JWT generate error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponce{Token: token})
}

type AddExpressionRequest struct {
	Expression string `json:"expression"`
}

type AddExpressionResponce struct {
	ID int `json:"id"`
}

// /calculate POST
func (s *Server) AddExpression(w http.ResponseWriter, r *http.Request) {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

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

	// достаем юзера из контекста запроса
	userID, ok := r.Context().Value(entities.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	exprID, err := s.db.CreateExpression(ctx, req.Expression, userID, entities.Accepted)
	if err != nil {
		http.Error(w, "Failed to create expression", http.StatusInternalServerError)
		return
	}
	logger.Info("Add expression", zap.Int("id", exprID), zap.String("expression", req.Expression))

	s.storage.AddExpression(s.db, exprID, req.Expression)

	resp := &AddExpressionResponce{
		ID: exprID,
	}

	w.WriteHeader(http.StatusCreated) // 201
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (AddExpression)", http.StatusInternalServerError) // 500
		return
	}
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
	ctx := s.ctx
	logger := logger.FromContext(ctx)

	// достаем юзера из контекста запроса
	userID, ok := r.Context().Value(entities.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	exprs, err := s.db.GetExpressionsByUser(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	logger.Info("Get expressions", zap.Any("expressions", resp))
}

type GetExpressionResponce struct {
	Expression localExpression `json:"expression"`
}

// /expression/:id GET
func (s *Server) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	ctx := s.ctx
	logger := logger.FromContext(ctx)

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

	// достаем юзера из контекста запроса
	userID, ok := r.Context().Value(entities.UserIDKey).(int)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	expr, err := s.db.GetExpressionByID(ctx, id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if expr == nil {
		http.Error(w, "Expression not found", http.StatusNotFound) // 404
		return
	}

	localExpr := localExpression{
		ID:         expr.ID,
		Expression: expr.Expression,
		Status:     expr.Status,
		Result:     expr.Result,
	}
	resp := GetExpressionResponce{Expression: localExpr}

	w.WriteHeader(http.StatusOK) // 200
	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response (GetExpressionByID)", http.StatusInternalServerError) // 500
		return
	}

	logger.Info("Get expression by id", zap.Any("expression", resp))
}

// gRPC
func (s *Server) GetTask(ctx context.Context, in *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	logCtx := s.ctx
	logger := logger.FromContext(logCtx)

	task := s.storage.GetTaskForAgent(s.db)
	if task == nil {
		return nil, status.Error(codes.NotFound, "no tasks available")
	}

	logger.Info("Get task for agent", zap.Any("id", task.ID))
	return &pb.GetTaskResponse{
		Id:        task.ID,
		Arg1:      task.Arg1,
		Arg2:      task.Arg2,
		Operation: task.Operation,
	}, nil
}

// gRPC
func (s *Server) SubmitResult(ctx context.Context, in *pb.SubmitResultRequest) (*pb.SubmitResultResponse, error) {
	logCtx := s.ctx
	logger := logger.FromContext(logCtx)

	s.storage.SubmitTaskResult(s.db, in)
	logger.Info("Recieved result from agent", zap.Any("id", in.Id), zap.Any("result", in.Result))

	return &pb.SubmitResultResponse{}, nil
}
