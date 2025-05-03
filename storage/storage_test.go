package storage

import (
	"djanGO/models"
	"strings"
	"testing"
)

func TestStorage(t *testing.T) {
	t.Run("Task operations", func(t *testing.T) {
		store := NewStorage()
		wrapper := NewStorageWrapper(store)

		task := &models.Task{
			ID:        "test-task",
			Operation: "+",
			Arg1:      2,
			Arg2:      3,
		}
		err := wrapper.AddTask(task)
		if err != nil {
			t.Errorf("Failed to add task: %v", err)
		}

		err = wrapper.AddTask(task)
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected error about task already existing, got: %v", err)
		}

		err = wrapper.UpdateTaskResult("test-task", 5)
		if err != nil && !strings.Contains(err.Error(), "выражение для задачи не найдено") {
			t.Errorf("Unexpected error updating task result: %v", err)
		}

		err = wrapper.UpdateTaskResult("non-existent", 0)
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected error about task not found, got: %v", err)
		}
	})

	t.Run("Expression operations", func(t *testing.T) {
		store := NewStorage()
		wrapper := NewStorageWrapper(store)

		expr := &models.Expression{
			ID:       "test-expr",
			Original: "2+2",
			Status:   "PENDING",
		}
		wrapper.AddExpression(expr)

		got, err := wrapper.GetExpression("test-expr")
		if err != nil {
			t.Errorf("Failed to get expression: %v", err)
		}
		if got.Original != "2+2" {
			t.Errorf("Expected expression 2+2, got %s", got.Original)
		}

		exprs, err := wrapper.GetAllExpressions()
		if err != nil {
			t.Errorf("Failed to get all expressions: %v", err)
		}
		if len(exprs) != 1 {
			t.Errorf("Expected 1 expression, got %d", len(exprs))
		}

		_, err = wrapper.GetExpression("non-existent")
		if err == nil || !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected error about expression not found, got: %v", err)
		}
	})
}
