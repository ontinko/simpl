package lexer

import (
	"simpl/errors"
	"simpl/tokens"
)

var singleChars = map[byte]tokens.TokenType{
	'+': tokens.PLUS,
	'-': tokens.MINUS,
	'*': tokens.STAR,
	'/': tokens.SLASH,
	'%': tokens.MODULO,
	';': tokens.SEMICOLON,
	'{': tokens.LEFT_BRACE,
	'}': tokens.RIGHT_BRACE,
	'(': tokens.LEFT_PAREN,
	')': tokens.RIGHT_PAREN,
}

func Tokenize(source string, filename string, line int) ([]tokens.Token, []errors.Error) {
	result := []tokens.Token{}
	errs := []errors.Error{}
	start := 0
	lineStart := 0
	sourceSize := len(source)
	for start < sourceSize {
		c := source[start]
		switch c {
		case '\n':
			line++
			start++
			lineStart = start
		case ' ':
			start++
		case '#':
			newStart := skipComment(&source, start)
			start = newStart
		case '+':
			next := peek(&source, start+1)
			var token tokens.Token
			if next == '=' {
				token = tokens.NewToken(tokens.PLUS_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			} else if next == '+' {
				token = tokens.NewToken(tokens.DOUBLE_PLUS, "", filename, line, start-lineStart+1)
				start += 2
			} else {
				token = tokens.NewToken(tokens.PLUS, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '-':
			next := peek(&source, start+1)
			var token tokens.Token
			if next == '=' {
				token = tokens.NewToken(tokens.MINUS_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			} else if next == '-' {
				token = tokens.NewToken(tokens.DOUBLE_MINUS, "", filename, line, start-lineStart+1)
				start += 2
			} else if isDigit(next) {
				var newStart int
				token, newStart = readNumber(&source, filename, line, start, lineStart)
				start = newStart
			} else {
				start++
				token = tokens.NewToken(tokens.MINUS, "", filename, line, start-lineStart+1)
			}
			result = append(result, token)
		case '*':
			next := peek(&source, start+1)
			var token tokens.Token
			if next == '=' {
				token = tokens.NewToken(tokens.STAR_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			} else {
				token = tokens.NewToken(tokens.STAR, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '/':
			next := peek(&source, start+1)
			var token tokens.Token
			if next == '=' {
				token = tokens.NewToken(tokens.SLASH_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			} else {
				token = tokens.NewToken(tokens.SLASH, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '%':
			next := peek(&source, start+1)
			var token tokens.Token
			if next == '=' {
				token = tokens.NewToken(tokens.MODULO_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			} else {
				token = tokens.NewToken(tokens.MODULO, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case ':':
			if peek(&source, start+1) == '=' {
				token := tokens.NewToken(tokens.COLON_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			}
			token := tokens.NewToken(tokens.UNPERMITTED, source[start:start+1], filename, line, start-lineStart+1)
			result = append(result, token)
			errs = append(errs, errors.Error{Message: "unpermitted character", Token: token, Type: errors.SyntaxError})
			start++
		case '=':
			var token tokens.Token
			if peek(&source, start+1) == '=' {
				token = tokens.NewToken(tokens.DOUBLE_EQUAL, "", filename, line, start-lineStart+1)
				start += 2
			} else {
				token = tokens.NewToken(tokens.EQUAL, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '>':
			var token tokens.Token
			if peek(&source, start+1) == '=' {
				token = tokens.NewToken(tokens.GREATER_EQUAL, "", filename, line, start-lineStart+1)
				start += 2
			} else {
				token = tokens.NewToken(tokens.GREATER, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '<':
			var token tokens.Token
			if peek(&source, start+1) == '=' {
				token = tokens.NewToken(tokens.LESS_EQUAL, "", filename, line, start-lineStart+1)
				start += 2
			} else {
				token = tokens.NewToken(tokens.LESS, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '!':
			var token tokens.Token
			if peek(&source, start+1) == '=' {
				token = tokens.NewToken(tokens.NOT_EQUAL, "", filename, line, start-lineStart+1)
				start += 2
			} else {
				token = tokens.NewToken(tokens.BANG, "", filename, line, start-lineStart+1)
				start++
			}
			result = append(result, token)
		case '|':
			if peek(&source, start+1) == '|' {
				token := tokens.NewToken(tokens.OR, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			}
			token := tokens.NewToken(tokens.UNPERMITTED, source[start:start+1], filename, line, start-lineStart+1)
			result = append(result, token)
			errs = append(errs, errors.Error{Message: "unpermitted character", Token: token, Type: errors.SyntaxError})
		case '&':
			if peek(&source, start+1) == '&' {
				token := tokens.NewToken(tokens.AND, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			}
			token := tokens.NewToken(tokens.UNPERMITTED, source[start:start+1], filename, line, start-lineStart+1)
			result = append(result, token)
			errs = append(errs, errors.Error{Message: "unpermitted character", Token: token, Type: errors.SyntaxError})
		default:
			if singleChars[c] != 0 {
				token := tokens.NewToken(singleChars[c], "", filename, line, start-lineStart+1)
				result = append(result, token)
				start++
			} else if isDigit(c) {
				token, newStart := readNumber(&source, filename, line, start, lineStart)
				result = append(result, token)
				start = newStart

			} else if isAlpha(c) {
				end := readAlphaNumeric(&source, start)
				var token tokens.Token
				switch source[start:end] {
				case "true":
					token = tokens.NewToken(tokens.TRUE, "", filename, line, start-lineStart+1)
				case "false":
					token = tokens.NewToken(tokens.FALSE, "", filename, line, start-lineStart+1)
				case "if":
					token = tokens.NewToken(tokens.IF, "", filename, line, start-lineStart+1)
				case "while":
					token = tokens.NewToken(tokens.WHILE, "", filename, line, start-lineStart+1)
				case "else":
					token = tokens.NewToken(tokens.ELSE, "", filename, line, start-lineStart+1)
				case "for":
					token = tokens.NewToken(tokens.FOR, "", filename, line, start-lineStart+1)
				case "break":
					token = tokens.NewToken(tokens.BREAK, "", filename, line, start-lineStart+1)
				case "continue":
					token = tokens.NewToken(tokens.CONTINUE, "", filename, line, start-lineStart+1)
				case "int":
					token = tokens.NewToken(tokens.INT_TYPE, "", filename, line, start-lineStart+1)
				case "bool":
					token = tokens.NewToken(tokens.BOOL_TYPE, "", filename, line, start-lineStart+1)
				default:
					token = tokens.NewToken(tokens.IDENTIFIER, source[start:end], filename, line, start-lineStart+1)
				}
				result = append(result, token)
				start = end
			} else {
				token := tokens.NewToken(tokens.UNPERMITTED, source[start:start+1], filename, line, start-lineStart+1)
				errs = append(errs, errors.Error{Message: "unpermitted character", Token: token, Type: errors.SyntaxError})
				result = append(result, token)
				start++
			}
		}
	}
	result = append(result, tokens.NewToken(tokens.EOF, "", filename, line, 0))
	return result, errs
}

func peek(source *string, i int) byte {
	if i >= len(*source) {
		return ' '
	}
	return (*source)[i]
}

func skipComment(source *string, start int) int {
	for start < len(*source) {
		if (*source)[start] == '\n' {
			break
		}
		start++
	}
	return start
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c == '_'
}

func readNumber(source *string, filename string, line, start int, lineStart int) (tokens.Token, int) {
	end := start + 1
	for isDigit(peek(source, end)) {
		end++
	}
	token := tokens.NewToken(tokens.NUMBER, (*source)[start:end], filename, line, start-lineStart+1)
	return token, end
}

func readAlphaNumeric(source *string, start int) int {
	end := start + 1
	for {
		c := (*source)[end]
		if !isDigit(c) && !isAlpha(c) {
			break
		}
		end++
	}
	return end
}

func readIdentifier(source *string, filename string, line, start int, lineStart int) (tokens.Token, int) {
	end := start + 1
	for {
		c := (*source)[end]
		if !isDigit(c) && !isAlpha(c) {
			break
		}
		end++
	}
	token := tokens.NewToken(tokens.IDENTIFIER, (*source)[start:end], filename, line+1, start-lineStart+1)
	return token, end
}
