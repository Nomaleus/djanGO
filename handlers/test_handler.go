package handlers

import (
	"djanGO/utils"
	"net/http"
)

func (h *Handler) TestToken(w http.ResponseWriter, r *http.Request) {
	userLogin, ok := r.Context().Value("user").(string)

	if !ok || userLogin == "" || userLogin == "guest" {
		utils.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"valid": false,
			"error": "Пользователь не авторизован",
		})
		return
	}

	headerLogin := r.Header.Get("X-User-Login")
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
