package server

import (
	"context"
	"djanGO/db"
	"djanGO/handlers"
	"djanGO/models"
	pb "djanGO/proto"
	"djanGO/storage"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskServer struct {
	pb.UnimplementedTaskServiceServer
	Storage  *storage.StorageWrapper
	Handlers *handlers.Handler
}

func NewTaskServer(store *storage.Storage, handlers *handlers.Handler) *TaskServer {
	storageWrapper := storage.NewStorageWrapper(store)

	return &TaskServer{
		Storage:  storageWrapper,
		Handlers: handlers,
	}
}

func (s *TaskServer) GetTask(context.Context, *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	tasks := s.Storage.GetAllTasks()

	var pbTasks []*pb.Task
	for _, task := range tasks {
		pbTasks = append(pbTasks, &pb.Task{
			Id:            task.ID,
			Arg1:          task.Arg1,
			Arg2:          task.Arg2,
			Operation:     task.Operation,
			OperationTime: int32(task.OperationTime),
			Status:        task.Status,
			Result:        task.Result,
			Error:         task.Error,
		})
	}

	return &pb.GetTaskResponse{
		Tasks: pbTasks,
	}, nil
}

func (s *TaskServer) GetAndLockPendingTask(_ context.Context, req *pb.GetAndLockPendingTaskRequest) (*pb.GetAndLockPendingTaskResponse, error) {
	workerID := fmt.Sprintf("worker-%d", req.WorkerId)

	task, err := s.Storage.GetAndLockPendingTask(workerID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get pending task: %v", err)
	}

	if task == nil {
		return &pb.GetAndLockPendingTaskResponse{
			TaskFound: false,
		}, nil
	}

	pbTask := &pb.Task{
		Id:            task.ID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operation,
		OperationTime: int32(task.OperationTime),
		Status:        task.Status,
		Result:        task.Result,
		Error:         task.Error,
	}

	return &pb.GetAndLockPendingTaskResponse{
		Task:      pbTask,
		TaskFound: true,
	}, nil
}

func (s *TaskServer) SubmitTaskResult(_ context.Context, req *pb.SubmitTaskResultRequest) (*pb.SubmitTaskResultResponse, error) {
	switch req := req.Request.(type) {
	case *pb.SubmitTaskResultRequest_Task:
		pbTask := req.Task

		taskID := pbTask.Id
		if taskID == "" {
			taskID = uuid.New().String()
		}

		if pbTask.Status == "ERROR" && pbTask.Error != "" {
			task, err := s.Storage.GetTask(taskID)
			if err != nil {
				return nil, status.Errorf(codes.NotFound, "Task not found: %v", err)
			}

			err = s.Storage.UpdateTaskError(taskID, pbTask.Error)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to update task error: %v", err)
			}

			expr, err := s.Storage.GetExpression(task.ExpressionID)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to get expression: %v", err)
			}

			_, err = db.DB.Exec(
				"UPDATE tasks SET status = 'ERROR' WHERE id = ?",
				pbTask.Error, taskID)
			if err != nil {
				fmt.Printf("Ошибка обновления статуса задачи в БД (gRPC): %v\n", err)
			} else {
				fmt.Printf("gRPC: Статус задачи %s обновлен на ERROR в БД\n", taskID)
			}

			_, err = db.DB.Exec(
				"UPDATE expressions SET status = 'ERROR' WHERE id = ?",
				pbTask.Error, expr.ID)
			if err != nil {
				fmt.Printf("Ошибка обновления статуса выражения в БД (gRPC): %v\n", err)
			} else {
				fmt.Printf("gRPC: Статус выражения %s обновлен на ERROR в БД\n", expr.ID)
			}

			return &pb.SubmitTaskResultResponse{
				Id: taskID,
			}, nil
		}

		task := &models.Task{
			ID:            taskID,
			Operation:     pbTask.Operation,
			Arg1:          pbTask.Arg1,
			Arg2:          pbTask.Arg2,
			Status:        "PENDING",
			OperationTime: int(pbTask.OperationTime),
		}

		if err := s.Storage.AddTask(task); err != nil {
			if err.Error() == "task already exists" {
				return nil, status.Errorf(codes.AlreadyExists, "Task already exists")
			}
			return nil, status.Errorf(codes.Internal, "Internal error: %v", err)
		}

		return &pb.SubmitTaskResultResponse{
			Id: taskID,
		}, nil

	case *pb.SubmitTaskResultRequest_Result:
		pbResult := req.Result

		err := s.Storage.UpdateTaskResult(pbResult.Id, pbResult.Result)
		if err != nil {
			if err.Error() == "task not found" {
				return nil, status.Errorf(codes.NotFound, "Task not found")
			}
			return nil, status.Errorf(codes.Internal, "Internal error: %v", err)
		}

		task, err := s.Storage.GetTask(pbResult.Id)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to get updated task: %v", err)
		}

		expr, err := s.Storage.GetExpression(task.ExpressionID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to get expression: %v", err)
		}

		_, err = db.DB.Exec(
			"UPDATE tasks SET status = 'COMPLETED', result = ? WHERE id = ?",
			pbResult.Result, pbResult.Id)
		if err != nil {
			fmt.Printf("Ошибка обновления задачи в БД (gRPC): %v\n", err)
		} else {
			fmt.Printf("gRPC: Задача %s обновлена в БД, результат=%f\n", pbResult.Id, pbResult.Result)
		}

		if expr.Status == "COMPLETED" {
			_, err = db.DB.Exec(
				"UPDATE expressions SET status = 'COMPLETED', result = ? WHERE id = ?",
				expr.Result, expr.ID)
			if err != nil {
				fmt.Printf("Ошибка обновления выражения в БД (gRPC): %v\n", err)
			} else {
				fmt.Printf("gRPC: Выражение %s обновлено в БД, результат=%f\n", expr.ID, expr.Result)
			}
		}

		return &pb.SubmitTaskResultResponse{
			Id:     pbResult.Id,
			Result: pbResult.Result,
		}, nil

	default:
		return nil, status.Errorf(codes.InvalidArgument, "Invalid request type")
	}
}
