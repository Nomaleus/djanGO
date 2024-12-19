package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"djanGO/handlers"
	"djanGO/utils"
)

func TestCalc(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
		err        bool
	}{
		{"2+2", 4, false},
		{"2-2", 0, false},
		{"3*3", 9, false},
		{"10/2", 5, false},
		{"5/0", 0, true},
		{"(3+2)*2", 10, false},
		{"(3+2)*(2+3)", 25, false},
		{"2 + (3 * 4)", 14, false},
		{"2 * (3 + 4)", 14, false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Evaluating: %s", test.expression), func(t *testing.T) {
			result, err := handlers.Calc(test.expression)

			if test.err {
				if err == nil {
					t.Errorf("Expected error for expression: %s, but got nil", test.expression)
				}
			} else {
				if err != nil {
					t.Errorf("Error evaluating expression: %s", test.expression)
				}
				if result != test.expected {
					t.Errorf("Expected %f for expression %s, but got %f", test.expected, test.expression, result)
				}
			}
		})
	}
}

func TestIsValidExpression(t *testing.T) {
	tests := []struct {
		expression string
		valid      bool
	}{
		{"2+2", true},
		{"2-2", true},
		{"3*3", true},
		{"10/2", true},
		{"(3+2)*2", true},
		{"abc+2", false},
		{"2+2$", false},
		{"2+2..", false},
		{"2+*2", false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Validating: %s", test.expression), func(t *testing.T) {
			result := utils.IsValidExpression(test.expression)
			if result != test.valid {
				t.Errorf("Expected %v for expression: %s, but got %v", test.valid, test.expression, result)
			}
		})
	}
}

func TestCalculateHandler(t *testing.T) {
	tests := []struct {
		expression string
		expected   string
		statusCode int
	}{
		{"2+2", `4.000000`, http.StatusOK},
		{"3*3", `9.000000`, http.StatusOK},
		{"10/2", `5.000000`, http.StatusOK},
		{"5/0", `Не дели на ноль!`, http.StatusUnprocessableEntity},
		{"(3+2)*2", `10.000000`, http.StatusOK},
		{"2+*", `Expression is not valid`, http.StatusUnprocessableEntity},
		{"abc+2", `Expression is not valid`, http.StatusUnprocessableEntity},
		{"", `Expression is required`, http.StatusBadRequest},
		{"(3+*2)", `Expression is not valid`, http.StatusUnprocessableEntity},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Processing: %s", test.expression), func(t *testing.T) {
			reqBody := fmt.Sprintf(`{"expression": "%s"}`, test.expression)
			req := httptest.NewRequest("POST", "/api/v1/calculate", strings.NewReader(reqBody))
			rr := httptest.NewRecorder()

			handlers.CalculateHandler(rr, req)

			res := rr.Result()
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					return
				}
			}(res.Body)

			if res.StatusCode != test.statusCode {
				t.Errorf("Expected status code %d, but got %d", test.statusCode, res.StatusCode)
			}

			var response map[string]string
			err := json.NewDecoder(res.Body).Decode(&response)
			if err != nil {
				t.Errorf("Error decoding response: %v", err)
			}

			if response["result"] != "" && response["result"] != test.expected {
				t.Errorf("Expected result %s, but got %s", test.expected, response["result"])
			}

			if response["error"] != "" && response["error"] != test.expected {
				t.Errorf("Expected error %s, but got %s", test.expected, response["error"])
			}
		})
	}
}
