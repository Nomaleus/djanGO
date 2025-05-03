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

	if strings.Contains(cleanExpr, "/0") {
		for i := 0; i < len(cleanExpr)-1; i++ {
			if cleanExpr[i] == '/' && cleanExpr[i+1] == '0' {
				if i+2 >= len(cleanExpr) || (!IsDigit(cleanExpr[i+2]) && cleanExpr[i+2] != '.') {
					return false
				}
			}
		}
	}

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
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
	}
}

func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
