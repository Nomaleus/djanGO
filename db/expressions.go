package db

import (
	"database/sql"
	"djanGO/models"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Expression struct {
	ID        string
	Text      string
	Status    string
	Result    float64
	Error     string
	UserLogin string
	CreatedAt time.Time
	UpdatedAt time.Time
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
	Dependencies  []string
	OperationTime int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func AddExpression(expression string, userLogin string) (int64, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var userID int64
	err = tx.QueryRow("SELECT id FROM users WHERE login = ?", userLogin).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("Пользователь '%s' не найден, использование ID=1 невозможно\n", userLogin)
			return 0, fmt.Errorf("пользователь '%s' не найден: %w", userLogin, err)
		} else {
			return 0, fmt.Errorf("ошибка получения ID пользователя: %w", err)
		}
	}

	expressionID := generateUUID()

	_, err = tx.Exec(
		"INSERT INTO expressions (id, user_id, original, status) VALUES (?, ?, ?, ?)",
		expressionID, userID, expression, "PENDING",
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления выражения: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("ошибка завершения транзакции: %w", err)
	}

	var lastID int64
	err = DB.QueryRow("SELECT rowid FROM expressions WHERE id = ?", expressionID).Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ROWID выражения: %w", err)
	}

	return lastID, nil
}

func generateUUID() string {
	timestamp := time.Now().UnixNano()
	randomPart := fmt.Sprintf("%d%d", timestamp, rand.Int63())

	return fmt.Sprintf("%s-%s-%s-%s-%s",
		randomPart[0:8],
		randomPart[8:12],
		randomPart[12:16],
		randomPart[16:20],
		randomPart[20:32])
}

func AddTask(task *Task) (int64, error) {
	var exprIDStr = task.ExpressionID

	var userID int64
	err := DB.QueryRow("SELECT user_id FROM expressions WHERE id = ?", exprIDStr).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения user_id для выражения: %w", err)
	}

	var dependenciesStr string
	if len(task.Dependencies) > 0 {
		for i, dep := range task.Dependencies {
			if i > 0 {
				dependenciesStr += ","
			}
			dependenciesStr += dep
		}
	}

	taskID := generateUUID()

	_, err = DB.Exec(
		`INSERT INTO tasks 
		(id, expression_id, user_id, task_order, operation, arg1, arg2, status, result, 
		operation_time, error, depends_on, arg1_source, arg2_source) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		taskID, exprIDStr, userID, task.Order, task.Operation,
		task.Arg1, task.Arg2, "PENDING", 0,
		task.OperationTime, "", dependenciesStr, "", "",
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления задачи: %w", err)
	}

	var lastID int64
	err = DB.QueryRow("SELECT MAX(ROWID) FROM tasks").Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID задачи: %w", err)
	}

	return lastID, nil
}

func GetTasksForExpression(expressionID string) ([]*Task, error) {
	rows, err := DB.Query(
		`SELECT id, expression_id, task_order, operation, arg1, arg2, result, status, error, 
		depends_on, operation_time, arg1_source, arg2_source
		FROM tasks WHERE expression_id = ?
		ORDER BY task_order ASC`, expressionID,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения задач для выражения: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var taskIDStr, exprIDStr string
		var dependenciesStr, arg1Source, arg2Source sql.NullString
		err := rows.Scan(
			&taskIDStr, &exprIDStr, &task.Order, &task.Operation, &task.Arg1, &task.Arg2,
			&task.Result, &task.Status, &task.Error, &dependenciesStr, &task.OperationTime,
			&arg1Source, &arg2Source,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования задачи: %w", err)
		}

		task.ID = taskIDStr
		task.ExpressionID = exprIDStr

		if dependenciesStr.Valid && dependenciesStr.String != "" {
			depIDs := strings.Split(dependenciesStr.String, ",")
			task.Dependencies = make([]string, len(depIDs))
			copy(task.Dependencies, depIDs)
		}

		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после сканирования задач: %w", err)
	}

	return tasks, nil
}

func GetExpressionsByUser(userLogin string) ([]*Expression, error) {
	var userID int64
	err := DB.QueryRow("SELECT id FROM users WHERE login = ?", userLogin).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*Expression{}, nil
		}
		return nil, fmt.Errorf("ошибка получения ID пользователя: %w", err)
	}

	query := `SELECT id, original, status, result, error, ? as user_login, created, created 
		FROM expressions WHERE user_id = ?
		ORDER BY created DESC`

	rows, err := DB.Query(query, userLogin, userID)
	if err != nil {
		fmt.Printf("Ошибка SQL запроса: %v\n", err)
		return nil, fmt.Errorf("ошибка получения выражений пользователя: %w", err)
	}
	defer rows.Close()

	var expressions []*Expression
	for rows.Next() {
		var expr Expression
		var resultNull sql.NullFloat64
		var errorNull sql.NullString

		err := rows.Scan(
			&expr.ID,
			&expr.Text,
			&expr.Status,
			&resultNull,
			&errorNull,
			&expr.UserLogin,
			&expr.CreatedAt,
			&expr.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Ошибка сканирования строки результата: %v\n", err)
			return nil, fmt.Errorf("ошибка сканирования выражения: %w", err)
		}

		if resultNull.Valid {
			expr.Result = resultNull.Float64
		}

		if errorNull.Valid {
			expr.Error = errorNull.String
		}

		expressions = append(expressions, &expr)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Ошибка после сканирования строк: %v\n", err)
		return nil, fmt.Errorf("ошибка после сканирования выражений: %w", err)
	}

	return expressions, nil
}

func CreateExpressionTables() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			original TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL DEFAULT 0,
			error TEXT DEFAULT '',
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы выражений: %w", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			expression_id TEXT NOT NULL,
			user_id INTEGER NOT NULL,
			operation TEXT NOT NULL,
			arg1 REAL NOT NULL,
			arg2 REAL NOT NULL,
			result REAL DEFAULT 0,
			status TEXT NOT NULL,
			error TEXT DEFAULT '',
			task_order INTEGER NOT NULL,
			depends_on TEXT DEFAULT '',
			operation_time INTEGER DEFAULT 100,
			arg1_source TEXT DEFAULT '',
			arg2_source TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (expression_id) REFERENCES expressions (id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы задач: %w", err)
	}

	return nil
}

func GetTasksByExpressionID(expressionID string) ([]*models.Task, error) {
	dbTasks, err := GetTasksForExpression(expressionID)
	if err != nil {
		return nil, err
	}

	var tasks []*models.Task
	for _, dbTask := range dbTasks {
		task := &models.Task{
			ID:            dbTask.ID,
			ExpressionID:  dbTask.ExpressionID,
			Operation:     dbTask.Operation,
			Arg1:          dbTask.Arg1,
			Arg2:          dbTask.Arg2,
			Result:        dbTask.Result,
			Status:        dbTask.Status,
			Error:         dbTask.Error,
			Order:         dbTask.Order,
			OperationTime: dbTask.OperationTime,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
