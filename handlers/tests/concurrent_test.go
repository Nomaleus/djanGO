package tests

import (
	"djanGO/handlers"
	"djanGO/models"
	"djanGO/storage"
	"sync"
	"testing"
	"time"
)

func TestConcurrentTaskProcessing(t *testing.T) {
	store := storage.NewStorage()

	tasks := []*models.Task{
		{
			ID:            "task1",
			Operation:     "+",
			Arg1:          2,
			Arg2:          3,
			OperationTime: 100,
		},
		{
			ID:            "task2",
			Operation:     "*",
			Arg1:          4,
			Arg2:          5,
			OperationTime: 200,
		},
		{
			ID:            "task3",
			Operation:     "/",
			Arg1:          10,
			Arg2:          2,
			OperationTime: 150,
		},
	}

	for _, task := range tasks {
		store.AddTask(task)
	}

	var wg sync.WaitGroup
	results := make(map[string]float64)
	var mu sync.Mutex

	for _, task := range tasks {
		wg.Add(1)
		go func(t *models.Task) {
			defer wg.Done()
			processor := handlers.NewTaskProcessor(t, store)
			result := processor.Process()
			mu.Lock()
			results[t.ID] = result
			mu.Unlock()
		}(task)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		expectedResults := map[string]float64{
			"task1": 5,
			"task2": 20,
			"task3": 5,
		}

		for id, expected := range expectedResults {
			mu.Lock()
			if result, ok := results[id]; !ok || result != expected {
				t.Errorf("Task %s: expected result %f, got %f", id, expected, result)
			}
			mu.Unlock()
		}

	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestConcurrentExpressionProcessing(t *testing.T) {
	store := storage.NewStorage()

	expressions := []string{
		"2+3*4",
		"10/2+5",
		"(3+4)*2",
	}

	var wg sync.WaitGroup
	for _, expr := range expressions {
		wg.Add(1)
		go func(expression string) {
			defer wg.Done()
			expr := &models.Expression{
				ID:       "test-" + expression,
				Original: expression,
				Status:   "PENDING",
			}

			processor := handlers.NewTaskProcessor(nil, store)
			tasks, err := processor.CreateTasks(expr)
			if err != nil {
				t.Errorf("Failed to create tasks for expression %s: %v", expression, err)
				return
			}

			var taskWg sync.WaitGroup
			for _, task := range tasks {
				taskWg.Add(1)
				go func(t *models.Task) {
					defer taskWg.Done()
					taskProcessor := handlers.NewTaskProcessor(t, store)
					result := taskProcessor.Process()
					store.UpdateTaskResult(t.ID, result)
				}(task)
			}
			taskWg.Wait()
		}(expr)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		tasks := store.GetAllTasks()
		if len(tasks) == 0 {
			t.Error("No tasks were created")
		}
		for _, task := range tasks {
			if task.Status != "COMPLETED" {
				t.Errorf("Task %s not completed", task.ID)
			}
		}

	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out")
	}
}
