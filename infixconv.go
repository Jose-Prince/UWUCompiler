package main

import (
	l "github.com/Jose-Prince/UWULexer/lib"
)

// Converts an infix expression into an array of tokens:
// (a|b)c -> [
// { val=( },
// { val=a },
// { operator = OR },
// { val=b },
// { val=) },
// { operator=AND },
// { val=c }
// ]
func InfixToTokens(infix string) []l.RX_Token {
	previousCanBeANDedTo := false
	tokens := []l.RX_Token{}

	for i := range len(infix) {
		currentChar := infix[i]
		var token l.RX_Token

		switch currentChar {
		case '|':
			token = l.CreateOperatorToken(l.OR)
			previousCanBeANDedTo = false
		case '*':
			token = l.CreateOperatorToken(l.ZERO_OR_MANY)
			previousCanBeANDedTo = false
		case '(':
			token = l.CreateValueToken('(')
			previousCanBeANDedTo = false
		case ')':
			token = l.CreateValueToken(')')
			previousCanBeANDedTo = true
		case '\\':
			token = l.CreateValueToken(rune(infix[i+1]))
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
