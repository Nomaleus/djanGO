package main

import (
	"djanGO/handlers"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api/v1/calculate", handlers.CalculateHandler)
	port := ":80"
	fmt.Println("Server started on http://localhost" + port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
