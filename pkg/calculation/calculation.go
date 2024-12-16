package calculation

import (
	"strconv"
	"strings"
)

// Приоретизация операций
var precedence = map[string]int{
	"+": 1,
	"-": 1,
	"*": 2,
	"/": 2,
}

// Основная вычисляющая функция
func Calc(expression string) (float64, error) {
	tokens := tokenize(expression)
	if len(tokens) == 0 {
		return 0, ErrEmptyExpression
	}

	rpn, err := toRPN(tokens)
	if err != nil {
		return 0, err
	}

	return evaluateRPN(rpn)
}

// Токенизация выражения
func tokenize(expression string) []string {
	var tokens []string
	var currentToken strings.Builder

	for _, char := range expression {
		if char == ' ' {
			continue
		} else if isOperator(string(char)) || char == '(' || char == ')' {
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		} else {
			currentToken.WriteRune(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

// Проверка, является ли символ оператором
func isOperator(token string) bool {
	_, exists := precedence[token]
	return exists
}

// Преобразование в обратную польскую нотацию (ОПН)
func toRPN(tokens []string) ([]string, error) {
	var output []string
	var stack []string

	for _, token := range tokens {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			output = append(output, token)
		} else if token == "(" {
			stack = append(stack, token)
		} else if token == ")" {
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, ErrNoOpeningParenthesis
			}
			stack = stack[:len(stack)-1]
		} else if isOperator(token) {
			for len(stack) > 0 && precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		} else {
			return nil, ErrInvalidExpression
		}
	}

	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, ErrNoClosingParenthesis
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return output, nil
}

// Вычисление результата из ОПН
func evaluateRPN(tokens []string) (float64, error) {
	var stack []float64

	for _, token := range tokens {
		if value, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, value)
		} else if isOperator(token) {
			if len(stack) < 2 {
				return 0, ErrShortExpression
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, ErrDevisionByZero
				}
				stack = append(stack, a/b)
			}
		} else {
			return 0, ErrInvalidExpression
		}
	}

	if len(stack) != 1 {
		return 0, ErrInvalidExpression
	}

	return stack[0], nil
}
