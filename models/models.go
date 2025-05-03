package models

import (
	"time"
)

type Expression struct {
	ID       string    `json:"id"`
	Original string    `json:"original"`
	Status   string    `json:"status"`
	Result   float64   `json:"result"`
	Error    string    `json:"error,omitempty"`
	Tasks    []*Task   `json:"tasks,omitempty"`
	Created  time.Time `json:"created"`
}

type Task struct {
	ID            string   `json:"id"`
	ExpressionID  string   `json:"expression_id,omitempty"`
	Order         int      `json:"order,omitempty"`
	Operation     string   `json:"operation"`
	Arg1          float64  `json:"arg1"`
	Arg2          float64  `json:"arg2"`
	Status        string   `json:"status"`
	Result        float64  `json:"result"`
	Completed     bool     `json:"-"`
	OperationTime int      `json:"operation_time"`
	Error         string   `json:"error,omitempty"`
	DependsOn     []string `json:"depends_on,omitempty"`
	Arg1Source    string   `json:"arg1_source,omitempty"`
	Arg2Source    string   `json:"arg2_source,omitempty"`
}

type ExpressionResponse struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Result  *float64  `json:"result,omitempty"`
	Error   string    `json:"error,omitempty"`
	Tasks   []*Task   `json:"tasks,omitempty"`
	Created time.Time `json:"created"`
	Text    string    `json:"text,omitempty"`
}

type ExpressionsResponse struct {
	Expressions []ExpressionResponse `json:"expressions"`
}

type ExpressionWrapper struct {
	Expression string `json:"expression"`
}
