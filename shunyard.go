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
var precedence = map[byte]int{
	'|': 2, // OR Operator
	'.': 3, // AND Operator
	'*': 1, // ZERO_OR_MORE
}

func toOperator(self byte) l.Optional[l.Operator] {
	log.Default().Printf("Trying to get operator from: %c", self)

	switch self {
	case '|':
		return l.CreateValue(l.OR)
	case '.':
		return l.CreateValue(l.AND)
	case '*':
		return l.CreateValue(l.ZERO_OR_MANY)
	default:
		return l.CreateNull[l.Operator]()
	}
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func tryToAppendWithPrecedence(stack *l.Stack[byte], operator byte, output *[]l.RX_Token) {
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
			poppedRune := stack.Pop().GetValue()

			op := toOperator(poppedRune)
			log.Default().Printf("Adding %c to output...", poppedRune)
			*output = append(*output, l.CreateOperatorToken(op.GetValue()))

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

func appendValueToOutput(infixExpr *string,
	currentChar *byte,
	i *int,
	previousCanBeANDedTo *bool,
	state *regexState,
	stack *shunStack,
	output *shunOutput,
	previousExprStack *l.ExprStack,
	negativeBuffer *strings.Builder,
) {
	log.Default().Printf("Iteration: (%c) %d != 0 && previousCanBeANDed: %t", *currentChar, *i, *previousCanBeANDedTo)
	if *i != 0 && *previousCanBeANDedTo {
		if *state == NORMAL || *state == IN_PARENTHESIS {
			log.Default().Printf("Trying to append '.' operator...")
			tryToAppendWithPrecedence(stack, '.', output)
		} else {
			log.Default().Printf("Trying to append '|' operator...")
			tryToAppendWithPrecedence(stack, '|', output)
		}
	}

	rangeStart := *currentChar
	if *state == IN_BRACKETS || *state == IN_NEGATIVE_BRACKETS {
		log.Default().Printf("Checking if the char (%c) is a range start...", *currentChar)
		previousExprStack.AppendTop(string(*currentChar))

		if isLetter(rangeStart) || isDigit(rangeStart) {
			nextChar := (*infixExpr)[*i+1]

			if nextChar == '-' {
				rangeEnd := (*infixExpr)[*i+2]
				isEndTheSameAsStart := (isLetter(rangeStart) && isLetter(rangeEnd)) || (isDigit(rangeStart) && isDigit(rangeStart))

				log.Default().Printf("The end char (%c) is the same type as start? %v", rangeEnd, isEndTheSameAsStart)
				if isEndTheSameAsStart {
					if rangeEnd < rangeStart {
						rangeEnd, rangeStart = rangeStart, rangeEnd
					}

					if *state == IN_BRACKETS {
						for j := byte(0); j <= (rangeEnd - rangeStart); j++ {
							if j >= 1 {
								tryToAppendWithPrecedence(stack, '|', output)
							}

							val := rune(rangeStart + j)
							log.Default().Printf("Adding %c to output...", val)
							*output = append(*output, l.CreateValueToken(val))
						}

						// We already parsed '-' and the other byte
						// So we need to ignore them
						*i += 2
						// continue
						return
					} else if *state == IN_NEGATIVE_BRACKETS {
						for j := byte(0); j <= (rangeEnd - rangeStart); j++ {
							val := rune(rangeStart + j)
							log.Default().Printf("Adding %c to negative buffer...", val)
							negativeBuffer.WriteRune(val)
						}

						// We already parsed '-' and the other byte
						// So we need to ignore them
						previousExprStack.AppendTop("-")
						previousExprStack.AppendTop(string((*infixExpr)[*i+2]))
						*i += 2
						// continue
						return
					}
				}
			}
		}
	}

	if *state == IN_NEGATIVE_BRACKETS {
		negativeBuffer.WriteByte(*currentChar)

	} else if *state == IN_BRACKETS || *state == IN_PARENTHESIS {
		expr := ""
		if !previousExprStack.IsEmpty() {
			expr = previousExprStack.Peek().GetValue()
		}

		log.Default().Printf("Appending %s to expression: %s", string(*currentChar), expr)
		previousExprStack.AppendTop(string(*currentChar))
	} else {
		expr := ""
		if !previousExprStack.IsEmpty() {
			expr = previousExprStack.Peek().GetValue()
		}

		log.Default().Printf("Changing previous expr from `%s` to `%s`", expr, string(*currentChar))
		previousExprStack.Pop()
		previousExprStack.Push(string(*currentChar))
	}

	log.Default().Printf("Adding %c to output...", *currentChar)
	*output = append(*output, l.CreateValueToken(rune(*currentChar)))
	*previousCanBeANDedTo = true

}

type regexState int

const (
	NORMAL               regexState = iota // We just started parsing the Regexp
	IN_BRACKETS                            // We are inside [ ]
	IN_NEGATIVE_BRACKETS                   // We are inside [^ ]
	IN_PARENTHESIS                         // We are inside ( )
)

type shunStack = l.Stack[byte]
type shunOutput = []l.RX_Token

func toPostFix(alph *Alphabet, infixExpression *string, stack *shunStack, output *shunOutput) {
	infixExpr := *infixExpression
	previousCanBeANDedTo := false
	state := NORMAL

	negativeBuffer := strings.Builder{}
	previousExprStack := l.ExprStack{}
	for i := 0; i < len(infixExpr); i++ {
		currentChar := infixExpr[i]
		log.Default().Printf("Currently checking: `%c`", currentChar)

		switch currentChar {
		case '|':
			if state == IN_NEGATIVE_BRACKETS {
				negativeBuffer.WriteByte('|')
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				if stack.Empty() {
					log.Default().Printf("Adding `%c` to stack!", currentChar)
					stack.Push(currentChar)
				} else {
					tryToAppendWithPrecedence(stack, currentChar, output)
				}
				previousCanBeANDedTo = false
			}

			previousExprStack.AppendTop("|")

		case '*':
			if state == IN_NEGATIVE_BRACKETS {
				negativeBuffer.WriteByte(currentChar)
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				tryToAppendWithPrecedence(stack, currentChar, output)
				previousCanBeANDedTo = true
			}
			previousExprStack.AppendTop("*")

		case '?':
			if state == IN_NEGATIVE_BRACKETS {
				negativeBuffer.WriteByte(currentChar)
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				log.Default().Printf("'?' found! Concatenating with epsilon...")

				// Concatenate previous expression with epsilon
				// And add * operator at the end
				tryToAppendWithPrecedence(stack, '|', output)
				*output = append(*output, l.CreateEpsilonToken())

				previousCanBeANDedTo = true
			}
			previousExprStack.AppendTop("?")

		case '(':
			if state == IN_NEGATIVE_BRACKETS {
				negativeBuffer.WriteByte('(')
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				if previousCanBeANDedTo {
					tryToAppendWithPrecedence(stack, '.', output)
				}

				stack.Push('(')
				previousCanBeANDedTo = false
				state = IN_PARENTHESIS

				expr := ""
				if !previousExprStack.IsEmpty() {
					expr = previousExprStack.Peek().GetValue()
				}

				log.Default().Printf("The previous expression before deleting is: %s", expr)
				previousExprStack.Pop()     // Deletes previous expression
				previousExprStack.Push("(") // Adds ( context
				previousExprStack.Push("")  // Adds inner ( ) context
			}

		case ')':
			if state == IN_NEGATIVE_BRACKETS {
				negativeBuffer.WriteByte(')')
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				log.Default().Printf("Popping until it finds: '('")
				for peeked := stack.Peek(); peeked.GetValue() != '('; peeked = stack.Peek() {
					val := stack.Pop()
					op := toOperator(val.GetValue()).GetValue()

					*output = append(*output, l.CreateOperatorToken(op))
				}

				// Popping '('
				stack.Pop()
				state = NORMAL
				previousExprStack.AppendTop(")")
				previousExprStack.Pop() // Popping inner ( ) context
			}

		case '[':
			if state == IN_BRACKETS {
				*output = append(*output, l.CreateValueToken('['))
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				if previousCanBeANDedTo {
					tryToAppendWithPrecedence(stack, '.', output)
				}

				stack.Push('[')
				nextChar := infixExpr[i+1]
				previousCanBeANDedTo = false

				if nextChar == '^' {
					state = IN_NEGATIVE_BRACKETS
					i++
				} else {
					state = IN_BRACKETS
				}

				expr := ""
				if !previousExprStack.IsEmpty() {
					expr = previousExprStack.Peek().GetValue()
				}
				log.Default().Printf("The previous expression before deleting is: %s", expr)
				previousExprStack.Pop() // Deletes previous expression
				if state == IN_NEGATIVE_BRACKETS {
					previousExprStack.AppendTop("[^")
				} else {
					previousExprStack.Push("[") // Adds [ context
				}
				previousExprStack.Push("") // Adds inner [ ] context
			}

		case ']':
			log.Default().Printf("Popping until it finds: '['")
			for peeked := stack.Peek(); peeked.GetValue() != '['; peeked = stack.Peek() {
				val := stack.Pop()
				op := toOperator(val.GetValue()).GetValue()

				*output = append(*output, l.CreateOperatorToken(op))
			}
			// Popping '['
			stack.Pop()
			previousExprStack.Pop() // Popping inner [ ] context
			previousExprStack.AppendTop("]")

			log.Default().Printf("Checking if IN_NEGATIVE_BRACKETS: %d == %d", state, IN_NEGATIVE_BRACKETS)
			if state == IN_NEGATIVE_BRACKETS {
				diff := alph.GetCharsNotIn(negativeBuffer.String())
				log.Default().Printf("Obtaining diff: `%s`", diff)

				for idx, val := range diff {
					if idx >= 1 {
						tryToAppendWithPrecedence(stack, '|', output)
					}

					log.Default().Printf("Appending %c to output...", val)
					*output = append(*output, l.CreateValueToken(rune(val)))
				}

				// Last '|' must be appended to output as well
				*output = append(*output, l.CreateOperatorToken(toOperator(stack.Pop().GetValue()).GetValue()))
			}

			negativeBuffer = strings.Builder{}
			state = NORMAL
			previousCanBeANDedTo = true

		case '+':
			if state == IN_NEGATIVE_BRACKETS {
				negativeBuffer.WriteByte(currentChar)
			} else if state == IN_BRACKETS {
				appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			} else {
				previousExpr := previousExprStack.Pop().GetValue()
				log.Default().Printf("'+' found! Getting previous expression... `%s`", previousExpr)

				// Concatenate previous expression with itself
				// And add * operator at the end
				toPostFix(alph, &previousExpr, &shunStack{}, output)
				tryToAppendWithPrecedence(stack, '*', output)
				tryToAppendWithPrecedence(stack, '.', output)

				previousExprStack.AppendTop("+")
				previousExprStack.Push("")
				previousCanBeANDedTo = true
			}

		case '\\':
			previousExprStack.AppendTop("\\")
			nextChar := infixExpr[i+1]
			previousExprStack.AppendTop(string(nextChar))
			log.Default().Printf("Escape sequence found! Adding %c as a char...", nextChar)
			if previousCanBeANDedTo {
				appendValueToOutput(&infixExpr, &nextChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
			}
			*output = append(*output, l.CreateValueToken(rune(nextChar)))
			i += 1

		default:
			appendValueToOutput(&infixExpr, &currentChar, &i, &previousCanBeANDedTo, &state, stack, output, &previousExprStack, &negativeBuffer)
		}
	}

	for !stack.Empty() {
		val := stack.Peek().GetValue()
		if val == '(' {
			break
		} else {
			stack.Pop()
		}
		op := toOperator(val)

		if op.HasValue() {
			log.Default().Printf("Adding %c to output...", val)
			*output = append(*output, l.CreateOperatorToken(op.GetValue()))
		}
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

func (alph *Alphabet) GetCharsNotIn(chars string) string {
	charsMap := make(map[rune]struct{})

	for _, rune := range chars {
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

func (alph Alphabet) ToPostfix(infixExpression string) []l.RX_Token {
	stack := l.Stack[byte]{}
	output := []l.RX_Token{}

	toPostFix(&alph, &infixExpression, &stack, &output)
	return output
}
