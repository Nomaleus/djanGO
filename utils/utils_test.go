package utils

import (
	"testing"
)

func TestIsValidExpression(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"123", true},
		{"123+456", true},
		{"123 + 456", true},
		{"(123+456)", true},
		{"(123+456)*789", true},
		{"2+2", true},
		{"2 + 2", true},
		{"2 + 2 * 3", true},
		{"(2 + 2) * 3", true},
		{"2++2", false},
		{"2 ++ 2", false},
		{"2 + + 2", false},
		{"2 + 2 +", false},
		{"+ 2 + 2", false},
		{"2 + 2 * ", false},
		{"* 2 + 2", false},
		{"2 + 2 * 3 /", false},
		{"/ 2 + 2 * 3", false},
		{"(2 + 2", false},
		{"2 + 2)", false},
		{"(2 + 2))", false},
		{"((2 + 2)", false},
		{"2 + 2 * 3 / 0", false},
		{"", false},
		{"abc", false},
		{"2 + 2 * abc", false},
		{"444(333)", false},
		{"(123)456", false},
		{"((123))888", false},
		{"5(123)", false},
		{"(1(2)3)", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := IsValidExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsValidExpression(%q) = %v, want %v", tt.expr, result, tt.expected)
			}
		})
	}
}
