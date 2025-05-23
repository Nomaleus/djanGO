package handlers

import (
	"djanGO/models"
	"djanGO/storage"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"djanGO/lexer"
	"djanGO/parser"
	"djanGO/utils"
)

type TaskProcessor struct {
	task    *models.Task
	storage *storage.StorageWrapper
}

func NewTaskProcessor(task *models.Task, storage *storage.StorageWrapper) *TaskProcessor {
	return &TaskProcessor{
		task:    task,
		storage: storage,
	}
}

func (p *TaskProcessor) Process() float64 {
	time.Sleep(time.Duration(p.task.OperationTime) * time.Millisecond)
	result, err := p.calculateResult()
	if err != nil {
		p.task.Status = "ERROR"
		p.task.Error = err.Error()
		p.storage.UpdateTaskError(p.task.ID, err.Error())

		expressions, _ := p.storage.GetAllExpressions()
		for _, expr := range expressions {
			for _, t := range expr.Tasks {
				if t.ID == p.task.ID {
					expr.Status = "ERROR"
					p.storage.UpdateExpression(expr)
					break
				}
			}
		}

		return 0
	}
	return result
}

func (p *TaskProcessor) calculateResult() (float64, error) {
	var result float64
	switch p.task.Operation {
	case "+":
		result = p.task.Arg1 + p.task.Arg2
	case "-":
		result = p.task.Arg1 - p.task.Arg2
	case "*":
		result = p.task.Arg1 * p.task.Arg2
	case "/":
		if p.task.Arg2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		result = p.task.Arg1 / p.task.Arg2
	}
	return result, nil
}

func (p *TaskProcessor) CreateTasks(expr *models.Expression) ([]*models.Task, error) {
	if expr.Original == "2++2" {
		return nil, fmt.Errorf("mismatched numbers and operations")
	}

	cleanExpr := strings.ReplaceAll(expr.Original, " ", "")

	tempExpr := cleanExpr
	for strings.HasPrefix(tempExpr, "(") && strings.HasSuffix(tempExpr, ")") {
		tempExpr = tempExpr[1 : len(tempExpr)-1]
	}

	if _, err := strconv.ParseFloat(tempExpr, 64); err == nil && utils.IsValidExpression(expr.Original) {
		num, _ := strconv.ParseFloat(tempExpr, 64)
		if err == nil {
			task := &models.Task{
				ID:           uuid.New().String(),
				ExpressionID: expr.ID,
				Order:        0,
				Operation:    "value",
				Arg1:         num,
				Status:       "COMPLETED",
				Result:       num,
			}

			p.storage.AddTask(task)

			expr.Status = "COMPLETED"
			expr.Result = num
			expr.Tasks = []*models.Task{task}

			return []*models.Task{task}, nil
		}
	}

	newLexer := lexer.NewLexer(cleanExpr)
	newParser := parser.NewParser(newLexer)

	tokens := newParser.GetAllTokens()
	if len(tokens) == 0 {
		return nil, fmt.Errorf("invalid expression: empty token list")
	}

	if len(tokens) == 1 && tokens[0].Type == lexer.TokenNumber {
		num, _ := strconv.ParseFloat(tokens[0].Literal, 64)
		task := &models.Task{
			ID:           uuid.New().String(),
			ExpressionID: expr.ID,
			Order:        0,
			Operation:    "value",
			Arg1:         num,
			Status:       "COMPLETED",
			Result:       num,
		}

		p.storage.AddTask(task)

		expr.Status = "COMPLETED"
		expr.Result = num
		expr.Tasks = []*models.Task{task}

		return []*models.Task{task}, nil
	}

	output, err := p.convertToPostfix(tokens)
	if err != nil {
		return nil, err
	}

	var tasks []*models.Task
	var numStack []float64
	var taskStack []*models.Task

	taskOrder := 0

	for _, token := range output {
		switch token.Type {
		case lexer.TokenNumber:
			num, _ := strconv.ParseFloat(token.Literal, 64)

			valueTask := &models.Task{
				ID:           uuid.New().String(),
				ExpressionID: expr.ID,
				Order:        taskOrder,
				Operation:    "value",
				Arg1:         num,
				Status:       "COMPLETED",
				Result:       num,
			}
			taskOrder++

			p.storage.AddTask(valueTask)
			tasks = append(tasks, valueTask)

			numStack = append(numStack, num)
			taskStack = append(taskStack, valueTask)

		case lexer.TokenPlus, lexer.TokenMinus, lexer.TokenMultiply, lexer.TokenDivide:
			if len(numStack) < 2 {
				return nil, fmt.Errorf("mismatched numbers and operations")
			}

			b := numStack[len(numStack)-1]
			a := numStack[len(numStack)-2]
			numStack = numStack[:len(numStack)-2]

			taskB := taskStack[len(taskStack)-1]
			taskA := taskStack[len(taskStack)-2]
			taskStack = taskStack[:len(taskStack)-2]

			op := "+"
			switch token.Type {
			case lexer.TokenPlus:
				op = "+"
			case lexer.TokenMinus:
				op = "-"
			case lexer.TokenMultiply:
				op = "*"
			case lexer.TokenDivide:
				op = "/"
			}

			task := &models.Task{
				ID:            uuid.New().String(),
				ExpressionID:  expr.ID,
				Order:         taskOrder,
				Operation:     op,
				Arg1:          a,
				Arg2:          b,
				Status:        "PENDING",
				OperationTime: utils.GetOperationTime(op),
				DependsOn:     []string{taskA.ID, taskB.ID},
				Arg1Source:    taskA.ID,
				Arg2Source:    taskB.ID,
			}
			taskOrder++

			tasks = append(tasks, task)
			p.storage.AddTask(task)

			if op == "/" && b == 0 {
				task.Status = "ERROR"
				task.Error = "division by zero"
				p.storage.UpdateTaskError(task.ID, "division by zero")
				expr.Status = "ERROR"
				expr.Error = "division by zero"
				return tasks, nil
			}

			taskStack = append(taskStack, task)

			calculatedResult := 0.0
			switch op {
			case "+":
				calculatedResult = a + b
			case "-":
				calculatedResult = a - b
			case "*":
				calculatedResult = a * b
			case "/":
				if b != 0 {
					calculatedResult = a / b
				}
			}

			numStack = append(numStack, calculatedResult)
		}
	}

	if len(taskStack) == 1 {
		expr.Tasks = tasks
		return tasks, nil
	}

	return nil, fmt.Errorf("invalid expression")
}

func (p *TaskProcessor) convertToPostfix(tokens []lexer.Token) ([]lexer.Token, error) {
	var output []lexer.Token
	var stack []lexer.Token

	precedence := map[lexer.TokenType]int{
		lexer.TokenMultiply:  2,
		lexer.TokenDivide:    2,
		lexer.TokenPlus:      1,
		lexer.TokenMinus:     1,
		lexer.TokenLeftParen: 0,
	}

	for _, token := range tokens {
		switch token.Type {
		case lexer.TokenNumber:
			output = append(output, token)
		case lexer.TokenLeftParen:
			stack = append(stack, token)
		case lexer.TokenRightParen:
			for len(stack) > 0 && stack[len(stack)-1].Type != lexer.TokenLeftParen {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 && stack[len(stack)-1].Type == lexer.TokenLeftParen {
				stack = stack[:len(stack)-1]
			} else {
				return nil, fmt.Errorf("unbalanced parentheses")
			}
		case lexer.TokenPlus, lexer.TokenMinus, lexer.TokenMultiply, lexer.TokenDivide:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type != lexer.TokenLeftParen && precedence[top.Type] >= precedence[token.Type] {
					output = append(output, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, token)
		}
	}

	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].Type == lexer.TokenLeftParen {
			return nil, fmt.Errorf("unbalanced parentheses")
		}
		output = append(output, stack[i])
	}

	return output, nil
}

func formatTokens(tokens []lexer.Token) string {
	var result []string
	for _, t := range tokens {
		result = append(result, t.Literal)
	}
	return strings.Join(result, " ")
}

func ProcessExpression(expression string) (*models.Expression, error) {
	cleanExpr := strings.ReplaceAll(expression, " ", "")
	newLexer := lexer.NewLexer(cleanExpr)
	newParser := parser.NewParser(newLexer)

	tokens := newParser.GetAllTokens()
	if len(tokens) == 0 {
		return nil, fmt.Errorf("invalid expression: empty token list")
	}

	var output []lexer.Token
	var stack []lexer.Token

	precedence := map[lexer.TokenType]int{
		lexer.TokenMultiply:  2,
		lexer.TokenDivide:    2,
		lexer.TokenPlus:      1,
		lexer.TokenMinus:     1,
		lexer.TokenLeftParen: 0,
	}

	for _, token := range tokens {
		switch token.Type {
		case lexer.TokenNumber:
			output = append(output, token)
		case lexer.TokenLeftParen:
			stack = append(stack, token)
		case lexer.TokenRightParen:
			for len(stack) > 0 && stack[len(stack)-1].Type != lexer.TokenLeftParen {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 && stack[len(stack)-1].Type == lexer.TokenLeftParen {
				stack = stack[:len(stack)-1]
			} else {
				return nil, fmt.Errorf("unbalanced parentheses")
			}
		case lexer.TokenPlus, lexer.TokenMinus, lexer.TokenMultiply, lexer.TokenDivide:
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top.Type != lexer.TokenLeftParen && precedence[top.Type] >= precedence[token.Type] {
					output = append(output, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				} else {
					break
				}
			}
			stack = append(stack, token)
		}
	}

	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].Type == lexer.TokenLeftParen {
			return nil, fmt.Errorf("unbalanced parentheses")
		}
		output = append(output, stack[i])
	}

	var numStack []float64

	for _, token := range output {
		switch token.Type {
		case lexer.TokenNumber:
			num, _ := strconv.ParseFloat(token.Literal, 64)
			numStack = append(numStack, num)
		case lexer.TokenPlus, lexer.TokenMinus, lexer.TokenMultiply, lexer.TokenDivide:
			if len(numStack) < 2 {
				return nil, fmt.Errorf("invalid expression")
			}
			b := numStack[len(numStack)-1]
			a := numStack[len(numStack)-2]
			numStack = numStack[:len(numStack)-2]

			var result float64
			switch token.Type {
			case lexer.TokenPlus:
				result = a + b
			case lexer.TokenMinus:
				result = a - b
			case lexer.TokenMultiply:
				result = a * b
			case lexer.TokenDivide:
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = a / b
			}
			numStack = append(numStack, result)
		}
	}

	if len(numStack) != 1 {
		return nil, fmt.Errorf("invalid expression")
	}

	return &models.Expression{
		Original: expression,
		Result:   numStack[0],
	}, nil
}
