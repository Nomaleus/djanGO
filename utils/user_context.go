package utils

import (
	"context"
	"fmt"
)

type contextKey string

const userCtxKey = contextKey("user")

func AddUserToContext(ctx context.Context, userLogin string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	fmt.Printf("DEBUG AddUserToContext: Добавлен пользователь '%s' в контекст\n", userLogin)
	return context.WithValue(ctx, userCtxKey, userLogin)
}

func GetUserFromContext(ctx context.Context) string {
	if ctx == nil {
		fmt.Println("DEBUG GetUserFromContext: Контекст nil")
		return ""
	}

	userLogin, ok := ctx.Value(userCtxKey).(string)
	if !ok {
		fmt.Println("DEBUG GetUserFromContext: Пользователь в контексте не найден")
		return ""
	}

	fmt.Printf("DEBUG GetUserFromContext: Найден пользователь '%s' в контексте\n", userLogin)
	return userLogin
}
