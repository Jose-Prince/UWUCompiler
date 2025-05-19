package regex

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Jose-Prince/UWULexer/lib"
)

func CreateCanvasExampleAFD() AFD {
	return AFD{
		InitialState:     "0",
		AcceptanceStates: lib.Set[AFDState]{"1": struct{}{}},
		Transitions: map[AFDState]map[AlphabetInput]AFDState{
			"0": {
				CreateValueToken('b'): "0",
				// CreateValueToken('c'): "5",
				CreateValueToken('a'): "2",
			},
			"2": {
				CreateValueToken('a'): "2",
				CreateValueToken('b'): "3",
			},
			"3": {
				CreateValueToken('a'): "2",
				CreateValueToken('b'): "1",
			},
			"1": {
				CreateValueToken('a'): "2",
				CreateValueToken('b'): "0",
			},
		},
	}
}

func validateAFD(expected *AFD, result *AFD) error {
	// Maps a state from expected into a state from result
	afdStatesMapper := make(map[string]string)
	afdStatesMapper[expected.InitialState] = result.InitialState

	evaluatedStates := lib.NewSet[string]()
	evaluationStack := lib.NewStack[string]()
	evaluationStack.Push(expected.InitialState)

	for !evaluationStack.Empty() {
		currentExp := evaluationStack.Pop().GetValue()
		if !evaluatedStates.Add(currentExp) {
			continue
		}

		if _, exists := afdStatesMapper[currentExp]; !exists {
			return errors.New(fmt.Sprintf(`Expected:
%s
Actual:
%s
The state '%s' doesn't have an equivalent on result AFD!
`,
				expected,
				result,
				currentExp,
			))
		}
		currentRes := afdStatesMapper[currentExp]

		expTransitions := expected.Transitions[currentExp]
		resTransitions := result.Transitions[currentRes]
		for input, nextState := range expTransitions {
			if _, exists := resTransitions[input]; !exists {
				return errors.New(fmt.Sprintf(`Expected:
%s
Actual:
%s
The state '%s' (mapped as %s) doesn't have a transition for input '%s' on result AFD!
`,
					expected,
					result,
					currentRes,
					currentExp,
					input.String(),
				))
			}

			afdStatesMapper[nextState] = resTransitions[input]
			evaluationStack.Push(nextState)
		}

	}

	return nil
}

func TestFullRegexToAFDFlow(t *testing.T) {
	regexStream := []RX_Token{
		CreateValueToken('a'),
		CreateValueToken('b'),
		CreateOperatorToken(OR),
		CreateOperatorToken(ZERO_OR_MANY),
		CreateValueToken('a'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
	}
	tree := ASTFromRegex(regexStream)
	expectedTree := CreateCanvasExampleTree()

	err := validateTree(t, expectedTree, tree)
	if err != nil {
		t.Fatal(err.Error())
	}

	expectedTable := CreateCanvasExampleTable()
	table := tree.ToTable()

	err = validateTable(t, expectedTable, table)
	if err != nil {
		t.Fatal(err.Error())
	}

	expectedAFD := CreateCanvasExampleAFD()
	afd := table.ToAFD()

	err = validateAFD(&expectedAFD, &afd)
	if err != nil {
		t.Fatal(err.Error())
	}
}
