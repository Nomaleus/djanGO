package tests

import (
	"djanGO/handlers"
	"djanGO/models"
	"djanGO/storage"
	"testing"
)

func TestTaskProcessor(t *testing.T) {
	tests := []struct {
		name          string
		task          *models.Task
		expectedValue float64
	}{
		{
			name: "Addition",
			task: &models.Task{
				Operation: "+",
				Arg1:      2,
				Arg2:      3,
			},
			expectedValue: 5,
		},
		{
			name: "Multiplication",
			task: &models.Task{
				Operation: "*",
				Arg1:      4,
				Arg2:      5,
			},
			expectedValue: 20,
		},
		{
			name: "Division",
			task: &models.Task{
				Operation: "/",
				Arg1:      10,
				Arg2:      2,
			},
			expectedValue: 5,
		},
		{
			name: "Division by zero",
			task: &models.Task{
				Operation: "/",
				Arg1:      10,
				Arg2:      0,
			},
			expectedValue: 0,
		},
	}

	store := storage.NewStorage()
	storeWrapper := storage.NewStorageWrapper(store)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := handlers.NewTaskProcessor(tt.task, storeWrapper)
			result := processor.Process()
			if result != tt.expectedValue {
				t.Errorf("Expected %f, got %f", tt.expectedValue, result)
			}
		})
	}
}

func TestCreateTasks(t *testing.T) {
	store := storage.NewStorage()
	storeWrapper := storage.NewStorageWrapper(store)

	t.Run("Simple expression", func(t *testing.T) {
		expr := &models.Expression{
			Original: "2+2",
		}
		processor := handlers.NewTaskProcessor(nil, storeWrapper)
		tasks, err := processor.CreateTasks(expr)

		if err != nil {
			t.Fatalf("CreateTasks() error = %v", err)
		}

		if len(tasks) != 3 {
			t.Fatalf("CreateTasks() returned %d tasks, want 3", len(tasks))
		}

		foundAdd := false
		for _, task := range tasks {
			if task.Operation == "+" {
				foundAdd = true
				if task.Arg1 != 2 || task.Arg2 != 2 {
					t.Errorf("Expected operation + with args [2 2], got [%v %v]", task.Arg1, task.Arg2)
				}
			}
		}

		if !foundAdd {
			t.Errorf("Expected to find a task with + operation")
		}
	})

	t.Run("Complex expression", func(t *testing.T) {
		expr := &models.Expression{
			Original: "2+3*4",
		}
		processor := handlers.NewTaskProcessor(nil, storeWrapper)
		tasks, err := processor.CreateTasks(expr)

		if err != nil {
			t.Fatalf("CreateTasks() error = %v", err)
		}

		if len(tasks) != 5 {
			t.Fatalf("CreateTasks() returned %d tasks, want 5", len(tasks))
		}

		foundMultiply := false
		foundAdd := false

		for _, task := range tasks {
			if task.Operation == "*" {
				foundMultiply = true
			}
			if task.Operation == "+" {
				foundAdd = true
			}
		}

		if !foundMultiply {
			t.Errorf("Expected to find a task with * operation")
		}
		if !foundAdd {
			t.Errorf("Expected to find a task with + operation")
		}
	})

	t.Run("Invalid expression", func(t *testing.T) {
		expr := &models.Expression{
			Original: "2++2",
		}
		processor := handlers.NewTaskProcessor(nil, storeWrapper)
		_, err := processor.CreateTasks(expr)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if err.Error() != "mismatched numbers and operations" {
			t.Errorf("Expected error containing \"mismatched numbers and operations\", got %v", err)
		}
	})
}
