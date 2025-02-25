package tests

import (
	"bytes"
	"djanGO/handlers"
	"djanGO/models"
	"djanGO/storage"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name           string
		expression     string
		expectedStatus int
		expectError    bool
		errorMessage   string
	}{
		{
			name:           "Valid simple expression",
			expression:     "2+2",
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "Valid complex expression",
			expression:     "(2+2)*3",
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "Empty expression",
			expression:     "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMessage:   "Invalid JSON",
		},
		{
			name:           "Invalid expression",
			expression:     "2++2",
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
			errorMessage:   "Invalid expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewStorage()
			handler := handlers.NewHandler(store)

			body := map[string]string{"expression": tt.expression}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/calculate", bytes.NewBuffer(jsonBody))
			rr := httptest.NewRecorder()

			handler.Calculate(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			json.NewDecoder(rr.Body).Decode(&response)

			if tt.expectError {
				if errMsg, ok := response["error"].(string); !ok || errMsg != tt.errorMessage {
					t.Errorf("Expected error message %q, got %q", tt.errorMessage, errMsg)
				}
			} else {
				if id, ok := response["id"].(string); !ok {
					t.Error("Expected id in response")
				} else if id == "" {
					t.Error("Expected non-empty id")
				}
			}
		})
	}
}

func TestExpressionCalculations(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   float64
	}{
		{
			name:       "Простое сложение",
			expression: "2+2",
			expected:   4,
		},
		{
			name:       "Выражение с приоритетом",
			expression: "2+2*2",
			expected:   6,
		},
		{
			name:       "Выражение со скобками",
			expression: "(2+2)*2",
			expected:   8,
		},
		{
			name:       "Сложное выражение",
			expression: "2+2*3+4",
			expected:   12,
		},
		{
			name:       "Сложное выражение со скобками",
			expression: "(2+2)*(3+4)",
			expected:   28,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := handlers.ProcessExpression(tt.expression)
			if err != nil {
				t.Fatalf("ProcessExpression() error = %v", err)
			}

			if expr.Result != tt.expected {
				t.Errorf("ProcessExpression() = %v, want %v", expr.Result, tt.expected)
			}
		})
	}
}

func TestTaskCreationAndExecution(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expected   float64
	}{
		{
			name:       "Простое сложение",
			expression: "2+2",
			expected:   4,
		},
		{
			name:       "Умножение",
			expression: "3*4",
			expected:   12,
		},
		{
			name:       "Выражение с приоритетом",
			expression: "2+2*2",
			expected:   6,
		},
		{
			name:       "Выражение со скобками",
			expression: "(2+2)*2",
			expected:   8,
		},
	}

	store := storage.NewStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &models.Expression{
				ID:       store.GetNextID(),
				Original: tt.expression,
				Status:   "PENDING",
			}
			store.AddExpression(expr)

			processor := handlers.NewTaskProcessor(nil, store)
			tasks, err := processor.CreateTasks(expr)
			if err != nil {
				t.Fatalf("CreateTasks() error = %v", err)
			}

			expr.Tasks = tasks

			for _, task := range tasks {
				procTask := handlers.NewTaskProcessor(task, store)
				result := procTask.Process()
				store.UpdateTaskResult(task.ID, result)
			}

			updatedExpr, _ := store.GetExpression(expr.ID)

			if updatedExpr.Result != tt.expected {
				t.Errorf("Expression result = %v, want %v", updatedExpr.Result, tt.expected)
			}
		})
	}
}
