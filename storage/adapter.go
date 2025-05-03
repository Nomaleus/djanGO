package storage

import (
	"djanGO/models"
	"time"
)

func ConvertModelToStorageTask(modelTask *models.Task) *Task {
	if modelTask == nil {
		return nil
	}

	return &Task{
		ID:            modelTask.ID,
		ExpressionID:  modelTask.ExpressionID,
		Operation:     modelTask.Operation,
		Arg1:          modelTask.Arg1,
		Arg2:          modelTask.Arg2,
		Result:        modelTask.Result,
		Status:        modelTask.Status,
		Error:         modelTask.Error,
		Order:         modelTask.Order,
		DependsOn:     modelTask.DependsOn,
		Arg1Source:    modelTask.Arg1Source,
		Arg2Source:    modelTask.Arg2Source,
		OperationTime: modelTask.OperationTime,
	}
}

func ConvertStorageToModelTask(storageTask *Task) *models.Task {
	if storageTask == nil {
		return nil
	}

	return &models.Task{
		ID:            storageTask.ID,
		ExpressionID:  storageTask.ExpressionID,
		Operation:     storageTask.Operation,
		Arg1:          storageTask.Arg1,
		Arg2:          storageTask.Arg2,
		Result:        storageTask.Result,
		Status:        storageTask.Status,
		Error:         storageTask.Error,
		Order:         storageTask.Order,
		DependsOn:     storageTask.DependsOn,
		Arg1Source:    storageTask.Arg1Source,
		Arg2Source:    storageTask.Arg2Source,
		OperationTime: storageTask.OperationTime,
	}
}

func ConvertModelToStorageExpression(modelExpr *models.Expression) *Expression {
	if modelExpr == nil {
		return nil
	}

	var tasks []*Task
	for _, task := range modelExpr.Tasks {
		tasks = append(tasks, ConvertModelToStorageTask(task))
	}

	created := modelExpr.Created.Format(time.RFC3339)

	return &Expression{
		ID:       modelExpr.ID,
		Original: modelExpr.Original,
		Status:   modelExpr.Status,
		Result:   modelExpr.Result,
		Error:    modelExpr.Error,
		Tasks:    tasks,
		Created:  created,
	}
}

func ConvertStorageToModelExpression(storageExpr *Expression) *models.Expression {
	if storageExpr == nil {
		return nil
	}

	var tasks []*models.Task
	for _, task := range storageExpr.Tasks {
		tasks = append(tasks, ConvertStorageToModelTask(task))
	}

	created, err := time.Parse(time.RFC3339, storageExpr.Created)
	if err != nil {
		created = time.Now()
	}

	return &models.Expression{
		ID:       storageExpr.ID,
		Original: storageExpr.Original,
		Status:   storageExpr.Status,
		Result:   storageExpr.Result,
		Error:    storageExpr.Error,
		Tasks:    tasks,
		Created:  created,
	}
}

func ConvertStorageTasksToModelTasks(storageTasks []*Task) []*models.Task {
	if storageTasks == nil {
		return nil
	}

	var tasks []*models.Task
	for _, task := range storageTasks {
		tasks = append(tasks, ConvertStorageToModelTask(task))
	}
	return tasks
}

func ConvertStorageExpressionsToModelExpressions(storageExprs []*Expression) []*models.Expression {
	if storageExprs == nil {
		return nil
	}

	var exprs []*models.Expression
	for _, expr := range storageExprs {
		exprs = append(exprs, ConvertStorageToModelExpression(expr))
	}
	return exprs
}
