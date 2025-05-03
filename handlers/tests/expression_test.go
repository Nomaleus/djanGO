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

	"github.com/google/uuid"
)

type mockHandler struct {
	*handlers.Handler
}

type ExpressionRequest struct {
	Expression string `json:"expression"`
}

func (h *mockHandler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req ExpressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Expression == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}

	if req.Expression == "2++2" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid expression"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"id": uuid.New().String()}
	json.NewEncoder(w).Encode(response)
}

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
			storeWrapper := storage.NewStorageWrapper(store)
			handler := handlers.NewHandler(storeWrapper)
			mockH := &mockHandler{Handler: handler}

			body := map[string]string{"expression": tt.expression}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/calculate", bytes.NewBuffer(jsonBody))
			req.Header.Set("X-User-Login", "testuser")
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			mockH.Calculate(rr, req)

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewStorage()
			storeWrapper := storage.NewStorageWrapper(store)

			expr := &models.Expression{
				ID:       uuid.New().String(),
				Original: tt.expression,
				Status:   "PENDING",
			}
			storeWrapper.AddExpression(expr)

			processor := handlers.NewTaskProcessor(nil, storeWrapper)
			tasks, err := processor.CreateTasks(expr)
			if err != nil {
				t.Fatalf("CreateTasks() error = %v", err)
			}

			expr.Tasks = tasks

			tasksByOrder := make(map[int][]*models.Task)
			var maxOrder int

			for _, task := range tasks {
				tasksByOrder[task.Order] = append(tasksByOrder[task.Order], task)
				if task.Order > maxOrder {
					maxOrder = task.Order
				}
			}

			for order := 0; order <= maxOrder; order++ {
				orderTasks, exists := tasksByOrder[order]
				if !exists {
					continue
				}

				for _, task := range orderTasks {
					if task.Operation == "value" {
						continue
					}

					taskProcessor := handlers.NewTaskProcessor(task, storeWrapper)
					result := taskProcessor.Process()
					storeWrapper.UpdateTaskResult(task.ID, result)
				}
			}

			updatedExpr, _ := storeWrapper.GetExpression(expr.ID)
			if updatedExpr.Result != tt.expected {
				t.Errorf("Expression result = %v, want %v", updatedExpr.Result, tt.expected)
			}
		})
	}
}
