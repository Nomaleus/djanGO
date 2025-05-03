package handlers

import (
	"djanGO/utils"
	"encoding/json"
	"net/http"
	"time"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Login   string `json:"login"`
	Success bool   `json:"success"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректные данные"})
		return
	}

	if req.Login == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Логин и пароль обязательны"})
		return
	}

	isValid, err := utils.AuthenticateUser(req.Login, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка сервера"})
		return
	}

	if !isValid {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
		return
	}

	token, err := utils.GenerateToken(req.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка генерации токена"})
		return
	}

	cookie := &http.Cookie{
		Name:     "user_login",
		Value:    req.Login,
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: false,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	response := `{"success":true,"login":"` + req.Login + `","token":"` + token + `"}`
	w.Write([]byte(response))
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Некорректный формат данных",
		})
		return
	}

	if len(req.Login) < 3 {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Логин должен содержать не менее 3 символов",
		})
		return
	}

	if len(req.Password) < 5 {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Пароль должен содержать не менее 5 символов",
		})
		return
	}

	err := utils.RegisterUser(req.Login, req.Password)
	if err != nil {
		if err.Error() == "пользователь с таким логином уже существует" {
			utils.WriteJSON(w, http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		} else {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при регистрации пользователя",
			})
		}
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{
		"success": "true",
		"message": "Пользователь успешно зарегистрирован",
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	cookie := &http.Cookie{
		Name:     "user_login",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-24 * time.Hour),
		HttpOnly: false,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Выход выполнен успешно",
	})
}

func (h *Handler) CheckToken(w http.ResponseWriter, r *http.Request) {
	userLogin, ok := r.Context().Value("user").(string)
	if !ok {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"valid": "false",
			"error": "Недействительный токен авторизации",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"valid": "true",
		"login": userLogin,
	})
}

func (h *Handler) ForceAuth(w http.ResponseWriter, r *http.Request) {
	login := r.URL.Query().Get("login")
	if login == "" {
		login = "admin"
	}

	sessionCookie := &http.Cookie{
		Name:     "session_id",
		Value:    login,
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
	}
	http.SetCookie(w, sessionCookie)

	userLoginCookie := &http.Cookie{
		Name:     "user_login",
		Value:    login,
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: false,
	}
	http.SetCookie(w, userLoginCookie)

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "application/json")

	w.Write([]byte(`{"success": true, "login": "` + login + `", "message": "Принудительная авторизация выполнена"}`))
}
