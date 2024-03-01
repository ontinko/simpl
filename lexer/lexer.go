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
	'=': tokens.EQUAL,
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
		if singleChars[c] != 0 {
			token := tokens.NewToken(singleChars[c], "", filename, line, start-lineStart+1)
			result = append(result, token)
			start++
			continue
		}
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
		case ':':
			if peek(&source, start+1) == '=' {
				token := tokens.NewToken(tokens.COLON_EQUAL, "", filename, line, start-lineStart+1)
				result = append(result, token)
				start += 2
				continue
			}
			token := tokens.NewToken(tokens.UNPERMITTED, source[start:start+1], filename, line, start-lineStart+1)
			result = append(result, token)
			errs = append(errs, errors.Error{Message: "Unpermitted character", Token: token, Type: errors.SyntaxError})
			start++
		default:
			if isDigit(c) {
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
				default:
					token = tokens.NewToken(tokens.IDENTIFIER, source[start:end], filename, line, start-lineStart+1)
				}
				result = append(result, token)
				start = end
			} else {
				token := tokens.NewToken(tokens.UNPERMITTED, source[start:start+1], filename, line, start-lineStart+1)
				errs = append(errs, errors.Error{Message: "Unpermitted character", Token: token, Type: errors.SyntaxError})
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
