package main

import (
	"math/rand"
	"testing"

	l "github.com/Jose-Prince/UWULexer/lib"
)

func generateExpectedPostfix(r *rand.Rand) []l.RX_Token {
	expressionCount := r.Intn(100)
	postfixExpr := []l.RX_Token{}
	possibleChars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789,.-;:_¿?¡!'{}+*|\"#$%&/()=[]<>°¬")
	getRandomRune := func() rune {
		return possibleChars[r.Intn(len(possibleChars))]
	}
	getRandomTwoOp := func() l.Operator {
		switch r.Intn(2) {
		case 0:
			return l.OR
		default:
			return l.AND
		}
	}
	getRandomOneOp := func() l.Operator {
		switch r.Intn(3) {
		case 0:
			return l.ZERO_OR_MANY
		case 1:
			return l.ONE_OR_MANY
		default:
			return l.OPTIONAL
		}
	}

	for i := range expressionCount {
		switch r.Intn(5) {
		default: // Simple two value expression
			a := l.CreateValueToken(getRandomRune())
			b := l.CreateValueToken(getRandomRune())
			op := l.CreateOperatorToken(getRandomTwoOp())

			postfixExpr = append(postfixExpr, a)
			postfixExpr = append(postfixExpr, b)
			postfixExpr = append(postfixExpr, op)
		}

		addOneOp := r.Intn(2) == 0
		if addOneOp {
			postfixExpr = append(postfixExpr, l.CreateOperatorToken(getRandomOneOp()))
		}

		if i > 0 {
			postfixExpr = append(postfixExpr, l.CreateOperatorToken(getRandomTwoOp()))
		}
	}

	return postfixExpr
}

func fromPostfixToInfix(postfix *[]l.RX_Token) []l.RX_Token {
	stack := l.Stack[[]l.RX_Token]{}

	for _, elem := range *postfix {
		if elem.IsOperator() {
			op := elem.GetOperator()
			switch op {
			case l.OR, l.AND:
				b := stack.Pop()
				a := stack.Pop()

				combined := []l.RX_Token{l.CreateOperatorToken(l.LEFT_PAREN)}

				for _, aElm := range a.GetValue() {
					combined = append(combined, aElm)
				}

				combined = append(combined, elem)

				for _, bElm := range b.GetValue() {
					combined = append(combined, bElm)
				}

				combined = append(combined, l.CreateOperatorToken(l.RIGHT_PAREN))
				stack.Push(combined)

			case l.ZERO_OR_MANY, l.ONE_OR_MANY, l.OPTIONAL:
				a := stack.Pop()

				combined := []l.RX_Token{l.CreateOperatorToken(l.LEFT_PAREN)}
				for _, aElm := range a.GetValue() {
					combined = append(combined, aElm)
				}
				combined = append(combined, l.CreateOperatorToken(l.RIGHT_PAREN))

				combined = append(combined, elem)
				stack.Push(combined)
			default:
				panic("No brackets/parenthesis or set negation are allowed when the expression is postfix!")
			}

		} else {
			stack.Push([]l.RX_Token{elem})
		}
	}

	return []l.RX_Token{}
}

func FuzzInfixToPostfix(f *testing.F) {
	f.Add(int64(69420))
	f.Fuzz(func(t *testing.T, seed int64) {
		source := rand.NewSource(seed)
		random := rand.New(source)

		expected := generateExpectedPostfix(random)
		infixExpr := fromPostfixToInfix(&expected)
		result := DEFAULT_ALPHABET.ToPostfix(&infixExpr)

		infixStr := fromTokenStreamToInfixString(&infixExpr)
		compareTokensStreams(t, infixStr, expected, result)
	})
}
