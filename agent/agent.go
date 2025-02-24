package agent

import (
	"bytes"
	"djanGO/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Agent struct {
	serverURL      string
	computingPower int
}

func (a *Agent) Start() {
	for i := 0; i < a.computingPower; i++ {
		go a.worker()
	}
}

func (a *Agent) worker() {
	for {
		task, err := a.getTask()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		result := a.compute(task)
		a.submitResult(task.ID, result)
	}
}

func (a *Agent) compute(task *models.Task) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

func (a *Agent) getTask() (*models.Task, error) {
	resp, err := http.Get(a.serverURL + "/task")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no tasks available")
	}

	var response struct {
		Task *models.Task `json:"task"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Task, nil
}

func (a *Agent) submitResult(id string, result float64) error {
	data := map[string]interface{}{
		"id":     id,
		"result": result,
	}

	_, err := http.Post(a.serverURL+"/submit", "application/json", jsonReader(data))
	return err
}

func jsonReader(v interface{}) io.Reader {
	data, _ := json.Marshal(v)
	return bytes.NewReader(data)
}
