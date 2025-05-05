package main

import (
	"djanGO/grpc/client"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
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

	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	workersCountStr := os.Getenv("WORKERS_COUNT")
	workersCount := 2

	if workersCountStr != "" {
		count, err := strconv.Atoi(workersCountStr)
		if err == nil && count > 0 {
			workersCount = count
		}
	}

	fmt.Printf("Starting %d gRPC worker(s) connecting to %s\n", workersCount, grpcAddr)

	for i := 1; i <= workersCount; i++ {
		go client.StartWorker(grpcAddr, i)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals
	fmt.Println("Shutting down workers...")
}
