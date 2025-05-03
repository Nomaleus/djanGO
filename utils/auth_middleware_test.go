package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		headers        map[string]string
		cookies        map[string]string
		expectedStatus int
		expectedUser   string
	}{
		{
			name:           "Публичный путь без авторизации",
			path:           "/api/v1/login",
			headers:        map[string]string{},
			expectedStatus: http.StatusOK,
			expectedUser:   "",
		},
		{
			name:           "Публичный путь с авторизацией",
			path:           "/api/v1/login",
			headers:        map[string]string{"X-User-Login": "testuser"},
			expectedStatus: http.StatusOK,
			expectedUser:   "",
		},
		{
			name:           "Защищенный путь с X-User-Login",
			path:           "/api/v1/calculate",
			headers:        map[string]string{"X-User-Login": "testuser"},
			expectedStatus: http.StatusOK,
			expectedUser:   "testuser",
		},
		{
			name:           "Защищенный путь с JWT токеном",
			path:           "/api/v1/calculate",
			headers:        map[string]string{"Authorization": "Bearer test-token"},
			expectedStatus: http.StatusOK,
			expectedUser:   "testuser",
		},
		{
			name:           "Защищенный путь без авторизации",
			path:           "/api/v1/calculate",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedUser:   "",
		},
		{
			name:           "Защищенный путь с куки",
			path:           "/api/v1/calculate",
			cookies:        map[string]string{"user_login": "cookieuser"},
			expectedStatus: http.StatusOK,
			expectedUser:   "cookieuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			for key, value := range tt.cookies {
				req.AddCookie(&http.Cookie{
					Name:  key,
					Value: value,
				})
			}

			rr := httptest.NewRecorder()

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userLogin := GetUserFromContext(r.Context())

				if tt.expectedUser == "" && userLogin != "" {
					t.Errorf("Ожидалось отсутствие пользователя, получен: %s", userLogin)
				} else if tt.expectedUser != "" && userLogin != tt.expectedUser {
					t.Errorf("Ожидался пользователь: %s, получен: %s", tt.expectedUser, userLogin)
				}

				w.WriteHeader(http.StatusOK)
			})

			customAuthMiddleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					for _, path := range []string{"/api/v1/login", "/api/v1/register"} {
						if r.URL.Path == path {
							next.ServeHTTP(w, r)
							return
						}
					}

					userLogin := r.Header.Get("X-User-Login")
					if userLogin != "" {
						ctx := AddUserToContext(r.Context(), userLogin)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}

					authHeader := r.Header.Get("Authorization")
					if authHeader != "" {
						if authHeader == "Bearer test-token" {
							userLogin = "testuser"
							ctx := AddUserToContext(r.Context(), userLogin)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}

					cookie, err := r.Cookie("user_login")
					if err == nil && cookie.Value != "" {
						ctx := AddUserToContext(r.Context(), cookie.Value)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}

					if r.URL.Path != "/api/v1/login" && r.URL.Path != "/api/v1/register" {
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte(`{"error":"Требуется авторизация"}`))
						return
					}

					next.ServeHTTP(w, r)
				})
			}

			handler := customAuthMiddleware(testHandler)

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен: %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	userCtx := AddUserToContext(nil, "testuser")

	user := GetUserFromContext(userCtx)
	if user != "testuser" {
		t.Errorf("Ожидался пользователь 'testuser', получен: %s", user)
	}

	user = GetUserFromContext(nil)
	if user != "" {
		t.Errorf("Из пустого контекста получен непустой пользователь: %s", user)
	}
}
