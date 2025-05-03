package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

var publicPaths = []string{
	"/api/v1/login",
	"/api/v1/register",
	"/api/v1/logout",
	"/api/v1/token",
	"/static/",
	"/login",
	"/register",
}

var allowedPaths = []string{
	"/api/v1/expressions",
	"/api/v1/calculate",
	"/api/v1/history",
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-Login, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		path := r.URL.Path

		userLogin := getUserFromRequest(r)

		if userLogin != "" {
			ctx := AddUserToContext(r.Context(), userLogin)
			r = r.WithContext(ctx)
		}

		for _, publicPath := range publicPaths {
			if strings.HasPrefix(path, publicPath) {
				next.ServeHTTP(w, r)
				return
			}
		}

		for _, allowedPath := range allowedPaths {
			if strings.HasPrefix(path, allowedPath) {
				if userLogin != "" {
					ctx := AddUserToContext(r.Context(), userLogin)
					r = r.WithContext(ctx)
					next.ServeHTTP(w, r)
					return
				}
			}
		}

		if userLogin == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Требуется авторизация",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getUserFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		login, err := ValidateToken(tokenString)
		if err == nil && login != "" {
			return login
		}
	}

	userLogin := r.Header.Get("X-User-Login")
	if userLogin != "" {
		return userLogin
	}

	cookie, err := r.Cookie("user_login")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}
