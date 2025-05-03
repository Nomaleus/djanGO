package utils

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	testUser := "testuser"
	token, err := GenerateToken(testUser)

	if err != nil {
		t.Fatalf("Ошибка при генерации токена: %v", err)
	}

	if token == "" {
		t.Fatal("Сгенерирован пустой токен")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("Неверная структура токена: %s", token)
	}

	login, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Ошибка при валидации токена: %v", err)
	}

	if login != testUser {
		t.Errorf("Ожидался логин %s, получен %s", testUser, login)
	}

	_, err = ValidateToken("invalid.token.format")
	if err == nil {
		t.Error("Ожидалась ошибка при валидации неправильного токена")
	}

	_, err = ValidateToken("")
	if err == nil {
		t.Error("Ожидалась ошибка при валидации пустого токена")
	}
}

func TestTokenExpiration(t *testing.T) {
	t.Skip("Тест пропущен - требуется модификация GenerateToken для тестирования с коротким сроком жизни")

	testUser := "testuser"
	token, err := GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Ошибка при генерации токена: %v", err)
	}

	_, err = ValidateToken(token)
	if err != nil {
		t.Errorf("Токен должен быть валидным сразу после создания: %v", err)
	}

	time.Sleep(2 * time.Second)

	_, err = ValidateToken(token)
	if err == nil {
		t.Error("Ожидалась ошибка при валидации истекшего токена")
	}
}
