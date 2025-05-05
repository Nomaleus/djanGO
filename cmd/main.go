package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"djanGO/db"
	"djanGO/grpc/server"
	"djanGO/handlers"
	"djanGO/storage"
	"djanGO/utils"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
)

var (
	computingPower int
	enableGRPC     bool
	grpcAddr       string
)

func init() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Ошибка получения пути к исполняемому файлу: %v\n", err)
	}

	execDir := filepath.Dir(execPath)
	projectRoot := filepath.Join(execDir, "..")

	envLoaded := false

	if err := godotenv.Load(); err == nil {
		envLoaded = true
		fmt.Println("Загружен .env из текущей директории")
	}

	if !envLoaded {
		envPath := filepath.Join(projectRoot, ".env")
		if err := godotenv.Load(envPath); err == nil {
			envLoaded = true
		} else {
			fmt.Printf("Не удалось загрузить .env из %s: %v\n", envPath, err)
		}
	}

	if !envLoaded {
		fmt.Println("Предупреждение: .env файл не найден, используются значения по умолчанию")
	}

	computingPower = utils.GetEnvInt("COMPUTING_POWER", 2)
	enableGRPC = os.Getenv("ENABLE_GRPC") == "true"
	grpcAddr = os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = ":50051"
	}
}

func startWorkers(store *storage.Storage) {
	wrapper := storage.NewStorageWrapper(store)

	for i := 0; i < computingPower; i++ {
		go worker(wrapper)
	}
}

func worker(storageWrapper *storage.StorageWrapper) {
	lastLogTime := time.Now().Add(-60 * time.Second)

	for {
		task, err := storageWrapper.GetPendingTask()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		if task == nil {
			now := time.Now()
			if now.Sub(lastLogTime) >= 30*time.Second {
				log.Printf("Worker local: Нет подходящих задач для выполнения")
				lastLogTime = now
			}
			time.Sleep(time.Second)
			continue
		}

		fmt.Printf("Воркер начинает выполнение задачи ID=%s (Operation=%s, Arg1=%f, Arg2=%f)\n",
			task.ID, task.Operation, task.Arg1, task.Arg2)
		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

		var result float64
		var opErr error

		switch task.Operation {
		case "+":
			result = task.Arg1 + task.Arg2
		case "-":
			result = task.Arg1 - task.Arg2
		case "*":
			result = task.Arg1 * task.Arg2
		case "/":
			if task.Arg2 == 0 {
				opErr = fmt.Errorf("division by zero")
				fmt.Printf("Ошибка деления на ноль в задаче ID=%s\n", task.ID)
				storageWrapper.UpdateTaskError(task.ID, opErr.Error())

				expressions, _ := storageWrapper.GetAllExpressions()
				for _, expr := range expressions {
					for _, t := range expr.Tasks {
						if t.ID == task.ID {
							expr.Status = "ERROR"
							fmt.Printf("Обновляем статус выражения ID=%s на ERROR\n", expr.ID)
							storageWrapper.UpdateExpression(expr)

							_, err = db.DB.Exec(
								"UPDATE expressions SET status = 'ERROR' WHERE id = ?",
								opErr.Error(), expr.ID)
							if err != nil {
								fmt.Printf("Ошибка обновления статуса выражения в БД: %v\n", err)
							}

							break
						}
					}
				}

				continue
			}
			result = task.Arg1 / task.Arg2
		}

		if opErr == nil {
			err = storageWrapper.UpdateTaskResult(task.ID, result)
			if err != nil {
				fmt.Printf("Ошибка обновления результата задачи ID=%s: %v\n", task.ID, err)
				continue
			}

			expr, err := storageWrapper.GetExpressionByTaskID(task.ID)
			if err == nil && expr != nil {
				if expr.Status == "COMPLETED" {

					_, err = db.DB.Exec(
						"UPDATE expressions SET status = 'COMPLETED', result = ? WHERE id = ?",
						expr.Result, expr.ID)
					if err != nil {
						fmt.Printf("Ошибка обновления выражения в БД: %v\n", err)
					} else {
						fmt.Printf("Результат выражения ID=%s успешно обновлен в БД\n", expr.ID)
					}
				} else {
					fmt.Println()
				}
			}

			_, err = db.DB.Exec(
				"UPDATE tasks SET status = 'COMPLETED', result = ? WHERE id = ?",
				result, task.ID)
		}
	}
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("djanGO v1.1.0 (с поддержкой JWT токенов для авторизации)")
		return
	}

	if err := db.InitDB(); err != nil {
		fmt.Printf("Ошибка инициализации базы данных: %v\n", err)
		return
	}
	defer db.CloseDB()

	store := storage.NewStorage()
	storageWrapper := storage.NewStorageWrapper(store)
	handler := handlers.NewHandler(storageWrapper)

	if enableGRPC {
		go func() {
			fmt.Printf("Starting gRPC server on %s\n", grpcAddr)
			if err := server.StartGRPCServer(grpcAddr, store, handler); err != nil {
				fmt.Printf("Failed to start gRPC server: %v\n", err)
			}
		}()
		fmt.Println("Используем внешние gRPC воркеры для обработки задач")
	} else {
		fmt.Printf("gRPC отключен, запускаем %d локальных воркеров\n", computingPower)
		startWorkers(store)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(utils.CORSMiddleware)

	r.Get("/", handler.Index)

	r.Get("/login", handler.LoginPage)
	r.Get("/register", handler.RegisterPage)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(utils.AuthMiddleware)

		r.Post("/login", handler.Login)
		r.Post("/register", handler.Register)
		r.Post("/logout", handler.Logout)
		r.Get("/check-token", handler.CheckToken)
		r.Post("/calculate", handler.Calculate)
		r.Get("/expressions", handler.GetExpressions)
		r.Get("/expressions/{id}", handler.GetExpressionByID)
		r.Get("/history", handler.GetUserHistory)
	})

	if !enableGRPC {
		r.Route("/internal", func(r chi.Router) {
			r.Get("/task", handler.GetTask)
			r.Post("/task", handler.SubmitTaskResult)
		})
	}

	staticDir := handler.StaticDir
	fmt.Printf("Static directory: %s\n", staticDir)

	keysFiles := []string{
		"login.html",
		"register.html",
		"index.html",
		"js/auth.js",
		"js/auth-fix.js",
		"js/calculator.js",
	}

	for _, file := range keysFiles {
		fullPath := filepath.Join(staticDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Printf("ВНИМАНИЕ: файл %s не найден\n", fullPath)
		}
	}

	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		path = strings.TrimPrefix(path, "/static/")
		fullPath := filepath.Join(staticDir, path)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	port = ":" + port

	fmt.Printf("HTTP server started on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
