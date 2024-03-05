package tokens

type TokenType int

const (
	PLUS TokenType = iota + 1
	MINUS
	STAR
	SLASH
	EQUAL
	MODULO

	DOUBLE_PLUS
	DOUBLE_MINUS

	PLUS_EQUAL
	MINUS_EQUAL
	STAR_EQUAL
	SLASH_EQUAL
	MODULO_EQUAL

	COLON_EQUAL
	SEMICOLON

	IDENTIFIER
	NUMBER

	TRUE
	FALSE

	DOUBLE_EQUAL
	NOT_EQUAL
	LESS
	LESS_EQUAL
	GREATER
	GREATER_EQUAL

	BANG
	OR
	AND

	IF
	ELSE
	WHILE
	FOR
	BREAK
	CONTINUE

	LEFT_BRACE
	RIGHT_BRACE
	LEFT_PAREN
	RIGHT_PAREN

	UNPERMITTED
	EOF
)

var Representations map[TokenType]string = map[TokenType]string{
	PLUS:   "+",
	MINUS:  "-",
	STAR:   "*",
	SLASH:  "/",
	MODULO: "%",

	DOUBLE_PLUS:  "++",
	DOUBLE_MINUS: "--",

	PLUS_EQUAL:   "+=",
	MINUS_EQUAL:  "-=",
	STAR_EQUAL:   "*=",
	SLASH_EQUAL:  "/=",
	MODULO_EQUAL: "%=",

	EQUAL:       "=",
	COLON_EQUAL: ":=",

	DOUBLE_EQUAL:  "==",
	NOT_EQUAL:     "!=",
	LESS:          "<",
	GREATER:       ">",
	LESS_EQUAL:    "<=",
	GREATER_EQUAL: ">=",

	LEFT_BRACE:  "{",
	RIGHT_BRACE: "}",
	LEFT_PAREN:  "(",
	RIGHT_PAREN: ")",

	BANG:     "!",
	TRUE:     "true",
	FALSE:    "false",
	OR:       "OR",
	AND:      "AND",
	IF:       "if",
	ELSE:     "else",
	WHILE:    "while",
	FOR:      "for",
	BREAK:    "break",
	CONTINUE: "continue",

	SEMICOLON: ";",
	EOF:       "EOF",
}

var Precedences map[TokenType]int = map[TokenType]int{
	EOF: -1,

	NUMBER:     1,
	IDENTIFIER: 1,
	TRUE:       1,
	FALSE:      1,

	PLUS:  2,
	MINUS: 2,
	STAR:  3,
	SLASH: 3,

	OR:  4,
	AND: 5,

	LESS:         5,
	GREATER:      5,
	DOUBLE_EQUAL: 5,
	NOT_EQUAL:    6,

	BANG: 7,
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
