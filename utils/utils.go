package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func IsDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func IsValidExpression(expr string) bool {
	if expr == "" {
		return false
	}

	cleanExpr := strings.ReplaceAll(expr, " ", "")

	if len(cleanExpr) > 0 {
		if strings.ContainsRune("+-*/", rune(cleanExpr[0])) || strings.ContainsRune("+-*/", rune(cleanExpr[len(cleanExpr)-1])) {
			return false
		}
	}

	for i := 0; i < len(expr)-1; i++ {
		if (expr[i] >= '0' && expr[i] <= '9') && expr[i+1] == '(' {
			return false
		}
		if expr[i] == ')' && (expr[i+1] >= '0' && expr[i+1] <= '9') {
			return false
		}
	}

	for _, c := range expr {
		if !strings.ContainsRune("0123456789+-*/() .", c) {
			return false
		}
	}

	count := 0
	for _, c := range expr {
		if c == '(' {
			count++
		} else if c == ')' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	if count != 0 {
		return false
	}

	for i := 0; i < len(cleanExpr)-1; i++ {
		if strings.ContainsRune("+-*/", rune(cleanExpr[i])) && strings.ContainsRune("+-*/", rune(cleanExpr[i+1])) {
			return false
		}
	}

	return true
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
