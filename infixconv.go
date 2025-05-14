package main

import (
	"math"

	"github.com/Jose-Prince/UWULexer/lib"
	"github.com/Jose-Prince/UWULexer/lib/regex"
)

type infixConverterState int

const (
	NORMAL               infixConverterState = iota // We just started parsing the Regexp
	IN_BRACKETS                                     // We are inside [ ]
	IN_NEGATIVE_BRACKETS                            // We are inside [^ ]
	IN_PARENTHESIS                                  // We are inside ( )
)

// Converts an infix expression into an array of tokens
func InfixToTokens(infix string) []regex.RX_Token {
	previousCanBeANDedTo := false
	tokens := []regex.RX_Token{}
	stateStack := lib.Stack[infixConverterState]{}
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

			if i+1 < len(runes) && i+2 < len(runes) && runes[i+1] == '-' { // If this is a range start...
				endRune := runes[i+2]

				if startRune > endRune { // It doesn't matter if the user writes A-Z or Z-A
					startRune, endRune = endRune, startRune
				}

				if previousCanBeANDedTo { // Concatenate with previous value on the bracket
					tokens = append(tokens, regex.CreateOperatorToken(regex.OR))
				}

				for j := rune(0); j <= (endRune - startRune); j++ {
					if j >= 1 {
						tokens = append(tokens, regex.CreateOperatorToken(regex.OR))
					}

					val := startRune + j
					tokens = append(tokens, regex.CreateValueToken(val))
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
					nextRune := currentRune
					if i+1 < len(runes) {
						nextRune = runes[i+1]
						i++
					}

					switch nextRune {
					case 'n':
						nextRune = '\n'
					case 't':
						nextRune = '\t'
					case 'r':
						nextRune = '\r'
					default:
					}

					if previousCanBeANDedTo {
						tokens = append(tokens, regex.CreateOperatorToken(regex.OR))
					}
					tokens = append(tokens, regex.CreateValueToken(nextRune))
					previousCanBeANDedTo = true

				case ']':
					stateStack.Pop()
					tokens = append(tokens, regex.CreateOperatorToken(regex.RIGHT_PAREN))
					previousCanBeANDedTo = true

				default:
					if previousCanBeANDedTo {
						tokens = append(tokens, regex.CreateOperatorToken(regex.OR))
					}
					tokens = append(tokens, regex.CreateValueToken(currentRune))
					previousCanBeANDedTo = true
				}
			}

		case IN_NEGATIVE_BRACKETS:
			startRune := currentRune
			if i+1 < len(runes) && i+2 < len(runes) && runes[i+1] == '-' { // If this is a range start...
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
					nextRune := currentRune
					if i+1 < len(runes) {
						nextRune = runes[i+1]
						i++
					}

					switch nextRune {
					case 'n':
						nextRune = '\n'
					case 't':
						nextRune = '\t'
					case 'r':
						nextRune = '\r'
					default:
					}
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
							tokens = append(tokens, regex.CreateOperatorToken(regex.OR))
						}
						tokens = append(tokens, regex.CreateValueToken(j))

						addedCount++
					}

					// In the end we create a right parenthesis to close the one we opened when reading [
					tokens = append(tokens, regex.CreateOperatorToken(regex.RIGHT_PAREN))
					stateStack.Pop()
					previousCanBeANDedTo = true

				default:
					negativeBuffer[currentRune] = struct{}{}
				}
			}

		default:
			var token regex.RX_Token
			switch currentRune {
			case '|':
				token = regex.CreateOperatorToken(regex.OR)
				previousCanBeANDedTo = false

			case '*':
				token = regex.CreateOperatorToken(regex.ZERO_OR_MANY)
				previousCanBeANDedTo = true

			case '(':
				if previousCanBeANDedTo {
					tokens = append(tokens, regex.CreateOperatorToken(regex.AND))
				}
				stateStack.Push(IN_PARENTHESIS)
				token = regex.CreateOperatorToken(regex.LEFT_PAREN)
				previousCanBeANDedTo = false

			case ')':
				if currentState == IN_PARENTHESIS {
					stateStack.Pop()
					token = regex.CreateOperatorToken(regex.RIGHT_PAREN)
					previousCanBeANDedTo = true
				} else {
					panic("Unclosed parenthesis found! Please check your regexes...")
				}

			case '[':
				if previousCanBeANDedTo {
					tokens = append(tokens, regex.CreateOperatorToken(regex.AND))
				}

				// We add a parenthesis instead since []
				// get's transformed into a lot of OR operations
				token = regex.CreateOperatorToken(regex.LEFT_PAREN)
				state := IN_BRACKETS

				nextRune := runes[i+1]
				if '^' == nextRune {
					state = IN_NEGATIVE_BRACKETS
					i++
				}
				stateStack.Push(state)
				previousCanBeANDedTo = false

			case '+':
				token = regex.CreateOperatorToken(regex.ONE_OR_MANY)
				previousCanBeANDedTo = true

			case '?':
				token = regex.CreateOperatorToken(regex.OPTIONAL)
				previousCanBeANDedTo = true

			case '\\':
				if previousCanBeANDedTo {
					tokens = append(tokens, regex.CreateOperatorToken(regex.AND))
				}

				nextRune := currentRune
				if i+1 < len(runes) {
					nextRune = runes[i+1]
					i++
				}

				switch nextRune {
				case 'n':
					nextRune = '\n'
				case 't':
					nextRune = '\t'
				case 'r':
					nextRune = '\r'
				default:
				}

				token = regex.CreateValueToken(rune(nextRune))
				previousCanBeANDedTo = true
			default:
				if previousCanBeANDedTo {
					tokens = append(tokens, regex.CreateOperatorToken(regex.AND))
				}

				token = regex.CreateValueToken(currentRune)
				previousCanBeANDedTo = true
			}

			tokens = append(tokens, token)
		}

	}

	return tokens
}
