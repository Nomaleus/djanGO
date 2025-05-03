package handlers

import (
	"djanGO/models"
	"djanGO/storage"

	"github.com/google/uuid"
)

func (h *Handler) CreateTestTask() *models.Task {
	task := &models.Task{
		ID:            uuid.New().String(),
		Operation:     "+",
		Arg1:          2,
		Arg2:          2,
		Status:        "PENDING",
		OperationTime: 1000,
	}
	h.Storage.AddTask(task)
	return task
}

func (h *Handler) CreateTestExpression() *models.Expression {
	expr := &models.Expression{
		ID:       uuid.New().String(),
		Original: "2+2",
		Status:   "PENDING",
	}
	h.Storage.AddExpression(expr)
	return expr
}

func (h *Handler) CleanupTestData() {
	store := storage.NewStorage()
	h.Storage = storage.NewStorageWrapper(store)
}
