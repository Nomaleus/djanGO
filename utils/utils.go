package utils

import "strings"

func IsDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func IsValidExpression(expression string) bool {
	for i := 0; i < len(expression); i++ {
		ch := expression[i]
		if !IsDigit(ch) && !strings.ContainsRune("+-*/().", rune(ch)) {
			return false
		}
	}

	if strings.Contains(expression, "..") {
		return false
	}

	if strings.Contains(expression, "+*") || strings.Contains(expression, "*/") ||
		strings.Contains(expression, "*+") || strings.Contains(expression, "/+") {
		return false
	}

	openBrackets := 0
	for _, ch := range expression {
		if ch == '(' {
			openBrackets++
		} else if ch == ')' {
			openBrackets--
		}
	}
	if openBrackets != 0 {
		return false
	}

	return true
}
