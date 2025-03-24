package main

import (
	"math"

	l "github.com/Jose-Prince/UWULexer/lib"
)

type infixConverterState int

const (
	NORMAL               infixConverterState = iota // We just started parsing the Regexp
	IN_BRACKETS                                     // We are inside [ ]
	IN_NEGATIVE_BRACKETS                            // We are inside [^ ]
	IN_PARENTHESIS                                  // We are inside ( )
)

// Converts an infix expression into an array of tokens
func InfixToTokens(infix string) []l.RX_Token {
	previousCanBeANDedTo := false
	tokens := []l.RX_Token{}
	stateStack := l.Stack[infixConverterState]{}
	stateStack.Push(NORMAL)

	// Contains all the characters that should NOT be added when a `l.SET_NEGATION` operator is closed
	negativeBuffer := make(map[rune]struct{})
	runes := []rune(infix)
	for i := 0; i < len(runes); i++ {
		currentRune := runes[i]

		currentState := stateStack.Peek().GetValue()
		switch currentState {
		case IN_BRACKETS:
			startRune := currentRune
			if runes[i+1] == '-' { // If this is a range start...
				endRune := runes[i+2]

				if startRune > endRune { // It doesn't matter if the user writes A-Z or Z-A
					startRune, endRune = endRune, startRune
				}

				if previousCanBeANDedTo { // Concatenate with previous value on the bracket
					tokens = append(tokens, l.CreateOperatorToken(l.OR))
				}

				for j := rune(0); j <= (endRune - startRune); j++ {
					if j >= 1 {
						tokens = append(tokens, l.CreateOperatorToken(l.OR))
					}

					val := startRune + j
					tokens = append(tokens, l.CreateValueToken(val))
				}

				// When we started parsing we where:
				// A-Z
				// ^
				// So we need to advance three runes...
				// A-Z
				//    ^
				// The last rune will be advanced by the `i++` of the for loop
				i += 2
			} else { // If not a range...
				switch currentRune {
				case '\\':
					nextRune := runes[i+1]
					i++

					if previousCanBeANDedTo {
						tokens = append(tokens, l.CreateOperatorToken(l.OR))
					}
					tokens = append(tokens, l.CreateValueToken(nextRune))

				case ']':
					stateStack.Pop()
					tokens = append(tokens, l.CreateOperatorToken(l.RIGHT_PAREN))
					previousCanBeANDedTo = true

				default:
					if previousCanBeANDedTo {
						tokens = append(tokens, l.CreateOperatorToken(l.OR))
					}
					tokens = append(tokens, l.CreateValueToken(currentRune))
				}
			}

		case IN_NEGATIVE_BRACKETS:
			startRune := currentRune
			if runes[i+1] == '-' { // If this is a range start...
				endRune := runes[i+2]

				if startRune > endRune { // It doesn't matter if the user writes A-Z or Z-A
					startRune, endRune = endRune, startRune
				}

				for j := rune(0); j <= (endRune - startRune); j++ {
					val := startRune + j
					negativeBuffer[val] = struct{}{}
				}

				// When we started parsing we where:
				// A-Z
				// ^
				// So we need to advance three runes...
				// A-Z
				//    ^
				// The last rune will be advanced by the `i++` of the for loop
				i += 2
			} else { // If not a range...
				switch currentRune {
				case '\\':
					nextRune := runes[i+1]
					i++
					negativeBuffer[nextRune] = struct{}{}

				case ']':
					// Since we reached the end of the set negation
					// now we need to add all the elements that are not in the negativeBuffer

					addedCount := int32(0)
					for j := int32(0); j <= math.MaxInt32; j++ {
						_, found := negativeBuffer[j]
						if found {
							continue
						}

						if addedCount > 0 {
							tokens = append(tokens, l.CreateOperatorToken(l.OR))
						}
						tokens = append(tokens, l.CreateValueToken(j))

						addedCount++
					}

					// In the end we create a right parenthesis to close the one we opened when reading [
					tokens = append(tokens, l.CreateOperatorToken(l.RIGHT_PAREN))
					stateStack.Pop()
					previousCanBeANDedTo = true

				default:
					negativeBuffer[currentRune] = struct{}{}
				}
			}

		default:
			var token l.RX_Token
			switch currentRune {
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
				stateStack.Push(IN_PARENTHESIS)
				token = l.CreateOperatorToken(l.LEFT_PAREN)
				previousCanBeANDedTo = false

			case ')':
				if currentState == IN_PARENTHESIS {
					stateStack.Pop()
					token = l.CreateOperatorToken(l.RIGHT_PAREN)
					previousCanBeANDedTo = true
				} else {
					panic("Unclosed parenthesis found! Please check your regexes...")
				}

			case '[':
				if previousCanBeANDedTo {
					tokens = append(tokens, l.CreateOperatorToken(l.AND))
				}

				// We add a parenthesis instead since []
				// get's transformed into a lot of OR operations
				token = l.CreateOperatorToken(l.LEFT_PAREN)
				state := IN_BRACKETS

				nextRune := runes[i+1]
				if '^' == nextRune {
					state = IN_NEGATIVE_BRACKETS
					i++
				}
				stateStack.Push(state)
				previousCanBeANDedTo = false

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

				token = l.CreateValueToken(currentRune)
				previousCanBeANDedTo = true
			}

			tokens = append(tokens, token)
		}

	}

	return tokens
}
