package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/faideww/glox/src/token"
)

type Scanner struct {
	source   string
	tokens   []token.Token
	start    int
	current  int
	line     int
	keywords map[string]token.TokenType
}

type ScannerError struct {
	line    int
	where   string
	message string
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s", e.line, e.where, e.message)
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source:  source,
		tokens:  make([]token.Token, 0),
		start:   0,
		current: 0,
		line:    1,
		keywords: map[string]token.TokenType{
			"and":    token.AND,
			"class":  token.CLASS,
			"else":   token.ELSE,
			"false":  token.FALSE,
			"fun":    token.FUN,
			"for":    token.FOR,
			"if":     token.IF,
			"nil":    token.NIL,
			"or":     token.OR,
			"print":  token.PRINT,
			"return": token.RETURN,
			"super":  token.SUPER,
			"this":   token.THIS,
			"true":   token.TRUE,
			"var":    token.VAR,
			"while":  token.WHILE,
		},
	}
}

func (s *Scanner) ScanTokens() ([]token.Token, error) {
	var err error
	for !s.atEnd() && err == nil {
		s.start = s.current
		err = s.scanToken()
	}
	if err != nil {
		return s.tokens, err
	}
	s.tokens = append(s.tokens, token.NewToken(token.EOF, "", nil, s.line))
	return s.tokens, nil
}

func (s *Scanner) atEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() error {
	c := s.advance()
	switch c {
	case '(':
		s.addToken(token.LEFT_PAREN)
	case ')':
		s.addToken(token.RIGHT_PAREN)
	case '{':
		s.addToken(token.LEFT_BRACE)
	case '}':
		s.addToken(token.RIGHT_BRACE)
	case ',':
		s.addToken(token.COMMA)
	case '.':
		s.addToken(token.DOT)
	case '-':
		s.addToken(token.MINUS)
	case '+':
		s.addToken(token.PLUS)
	case ';':
		s.addToken(token.SEMICOLON)
	case '*':
		s.addToken(token.STAR)
	case '?':
		s.addToken(token.QMARK)
	case ':':
		s.addToken(token.COLON)
	case '!':
		if s.match('=') {
			s.addToken(token.BANG_EQUAL)
		} else {
			s.addToken(token.BANG)
		}
	case '=':
		if s.match('=') {
			s.addToken(token.EQUAL_EQUAL)
		} else {
			s.addToken(token.EQUAL)
		}
	case '<':
		if s.match('=') {
			s.addToken(token.LESS_EQUAL)
		} else {
			s.addToken(token.LESS)
		}
	case '>':
		if s.match('=') {
			s.addToken(token.GREATER_EQUAL)
		} else {
			s.addToken(token.GREATER)
		}
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.atEnd() {
				s.advance()
				// comments are ignored in the parser, so we don't add a token for them
			}
		} else if s.match('*') {
			nestLevel := 1
			for nestLevel > 0 && !s.atEnd() {
				if s.match('*') && s.match('/') {
					nestLevel--
				} else if s.match('/') && s.match('*') {
					nestLevel++
				} else {
					s.advance()
				}
				// comments are ignored
			}
		} else {
			s.addToken(token.SLASH)
		}
	case ' ':
		fallthrough
	case '\r':
		fallthrough
	case '\t':
		break
	case '\n':
		s.line++
	case '"':
		err := s.string()
		if err != nil {
			return err
		}
	default:
		if isDigit(c) {
			s.number()
		} else if isAlpha(c) {
			s.identifier()
		} else {
			return &ScannerError{s.line, "", fmt.Sprintf("Unexpected character '%s'\n", string(c))}
		}
	}
	return nil
}

func (s *Scanner) advance() rune {
	r := s.source[s.current]
	s.current++
	return rune(r)
}

func (s *Scanner) addToken(t token.TokenType) {
	s.addTokenWithLiteral(t, nil)
}

func (s *Scanner) addTokenWithLiteral(t token.TokenType, literal token.LiteralObject) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, token.NewToken(t, text, literal, s.line))
}

func (s *Scanner) match(expected rune) bool {
	if s.atEnd() {
		return false
	}
	if rune(s.source[s.current]) != expected {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.atEnd() {
		return rune(0)
	}
	return rune(s.source[s.current])

}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return rune(0)
	}
	return rune(s.source[s.current+1])
}

func (s *Scanner) string() error {
	for s.peek() != '"' && !s.atEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}
	if s.atEnd() {
		return &ScannerError{s.line, "", "Unterminated string"}
	}

	s.advance()

	str := s.source[s.start+1 : s.current-1]
	s.addTokenWithLiteral(token.STRING, str)
	return nil
}

func isDigit(c rune) bool {
	return strings.ContainsRune("0123456789", c)
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advance()
	}
	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()
		for isDigit(s.peek()) {
			s.advance()
		}
	}

	value, err := strconv.ParseFloat(s.source[s.start:s.current], 64)
	if err != nil {
		// it shouldn't be possible to reach this, we've already scanned the number. if we do: panic
		panic(err)
	}

	s.addTokenWithLiteral(token.NUMBER, value)
}

func (s *Scanner) identifier() {
	for isAlphanumeric(s.peek()) {
		s.advance()
	}

	value := s.source[s.start:s.current]
	tokenType, ok := s.keywords[value]
	if !ok {
		tokenType = token.IDENTIFIER
	}

	s.addToken(tokenType)
}

func isAlpha(c rune) bool {
	return strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", c)
}

func isAlphanumeric(c rune) bool {
	return isDigit(c) || isAlpha(c)
}
