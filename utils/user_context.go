package utils

import (
	"context"
)

type contextKey string

const userCtxKey = contextKey("user")

func AddUserToContext(ctx context.Context, userLogin string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, userCtxKey, userLogin)
}

func GetUserFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	userLogin, ok := ctx.Value(userCtxKey).(string)
	if !ok {
		return ""
	}

	return userLogin
}
