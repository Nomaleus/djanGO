package lexer

import (
	"djanGO/utils"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	TokenEOF     TokenType
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:    input,
		TokenEOF: TokenEOF,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	var tok Token

	switch l.ch {
	case '+':
		tok = Token{Type: TokenPlus, Value: string(l.ch)}
	case '-':
		tok = Token{Type: TokenMinus, Value: string(l.ch)}
	case '*':
		tok = Token{Type: TokenMultiply, Value: string(l.ch)}
	case '/':
		tok = Token{Type: TokenDivide, Value: string(l.ch)}
	case '(':
		tok = Token{Type: TokenLeftParen, Value: string(l.ch)}
	case ')':
		tok = Token{Type: TokenRightParen, Value: string(l.ch)}
	case 0:
		tok = Token{Type: TokenEOF, Value: ""}
	default:
		if utils.IsDigit(l.ch) || l.ch == '.' {
			tok.Type = TokenNumber
			tok.Value = l.readNumber()
			return tok
		} else {
			return Token{Type: TokenEOF, Value: ""}
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) readNumber() string {
	position := l.position
	hasDot := false

	for utils.IsDigit(l.ch) || (l.ch == '.' && !hasDot) {
		if l.ch == '.' {
			hasDot = true
		}
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
