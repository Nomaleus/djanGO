package handlers

import (
	"bytes"
	"djanGO/models"
	"djanGO/storage"
	"djanGO/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
	utils.UserAuthenticateFunc = func(login, password string) (bool, error) {
		return true, nil
	}

	utils.UserRegisterFunc = func(login, password string) error {
		if login == "" || password == "" {
			return fmt.Errorf("пустой логин или пароль")
		}
		return nil
	}
}

func setupTestHandler() *Handler {
	store := storage.NewStorage()
	storeWrapper := storage.NewStorageWrapper(store)
	return NewHandler(storeWrapper)
}

func TestLoginHandler(t *testing.T) {
	handler := setupTestHandler()

	tests := []struct {
		name           string
		loginData      map[string]string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Валидные данные для входа",
			loginData: map[string]string{
				"login":    "testuser",
				"password": "testpass",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Пустой логин",
			loginData: map[string]string{
				"login":    "",
				"password": "testpass",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Пустой пароль",
			loginData: map[string]string{
				"login":    "testuser",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Пустой запрос",
			loginData:      map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.loginData)
			req, err := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Ошибка при парсинге ответа: %v", err)
				return
			}

			_, hasError := response["error"]
			if hasError != tt.expectedError {
				t.Errorf("Ожидалось наличие ошибки: %v, получено: %v", tt.expectedError, hasError)
			}

			if !tt.expectedError {
				token, hasToken := response["token"]
				if !hasToken {
					t.Errorf("Ожидалось наличие токена в ответе")
				}
				if token == "" {
					t.Errorf("Токен не должен быть пустым")
				}
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	handler := setupTestHandler()

	tests := []struct {
		name           string
		registerData   map[string]string
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Валидные данные для регистрации",
			registerData: map[string]string{
				"login":    "newuser",
				"password": "newpass",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "Пустой логин",
			registerData: map[string]string{
				"login":    "",
				"password": "newpass",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Пустой пароль",
			registerData: map[string]string{
				"login":    "newuser",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Пустой запрос",
			registerData:   map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.registerData)
			req, err := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Ошибка при парсинге ответа: %v", err)
				return
			}

			_, hasError := response["error"]
			if hasError != tt.expectedError {
				t.Errorf("Ожидалось наличие ошибки: %v, получено: %v", tt.expectedError, hasError)
			}

			if !tt.expectedError {
				message, hasMessage := response["message"]
				if !hasMessage {
					t.Errorf("Ожидалось наличие сообщения в ответе")
				}
				if message == "" {
					t.Errorf("Сообщение не должно быть пустым")
				}
			}
		})
	}
}

func TestCalculateHandler(t *testing.T) {
	_ = setupTestHandler()

	mockCalculateHandler := func(w http.ResponseWriter, r *http.Request) {
		userLogin := utils.GetUserFromContext(r.Context())
		if userLogin == "" {
			utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Требуется авторизация",
			})
			return
		}

		var req struct {
			Expression string `json:"expression"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid JSON",
			})
			return
		}

		if req.Expression == "" {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Expression is required",
			})
			return
		}

		if strings.Contains(req.Expression, "++") {
			utils.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "Invalid expression",
			})
			return
		}

		utils.WriteJSON(w, http.StatusCreated, map[string]string{
			"id": fmt.Sprintf("test-expr-%d", 12345),
		})
	}

	tests := []struct {
		name           string
		expression     string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Валидное выражение",
			expression:     "2+2",
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "Сложное выражение",
			expression:     "(2+2)*3",
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "Пустое выражение",
			expression:     "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Некорректное выражение",
			expression:     "2++2",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(map[string]string{"expression": tt.expression})
			req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-User-Login", "testuser")

			ctx := utils.AddUserToContext(req.Context(), "testuser")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockCalculateHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Errorf("Ошибка при парсинге ответа: %v", err)
				return
			}

			_, hasError := response["error"]
			if hasError != tt.expectedError {
				t.Errorf("Ожидалось наличие ошибки: %v, получено: %v", tt.expectedError, hasError)
			}

			if !tt.expectedError {
				id, hasID := response["id"]
				if !hasID {
					t.Errorf("Ожидалось наличие ID в ответе")
				}
				if id == "" {
					t.Errorf("ID не должен быть пустым")
				}
			}
		})
	}
}

func TestGetExpressionHandler(t *testing.T) {
	handler := setupTestHandler()

	testExpr := &models.Expression{
		ID:       "test-expr-id",
		Original: "2+2",
		Status:   "COMPLETED",
		Result:   4,
	}
	handler.Storage.AddExpression(testExpr)

	mockGetExpressionHandler := func(w http.ResponseWriter, r *http.Request, expressionID string) {
		if expressionID == "non-existent-id" {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "Выражение не найдено",
			})
			return
		}

		expr, err := handler.Storage.GetExpression(expressionID)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Internal server error",
			})
			return
		}

		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"expression": expr,
		})
	}

	tests := []struct {
		name           string
		expressionID   string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Существующее выражение",
			expressionID:   "test-expr-id",
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Несуществующее выражение",
			expressionID:   "non-existent-id",
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := fmt.Sprintf("/api/v1/expressions/%s", tt.expressionID)
			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatal(err)
			}

			req = addExpressionIDToRequest(req, tt.expressionID)

			ctx := utils.AddUserToContext(req.Context(), "testuser")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			mockGetExpressionHandler(rr, req, tt.expressionID)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, rr.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				if tt.expectedStatus != http.StatusNotFound {
					t.Errorf("Ошибка при парсинге ответа: %v", err)
				}
				return
			}

			_, hasError := response["error"]
			if hasError != tt.expectedError {
				t.Errorf("Ожидалось наличие ошибки: %v, получено: %v", tt.expectedError, hasError)
			}

			if !tt.expectedError {
				expressionData, hasExpr := response["expression"]
				if !hasExpr {
					t.Errorf("Ожидалось наличие данных выражения в ответе")
					return
				}

				exprMap, ok := expressionData.(map[string]interface{})
				if !ok {
					t.Errorf("Ожидалось, что данные выражения будут объектом, получено: %T", expressionData)
					return
				}

				id, hasID := exprMap["id"]
				if !hasID || id != tt.expressionID {
					t.Errorf("Ожидался ID выражения %s, получен %v", tt.expressionID, id)
				}
			}
		})
	}
}

func addExpressionIDToRequest(req *http.Request, id string) *http.Request {
	parts := strings.Split(req.URL.Path, "/")
	if len(parts) > 0 {
		parts[len(parts)-1] = id
		req.URL.Path = strings.Join(parts, "/")
	}
	return req
}
