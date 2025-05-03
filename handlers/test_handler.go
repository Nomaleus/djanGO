package handlers

import (
	"djanGO/utils"
	"fmt"
	"net/http"
)

func (h *Handler) TestToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handler TestToken вызван")

	userLogin, ok := r.Context().Value("user").(string)
	fmt.Printf("TestToken: пользователь из контекста: %s\n", userLogin)

	if !ok || userLogin == "" || userLogin == "guest" {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"valid": false,
			"error": "Пользователь не авторизован",
		})
		return
	}

	headerLogin := r.Header.Get("X-User-Login")
	fmt.Printf("TestToken: заголовок X-User-Login: %s\n", headerLogin)

	fmt.Printf("TestToken: пользователь авторизован: %s\n", userLogin)

	var source string
	if headerLogin != "" {
		source = "header"
	} else {
		source = "cookie"
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"login": userLogin,
		"info": map[string]string{
			"source":       source,
			"header_login": headerLogin,
		},
	})
}
