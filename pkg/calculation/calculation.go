package calculation

import (
	"strconv"
	"strings"
	"unicode"
)

var priority = map[string]int{
	"+": 1,
	"-": 1,
	"*": 2,
	"/": 2,
	"~": 3, // unary -
}

func ToRPN(tokens []string) ([]string, error) {
	var stack []string
	var out []string

	if len(tokens) == 0 {
		return nil, ErrEmptyExpression
	}

	if len(tokens) <= 2 {
		return nil, ErrShortExpression
	}

	for i, token := range tokens {
		switch {
		case isNum(token):
			out = append(out, token)
		case token == "(":
			stack = append(stack, token)
		case token == ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				out = append(out, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, ErrNoOpeningParenthesis
			}
			stack = stack[:len(stack)-1]
		case isOperation(token):
			if token == "-" && (i == 0 || tokens[i-1] == "(" || isOperation(tokens[i-1])) {
				token = "~"
			}

			for len(stack) > 0 && isOperation(stack[len(stack)-1]) && (priority[stack[len(stack)-1]] >= priority[token]) {
				out = append(out, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		default:
			return nil, ErrInvalidExpression
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrNoClosingParenthesis
		}
		out = append(out, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return out, nil
}

func Tokenize(expression string) []string {
	var tokens []string
	var buffer strings.Builder

	for _, char := range expression {
		if unicode.IsDigit(char) || char == '.' {
			buffer.WriteRune(char)
		} else {
			if buffer.Len() > 0 {
				tokens = append(tokens, buffer.String())
				buffer.Reset()
			}
			if !unicode.IsSpace(char) {
				tokens = append(tokens, string(char))
			}
		}
	}

	if buffer.Len() > 0 {
		tokens = append(tokens, buffer.String())
	}

	return tokens
}

func isOperation(token string) bool {
	_, exists := priority[token]
	return exists
}

func isNum(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

// Вычисляет новую таску для заданного ОПН и текущего стека
func NextTask(rpn []string, stack []string) (arg1, arg2 string, operation string, newRPN []string, newStack []string, err error) {
	if len(rpn) == 0 {
		err = ErrEmptyExpression
		return
	}

	element := rpn[0]
	newRPN = rpn[1:]

	switch {
	case isNum(element):
		newStack = append(stack, element)
		return NextTask(newRPN, newStack)

	case isOperation(element):
		if len(stack) < 2 {
			err = ErrShortExpression
			return
		}
		arg2, arg1 = stack[len(stack)-1], stack[len(stack)-2]
		newStack = stack[:len(stack)-2]
		operation = element
		return arg1, arg2, operation, newRPN, newStack, nil

	default:
		err = ErrInvalidExpression
		return
	}
}
