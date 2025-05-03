package handlers

import (
	"djanGO/db"
	"djanGO/models"
	"djanGO/storage"
	"djanGO/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type Request struct {
	Expression string `json:"expression"`
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	userLogin := utils.GetUserFromContext(r.Context())
	fmt.Printf("DEBUG Calculate: Пользователь из функции GetUserFromContext: '%s'\n", userLogin)

	fmt.Println("DEBUG Calculate: Заголовки запроса:")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", name, value)
		}
	}

	for _, cookie := range r.Cookies() {
		fmt.Printf("DEBUG Calculate: Cookie %s: %s\n", cookie.Name, cookie.Value)
	}

	if userLogin == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "Требуется авторизация",
		})
		return
	}

	fmt.Printf("Расчет выражения для пользователя: %s\n", userLogin)

	var req Request
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

	if !utils.IsValidExpression(req.Expression) {
		utils.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{
			"error": "Invalid expression",
		})
		return
	}

	_, err := db.AddExpression(req.Expression, userLogin)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при сохранении выражения: " + err.Error(),
		})
		return
	}

	expressions, err := db.GetExpressionsByUser(userLogin)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении выражения: " + err.Error(),
		})
		return
	}

	if len(expressions) == 0 {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Не удалось найти добавленное выражение",
		})
		return
	}

	exprUUID := expressions[0].ID

	expr := &models.Expression{
		ID:       exprUUID,
		Original: req.Expression,
		Status:   "PENDING",
	}

	h.Storage.AddExpression(expr)

	processor := NewTaskProcessor(nil, h.Storage)
	tasks, err := processor.CreateTasks(expr)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	expr.Tasks = tasks

	for _, task := range tasks {
		dbTask := &db.Task{
			ExpressionID:  exprUUID,
			Operation:     task.Operation,
			Arg1:          task.Arg1,
			Arg2:          task.Arg2,
			Status:        task.Status,
			Order:         task.Order,
			OperationTime: task.OperationTime,
		}

		_, err := db.AddTask(dbTask)
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при сохранении задачи: " + err.Error(),
			})
			return
		}
	}

	if expr.Status == "ERROR" {
		for _, task := range tasks {
			if task.Status == "ERROR" && task.Error == "division by zero" {
				utils.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{
					"error": "Division by zero",
				})
				return
			}
		}
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"id": expr.ID,
	})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}

	fmt.Printf("Получение выражения по ID: %s\n", id)

	expr, err := h.Storage.GetExpression(id)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Выражение не найдено", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if expr.Status == "ERROR" {
		for _, task := range expr.Tasks {
			if task.Status == "ERROR" && task.Error == "division by zero" {
				utils.WriteJSON(w, http.StatusUnprocessableEntity, map[string]string{
					"error": "Division by zero",
				})
				return
			}
		}
	}
	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"expression": expr,
	})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userLogin := r.Header.Get("X-User-Login")
	if userLogin == "" {
		userLogin = r.URL.Query().Get("user_login")
	}

	if userLogin == "" {
		userLogin = "admin"
	}

	expressions, err := db.GetExpressionsByUser(userLogin)
	if err != nil {
		fmt.Printf("Ошибка при получении выражений: %v\n", err)
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Внутренняя ошибка сервера",
		})
		return
	}

	response := models.ExpressionsResponse{
		Expressions: make([]models.ExpressionResponse, 0, len(expressions)),
	}

	for _, expr := range expressions {
		var resultPtr *float64
		if expr.Status == "COMPLETED" {
			result := expr.Result
			resultPtr = &result
		}

		idStr := fmt.Sprintf("%v", expr.ID)

		response.Expressions = append(response.Expressions, models.ExpressionResponse{
			ID:      idStr,
			Status:  expr.Status,
			Result:  resultPtr,
			Error:   expr.Error,
			Created: expr.CreatedAt,
			Text:    expr.Text,
		})
	}

	if response.Expressions == nil {
		response.Expressions = []models.ExpressionResponse{}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) GetUserHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userLogin := utils.GetUserFromContext(r.Context())
	if userLogin == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "Требуется авторизация",
		})
		return
	}

	fmt.Printf("GetUserHistory для пользователя: %s\n", userLogin)

	expressions, err := db.GetExpressionsByUser(userLogin)
	if err != nil {
		fmt.Printf("Ошибка при получении выражений: %v\n", err)
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Внутренняя ошибка сервера",
		})
		return
	}

	response := models.ExpressionsResponse{
		Expressions: make([]models.ExpressionResponse, 0, len(expressions)),
	}

	for i, expr := range expressions {
		fmt.Printf("Обработка выражения [%d]: ID=%v, Status=%s\n", i, expr.ID, expr.Status)

		var resultPtr *float64
		if expr.Status == "COMPLETED" {
			result := expr.Result
			resultPtr = &result
		}

		response.Expressions = append(response.Expressions, models.ExpressionResponse{
			ID:      expr.ID,
			Status:  expr.Status,
			Result:  resultPtr,
			Error:   expr.Error,
			Created: expr.CreatedAt,
			Text:    expr.Text,
		})
	}

	if response.Expressions == nil {
		response.Expressions = []models.ExpressionResponse{}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
