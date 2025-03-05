package agent

import "errors"

var (
	ErrFailedToConnect = errors.New("failed to connect orchestrator")
	//
	ErrDevisionByZero   = errors.New("devision by zero")
	ErrInvalidOperator  = errors.New("operator is not a number")
	ErrInvalidOperation = errors.New("invalid operation")
)
