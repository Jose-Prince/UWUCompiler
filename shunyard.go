package main

// You can rename package names when importing them!
// Here the "l" alias is being used!
import (
	"log"
	"slices"
	"strings"

	l "github.com/Jose-Prince/UWULexer/lib"
)

// Maps an operator in the form of a rune into a precedence number.
// Smaller means it has more priority
// Shunting yard only works with these 3 operator types!
var precedence = map[l.Operator]int{
	l.OR:           2, // OR Operator
	l.AND:          3, // AND Operator
	l.ZERO_OR_MANY: 1, // ZERO_OR_MORE
}

func isDigit(t *l.RX_Token) bool {
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

func isLetter(t *l.RX_Token) bool {
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

func tryToAppendWithPrecedence(stack *shunStack, operator l.Operator, output *[]l.RX_Token) {
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

			log.Default().Printf("Adding %s to output...", op.ToString())
			*output = append(*output, l.CreateOperatorToken(op))

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

func appendValueToOutput(infixExpr *[]l.RX_Token,
	currentToken *l.RX_Token,
	i *int,
	previousCanBeANDedTo *bool,
	state *regexState,
	stack *shunStack,
	output *shunOutput,
	previousExprStack *l.ExprStack,
	negativeBuffer *[]*l.RX_Token,
) {
	log.Default().Printf("Iteration: (%s) %d != 0 && previousCanBeANDed: %t", currentToken.ToString(), *i, *previousCanBeANDedTo)
	if *i != 0 && *previousCanBeANDedTo {
		if *state == NORMAL || *state == IN_PARENTHESIS {
			log.Default().Printf("Trying to append '.' operator...")
			tryToAppendWithPrecedence(stack, l.AND, output)
		} else {
			log.Default().Printf("Trying to append '|' operator...")
			tryToAppendWithPrecedence(stack, l.OR, output)
		}
	}

	rangeStart := currentToken
	if *state == IN_BRACKETS || *state == IN_NEGATIVE_BRACKETS {
		log.Default().Printf("Checking if the token (%s) is a range start...", currentToken.ToString())
		previousExprStack.AppendTop(*currentToken)

		if isLetter(rangeStart) || isDigit(rangeStart) {
			nextToken := (*infixExpr)[*i+1]

			if rangeOp := l.CreateValueToken('-'); nextToken.Equals(&rangeOp) {
				rangeEnd := &(*infixExpr)[*i+2]
				isEndTheSameAsStart := (isLetter(rangeStart) && isLetter(rangeEnd)) || (isDigit(rangeStart) && isDigit(rangeStart))

				log.Default().Printf("The end token (%s) is the same type as start? %v", rangeEnd.ToString(), isEndTheSameAsStart)
				if isEndTheSameAsStart {
					if rangeEnd.GetValue().GetValue() < rangeStart.GetValue().GetValue() {
						rangeEnd, rangeStart = rangeStart, rangeEnd
					}

					startRune := rangeStart.GetValue().GetValue()
					endRune := rangeEnd.GetValue().GetValue()

					if *state == IN_BRACKETS {
						for j := rune(0); j <= (endRune - startRune); j++ {
							if j >= 1 {
								tryToAppendWithPrecedence(stack, l.OR, output)
							}

							val := startRune + j
							log.Default().Printf("Adding %c to output...", val)
							*output = append(*output, l.CreateValueToken(val))
						}

						// We already parsed '-' and the other byte
						// So we need to ignore them
						// *i += 2
						// continue
						// return
					} else if *state == IN_NEGATIVE_BRACKETS {
						for j := rune(0); j <= (endRune - startRune); j++ {
							val := startRune + j
							log.Default().Printf("Adding %c to negative buffer...", val)
							valTk := l.CreateValueToken(val)
							*negativeBuffer = append(*negativeBuffer, &valTk)
						}

						// continue
						// return
					}

					// We already parsed '-' and the other byte
					// So we need to ignore them
					rangeTk := l.CreateValueToken('-')
					previousExprStack.AppendTop(rangeTk)
					previousExprStack.AppendTop(*rangeEnd)
					*i += 2
					return
				}
			}
		}
	}

	if *state == IN_NEGATIVE_BRACKETS {
		*negativeBuffer = append(*negativeBuffer, currentToken)

	} else if *state == IN_BRACKETS || *state == IN_PARENTHESIS {
		var expr l.ExprStackItem
		if !previousExprStack.IsEmpty() {
			expr = previousExprStack.Peek().GetValue()
		}

		log.Default().Printf("Appending %s to expression: %s", currentToken.ToString(), l.ExprStackItem_ToString(&expr))
		previousExprStack.AppendTop(*currentToken)
	} else {
		var expr l.ExprStackItem
		if !previousExprStack.IsEmpty() {
			expr = previousExprStack.Peek().GetValue()
		}

		var newExpr l.ExprStackItem = []l.RX_Token{*currentToken}
		log.Default().Printf("Changing previous expr from `%s` to `%s`", l.ExprStackItem_ToString(&expr), l.ExprStackItem_ToString(&newExpr))
		previousExprStack.Pop()
		previousExprStack.Push(newExpr)
	}

	log.Default().Printf("Adding %s to output...", currentToken.ToString())
	*output = append(*output, *currentToken)
	*previousCanBeANDedTo = true

}

type regexState int

const (
	NORMAL               regexState = iota // We just started parsing the Regexp
	IN_BRACKETS                            // We are inside [ ]
	IN_NEGATIVE_BRACKETS                   // We are inside [^ ]
	IN_PARENTHESIS                         // We are inside ( )
)

type shunStack = l.Stack[l.Operator]
type shunOutput = []l.RX_Token

func toPostFix(alph *Alphabet, infixExpression *[]l.RX_Token, stack *shunStack, output *shunOutput) {
	infixExpr := *infixExpression
	previousCanBeANDedTo := false
	state := NORMAL

	// Contains all the characters that should NOT be added when a `l.SET_NEGATION` operator is closed
	negativeBuffer := []*l.RX_Token{}
	previousExprStack := l.ExprStack{}
	for i := 0; i < len(infixExpr); i++ {
		currentToken := infixExpr[i]
		log.Default().Printf("Currently checking: `%s`", currentToken.ToString())

		if currentToken.IsOperator() {
			op := currentToken.GetOperator()
			switch op {
			case l.OR:
				if state == IN_NEGATIVE_BRACKETS {
					negativeBuffer = append(negativeBuffer, &currentToken)
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					if stack.Empty() {
						log.Default().Printf("Adding `%s` to stack!", currentToken.ToString())
						stack.Push(op)
					} else {
						tryToAppendWithPrecedence(stack, op, output)
					}
					previousCanBeANDedTo = false
				}

				previousExprStack.AppendTop(currentToken)

			case l.ZERO_OR_MANY:
				if state == IN_NEGATIVE_BRACKETS {
					negativeBuffer = append(negativeBuffer, &currentToken)
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					tryToAppendWithPrecedence(stack, op, output)
					previousCanBeANDedTo = true
				}
				previousExprStack.AppendTop(currentToken)

			case l.OPTIONAL:
				if state == IN_NEGATIVE_BRACKETS {
					negativeBuffer = append(negativeBuffer, &currentToken)
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					log.Default().Printf("'?' found! Concatenating with epsilon...")

					// Concatenate previous expression with epsilon
					// And add * operator at the end
					tryToAppendWithPrecedence(stack, l.OR, output)
					*output = append(*output, l.CreateEpsilonToken())

					previousCanBeANDedTo = true
				}
				previousExprStack.AppendTop(currentToken)

			case l.LEFT_PAREN:
				if state == IN_NEGATIVE_BRACKETS {
					negativeBuffer = append(negativeBuffer, &currentToken)
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					if previousCanBeANDedTo {
						tryToAppendWithPrecedence(stack, l.AND, output)
					}

					stack.Push(l.LEFT_PAREN)
					previousCanBeANDedTo = false
					state = IN_PARENTHESIS

					var expr l.ExprStackItem
					if !previousExprStack.IsEmpty() {
						expr = previousExprStack.Peek().GetValue()
					}

					log.Default().Printf("The previous expression before deleting is: %s", l.ExprStackItem_ToString(&expr))
					previousExprStack.Pop() // Deletes previous expression
					var parenCtx l.ExprStackItem = []l.RX_Token{currentToken}
					previousExprStack.Push(parenCtx)       // Adds ( context
					previousExprStack.Push([]l.RX_Token{}) // Adds inner ( ) context
				}
			case l.RIGHT_PAREN:
				if state == IN_NEGATIVE_BRACKETS {
					negativeBuffer = append(negativeBuffer, &currentToken)
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					log.Default().Printf("Popping until it finds: '('")
					for peeked := stack.Peek(); peeked.GetValue() != l.OR; peeked = stack.Peek() {
						val := stack.Pop()
						op := val.GetValue()

						*output = append(*output, l.CreateOperatorToken(op))
					}

					// Popping '('
					stack.Pop()
					state = NORMAL
					previousExprStack.Pop() // Popping inner ( ) context
					previousExprStack.AppendTop(currentToken)
				}

			case l.LEFT_BRACKET:
				if state == IN_NEGATIVE_BRACKETS {
					*output = append(*output, l.CreateValueToken('['))
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					if previousCanBeANDedTo {
						tryToAppendWithPrecedence(stack, l.AND, output)
					}

					stack.Push(op)
					previousCanBeANDedTo = false
					state = IN_BRACKETS

					var expr l.ExprStackItem
					if !previousExprStack.IsEmpty() {
						expr = previousExprStack.Peek().GetValue()
					}
					log.Default().Printf("The previous expression before deleting is: %s", l.ExprStackItem_ToString(&expr))
					previousExprStack.Pop() // Deletes previous expression

					var parenCtx l.ExprStackItem = []l.RX_Token{currentToken}
					previousExprStack.Push(parenCtx)       // Adds [ context
					previousExprStack.Push([]l.RX_Token{}) // Adds inner [ ] context
				}
			case l.SET_NEGATION:
				if state != NORMAL {
					panic("Invalid Regular Expression! A set negation can't be inside brackets or another set negation!")
				} else if state != IN_PARENTHESIS {
					panic("Invalid Regular Expression! A set negation can't be inside brackets or another set negation!")

				}

				if previousCanBeANDedTo {
					tryToAppendWithPrecedence(stack, l.AND, output)
				}

				stack.Push(op)
				previousCanBeANDedTo = false
				state = IN_NEGATIVE_BRACKETS

				var expr l.ExprStackItem
				if !previousExprStack.IsEmpty() {
					expr = previousExprStack.Peek().GetValue()
				}
				log.Default().Printf("The previous expression before deleting is: %s", l.ExprStackItem_ToString(&expr))
				previousExprStack.Pop() // Deletes previous expression

				var parenCtx l.ExprStackItem = []l.RX_Token{currentToken}
				previousExprStack.Push(parenCtx)       // Adds [ context
				previousExprStack.Push([]l.RX_Token{}) // Adds inner [ ] context

			case l.RIGHT_BRACKET:
				log.Default().Printf("Popping until it finds: '['")
				for peeked := stack.Peek(); peeked.GetValue() != l.LEFT_BRACKET; peeked = stack.Peek() {
					val := stack.Pop()
					op := val.GetValue()

					*output = append(*output, l.CreateOperatorToken(op))
				}
				// Popping '['
				stack.Pop()
				previousExprStack.Pop() // Popping inner [ ] context
				previousExprStack.AppendTop(currentToken)

				log.Default().Printf("Checking if IN_NEGATIVE_BRACKETS: %d == %d", state, IN_NEGATIVE_BRACKETS)
				if state == IN_NEGATIVE_BRACKETS {
					diff := alph.GetCharsNotIn(&negativeBuffer)
					log.Default().Printf("Obtaining diff: `%s`", diff)

					for idx, val := range diff {
						if idx >= 1 {
							tryToAppendWithPrecedence(stack, l.OR, output)
						}

						log.Default().Printf("Appending %c to output...", val)
						*output = append(*output, l.CreateValueToken(rune(val)))
					}

					// Last '|' must be appended to output as well
					*output = append(*output, l.CreateOperatorToken(stack.Pop().GetValue()))
				}

				negativeBuffer = []*l.RX_Token{}
				state = NORMAL
				previousCanBeANDedTo = true

			case l.ONE_OR_MANY:
				if state == IN_NEGATIVE_BRACKETS {
					negativeBuffer = append(negativeBuffer, &currentToken)
				} else if state == IN_BRACKETS {
					appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
				} else {
					previousExpr := previousExprStack.Pop().GetValue()
					log.Default().Printf("'+' found! Getting previous expression... `%s`", l.ExprStackItem_ToString(&previousExpr))

					// Concatenate previous expression with itself
					// And add * operator at the end
					tryToAppendWithPrecedence(stack, l.AND, output)
					toPostFix(alph, &previousExpr, &shunStack{}, output)
					tryToAppendWithPrecedence(stack, l.ZERO_OR_MANY, output)

					previousExprStack.AppendTop(l.CreateOperatorToken(op))
					previousExprStack.Push([]l.RX_Token{})
					previousCanBeANDedTo = true
				}

			default:
				appendValueToOutput(&infixExpr, &currentToken, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			}
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
		*output = append(*output, l.CreateOperatorToken(op))
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

func (alph *Alphabet) GetCharsNotIn(tokens *[]*l.RX_Token) string {
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

func (alph Alphabet) ToPostfix(infixExpression *[]l.RX_Token) []l.RX_Token {
	stack := shunStack{}
	output := []l.RX_Token{}

	toPostFix(&alph, infixExpression, &stack, &output)
	return output
}
