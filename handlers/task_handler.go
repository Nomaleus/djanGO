package handlers

import (
	"djanGO/utils"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

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

func (h *Handler) ProcessExpression(w http.ResponseWriter, r *http.Request) {
	var expResponse models.ExpressionResponse
	var request models.ExpressionWrapper

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	log.Printf("Получен запрос на обработку выражения: %s", string(body))

	err = json.Unmarshal(body, &request)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	if request.Expression == "" {
		http.Error(w, "Expression cannot be empty", http.StatusBadRequest)
		return
	}

	if !utils.IsValidExpression(request.Expression) {
		log.Printf("Недопустимое выражение: %s", request.Expression)
		http.Error(w, "Invalid expression", http.StatusBadRequest)
		return
	}

	expr := &models.Expression{
		ID:       uuid.New().String(),
		Original: request.Expression,
		Status:   "PENDING",
		Created:  time.Now(),
	}

	h.Storage.AddExpression(expr)

	log.Printf("Создано новое выражение с ID: %s, выражение: %s", expr.ID, expr.Original)

	dummyTask := &models.Task{}
	processor := NewTaskProcessor(dummyTask, h.Storage)
	tasks, err := processor.CreateTasks(expr)
	if err != nil {
		log.Printf("Ошибка при создании задач для выражения %s: %v", expr.ID, err)
		expr.Status = "ERROR"
		expr.Error = err.Error()
		h.Storage.UpdateExpression(expr)

		expResponse = models.ExpressionResponse{
			ID:      expr.ID,
			Status:  "ERROR",
			Error:   err.Error(),
			Tasks:   nil,
			Created: expr.Created,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expResponse)
		return
	}

	log.Printf("Создано %d задач для выражения %s", len(tasks), expr.ID)

	if expr.Status == "COMPLETED" {
		log.Printf("Выражение %s уже завершено, результат: %f", expr.ID, expr.Result)

		expResponse = models.ExpressionResponse{
			ID:      expr.ID,
			Status:  "COMPLETED",
			Result:  &expr.Result,
			Tasks:   tasks,
			Created: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expResponse)
		return
	}

	expResponse = models.ExpressionResponse{
		ID:      expr.ID,
		Status:  "PENDING",
		Tasks:   tasks,
		Created: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(expResponse)
}

func (h *Handler) GetExpression(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "Expression ID is required", http.StatusBadRequest)
		return
	}

	expr, err := h.Storage.GetExpression(id)
	if err != nil {
		log.Printf("Выражение с ID %s не найдено", id)
		http.Error(w, "Expression not found", http.StatusNotFound)
		return
	}

	tasks := make([]*models.Task, 0)
	if expr.Tasks != nil {
		tasks = expr.Tasks
	} else {
		allTasks := h.Storage.GetAllTasks()
		for _, task := range allTasks {
			if task.ExpressionID == id {
				tasks = append(tasks, task)
			}
		}
	}

	log.Printf("Запрошено выражение %s, статус: %s", id, expr.Status)

	var response models.ExpressionResponse
	response.ID = expr.ID
	response.Status = expr.Status
	response.Tasks = tasks
	response.Created = expr.Created

	if expr.Status == "COMPLETED" {
		response.Result = &expr.Result
		log.Printf("Возвращается результат для выражения %s: %f", id, expr.Result)
	} else if expr.Status == "ERROR" {
		response.Error = expr.Error
		log.Printf("Возвращается ошибка для выражения %s: %s", id, expr.Error)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
