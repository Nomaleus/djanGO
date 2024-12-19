package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"djanGO/lexer"
	"djanGO/parser"
	"djanGO/utils"
)

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Expression == "" {
		http.Error(w, `{"error": "Expression is required"}`, http.StatusBadRequest)
		return
	}

	if !utils.IsValidExpression(req.Expression) {
		http.Error(w, `{"error": "Expression is not valid"}`, http.StatusUnprocessableEntity)
		return
	}

	result, err := Calc(req.Expression)
	if err != nil {
		if strings.Contains(err.Error(), "дели на ноль") {
			http.Error(w, `{"error": "Не дели на ноль!"}`, http.StatusUnprocessableEntity)
		} else {
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	resp := Response{
		Result: fmt.Sprintf("%.6f", result),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

func Calc(expression string) (float64, error) {
	newLexer := lexer.NewLexer(expression)
	newParser := parser.NewParser(newLexer)
	result, err := newParser.ParseExpression()
	if err != nil {
		return 0, err
	}
	if newParser.CurToken.Type != newLexer.TokenEOF {
		return 0, errors.New("Неожиданный токен после конца expression")
	}
	return result, nil
}
