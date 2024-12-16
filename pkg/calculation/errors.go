package calculation

import "errors"

var (
	ErrEmptyExpression      = errors.New("empty expression")
	ErrShortExpression      = errors.New("too short expression")
	ErrNoOpeningParenthesis = errors.New("no opening parenthesis")
	ErrNoClosingParenthesis = errors.New("no closing parenthesis")
	ErrDevisionByZero       = errors.New("devision by zero")
	ErrInvalidExpression    = errors.New("expression is not valid")
	ErrInternalServer       = errors.New("internal server error")
)
