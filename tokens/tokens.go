package tokens

type TokenType int

const (
	UNPERMITTED TokenType = iota
	EOF

	COMMA

	PLUS
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

	INT_TYPE
	BOOL_TYPE

	DEF
	RETURN
)

var Representations map[TokenType]string = map[TokenType]string{
	COMMA:  ",",
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

	INT_TYPE:  "int",
	BOOL_TYPE: "bool",

	DEF:    "def",
	RETURN: "return",

	SEMICOLON: ";",
	EOF:       "EOF",
}

var Precedences map[TokenType]int = map[TokenType]int{
	EOF: -1,

	NUMBER:     1,
	IDENTIFIER: 1,
	TRUE:       1,
	FALSE:      1,

	LESS:         2,
	GREATER:      2,
	DOUBLE_EQUAL: 2,
	NOT_EQUAL:    2,

	PLUS:   3,
	MINUS:  3,
	STAR:   4,
	SLASH:  4,
	MODULO: 5,

	OR:  6,
	AND: 7,

	BANG: 8,
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
