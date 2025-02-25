package handlers

import (
	"djanGO/models"
	"djanGO/storage"
	"djanGO/utils"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi"
)

type Request struct {
	Expression string `json:"expression"`
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
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

	expr := &models.Expression{
		ID:       uuid.New().String(),
		Original: req.Expression,
		Status:   "PENDING",
	}

	processor := NewTaskProcessor(nil, h.Storage)
	tasks, err := processor.CreateTasks(expr)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	expr.Tasks = tasks
	h.Storage.AddExpression(expr)

	for _, task := range tasks {
		go func(t *models.Task) {
			processor := NewTaskProcessor(t, h.Storage)
			result := processor.Process()
			h.Storage.UpdateTaskResult(t.ID, result)
		}(task)
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"id": expr.ID,
	})
}

func (h *Handler) GetExpressionByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}

	expr, err := h.Storage.GetExpression(id)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Выражение не найдено", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, models.ExpressionWrapper{
		Expression: models.ExpressionResponse{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		},
	})
}

func (h *Handler) GetExpressions(w http.ResponseWriter, r *http.Request) {
	expressions, err := h.Storage.GetAllExpressions()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := models.ExpressionsResponse{
		Expressions: make([]models.ExpressionResponse, len(expressions)),
	}

	for i, expr := range expressions {
		response.Expressions[i] = models.ExpressionResponse{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		}
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
