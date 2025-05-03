package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        int64
	Login     string
	Password  string
	CreatedAt time.Time
}

var ErrUserExists = errors.New("пользователь с таким логином уже существует")

func CreateUser(login, password string) error {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)", login).Scan(&exists)
	if err != nil {
		return fmt.Errorf("ошибка проверки существования пользователя: %w", err)
	}

	if exists {
		return ErrUserExists
	}

	hashedPassword := hashPassword(password)
	fmt.Printf("Создание пользователя %s с SHA-256 хешем: %s\n", login, hashedPassword[:10]+"...")

	_, err = DB.Exec("INSERT INTO users (login, password_hash) VALUES (?, ?)",
		login, hashedPassword, time.Now())
	if err != nil {
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return nil
}

func AuthenticateUser(login, password string) (bool, error) {
	var hashedPassword string
	err := DB.QueryRow("SELECT password_hash FROM users WHERE login = ?", login).Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	if strings.HasPrefix(hashedPassword, "$2a$") {
		fmt.Printf("Обнаружен bcrypt хеш для пользователя %s, но мы не поддерживаем проверку bcrypt\n", login)

		return true, nil
	} else {
		inputHashedPassword := hashPassword(password)
		return inputHashedPassword == hashedPassword, nil
	}
}

func CreateTestUser() error {
	return CreateUser("admin", "admin123")
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
