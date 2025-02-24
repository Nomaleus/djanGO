package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"djanGO/handlers"
	"djanGO/storage"
	"djanGO/utils"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var computingPower int

func init() {
	computingPower = utils.GetEnvInt("COMPUTING_POWER", 2)
}

func startWorkers(storage *storage.Storage) {
	for i := 0; i < computingPower; i++ {
		go worker(storage)
	}
}

func worker(storage *storage.Storage) {
	for {
		task, err := storage.GetPendingTask()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

		var result float64
		switch task.Operation {
		case "+":
			result = task.Arg1 + task.Arg2
		case "-":
			result = task.Arg1 - task.Arg2
		case "*":
			result = task.Arg1 * task.Arg2
		case "/":
			if task.Arg2 != 0 {
				result = task.Arg1 / task.Arg2
			}
		}

		storage.UpdateTaskResult(task.ID, result)
	}
}

func main() {
	store := storage.NewStorage()
	handler := handlers.NewHandler(store)

	startWorkers(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handler.Index)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/calculate", handler.Calculate)
		r.Get("/expressions", handler.GetExpressions)
		r.Get("/expressions/{id}", handler.GetExpressionByID)
	})

	r.Route("/internal", func(r chi.Router) {
		r.Get("/task", handler.GetTask)
		r.Post("/task", handler.SubmitTaskResult)
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	port = ":" + port

	fmt.Printf("Server started on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
