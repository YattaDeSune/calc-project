package calculation

import (
	"testing"
)

func TestToRPN(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testCases := []struct {
			name      string
			tokens    []string
			expected  []string
			expectErr bool
		}{
			{
				name:      "simple expression",
				tokens:    []string{"52.2", "/", "2"},
				expected:  []string{"52.2", "2", "/"},
				expectErr: false,
			},
			{
				name:      "with priority",
				tokens:    []string{"52.2", "+", "23", "*", "2"},
				expected:  []string{"52.2", "23", "2", "*", "+"},
				expectErr: false,
			},
			{
				name:      "with parenthesis",
				tokens:    []string{"11", "*", "(", "2", "+", "4", "-", "1", ")"},
				expected:  []string{"11", "2", "4", "+", "1", "-", "*"},
				expectErr: false,
			},
			{
				name:      "unary minus",
				tokens:    []string{"(", "-", "2", ")", "*", "3"},
				expected:  []string{"2", "~", "3", "*"},
				expectErr: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := ToRPN(tc.tokens)
				if (err != nil) != tc.expectErr {
					t.Errorf("expected error %v, but got %v", tc.expectErr, err)
				}
				if !equalSlices(got, tc.expected) {
					t.Errorf("expected %v, but got %v", tc.expected, got)
				}
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		testCases := []struct {
			name        string
			tokens      []string
			expectedErr error
		}{
			{
				name:        "empty expression",
				tokens:      []string{},
				expectedErr: ErrEmptyExpression,
			},
			{
				name:        "short expression",
				tokens:      []string{"2", "+"},
				expectedErr: ErrShortExpression,
			},
			{
				name:        "no opening parenthesis",
				tokens:      []string{"2", "+", "4", ")"},
				expectedErr: ErrNoOpeningParenthesis,
			},
			{
				name:        "no closing parenthesis",
				tokens:      []string{"2", "+", "(", "4", "*", "2"},
				expectedErr: ErrNoClosingParenthesis,
			},
			{
				name:        "invalid expression",
				tokens:      []string{"%", "+", "4"},
				expectedErr: ErrInvalidExpression,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := ToRPN(tc.tokens)
				if err != tc.expectedErr {
					t.Errorf("expected error %v, but got %v", tc.expectedErr, err)
				}
			})
		}
	})
}

func TestTokenize(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		expected   []string
	}{
		{
			name:       "simple expression",
			expression: "52.2/2",
			expected:   []string{"52.2", "/", "2"},
		},
		{
			name:       "with spaces",
			expression: "52.2 + 23 * 2",
			expected:   []string{"52.2", "+", "23", "*", "2"},
		},
		{
			name:       "with parenthesis",
			expression: "11 * (2+4-1)",
			expected:   []string{"11", "*", "(", "2", "+", "4", "-", "1", ")"},
		},
		{
			name:       "floating point numbers",
			expression: "3.14 + 2.71",
			expected:   []string{"3.14", "+", "2.71"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Tokenize(tc.expression)
			if !equalSlices(got, tc.expected) {
				t.Errorf("expected %v, but got %v", tc.expected, got)
			}
		})
	}
}

func TestNextTask(t *testing.T) {
	testCases := []struct {
		name          string
		rpn           []string
		stack         []string
		expectedArg1  string
		expectedArg2  string
		expectedOp    string
		expectedRPN   []string
		expectedStack []string
		expectErr     bool
	}{
		{
			name:          "simple operation",
			rpn:           []string{"2", "3", "+"},
			stack:         []string{},
			expectedArg1:  "2",
			expectedArg2:  "3",
			expectedOp:    "+",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     false,
		},
		{
			name:          "unary operation",
			rpn:           []string{"2", "~"},
			stack:         []string{},
			expectedArg1:  "2",
			expectedArg2:  "",
			expectedOp:    "~",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     false,
		},
		{
			name:          "not enough operands",
			rpn:           []string{"+"},
			stack:         []string{"2"},
			expectedArg1:  "",
			expectedArg2:  "",
			expectedOp:    "",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     true,
		},
		{
			name:          "multiplication operation",
			rpn:           []string{"2", "3", "*"},
			stack:         []string{},
			expectedArg1:  "2",
			expectedArg2:  "3",
			expectedOp:    "*",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     false,
		},
		{
			name:          "division operation",
			rpn:           []string{"6", "2", "/"},
			stack:         []string{},
			expectedArg1:  "6",
			expectedArg2:  "2",
			expectedOp:    "/",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     false,
		},
		{
			name:          "subtraction operation",
			rpn:           []string{"5", "3", "-"},
			stack:         []string{},
			expectedArg1:  "5",
			expectedArg2:  "3",
			expectedOp:    "-",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     false,
		},
		{
			name:          "not enough operands for unary operation",
			rpn:           []string{"~"},
			stack:         []string{},
			expectedArg1:  "",
			expectedArg2:  "",
			expectedOp:    "",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     true,
		},
		{
			name:          "empty rpn",
			rpn:           []string{},
			stack:         []string{},
			expectedArg1:  "",
			expectedArg2:  "",
			expectedOp:    "",
			expectedRPN:   []string{},
			expectedStack: []string{},
			expectErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			arg1, arg2, op, newRPN, newStack, err := NextTask(tc.rpn, tc.stack)
			if (err != nil) != tc.expectErr {
				t.Errorf("expected error %v, but got %v", tc.expectErr, err)
			}
			if arg1 != tc.expectedArg1 || arg2 != tc.expectedArg2 || op != tc.expectedOp {
				t.Errorf("expected args %s, %s and op %s, but got %s, %s and %s", tc.expectedArg1, tc.expectedArg2, tc.expectedOp, arg1, arg2, op)
			}
			if !equalSlices(newRPN, tc.expectedRPN) || !equalSlices(newStack, tc.expectedStack) {
				t.Errorf("expected RPN %v and stack %v, but got %v and %v", tc.expectedRPN, tc.expectedStack, newRPN, newStack)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
