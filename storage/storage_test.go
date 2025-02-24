package storage

import (
	"djanGO/models"
	"testing"
)

func TestStorage(t *testing.T) {
	t.Run("Task operations", func(t *testing.T) {
		store := NewStorage()

		task := &models.Task{
			ID:        "test-task",
			Operation: "+",
			Arg1:      2,
			Arg2:      3,
		}
		err := store.AddTask(task)
		if err != nil {
			t.Errorf("Failed to add task: %v", err)
		}

		err = store.AddTask(task)
		if err != ErrTaskExists {
			t.Errorf("Expected ErrTaskExists, got %v", err)
		}

		err = store.UpdateTaskResult("test-task", 5)
		if err != nil {
			t.Errorf("Failed to update task result: %v", err)
		}

		err = store.UpdateTaskResult("non-existent", 0)
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})

	t.Run("Expression operations", func(t *testing.T) {
		store := NewStorage()

		expr := &models.Expression{
			ID:       "test-expr",
			Original: "2+2",
			Status:   "PENDING",
		}
		store.AddExpression(expr)

		got, err := store.GetExpression("test-expr")
		if err != nil {
			t.Errorf("Failed to get expression: %v", err)
		}
		if got.Original != "2+2" {
			t.Errorf("Expected expression 2+2, got %s", got.Original)
		}

		exprs, err := store.GetAllExpressions()
		if err != nil {
			t.Errorf("Failed to get all expressions: %v", err)
		}
		if len(exprs) != 1 {
			t.Errorf("Expected 1 expression, got %d", len(exprs))
		}

		_, err = store.GetExpression("non-existent")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got %v", err)
		}
	})
}
