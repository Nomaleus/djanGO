package db

import (
	"database/sql"
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
	fmt.Printf("Добавление выражения '%s' для пользователя: %s\n", expression, userLogin)

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

	fmt.Printf("Найден пользователь '%s' с ID=%d\n", userLogin, userID)

	expressionID := generateUUID()
	fmt.Printf("Сгенерирован ID выражения: %s\n", expressionID)

	_, err = tx.Exec(
		"INSERT INTO expressions (id, user_id, original, status, created) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)",
		expressionID, userID, expression, "PENDING",
	)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления выражения: %w", err)
	}

	fmt.Printf("Выражение добавлено в базу данных: id=%s, user_id=%d, expression=%s\n",
		expressionID, userID, expression)

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("ошибка завершения транзакции: %w", err)
	}

	var lastID int64
	err = DB.QueryRow("SELECT MAX(ROWID) FROM expressions").Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID выражения: %w", err)
	}

	fmt.Printf("Возвращаем ROWID=%d для выражения с id=%s\n", lastID, expressionID)

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
		depends_on, operation_time
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
		var dependenciesStr sql.NullString
		err := rows.Scan(
			&taskIDStr, &exprIDStr, &task.Order, &task.Operation, &task.Arg1, &task.Arg2,
			&task.Result, &task.Status, &task.Error, &dependenciesStr, &task.OperationTime,
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
	fmt.Printf("Запрос выражений для пользователя: %s\n", userLogin)

	var userID int64
	err := DB.QueryRow("SELECT id FROM users WHERE login = ?", userLogin).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("Пользователь '%s' не найден в базе данных, возвращаем пустой список\n", userLogin)
			return []*Expression{}, nil
		}
		fmt.Printf("Ошибка получения ID пользователя '%s': %v\n", userLogin, err)
		return nil, fmt.Errorf("ошибка получения ID пользователя: %w", err)
	}

	fmt.Printf("Найден пользователь '%s' с ID=%d\n", userLogin, userID)

	query := `SELECT id, original as text, status, result, error, ? as user_login, created as created_at, created as updated_at 
		FROM expressions WHERE user_id = ?
		ORDER BY created DESC`
	fmt.Printf("SQL запрос: %s\nПараметры: userLogin=%s, userID=%d\n", query, userLogin, userID)

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

		fmt.Printf("Найдено выражение: ID=%v, Текст=%s, Статус=%s\n", expr.ID, expr.Text, expr.Status)
		expressions = append(expressions, &expr)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Ошибка после сканирования строк: %v\n", err)
		return nil, fmt.Errorf("ошибка после сканирования выражений: %w", err)
	}

	fmt.Printf("Всего найдено выражений: %d\n", len(expressions))
	return expressions, nil
}

func CreateExpressionTables() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL DEFAULT 0,
			error TEXT DEFAULT '',
			user_login TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы выражений: %w", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expression_id INTEGER NOT NULL,
			operation TEXT NOT NULL,
			arg1 REAL NOT NULL,
			arg2 REAL NOT NULL,
			result REAL DEFAULT 0,
			status TEXT NOT NULL,
			error TEXT DEFAULT '',
			task_order INTEGER NOT NULL,
			depends_on TEXT DEFAULT '',
			operation_time INTEGER DEFAULT 100,
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
