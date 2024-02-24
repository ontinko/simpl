package tokens

import "fmt"

type TokenType int

const (
	PLUS TokenType = iota + 1
	MINUS
	STAR
	SLASH
	EQUAL
	COLON_EQUAL
	SEMICOLON

	IDENTIFIER
	NUMBER

	LEFT_BRACE
	RIGHT_BRACE
	LEFT_PAREN
	RIGHT_PAREN

	UNPERMITTED
	EOF
)

var Representations map[TokenType]string = map[TokenType]string{
	PLUS:        "PLUS: +",
	MINUS:       "MINUS: -",
	STAR:        "STAR: *",
	SLASH:       "SLASH: /",
	EQUAL:       "EQUAL: =",
	COLON_EQUAL: "COLON_EQUAL: :=",
	SEMICOLON:   "SEMICOLON: ;",
	LEFT_BRACE:  "LEFT_BRACE: {",
	RIGHT_BRACE: "RIGHT_BRACE: }",
	LEFT_PAREN:  "LEFT_PAREN: (",
	RIGHT_PAREN: "RIGHT_PAREN: )",
    EOF: "EOF",
}

var Precedences map[TokenType]int = map[TokenType]int{
	PLUS:  3,
	MINUS: 3,
	STAR:  4,
	SLASH: 4,
    EQUAL: 2,
    COLON_EQUAL: 2,
    NUMBER: 1,
    IDENTIFIER: 1,
    EOF: -1,
}

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Char  int
}

func NewToken(tType TokenType, value string, line, char int) Token {
	return Token{tType, value, line, char}
}

func (t *Token) Print() {
	value, found := Representations[t.Type]
	if found {
		fmt.Println(value)
		return
	}
	switch t.Type {
	case IDENTIFIER:
		fmt.Printf("IDENTIFIER: %s\n", t.Value)
	case NUMBER:
		fmt.Printf("NUMBER: %s\n", t.Value)
	default:
		fmt.Printf("UNPERMITTED: %s\n", t.Value)
	}
}
