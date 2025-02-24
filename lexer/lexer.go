package lexer

import (
	"djanGO/utils"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenNumber
	TokenOperator
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenLeftParen
	TokenRightParen
)

type Token struct {
	Type    TokenType
	Literal string
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
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

	if l.ch == 0 {
		return Token{Type: TokenEOF}
	}

	switch l.ch {
	case '+':
		l.readChar()
		return Token{Type: TokenPlus, Literal: "+"}
	case '-':
		l.readChar()
		return Token{Type: TokenMinus, Literal: "-"}
	case '*':
		l.readChar()
		return Token{Type: TokenMultiply, Literal: "*"}
	case '/':
		l.readChar()
		return Token{Type: TokenDivide, Literal: "/"}
	case '(':
		l.readChar()
		return Token{Type: TokenLeftParen, Literal: "("}
	case ')':
		l.readChar()
		return Token{Type: TokenRightParen, Literal: ")"}
	default:
		if utils.IsDigit(l.ch) {
			literal := l.readNumber()
			return Token{Type: TokenNumber, Literal: literal}
		}
		l.readChar()
		return l.NextToken()
	}
}

func (l *Lexer) readNumber() string {
	position := l.position
	for utils.IsDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) Reset() {
	l.position = 0
	l.readPosition = 0
	l.ch = 0
	l.readChar()
}

func (l *Lexer) GetAllTokens() []Token {
	savedPos := l.position
	savedReadPos := l.readPosition
	savedCh := l.ch

	l.Reset()

	var tokens []Token
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.Type == TokenEOF {
			break
		}
	}

	l.position = savedPos
	l.readPosition = savedReadPos
	l.ch = savedCh

	return tokens
}
