package utils

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func IsDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func IsValidExpression(expr string) bool {
	expr = strings.ReplaceAll(expr, " ", "")

	validChars := regexp.MustCompile(`^[0-9+\-*/().]+$`)
	if !validChars.MatchString(expr) {
		return false
	}

	if regexp.MustCompile(`\.{2,}|^\.|\.$|\d*\.\d*\.`).MatchString(expr) {
		return false
	}

	if regexp.MustCompile(`^[+\-*/]|[+\-*/]$`).MatchString(expr) {
		return false
	}

	if regexp.MustCompile(`[+\-*/][+\-*/]`).MatchString(expr) {
		return false
	}

	brackets := 0
	for _, ch := range expr {
		if ch == '(' {
			brackets++
		} else if ch == ')' {
			brackets--
			if brackets < 0 {
				return false
			}
		}
	}
	return brackets == 0
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
