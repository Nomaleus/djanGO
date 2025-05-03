package client

import (
	"context"
	"djanGO/models"
	pb "djanGO/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TaskClient struct {
	conn   *grpc.ClientConn
	client pb.TaskServiceClient
}

func NewTaskClient(serverAddr string) (*TaskClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewTaskServiceClient(conn)

	return &TaskClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *TaskClient) Close() error {
	return c.conn.Close()
}

func (c *TaskClient) GetTasks(ctx context.Context) ([]*models.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.client.GetTask(ctx, &pb.GetTaskRequest{})
	if err != nil {
		return nil, err
	}

	var tasks []*models.Task
	for _, pbTask := range response.Tasks {
		tasks = append(tasks, &models.Task{
			ID:            pbTask.Id,
			Arg1:          pbTask.Arg1,
			Arg2:          pbTask.Arg2,
			Operation:     pbTask.Operation,
			OperationTime: int(pbTask.OperationTime),
			Status:        pbTask.Status,
			Result:        pbTask.Result,
			Error:         pbTask.Error,
		})
	}

	return tasks, nil
}

func (c *TaskClient) GetPendingTask(ctx context.Context, workerID int) (*models.Task, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	response, err := c.client.GetAndLockPendingTask(ctx, &pb.GetAndLockPendingTaskRequest{
		WorkerId: int32(workerID),
	})
	if err != nil {
		return nil, false, err
	}

	if !response.TaskFound {
		return nil, false, nil
	}

	task := &models.Task{
		ID:            response.Task.Id,
		Arg1:          response.Task.Arg1,
		Arg2:          response.Task.Arg2,
		Operation:     response.Task.Operation,
		OperationTime: int(response.Task.OperationTime),
		Status:        response.Task.Status,
		Result:        response.Task.Result,
		Error:         response.Task.Error,
	}

	return task, true, nil
}

func (c *TaskClient) SubmitNewTask(ctx context.Context, task *models.Task) (string, float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pbTask := &pb.Task{
		Id:            task.ID,
		Arg1:          task.Arg1,
		Arg2:          task.Arg2,
		Operation:     task.Operation,
		OperationTime: int32(task.OperationTime),
	}

	response, err := c.client.SubmitTaskResult(ctx, &pb.SubmitTaskResultRequest{
		Request: &pb.SubmitTaskResultRequest_Task{
			Task: pbTask,
		},
	})
	if err != nil {
		return "", 0, err
	}

	return response.Id, response.Result, nil
}

func (c *TaskClient) SubmitTaskResult(ctx context.Context, taskID string, result float64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.SubmitTaskResult(ctx, &pb.SubmitTaskResultRequest{
		Request: &pb.SubmitTaskResultRequest_Result{
			Result: &pb.TaskResult{
				Id:     taskID,
				Result: result,
			},
		},
	})

	return err
}

func (c *TaskClient) SubmitTaskError(ctx context.Context, taskID string, errorMsg string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	task := &pb.Task{
		Id:     taskID,
		Error:  errorMsg,
		Status: "ERROR",
	}

	_, err := c.client.SubmitTaskResult(ctx, &pb.SubmitTaskResultRequest{
		Request: &pb.SubmitTaskResultRequest_Task{
			Task: task,
		},
	})

	return err
}
