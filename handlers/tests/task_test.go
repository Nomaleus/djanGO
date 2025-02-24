package tests

import (
	"bytes"
	"djanGO/handlers"
	"djanGO/storage"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskEndpoints(t *testing.T) {
	store := storage.NewStorage()
	handler := handlers.NewHandler(store)

	t.Run("Submit task", func(t *testing.T) {
		taskReq := map[string]interface{}{
			"task": map[string]interface{}{
				"id":             "test-task",
				"arg1":           "2",
				"arg2":           "3",
				"operation":      "*",
				"operation_time": 1000,
			},
		}
		jsonBody, _ := json.Marshal(taskReq)
		req := httptest.NewRequest("POST", "/internal/task", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		handler.SubmitTaskResult(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&response)

		if response["id"] != "test-task" {
			t.Errorf("Expected task ID test-task, got %v", response["id"])
		}
		if response["result"] != float64(6) {
			t.Errorf("Expected result 6, got %v", response["result"])
		}
	})

	t.Run("Get tasks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/internal/task", nil)
		rr := httptest.NewRecorder()

		handler.GetTask(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %d", rr.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&response)

		tasks, ok := response["tasks"].([]interface{})
		if !ok {
			t.Fatal("Expected tasks array in response")
		}

		if len(tasks) == 0 {
			t.Error("Expected at least one task")
		}

		task := tasks[0].(map[string]interface{})
		expectedFields := []string{"id", "arg1", "arg2", "operation", "operation_time"}
		for _, field := range expectedFields {
			if _, ok := task[field]; !ok {
				t.Errorf("Task missing field %s", field)
			}
		}
	})
}
