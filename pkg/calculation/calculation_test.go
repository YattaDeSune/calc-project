package calculation

import (
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"3 + 4", []string{"3", "+", "4"}},
		{"3.5 - 2.1", []string{"3.5", "-", "2.1"}},
		{"(3 + 4) * 2", []string{"(", "3", "+", "4", ")", "*", "2"}},
		{"-3 + 4", []string{"-", "3", "+", "4"}},
		{"3 + 4 * 2 / (1 - 5)", []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")"}},
		{"3 + 4 * 2 / (1 - 5)", []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")"}},
		{"3 + 4 * 2 / (1 - 5)", []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")"}},
		{"3 + 4 * 2 / (1 - 5)", []string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Tokenize(tt.input)
			if !equalSlices(result, tt.expected) {
				t.Errorf("Tokenize(%q) = %v, expected %v", tt.input, result, tt.expected)
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

func TestToRPN(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
		err      error
	}{
		{[]string{"3", "+", "4"}, []string{"3", "4", "+"}, nil},
		{[]string{"3", "+", "4", "*", "2"}, []string{"3", "4", "2", "*", "+"}, nil},
		{[]string{"(", "3", "+", "4", ")", "*", "2"}, []string{"3", "4", "+", "2", "*"}, nil},
		{[]string{"-", "3", "+", "4"}, []string{"3", "~", "4", "+"}, nil},
		{[]string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5", ")"}, []string{"3", "4", "2", "*", "1", "5", "-", "/", "+"}, nil},
		{[]string{"3", "+", "4", "*", "2", "/", "(", "1", "-", "5"}, nil, ErrNoClosingParenthesis},
		{[]string{"3", "+", "4", "*", "2", "/", "1", "-", "5", ")"}, nil, ErrNoOpeningParenthesis},
		{[]string{"3", "+"}, nil, ErrShortExpression},
		{[]string{}, nil, ErrEmptyExpression},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.input, " "), func(t *testing.T) {
			result, err := ToRPN(tt.input)
			if err != tt.err {
				t.Errorf("ToRPN(%v) error = %v, expected %v", tt.input, err, tt.err)
			}
			if !equalSlices(result, tt.expected) {
				t.Errorf("ToRPN(%v) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
