package storage

import (
	"djanGO/models"
	"fmt"
	"sync"
)

type Storage struct {
	expressions map[string]*models.Expression
	tasks       map[string]*models.Task
	lastID      int
	mu          sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		expressions: make(map[string]*models.Expression),
		tasks:       make(map[string]*models.Task),
		lastID:      0,
		mu:          sync.RWMutex{},
	}
}

func (s *Storage) GetNextID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastID++
	return fmt.Sprintf("%d", s.lastID)
}

func (s *Storage) AddExpression(expr *models.Expression) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expressions[expr.ID] = expr
}

func (s *Storage) GetExpression(id string) (*models.Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expr, exists := s.expressions[id]
	if !exists {
		return nil, ErrNotFound
	}
	return expr, nil
}

func (s *Storage) GetAllExpressions() ([]*models.Expression, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expressions := make([]*models.Expression, 0, len(s.expressions))
	for _, expr := range s.expressions {
		expressions = append(expressions, expr)
	}
	return expressions, nil
}

func (s *Storage) GetNextPendingTask() *models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if task.Status == "PENDING" {
			return task
		}
	}
	return nil
}

func (s *Storage) GetPendingTask() (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, task := range s.tasks {
		if task.Status == "PENDING" {
			task.Status = "PROCESSING"
			return task, nil
		}
	}
	return nil, ErrNoTasks
}

func (s *Storage) UpdateTaskResult(taskID string, result float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return ErrNotFound
	}

	task.Result = result
	task.Status = "COMPLETED"

	for _, expr := range s.expressions {
		for _, t := range expr.Tasks {
			if t.ID == taskID {
				allCompleted := true

				for _, task := range expr.Tasks {
					if task.Status != "COMPLETED" {
						allCompleted = false
						break
					}
				}

				if allCompleted {
					expr.Status = "COMPLETED"
					expr.Result = expr.Tasks[len(expr.Tasks)-1].Result
				}
				break
			}
		}
	}

	return nil
}

func (s *Storage) AddTask(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[task.ID]; exists {
		return ErrTaskExists
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *Storage) GetAllTasks() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *Storage) GetTask(id string) (*models.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return task, nil
}

func (s *Storage) UpdateTask(task *models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.tasks[task.ID]
	if !ok {
		return fmt.Errorf("task %s not found", task.ID)
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *Storage) UpdateExpression(expr *models.Expression) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.expressions[expr.ID]
	if !ok {
		return fmt.Errorf("expression %s not found", expr.ID)
	}
	s.expressions[expr.ID] = expr
	return nil
}
