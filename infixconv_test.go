package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	l "github.com/Jose-Prince/UWUCompiler/lib/regex"
)

func PrintSideBySide(t *testing.T, markedIdx int, expected []l.RX_Token, result []l.RX_Token) {
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

		if i < len(expected) {
			elem := expected[i].String()
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, elem))
		} else {
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, "<N/A>"))
		}

		if i < len(result) {
			elem := result[i].String()
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
		if i >= len(result) {
			t.Logf("ORIGINAL: %s", originalInfix)
			t.Logf("EXPECTED (%s) != RESULT: (< No value on idx >) IDX: %d", elem.String(), i)
			PrintSideBySide(t, i, expected, result)
			t.FailNow()
		}

		resultElem := result[i]

		if !elem.Equals(&resultElem) {
			t.Logf("ORIGINAL: %s", originalInfix)
			t.Logf("EXPECTED (%s) != RESULT: (%s) IDX: %d", elem.String(), resultElem.String(), i)
			PrintSideBySide(t, i, expected, result)
			t.FailNow()
		}
	}
}

func TestSimpleExpression(t *testing.T) {
	infix := "(a|b)c"
	result := InfixToTokens(infix)
	expected := []l.RX_Token{
		l.CreateOperatorToken(l.LEFT_PAREN),
		l.CreateValueToken('a'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('b'),
		l.CreateOperatorToken(l.RIGHT_PAREN),
		l.CreateOperatorToken(l.AND),
		l.CreateValueToken('c'),
	}

	compareTokensStreams(t, infix, expected, result)
}

func fromTokenStreamToInfixString(stream []l.RX_Token) string {
	b := strings.Builder{}

	for _, elem := range stream {
		if elem.IsOperator() {
			switch elem.GetOperator() {
			case l.OR:
				b.WriteByte('|')
			case l.ZERO_OR_MANY:
				b.WriteByte('*')
			case l.ONE_OR_MANY:
				b.WriteByte('+')
			case l.OPTIONAL:
				b.WriteByte('?')
			case l.LEFT_PAREN:
				b.WriteByte('(')
			case l.RIGHT_PAREN:
				b.WriteByte(')')
			case l.AND:
				// Ignore it since it's implicit...
			default:
				b.WriteString("<INVALID OPERATOR>")
			}

		} else {
			rune := elem.GetValue().GetValue()
			switch rune {
			case '|', '*', '.', '(', ')', '[', ']', '+', '?':
				b.WriteRune('\\')
			default:
			}
			b.WriteRune(rune)
		}
	}

	return b.String()
}

func generateExpectedInfix(random *rand.Rand) []l.RX_Token {
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
	getRandomOneOp := func() l.Operator {
		switch random.Intn(3) {
		case 0:
			return l.ZERO_OR_MANY
		case 1:
			return l.ONE_OR_MANY
		default:
			return l.OPTIONAL
		}
	}

	for i := range expressionCount {
		switch random.Intn(5) {
		case 0: // Between parenthesis
			a := l.CreateValueToken(getRandomRune())
			b := l.CreateValueToken(getRandomRune())
			op := l.CreateOperatorToken(getRandomTwoOp())

			tokens = append(tokens, l.CreateOperatorToken(l.LEFT_PAREN))
			tokens = append(tokens, a)
			tokens = append(tokens, op)
			tokens = append(tokens, b)
			tokens = append(tokens, l.CreateOperatorToken(l.RIGHT_PAREN))

		case 1: // Nested parenthesis
			a := l.CreateValueToken(getRandomRune())
			b := l.CreateValueToken(getRandomRune())
			c := l.CreateValueToken(getRandomRune())
			d := l.CreateValueToken(getRandomRune())
			op1 := l.CreateOperatorToken(getRandomTwoOp())
			op2 := l.CreateOperatorToken(getRandomTwoOp())
			op3 := l.CreateOperatorToken(getRandomTwoOp())

			tokens = append(tokens, l.CreateOperatorToken(l.LEFT_PAREN))
			tokens = append(tokens, l.CreateOperatorToken(l.LEFT_PAREN))
			tokens = append(tokens, a)
			tokens = append(tokens, op1)
			tokens = append(tokens, b)
			tokens = append(tokens, l.CreateOperatorToken(l.RIGHT_PAREN))

			tokens = append(tokens, op3)

			tokens = append(tokens, l.CreateOperatorToken(l.LEFT_PAREN))
			tokens = append(tokens, c)
			tokens = append(tokens, op2)
			tokens = append(tokens, d)
			tokens = append(tokens, l.CreateOperatorToken(l.RIGHT_PAREN))
			tokens = append(tokens, l.CreateOperatorToken(l.RIGHT_PAREN))

		default: // Simple two value expression
			a := l.CreateValueToken(getRandomRune())
			b := l.CreateValueToken(getRandomRune())
			op := l.CreateOperatorToken(getRandomTwoOp())

			tokens = append(tokens, a)
			tokens = append(tokens, op)
			tokens = append(tokens, b)
		}

		if i+1 < expressionCount {
			addOneOp := random.Intn(2) == 0
			if addOneOp {
				tokens = append(tokens, l.CreateOperatorToken(getRandomOneOp()))
			}

			tokens = append(tokens, l.CreateOperatorToken(getRandomTwoOp()))
		}
	}

	addOneOp := random.Intn(2) == 0
	if addOneOp {
		tokens = append(tokens, l.CreateOperatorToken(getRandomOneOp()))
	}

	return tokens
}

func FuzzFromInfixToRegex(f *testing.F) {
	f.Add(int64(69420))
	f.Fuzz(func(t *testing.T, seed int64) {
		source := rand.NewSource(seed)
		random := rand.New(source)

		expected := generateExpectedInfix(random)
		infix := fromTokenStreamToInfixString(expected)
		result := InfixToTokens(infix)

		compareTokensStreams(t, infix, expected, result)
	})
}

func TestPythonExample(t *testing.T) {
	infix := "[0-9]+"
	result := InfixToTokens(infix)
	expected := []l.RX_Token{
		l.CreateOperatorToken(l.LEFT_PAREN),
		l.CreateValueToken('0'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('1'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('2'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('3'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('4'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('5'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('6'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('7'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('8'),
		l.CreateOperatorToken(l.OR),
		l.CreateValueToken('9'),
		l.CreateOperatorToken(l.RIGHT_PAREN),
		l.CreateOperatorToken(l.ONE_OR_MANY),
	}

	compareTokensStreams(t, infix, expected, result)

}
