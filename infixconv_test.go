package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	l "github.com/Jose-Prince/UWULexer/lib"
)

func printSideBySide(t *testing.T, markedIdx int, expected []l.RX_Token, result []l.RX_Token) {
	maxLength := max(len(expected), len(result))
	header1 := "EXPECTED:"
	header2 := "VALUE:"
	header3 := "IDX:"
	maxLeftLength := 20
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("\n%-*s%-*s%s\n", maxLeftLength, header1, maxLeftLength, header2, header3))

	for i := range maxLength {

		if i == markedIdx {
			b.WriteString("\033[31m")
		}

		if i < len(expected)-1 {
			elem := expected[i].ToString()
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, elem))
		} else {
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, "<N/A>"))
		}

		if i < len(result)-1 {
			elem := result[i].ToString()
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, elem))
		} else {
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, "<N/A>"))
		}

		b.WriteString(fmt.Sprintf("%-*d", maxLeftLength, i))
		b.WriteString("\033[0m")
		b.WriteRune('\n')
	}

	t.Log(b.String())
}

func compareTokensStreams(t *testing.T, originalInfix string, expected []l.RX_Token, result []l.RX_Token) {
	for i, elem := range expected {
		resultElem := result[i]

		if !elem.Equals(&resultElem) {
			t.Logf("ORIGINAL: %s", originalInfix)
			t.Logf("EXPECTED (%s) != RESULT: (%s) IDX: %d", elem.ToString(), resultElem.ToString(), i)
			printSideBySide(t, i, expected, result)
			t.FailNow()
		}
	}
}

func TestSimpleExpression(t *testing.T) {
	infix := "(a|b)c"
	result := InfixToTokens(infix)
	expected := []l.RX_Token{
		l.CreateValueToken('('),
		l.CreateValueToken('a'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('b'),
		l.CreateValueToken(')'),
		l.CreateOperatorToken(l.AND),
		l.CreateValueToken('c'),
	}

	compareTokensStreams(t, infix, expected, result)
}

func fromTokenStreamToInfix(stream []l.RX_Token) string {
	b := strings.Builder{}

	for _, elem := range stream {
		if elem.IsOperator() {
			switch elem.GetOperator() {
			case l.OR:
				b.WriteByte('|')
			case l.ZERO_OR_MANY:
				b.WriteByte('*')
			default:
			}

		} else {
			rune := elem.GetValue().GetValue()
			switch rune {
			case '|', '*', '.', '(', ')', '[', ']':
				b.WriteRune('\\')
			default:
			}
			b.WriteRune(rune)
		}
	}

	return b.String()
}

func generateExpected(random *rand.Rand) []l.RX_Token {
	expressionCount := random.Intn(100)
	tokens := []l.RX_Token{}
	possibleChars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789,.-;:_¿?¡!'{}+*|\"#$%&/()=[]<>°¬")
	getRandomRune := func() rune {
		return possibleChars[random.Intn(len(possibleChars))]
	}
	getRandomTwoOp := func() l.Operator {
		switch random.Intn(2) {
		case 0:
			return l.OR
		default:
			return l.AND
		}
	}

	for i := range expressionCount {
		switch random.Intn(5) {
		default: // Simple two value expression
			a := l.CreateValueToken(getRandomRune())
			b := l.CreateValueToken(getRandomRune())
			op := l.CreateOperatorToken(getRandomTwoOp())

			tokens = append(tokens, a)
			tokens = append(tokens, op)
			tokens = append(tokens, b)
		}

		if i+1 < expressionCount {
			addStar := random.Intn(2) == 0
			if addStar {
				tokens = append(tokens, l.CreateOperatorToken(l.ZERO_OR_MANY))
			}

			tokens = append(tokens, l.CreateOperatorToken(getRandomTwoOp()))
		}
	}
	addStar := random.Intn(2) == 0
	if addStar {
		tokens = append(tokens, l.CreateOperatorToken(l.ZERO_OR_MANY))
	}

	return tokens
}

func FuzzInfixExpr(f *testing.F) {
	f.Add(int64(69420))
	f.Fuzz(func(t *testing.T, seed int64) {
		source := rand.NewSource(seed)
		random := rand.New(source)

		expected := generateExpected(random)
		infix := fromTokenStreamToInfix(expected)
		result := InfixToTokens(infix)

		compareTokensStreams(t, infix, expected, result)
	})
}
