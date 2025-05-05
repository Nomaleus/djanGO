package handlers

import (
	"djanGO/storage"
	"net/http"
	"os"
	"path/filepath"
)

type Handler struct {
	Storage   *storage.StorageWrapper
	StaticDir string
}

func NewHandler(store *storage.StorageWrapper) *Handler {
	if _, err := os.Stat("static"); err == nil {
		absolutePath, _ := filepath.Abs("static")
		return &Handler{
			Storage:   store,
			StaticDir: absolutePath,
		}
	}

	if _, err := os.Stat("../static"); err == nil {
		absolutePath, _ := filepath.Abs("../static")
		return &Handler{
			Storage:   store,
			StaticDir: absolutePath,
		}
	}

	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)

		if filepath.Base(execDir) == "bin" {
			projectRoot := filepath.Dir(execDir)
			staticDir := filepath.Join(projectRoot, "static")

			if _, err := os.Stat(staticDir); err == nil {
				return &Handler{
					Storage:   store,
					StaticDir: staticDir,
				}
			}
		} else {
			staticDir := filepath.Join(execDir, "static")

			if _, err := os.Stat(staticDir); err == nil {
				return &Handler{
					Storage:   store,
					StaticDir: staticDir,
				}
			}

			projectRoot := filepath.Dir(execDir)
			staticDir = filepath.Join(projectRoot, "static")

			if _, err := os.Stat(staticDir); err == nil {
				return &Handler{
					Storage:   store,
					StaticDir: staticDir,
				}
			}
		}
	}

	return &Handler{
		Storage:   store,
		StaticDir: "static",
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(h.StaticDir, "index.html")

	http.ServeFile(w, r, indexPath)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	loginPath := filepath.Join(h.StaticDir, "login.html")

	http.ServeFile(w, r, loginPath)
}

func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	registerPath := filepath.Join(h.StaticDir, "register.html")

	http.ServeFile(w, r, registerPath)
}
