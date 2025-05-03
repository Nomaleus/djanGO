package storage

import (
	"fmt"
	"log"
	"sync"
)

const (
	TaskStatusPending    = "PENDING"
	TaskStatusInProgress = "PROCESSING"
	TaskStatusCompleted  = "COMPLETED"
	TaskStatusError      = "ERROR"

	ExpressionStatusCompleted = "COMPLETED"
	ExpressionStatusError     = "ERROR"
)

type Expression struct {
	ID       string
	Original string
	Status   string
	Result   float64
	Error    string
	Tasks    []*Task
	Created  string
}

type Task struct {
	ID            string
	ExpressionID  string
	Operation     string
	Arg1          float64
	Arg2          float64
	Result        float64
	Status        string
	Error         string
	Order         int
	DependsOn     []string
	Arg1Source    string
	Arg2Source    string
	OperationTime int
}

type Storage struct {
	mu sync.RWMutex

	expressions map[string]*Expression

	tasks map[string]*Task
}

func NewStorage() *Storage {
	return &Storage{
		expressions: make(map[string]*Expression),
		tasks:       make(map[string]*Task),
	}
}

func (s *Storage) AddExpression(expr *Expression) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expressions[expr.ID] = expr
}

func (s *Storage) GetExpression(id string) (*Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expr, ok := s.expressions[id]
	if !ok {
		return nil, fmt.Errorf("expression not found")
	}
	return expr, nil
}

func (s *Storage) GetAllExpressions() ([]*Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Expression, 0, len(s.expressions))
	for _, expr := range s.expressions {
		result = append(result, expr)
	}
	return result, nil
}

func (s *Storage) GetAndLockPendingTask(workerID string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, task := range s.tasks {
		if task.Status == TaskStatusPending {
			allDepsCompleted := true
			for _, depID := range task.DependsOn {
				depTask, exists := s.tasks[depID]
				if !exists || depTask.Status != TaskStatusCompleted {
					allDepsCompleted = false
					break
				}
			}

			if allDepsCompleted {
				task.Status = TaskStatusInProgress
				log.Printf("Worker %s: Получена задача ID=%s для выполнения", workerID, id)
				return task, nil
			}
		}
	}

	return nil, nil
}

func (s *Storage) GetPendingTask() (*Task, error) {
	return s.GetAndLockPendingTask("local")
}

func (s *Storage) UpdateTaskResult(taskID string, result float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found")
	}

	task.Result = result
	task.Status = TaskStatusCompleted

	expr, exists := s.expressions[task.ExpressionID]
	if !exists {
		return fmt.Errorf("выражение для задачи не найдено")
	}

	allCompleted := true
	for _, t := range expr.Tasks {
		if t.Status != TaskStatusCompleted {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		var lastTask *Task
		maxOrder := -1

		for _, t := range expr.Tasks {
			if t.Order > maxOrder {
				maxOrder = t.Order
				lastTask = t
			}
		}

		if lastTask != nil {
			expr.Status = ExpressionStatusCompleted
			expr.Result = lastTask.Result
		}
	}

	return nil
}

func (s *Storage) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; exists {
		return fmt.Errorf("task already exists")
	}

	s.tasks[task.ID] = task

	if expr, exists := s.expressions[task.ExpressionID]; exists {
		expr.Tasks = append(expr.Tasks, task)
	}

	return nil
}

func (s *Storage) GetAllTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		result = append(result, task)
	}
	return result
}

func (s *Storage) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}
	return task, nil
}

func (s *Storage) UpdateTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; !exists {
		return fmt.Errorf("task not found")
	}

	s.tasks[task.ID] = task
	return nil
}

func (s *Storage) UpdateExpression(expr *Expression) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.expressions[expr.ID]; !exists {
		return fmt.Errorf("expression not found")
	}

	s.expressions[expr.ID] = expr
	return nil
}

func (s *Storage) UpdateTaskError(taskID string, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found")
	}

	task.Status = TaskStatusError
	task.Error = errorMsg

	expr, exists := s.expressions[task.ExpressionID]
	if exists {
		expr.Status = ExpressionStatusError
		expr.Error = errorMsg
	}

	return nil
}

func (s *Storage) GetNextID() string {
	uid := make([]byte, 16)
	for i := range uid {
		uid[i] = byte(i)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", uid[0:4], uid[4:6], uid[6:8], uid[8:10], uid[10:])
}
