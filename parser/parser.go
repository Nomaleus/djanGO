package parser

import (
	"errors"
	"fmt"
	"strconv"

	"djanGO/lexer"
)

type Parser struct {
	lexer     *lexer.Lexer
	CurToken  lexer.Token
	peekToken lexer.Token
}

func NewParser(lexer *lexer.Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.CurToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseExpression() (float64, error) {
	return p.parseExpression()
}

func (p *Parser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for p.CurToken.Type == lexer.TokenPlus || p.CurToken.Type == lexer.TokenMinus {
		operator := p.CurToken.Type
		p.nextToken()
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if operator == lexer.TokenPlus {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *Parser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	for p.CurToken.Type == lexer.TokenMultiply || p.CurToken.Type == lexer.TokenDivide {
		operator := p.CurToken.Type
		p.nextToken()
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if operator == lexer.TokenMultiply {
			left *= right
		} else {
			if right == 0 {
				return 0, errors.New("Не дели на ноль!")
			}
			left /= right
		}
	}
	return left, nil
}

func (p *Parser) parseFactor() (float64, error) {
	switch p.CurToken.Type {
	case lexer.TokenNumber:
		value := p.CurToken.Value
		p.nextToken()
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("Неправильное число: %s", value)
		}
		return num, nil
	case lexer.TokenLeftParen:
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		if p.CurToken.Type != lexer.TokenRightParen {
			return 0, errors.New("Отсутствует закрывающая скобка")
		}
		p.nextToken()
		return expr, nil
	case lexer.TokenMinus:
		p.nextToken()
		value, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return -value, nil
	default:
		return 0, fmt.Errorf("Неожиданный токен: %s", p.CurToken.Value)
	}
}
