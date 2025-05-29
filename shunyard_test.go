package main

import (
	"math/rand"
	"testing"

	l "github.com/Jose-Prince/UWUCompiler/lib"
	reg "github.com/Jose-Prince/UWUCompiler/lib/regex"
)

func generateExpectedPostfix(r *rand.Rand) []reg.RX_Token {
	expressionCount := r.Intn(2) + 1 // Minimum of 1 expressions
	postfixExpr := []reg.RX_Token{}
	possibleChars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789,.-;:_¿?¡!'{}+*|\"#$%&/()=[]<>°¬")
	getRandomRune := func() rune {
		return possibleChars[r.Intn(len(possibleChars))]
	}
	getRandomTwoOp := func() reg.Operator {
		switch r.Intn(2) {
		case 0:
			return reg.OR
		default:
			return reg.AND
		}
	}

	for i := range expressionCount {
		switch r.Intn(5) {
		default: // Simple two value expression
			a := reg.CreateValueToken(getRandomRune())
			b := reg.CreateValueToken(getRandomRune())
			op := reg.CreateOperatorToken(getRandomTwoOp())

			postfixExpr = append(postfixExpr, a)
			postfixExpr = append(postfixExpr, b)
			postfixExpr = append(postfixExpr, op)
		}

		addOneOp := r.Intn(2) == 0
		if addOneOp {
			postfixExpr = append(postfixExpr, reg.CreateOperatorToken(reg.ZERO_OR_MANY))
		}

		if i > 0 {
			postfixExpr = append(postfixExpr, reg.CreateOperatorToken(getRandomTwoOp()))
		}
	}

	return postfixExpr
}

func fromPostfixToInfix(postfix []reg.RX_Token) []reg.RX_Token {
	stack := l.Stack[[]reg.RX_Token]{}

	for _, elem := range postfix {
		if elem.IsOperator() {
			op := elem.GetOperator()
			switch op {
			case reg.OR, reg.AND:
				b := stack.Pop()
				a := stack.Pop()

				combined := []reg.RX_Token{reg.CreateOperatorToken(reg.LEFT_PAREN)}
				combined = append(combined, a.GetValue()...)
				combined = append(combined, elem)
				combined = append(combined, b.GetValue()...)
				combined = append(combined, reg.CreateOperatorToken(reg.RIGHT_PAREN))

				stack.Push(combined)

			case reg.ZERO_OR_MANY, reg.ONE_OR_MANY, reg.OPTIONAL:
				a := stack.Pop()

				combined := []reg.RX_Token{reg.CreateOperatorToken(reg.LEFT_PAREN)}
				combined = append(combined, a.GetValue()...)
				combined = append(combined, reg.CreateOperatorToken(reg.RIGHT_PAREN))

				combined = append(combined, elem)
				stack.Push(combined)
			default:
				panic("No brackets/parenthesis or set negation are allowed when the expression is postfix!")
			}

		} else {
			stack.Push([]reg.RX_Token{elem})
		}
	}

	return stack.Pop().GetValue()
}

func FuzzInfixToPostfix(f *testing.F) {
	f.Add(int64(69420))
	f.Fuzz(func(t *testing.T, seed int64) {
		source := rand.NewSource(seed)
		random := rand.New(source)

		expected := generateExpectedPostfix(random)
		infixExpr := fromPostfixToInfix(expected)
		infixStr := fromTokenStreamToInfixString(infixExpr)

		result := DEFAULT_ALPHABET.ToPostfix(&infixExpr)
		compareTokensStreams(t, infixStr, expected, result)
	})
}

func TestDummyTokens(t *testing.T) {
	dummyCode := "Hello"
	expected := []reg.RX_Token{
		reg.CreateValueToken('a'),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateDummyToken(reg.DummyInfo{Code: dummyCode}),
		reg.CreateOperatorToken(reg.AND),
	}
	infix := []reg.RX_Token{
		reg.CreateValueToken('a'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateDummyToken(reg.DummyInfo{Code: dummyCode}),
	}
	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "a|b (Dummy token)", expected, result)
}

func TestZeroOrManyOperator(t *testing.T) {
	expected := []reg.RX_Token{
		reg.CreateValueToken('a'),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('a'),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateOperatorToken(reg.ZERO_OR_MANY),
		reg.CreateOperatorToken(reg.AND),
	}
	infix := []reg.RX_Token{
		reg.CreateOperatorToken(reg.LEFT_PAREN),
		reg.CreateValueToken('a'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.RIGHT_PAREN),
		reg.CreateOperatorToken(reg.ONE_OR_MANY),
	}
	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "(a|b)+", expected, result)
}

func TestOptionalOperator(t *testing.T) {
	expected := []reg.RX_Token{
		reg.CreateValueToken('a'),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateEpsilonToken(),
		reg.CreateOperatorToken(reg.OR),
	}
	infix := []reg.RX_Token{
		reg.CreateOperatorToken(reg.LEFT_PAREN),
		reg.CreateValueToken('a'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.RIGHT_PAREN),
		reg.CreateOperatorToken(reg.OPTIONAL),
	}
	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "(ab)?", expected, result)
}

func TestCanvasExample(t *testing.T) {
	infix := []reg.RX_Token{
		reg.CreateOperatorToken(reg.LEFT_PAREN),
		reg.CreateValueToken('a'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.RIGHT_PAREN),
		reg.CreateOperatorToken(reg.ZERO_OR_MANY),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateValueToken('a'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateValueToken('b'),
	}
	expected := []reg.RX_Token{
		reg.CreateValueToken('a'),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateOperatorToken(reg.ZERO_OR_MANY),
		reg.CreateValueToken('a'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.AND),
		reg.CreateValueToken('b'),
		reg.CreateOperatorToken(reg.AND),
	}

	result := DEFAULT_ALPHABET.ToPostfix(&infix)
	compareTokensStreams(t, "(a|b)*abb", expected, result)
}

func TestPythonFromRegex(t *testing.T) {
	infix := "[0-9]+"
	infixExpr := DEFAULT_ALPHABET.InfixToTokens(infix)
	expected := []reg.RX_Token{
		reg.CreateOperatorToken(reg.LEFT_PAREN),
		reg.CreateValueToken('0'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('1'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('2'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('3'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('4'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('5'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('6'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('7'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('8'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('9'),
		reg.CreateOperatorToken(reg.RIGHT_PAREN),
		reg.CreateOperatorToken(reg.ONE_OR_MANY),
	}

	compareTokensStreams(t, infix, expected, infixExpr)

	expectedRes := []reg.RX_Token{
		reg.CreateValueToken('0'),
		reg.CreateValueToken('1'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('2'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('3'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('4'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('5'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('6'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('7'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('8'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('9'),
		reg.CreateOperatorToken(reg.OR),

		reg.CreateValueToken('0'),
		reg.CreateValueToken('1'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('2'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('3'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('4'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('5'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('6'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('7'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('8'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateValueToken('9'),
		reg.CreateOperatorToken(reg.OR),
		reg.CreateOperatorToken(reg.ZERO_OR_MANY),
		reg.CreateOperatorToken(reg.AND),
	}

	result := DEFAULT_ALPHABET.ToPostfix(&infixExpr)
	compareTokensStreams(t, infix, expectedRes, result)
}

// func TestFuzzFail(t *testing.T) {
// 	source := rand.NewSource(int64(69326))
// 	random := rand.New(source)
//
// 	expected := generateExpectedPostfix(random)
// 	infixExpr := fromPostfixToInfix(expected)
// 	infixStr := fromTokenStreamToInfixString(infixExpr)
//
// 	result := DEFAULT_ALPHABET.ToPostfix(&infixExpr)
// 	compareTokensStreams(t, infixStr, expected, result)
// }
