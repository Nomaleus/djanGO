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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := handlers.NewTaskProcessor(tt.task, store)
			result := processor.Process()
			if result != tt.expectedValue {
				t.Errorf("Expected %f, got %f", tt.expectedValue, result)
			}
		})
	}
}

func TestCreateTasks(t *testing.T) {
	store := storage.NewStorage()

	t.Run("Simple expression", func(t *testing.T) {
		expr := &models.Expression{
			Original: "2+2",
		}
		processor := handlers.NewTaskProcessor(nil, store)
		tasks, err := processor.CreateTasks(expr)

		if err != nil {
			t.Fatalf("CreateTasks() error = %v", err)
		}

		if len(tasks) != 1 {
			t.Fatalf("CreateTasks() returned %d tasks, want 1", len(tasks))
		}

		task := tasks[0]
		if task.Operation != "+" {
			t.Errorf("Expected operation +, got %s", task.Operation)
		}

		if task.Arg1 != 2 || task.Arg2 != 2 {
			t.Errorf("Expected args [2 2], got %v, %v", task.Arg1, task.Arg2)
		}
	})

	t.Run("Complex expression", func(t *testing.T) {
		expr := &models.Expression{
			Original: "2+3*4",
		}
		processor := handlers.NewTaskProcessor(nil, store)
		tasks, err := processor.CreateTasks(expr)

		if err != nil {
			t.Fatalf("CreateTasks() error = %v", err)
		}

		if len(tasks) != 2 {
			t.Fatalf("CreateTasks() returned %d tasks, want 2", len(tasks))
		}

		if tasks[0].Operation != "*" {
			t.Errorf("Expected operation *, got %s", tasks[0].Operation)
		}

		if tasks[0].Arg1 != 3 || tasks[0].Arg2 != 4 {
			t.Errorf("Expected args [3 4], got %v, %v", tasks[0].Arg1, tasks[0].Arg2)
		}
	})

	t.Run("Invalid expression", func(t *testing.T) {
		expr := &models.Expression{
			Original: "2++2",
		}
		processor := handlers.NewTaskProcessor(nil, store)
		_, err := processor.CreateTasks(expr)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if err.Error() != "mismatched numbers and operations" {
			t.Errorf("Expected error containing \"mismatched numbers and operations\", got %v", err)
		}
	})
}
