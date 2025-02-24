package handlers

import (
	"djanGO/utils"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"djanGO/models"
	"djanGO/storage"
)

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	tasks := h.Storage.GetAllTasks()
	var taskResponses []map[string]interface{}
	for _, task := range tasks {
		taskResponses = append(taskResponses, map[string]interface{}{
			"id":             task.ID,
			"arg1":           task.Arg1,
			"arg2":           task.Arg2,
			"operation":      task.Operation,
			"operation_time": utils.GetOperationTime(task.Operation),
		})
	}
	response := map[string]interface{}{
		"tasks": taskResponses,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *Handler) SubmitTaskResult(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var taskReq struct {
		Task struct {
			ID            string `json:"id"`
			Arg1          string `json:"arg1"`
			Arg2          string `json:"arg2"`
			Operation     string `json:"operation"`
			OperationTime int    `json:"operation_time"`
		} `json:"task"`
	}

	if err := json.Unmarshal(body, &taskReq); err == nil && taskReq.Task.ID != "" {
		arg1, err1 := strconv.ParseFloat(taskReq.Task.Arg1, 64)
		arg2, err2 := strconv.ParseFloat(taskReq.Task.Arg2, 64)
		if err1 != nil || err2 != nil {
			http.Error(w, "Invalid arguments", http.StatusInternalServerError)
			return
		}

		taskID := taskReq.Task.ID
		if taskID == "" {
			taskID = uuid.New().String()
		}

		task := &models.Task{
			ID:            taskID,
			Operation:     taskReq.Task.Operation,
			Arg1:          arg1,
			Arg2:          arg2,
			Status:        "PENDING",
			OperationTime: taskReq.Task.OperationTime,
		}

		if err := h.Storage.AddTask(task); err != nil {
			if err == storage.ErrTaskExists {
				http.Error(w, "Task already exists", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		resultChan := make(chan float64)

		go func() {
			processor := NewTaskProcessor(task, h.Storage)
			result := processor.Process()
			resultChan <- result
		}()

		select {
		case result := <-resultChan:
			response := map[string]interface{}{
				"id":     taskID,
				"result": result,
			}
			utils.WriteJSON(w, http.StatusOK, response)
		case <-r.Context().Done():
			http.Error(w, "Client disconnected", http.StatusInternalServerError)
		}
		return
	}

	var request struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Invalid data", http.StatusInternalServerError)
		return
	}

	err = h.Storage.UpdateTaskResult(request.ID, request.Result)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":     request.ID,
		"result": request.Result,
	}
	utils.WriteJSON(w, http.StatusOK, response)
}
