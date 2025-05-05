package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"djanGO/handlers"
	"djanGO/models"
	"djanGO/storage"
	"djanGO/utils"
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

func MockCalculate(w http.ResponseWriter, r *http.Request, storageWrapper *storage.StorageWrapper) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

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
			"error": "Invalid JSON",
		})
		return
	}

	expr := &models.Expression{
		ID:       fmt.Sprintf("test-%d", os.Getpid()),
		Original: req.Expression,
		Status:   "COMPLETED",
		Result:   4,
	}

	storageWrapper.AddExpression(expr)

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"id": expr.ID,
	})
}

func MockGetExpression(w http.ResponseWriter, r *http.Request, storageWrapper *storage.StorageWrapper) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	id := parts[len(parts)-1]

	if id == "" {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "ID не указан",
		})
		return
	}

	expr, err := storageWrapper.GetExpression(id)
	if err != nil {
		if err == storage.ErrNotFound {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{
				"error": "Выражение не найдено",
			})
			return
		}
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"expression": expr,
	})
}

func setupIntegrationTest() (*httptest.Server, func()) {
	store := storage.NewStorage()
	storageWrapper := storage.NewStorageWrapper(store)

	handler := handlers.NewHandler(storageWrapper)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/register", handler.Register)
	mux.HandleFunc("/api/v1/login", handler.Login)

	mux.HandleFunc("/api/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			login, err := utils.ValidateToken(tokenString)
			if err == nil && login != "" {
				ctx := utils.AddUserToContext(r.Context(), login)
				r = r.WithContext(ctx)
			}
		}
		MockCalculate(w, r, storageWrapper)
	})

	mux.HandleFunc("/api/v1/expressions/", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			login, err := utils.ValidateToken(tokenString)
			if err == nil && login != "" {
				ctx := utils.AddUserToContext(r.Context(), login)
				r = r.WithContext(ctx)
			}
		}
		MockGetExpression(w, r, storageWrapper)
	})

	server := httptest.NewServer(mux)

	cleanup := func() {
		server.Close()
	}

	return server, cleanup
}

func TestIntegrationAPI(t *testing.T) {
	server, cleanup := setupIntegrationTest()
	defer cleanup()

	testUser := fmt.Sprintf("testuser_%d", os.Getpid())
	testPassword := "testpass123"

	registerPayload := map[string]string{
		"login":    testUser,
		"password": testPassword,
	}

	t.Run("Регистрация пользователя", func(t *testing.T) {
		jsonData, _ := json.Marshal(registerPayload)
		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/register", server.URL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)

		if err != nil {
			t.Fatalf("Ошибка при отправке запроса на регистрацию: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			var errorResp map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
				t.Fatalf("Ошибка регистрации: %v, код %d", errorResp["error"], resp.StatusCode)
			} else {
				t.Fatalf("Ошибка регистрации, код %d", resp.StatusCode)
			}
		}
	})

	t.Run("Вход пользователя", func(t *testing.T) {
		jsonData, _ := json.Marshal(registerPayload)
		resp, err := http.Post(
			fmt.Sprintf("%s/api/v1/login", server.URL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)

		if err != nil {
			t.Fatalf("Ошибка при отправке запроса на вход: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Ошибка входа, код %d", resp.StatusCode)
		}

		var loginResp struct {
			Success bool   `json:"success"`
			Login   string `json:"login"`
			Token   string `json:"token"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
			t.Fatalf("Ошибка при чтении ответа: %v", err)
		}

		jwt := loginResp.Token
		if jwt == "" {
			t.Fatal("JWT токен отсутствует в ответе")
		}

		t.Run("Расчет выражения", func(t *testing.T) {
			expression := "2+2"
			expressionPayload := map[string]string{
				"expression": expression,
			}
			jsonData, _ := json.Marshal(expressionPayload)

			req, _ := http.NewRequest(
				"POST",
				fmt.Sprintf("%s/api/v1/calculate", server.URL),
				bytes.NewBuffer(jsonData),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Ошибка при отправке запроса на расчет: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				t.Fatalf("Ошибка расчета выражения, код %d", resp.StatusCode)
			}

			var calcResp struct {
				ID      string `json:"id"`
				Message string `json:"message"`
				Error   string `json:"error"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&calcResp); err != nil {
				t.Fatalf("Ошибка при чтении ответа: %v", err)
			}

			expressionID := calcResp.ID
			if expressionID == "" {
				t.Fatalf("Ошибка: ID выражения не получен")
			}

			t.Run("Получение результата выражения", func(t *testing.T) {
				req, _ := http.NewRequest(
					"GET",
					fmt.Sprintf("%s/api/v1/expressions/%s", server.URL, expressionID),
					nil,
				)
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("Ошибка при отправке запроса на получение выражения: %v", err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Fatalf("Ошибка получения выражения, код %d", resp.StatusCode)
				}

				var exprResp struct {
					Expression models.Expression `json:"expression"`
					Error      string            `json:"error"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&exprResp); err != nil {
					t.Fatalf("Ошибка при чтении ответа: %v", err)
				}

				if exprResp.Expression.ID != expressionID {
					t.Errorf("Ожидался ID %s, получен %s", expressionID, exprResp.Expression.ID)
				}
			})
		})
	})
}
