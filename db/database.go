package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB *sql.DB
)

func InitDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("ошибка получения текущей директории: %w", err)
		}
		dbPath = filepath.Join(dir, "djanGO.db")
	}

	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("ошибка создания директории для базы данных: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("ошибка соединения с базой данных: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ошибка проверки соединения с базой данных: %w", err)
	}

	DB = db
	fmt.Printf("База данных успешно инициализирована: %s\n", dbPath)

	if err := createTables(); err != nil {
		return fmt.Errorf("ошибка создания таблиц: %w", err)
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Printf("Ошибка закрытия соединения с базой данных: %v", err)
		}
	}
}

func createTables() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы пользователей: %w", err)
	}

	var count int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return fmt.Errorf("ошибка проверки количества пользователей: %w", err)
	}

	if count == 0 {
		fmt.Println("Создание тестового пользователя admin...")
		if err := CreateTestUser(); err != nil {
			return fmt.Errorf("ошибка создания тестового пользователя: %w", err)
		}
		fmt.Println("Тестовый пользователь admin успешно создан")
	}

	if err := CreateExpressionTables(); err != nil {
		return fmt.Errorf("ошибка создания таблиц для выражений: %w", err)
	}

	return nil
}
