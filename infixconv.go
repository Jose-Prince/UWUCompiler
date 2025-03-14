package main

import (
	l "github.com/Jose-Prince/UWULexer/lib"
)

// Converts an infix expression into an array of tokens
func InfixToTokens(infix string) []l.RX_Token {
	previousCanBeANDedTo := false
	tokens := []l.RX_Token{}

	runes := []rune(infix)
	for i := 0; i < len(runes); i++ {
		currentChar := runes[i]
		var token l.RX_Token

		switch currentChar {
		case '|':
			token = l.CreateOperatorToken(l.OR)
			previousCanBeANDedTo = false

		case '*':
			token = l.CreateOperatorToken(l.ZERO_OR_MANY)
			previousCanBeANDedTo = true

		case '(':
			if previousCanBeANDedTo {
				tokens = append(tokens, l.CreateOperatorToken(l.AND))
			}
			token = l.CreateOperatorToken(l.LEFT_PAREN)
			previousCanBeANDedTo = false

		case ')':
			token = l.CreateOperatorToken(l.RIGHT_PAREN)
			previousCanBeANDedTo = true

		case '[':
			if previousCanBeANDedTo {
				tokens = append(tokens, l.CreateOperatorToken(l.AND))
			}

			token = l.CreateOperatorToken(l.LEFT_BRACKET)

			nextRune := runes[i+1]
			if '^' == nextRune {
				token = l.CreateOperatorToken(l.SET_NEGATION)
				i++
			}

			previousCanBeANDedTo = false

		case ']':
			token = l.CreateOperatorToken(l.RIGHT_BRACKET)
			previousCanBeANDedTo = true

		case '+':
			token = l.CreateOperatorToken(l.ONE_OR_MANY)
			previousCanBeANDedTo = true

		case '?':
			token = l.CreateOperatorToken(l.OPTIONAL)
			previousCanBeANDedTo = true

		case '\\':
			if previousCanBeANDedTo {
				tokens = append(tokens, l.CreateOperatorToken(l.AND))
			}

			token = l.CreateValueToken(rune(runes[i+1]))
			i++
			previousCanBeANDedTo = true
		default:
			if previousCanBeANDedTo {
				tokens = append(tokens, l.CreateOperatorToken(l.AND))
			}

			token = l.CreateValueToken(rune(currentChar))
			previousCanBeANDedTo = true
		}

		tokens = append(tokens, token)
	}

	return tokens
}
