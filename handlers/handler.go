package handlers

import (
	"djanGO/storage"
	"net/http"
)

type Handler struct {
	Storage *storage.Storage
}

func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{
		Storage: storage,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
