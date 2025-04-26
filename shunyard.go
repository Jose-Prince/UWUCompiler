package main

// You can rename package names when importing them!
// Here the "l" alias is being used!
import (
	"fmt"
	"log"
	"slices"
	"strings"

	l "github.com/Jose-Prince/UWULexer/lib"
	reg "github.com/Jose-Prince/UWULexer/lib/regex"
)

// Maps an operator in the form of a rune into a precedence number.
// Smaller means it has more priority
// Shunting yard only works with these 3 operator types!
var precedence = map[reg.Operator]int{
	reg.OR:           2, // OR Operator
	reg.AND:          3, // AND Operator
	reg.ZERO_OR_MANY: 1, // ZERO_OR_MORE
}

func isDigit(t *reg.RX_Token) bool {
	if !t.IsValue() {
		return false
	}

	tValue := t.GetValue()
	if !tValue.HasValue() {
		return false
	}

	b := tValue.GetValue()
	return b >= '0' && b <= '9'
}

func isLetter(t *reg.RX_Token) bool {
	if !t.IsValue() {
		return false
	}

	tValue := t.GetValue()
	if !tValue.HasValue() {
		return false
	}

	b := tValue.GetValue()
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func tryToAppendWithPrecedence(stack *shunStack, operator reg.Operator, output *[]reg.RX_Token) {
	if stack.Empty() {
		log.Default().Printf("Adding %c to stack!", operator)
		stack.Push(operator)
		return
	}

	top := stack.Peek()
	currentPrecedence := precedence[operator]
	stackPrecedence, found := precedence[top.GetValue()]

	log.Default().Printf("Checking if it can add operator directly %d > %d...", stackPrecedence, currentPrecedence)
	if !found || stackPrecedence > currentPrecedence {
		log.Default().Printf("Adding %c to stack!", operator)
		stack.Push(operator)
	} else {
		for stackPrecedence <= currentPrecedence {
			op := stack.Pop().GetValue()

			log.Default().Printf("Adding %s to output...", op.String())
			*output = append(*output, reg.CreateOperatorToken(op))

			if stack.Empty() {
				break
			}

			top := stack.Peek()
			stackPrecedence, found = precedence[top.GetValue()]
			if !found {
				break
			}
		}

		log.Default().Printf("Adding %c to stack!", operator)
		stack.Push(operator)
	}
}

func appendValueToOutput(
	currentToken *reg.RX_Token,
	previousCanBeANDedTo *bool,
	output *shunOutput,
) {
	log.Default().Printf("Adding %s to output...", currentToken.String())
	*output = append(*output, *currentToken)
	*previousCanBeANDedTo = true
}

type shunStack = l.Stack[reg.Operator]
type shunOutput = []reg.RX_Token

func toPostFix(alph *Alphabet, infixExpression *[]reg.RX_Token, stack *shunStack, output *shunOutput) {
	infixExpr := *infixExpression
	previousCanBeANDedTo := false

	previousExprStack := reg.ExprStack{}
	for i := 0; i < len(infixExpr); i++ {
		currentToken := infixExpr[i]
		log.Default().Printf("Currently checking: `%s`", currentToken.String())

		if currentToken.IsOperator() {
			op := currentToken.GetOperator()
			switch op {
			case reg.OR, reg.AND:
				if stack.Empty() {
					log.Default().Printf("Adding `%s` to stack!", currentToken.String())
					stack.Push(op)
				} else {
					tryToAppendWithPrecedence(stack, op, output)
				}
				previousCanBeANDedTo = false
				previousExprStack.AppendTop(currentToken)

			case reg.ZERO_OR_MANY:
				tryToAppendWithPrecedence(stack, op, output)
				previousCanBeANDedTo = true
				previousExprStack.AppendTop(currentToken)

			case reg.OPTIONAL:
				log.Default().Printf("'?' found! Concatenating with epsilon...")

				// Concatenate previous expression with epsilon
				// And add * operator at the end
				tryToAppendWithPrecedence(stack, reg.OR, output)
				*output = append(*output, reg.CreateEpsilonToken())

				previousCanBeANDedTo = true
				previousExprStack.AppendTop(currentToken)

			case reg.LEFT_PAREN:
				if previousCanBeANDedTo {
					tryToAppendWithPrecedence(stack, reg.AND, output)
				}

				stack.Push(reg.LEFT_PAREN)
				previousCanBeANDedTo = false

				var expr reg.ExprStackItem
				if !previousExprStack.IsEmpty() {
					expr = previousExprStack.Peek().GetValue()
				}

				log.Default().Printf("The previous expression before deleting is: %s", reg.ExprStackItem_ToString(&expr))
				previousExprStack.Pop() // Deletes previous expression
				var parenCtx reg.ExprStackItem = []reg.RX_Token{currentToken}
				previousExprStack.Push(parenCtx)         // Adds ( context
				previousExprStack.Push([]reg.RX_Token{}) // Adds inner ( ) context

			case reg.RIGHT_PAREN:
				log.Default().Printf("Popping until it finds: '('")
				for peeked := stack.Peek(); peeked.GetValue() != reg.LEFT_PAREN; peeked = stack.Peek() {
					val := stack.Pop()
					op := val.GetValue()

					*output = append(*output, reg.CreateOperatorToken(op))
				}

				// Popping '('
				stack.Pop()
				previousCanBeANDedTo = true
				previousExprStack.Pop() // Popping inner ( ) context
				previousExprStack.AppendTop(currentToken)

			case reg.ONE_OR_MANY:
				previousExpr := previousExprStack.Pop().GetValue()
				log.Default().Printf("'+' found! Getting previous expression... `%s`", reg.ExprStackItem_ToString(&previousExpr))

				// Concatenate previous expression with itself
				// And add * operator at the end
				toPostFix(alph, &previousExpr, &shunStack{}, output)
				tryToAppendWithPrecedence(stack, reg.ZERO_OR_MANY, output)
				tryToAppendWithPrecedence(stack, reg.AND, output)

				previousExprStack.AppendTop(reg.CreateOperatorToken(op))
				previousExprStack.Push([]reg.RX_Token{})
				previousCanBeANDedTo = true

			default:
				panic(fmt.Sprintf("Unrecognized operator `%s`!", currentToken.String()))
			}
		} else {
			previousExprStack.Pop()
			previousExprStack.Push([]reg.RX_Token{currentToken})
			appendValueToOutput(&currentToken, &previousCanBeANDedTo, output)
		}
	}

	for !stack.Empty() {
		val := stack.Peek().GetValue()
		if val == '(' {
			break
		} else {
			stack.Pop()
		}
		op := val

		log.Default().Printf("Adding %c to output...", val)
		*output = append(*output, reg.CreateOperatorToken(op))
	}
}

type Alphabet map[rune]struct{}

// Creates a new alphabet from a string
func NewAlphabetFromString(chars string) Alphabet {
	output := Alphabet{}
	for _, rune := range chars {
		output[rune] = struct{}{}
	}

	return output
}

func (alph *Alphabet) GetCharsNotIn(tokens *[]*reg.RX_Token) string {
	charsMap := make(map[rune]struct{})

	for _, token := range *tokens {
		rune := token.GetValue().GetValue()
		charsMap[rune] = struct{}{}
	}

	runes := []rune{}
	for rune := range *alph {
		_, found := charsMap[rune]

		if !found {
			runes = append(runes, rune)
		}
	}
	slices.Sort(runes)

	out := strings.Builder{}
	for _, rune := range runes {
		out.WriteRune(rune)
	}

	return out.String()
}

// Contains all the basic characters that could be inputted on a string
// You can define you're own alphabet
var DEFAULT_ALPHABET = NewAlphabetFromString("abcdefghijklmnñopqrstuvwxyz0123456789:;\"\\'`,._{[()]}*+?¿¡!@#$%&/=~|")

func (alph Alphabet) ToPostfix(infixExpression *[]reg.RX_Token) []reg.RX_Token {
	stack := shunStack{}
	output := []reg.RX_Token{}

	toPostFix(&alph, infixExpression, &stack, &output)
	return output
}
