package client

import (
	"context"
	"djanGO/models"
	"fmt"
	"log"
	"time"
)

func StartWorker(grpcAddr string, workerId int) {
	log.Printf("Starting gRPC worker #%d, connecting to %s", workerId, grpcAddr)

	client, clientErr := NewTaskClient(grpcAddr)
	if clientErr != nil {
		log.Fatalf("Worker #%d failed to create client: %v", workerId, clientErr)
	}
	defer client.Close()

	for {
		task, taskFound, getTaskErr := client.GetPendingTask(context.Background(), workerId)
		if getTaskErr != nil {
			time.Sleep(time.Second)
			continue
		}

		if !taskFound {
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Worker #%d взял задачу %s: %s", workerId, task.ID, taskDescription(task))

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
				log.Printf("Worker #%d: ошибка деления на ноль в задаче %s", workerId, task.ID)
			} else {
				result = task.Arg1 / task.Arg2
			}
		default:
			log.Printf("Worker #%d: неизвестная операция: %s", workerId, task.Operation)
		}

		if opErr != nil {
			log.Printf("Worker #%d задача %s завершилась с ошибкой: %v", workerId, task.ID, opErr)
		} else {
			if submitErr := client.SubmitTaskResult(context.Background(), task.ID, result); submitErr != nil {
				log.Printf("Worker #%d не удалось отправить результат для задачи %s: %v", workerId, task.ID, submitErr)
			} else {
				log.Printf("Worker #%d решил задачу %s с результатом: %v", workerId, task.ID, result)
			}
		}
	}
}

func taskDescription(task *models.Task) string {
	return fmt.Sprintf("%v %s %v", task.Arg1, task.Operation, task.Arg2)
}
