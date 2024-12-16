package calculation

import "testing"

func TestCalc(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testCases := []struct {
			name       string
			expression string
			expected   float64
		}{
			{
				name:       "simple expression",
				expression: "52.2/2",
				expected:   26.1,
			},
			{
				name:       "with priority",
				expression: "52.2+23*2",
				expected:   98.2,
			},
			{
				name:       "with parenthesis",
				expression: "11 * (2+4-1)",
				expected:   55,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, _ := Calc(tc.expression)
				if got != tc.expected {
					t.Errorf("expected %g, but got %g", tc.expected, got)
				}
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		testCases := []struct {
			name        string
			expression  string
			expectedErr error
		}{
			{
				name:        "empty expression",
				expression:  "",
				expectedErr: ErrEmptyExpression,
			},
			{
				name:        "short expression",
				expression:  "2+",
				expectedErr: ErrShortExpression,
			},
			{
				name:        "no opening parenthesis",
				expression:  "2+4)",
				expectedErr: ErrNoOpeningParenthesis,
			},
			{
				name:        "no closing parenthesis",
				expression:  "2+(4*2",
				expectedErr: ErrNoClosingParenthesis,
			},
			{
				name:        "devision by zero",
				expression:  "6/0",
				expectedErr: ErrDevisionByZero,
			},
			{
				name:        "invalid expression",
				expression:  "%+4",
				expectedErr: ErrInvalidExpression,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := Calc(tc.expression)
				if err != tc.expectedErr {
					t.Errorf("expected %g, but got %g", tc.expectedErr, err)
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
			expression: "3 + 4",
			expected:   []string{"3", "+", "4"},
		},
		{
			name:       "with parentheses",
			expression: "(1 + 2) * 3",
			expected:   []string{"(", "1", "+", "2", ")", "*", "3"},
		},
		{
			name:       "complex expression",
			expression: "10 + 2 * (3 - 1)",
			expected:   []string{"10", "+", "2", "*", "(", "3", "-", "1", ")"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tokenize(tc.expression)
			if !equalSlices(got, tc.expected) {
				t.Errorf("expected %v, but got %v", tc.expected, got)
			}
		})
	}
}

func TestToRPN(t *testing.T) {
	testCases := []struct {
		name      string
		tokens    []string
		expected  []string
		expectErr bool
	}{
		{
			name:     "simple expression",
			tokens:   []string{"3", "+", "4"},
			expected: []string{"3", "4", "+"},
		},
		{
			name:     "with parentheses",
			tokens:   []string{"(", "1", "+", "2", ")", "*", "3"},
			expected: []string{"1", "2", "+", "3", "*"},
		},
		{
			name:     "complex expression",
			tokens:   []string{"10", "+", "2", "*", "(", "3", "-", "1", ")"},
			expected: []string{"10", "2", "3", "1", "-", "*", "+"},
		},
		{
			name:      "mismatched parentheses",
			tokens:    []string{"(", "2", "+", "3"},
			expectErr: true,
		},
		{
			name:      "no opening parentheses",
			tokens:    []string{")", "2", "+", "3"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := toRPN(tc.tokens)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got %v", err)
				}
				if !equalSlices(got, tc.expected) {
					t.Errorf("expected %v, but got %v", tc.expected, got)
				}
			}
		})
	}
}

func TestEvaluateRPN(t *testing.T) {
	testCases := []struct {
		name      string
		tokens    []string
		expected  float64
		expectErr error
	}{
		{
			name:     "simple evaluation",
			tokens:   []string{"3", "4", "+"},
			expected: 7,
		},
		{
			name:     "simple subtraction",
			tokens:   []string{"4", "2", "-"},
			expected: 2,
		},
		{
			name:     "with multiplication",
			tokens:   []string{"2", "3", "*", "4", "+"},
			expected: 10,
		},
		{
			name:     "division",
			tokens:   []string{"6", "2", "/"},
			expected: 3,
		},
		{
			name:      "division by zero",
			tokens:    []string{"6", "0", "/"},
			expectErr: ErrDevisionByZero,
		},
		{
			name:      "invalid RPN",
			tokens:    []string{"3", "+"},
			expectErr: ErrShortExpression,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := evaluateRPN(tc.tokens)
			if tc.expectErr != nil {
				if err == nil {
					t.Errorf("expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, but got %v", err)
				}
				if got != tc.expected {
					t.Errorf("expected %g, but got %g", tc.expected, got)
				}
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
