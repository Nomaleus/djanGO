package models

type Expression struct {
	ID       string  `json:"id"`
	Original string  `json:"original"`
	Status   string  `json:"status"`
	Result   float64 `json:"result"`
	Tasks    []*Task `json:"tasks,omitempty"`
}

type Task struct {
	ID            string  `json:"id"`
	Operation     string  `json:"operation"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Status        string  `json:"status"`
	Result        float64 `json:"result"`
	Completed     bool    `json:"-"`
	OperationTime int     `json:"operation_time"`
}

type ExpressionResponse struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

type ExpressionsResponse struct {
	Expressions []ExpressionResponse `json:"expressions"`
}

type ExpressionWrapper struct {
	Expression ExpressionResponse `json:"expression"`
}
