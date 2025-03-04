package calculation

import "errors"

var (
	ErrEmptyExpression      = errors.New("empty expression")
	ErrShortExpression      = errors.New("too short expression")
	ErrNoOpeningParenthesis = errors.New("no opening parenthesis")
	ErrNoClosingParenthesis = errors.New("no closing parenthesis")
	ErrInvalidExpression    = errors.New("expression is not valid")
)
