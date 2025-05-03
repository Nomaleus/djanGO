package storage

import (
	"djanGO/db"
	"djanGO/models"
	"fmt"
)

type StorageWrapper struct {
	storage *Storage
}

func NewStorageWrapper(storage *Storage) *StorageWrapper {
	return &StorageWrapper{
		storage: storage,
	}
}

func (w *StorageWrapper) AddExpression(expr *models.Expression) {
	storageExpr := ConvertModelToStorageExpression(expr)
	w.storage.AddExpression(storageExpr)
}

func (w *StorageWrapper) GetExpression(id string) (*models.Expression, error) {
	storageExpr, err := w.storage.GetExpression(id)
	if err != nil {
		return nil, err
	}
	return ConvertStorageToModelExpression(storageExpr), nil
}

func (w *StorageWrapper) GetAllExpressions() ([]*models.Expression, error) {
	storageExprs, err := w.storage.GetAllExpressions()
	if err != nil {
		return nil, err
	}
	return ConvertStorageExpressionsToModelExpressions(storageExprs), nil
}

func (w *StorageWrapper) GetAndLockPendingTask(workerID string) (*models.Task, error) {
	storageTask, err := w.storage.GetAndLockPendingTask(workerID)
	if err != nil {
		return nil, err
	}
	if storageTask == nil {
		return nil, nil
	}
	return ConvertStorageToModelTask(storageTask), nil
}

func (w *StorageWrapper) GetPendingTask() (*models.Task, error) {
	storageTask, err := w.storage.GetPendingTask()
	if err != nil {
		return nil, err
	}
	if storageTask == nil {
		return nil, nil
	}
	return ConvertStorageToModelTask(storageTask), nil
}

func (w *StorageWrapper) UpdateTaskResult(taskID string, result float64) error {
	return w.storage.UpdateTaskResult(taskID, result)
}

func (w *StorageWrapper) AddTask(task *models.Task) error {
	storageTask := ConvertModelToStorageTask(task)
	return w.storage.AddTask(storageTask)
}

func (w *StorageWrapper) GetAllTasks() []*models.Task {
	storageTasks := w.storage.GetAllTasks()
	return ConvertStorageTasksToModelTasks(storageTasks)
}

func (w *StorageWrapper) GetTask(id string) (*models.Task, error) {
	storageTask, err := w.storage.GetTask(id)
	if err != nil {
		return nil, err
	}
	return ConvertStorageToModelTask(storageTask), nil
}

func (w *StorageWrapper) UpdateTask(task *models.Task) error {
	storageTask := ConvertModelToStorageTask(task)
	return w.storage.UpdateTask(storageTask)
}

func (w *StorageWrapper) UpdateExpression(expr *models.Expression) error {
	storageExpr := ConvertModelToStorageExpression(expr)
	return w.storage.UpdateExpression(storageExpr)
}

func (w *StorageWrapper) UpdateTaskError(taskID string, errorMsg string) error {
	return w.storage.UpdateTaskError(taskID, errorMsg)
}

func (w *StorageWrapper) GetExpressionsByUser(userLogin string) ([]*models.Expression, error) {
	dbExpressions, err := db.GetExpressionsByUser(userLogin)
	if err != nil {
		return nil, err
	}

	var expressions []*models.Expression
	for _, dbExpr := range dbExpressions {
		dbTasks, err := db.GetTasksForExpression(dbExpr.ID)
		if err != nil {
			return nil, err
		}

		var tasks []*models.Task
		for _, dbTask := range dbTasks {
			task := &models.Task{
				ID:            dbTask.ID,
				ExpressionID:  dbExpr.ID,
				Operation:     dbTask.Operation,
				Arg1:          dbTask.Arg1,
				Arg2:          dbTask.Arg2,
				Status:        dbTask.Status,
				Result:        dbTask.Result,
				Error:         dbTask.Error,
				Order:         dbTask.Order,
				OperationTime: dbTask.OperationTime,
			}
			tasks = append(tasks, task)
		}

		expr := &models.Expression{
			ID:       dbExpr.ID,
			Original: dbExpr.Text,
			Status:   dbExpr.Status,
			Result:   dbExpr.Result,
			Error:    dbExpr.Error,
			Tasks:    tasks,
			Created:  dbExpr.CreatedAt,
		}

		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (w *StorageWrapper) GetExpressionByTaskID(taskID string) (*models.Expression, error) {
	storageTask, err := w.storage.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении задачи: %w", err)
	}

	storageExpr, err := w.storage.GetExpression(storageTask.ExpressionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении выражения: %w", err)
	}

	return ConvertStorageToModelExpression(storageExpr), nil
}
