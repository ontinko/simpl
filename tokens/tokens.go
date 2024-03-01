package tokens

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

	TRUE
	FALSE

	BANG
	OR
	AND

	LEFT_BRACE
	RIGHT_BRACE
	LEFT_PAREN
	RIGHT_PAREN

	UNPERMITTED
	EOF
)

var Representations map[TokenType]string = map[TokenType]string{
	PLUS:        "+",
	MINUS:       "-",
	STAR:        "*",
	SLASH:       "/",
	EQUAL:       "=",
	COLON_EQUAL: ":=",
	SEMICOLON:   ";",
	LEFT_BRACE:  "{",
	RIGHT_BRACE: "}",
	LEFT_PAREN:  "(",
	RIGHT_PAREN: ")",
	BANG:        "!",
	TRUE:        "true",
	FALSE:       "false",
	OR:          "OR",
	AND:         "AND",
	EOF:         "EOF",
}

var Precedences map[TokenType]int = map[TokenType]int{
	PLUS:       2,
	MINUS:      2,
	STAR:       3,
	SLASH:      3,
	OR:         4,
	AND:        5,
	BANG:       6,
	NUMBER:     1,
	IDENTIFIER: 1,
	TRUE:       1,
	FALSE:      1,
	EOF:        -1,
}

type Token struct {
	Type     TokenType
	Value    string
	Line     int
	Char     int
	Filename string
}

func NewToken(tType TokenType, value string, filename string, line, char int) Token {
	return Token{tType, value, line, char, filename}
}

func (t *Token) View() string {
	value, found := Representations[t.Type]
	if found {
		return value
	}
	return t.Value
}
