package main

import (
	"bufio"
	"math"
	"os"

	"github.com/Jose-Prince/UWULexer/lib"
)

type afdLeafInfo struct {
	Code     string
	NewState lib.AFDState
}

type afdSwitch struct {
	// The first key is the state
	// The second key is the input
	// The third key is the nextState and the code to write
	Transitions map[lib.AFDState]map[rune]afdLeafInfo

	InitialState     lib.AFDState
	AcceptanceStates lib.Set[lib.AFDState]
}

func simplifyIntoSwitch(afd *lib.AFD) afdSwitch {
	sw := afdSwitch{
		InitialState:     afd.InitialState,
		AcceptanceStates: afd.AcceptanceStates,
		Transitions:      make(map[lib.AFDState]map[rune]afdLeafInfo),
	}
	_simplifyIntoSwitch(afd, afd.InitialState, &sw)
	return sw
}

type tranInput struct {
	Input lib.AlphabetInput
	State lib.AFDState
}

func getChildrenWithDummyTransitions(afd *lib.AFD, state lib.AFDState) []tranInput {
	children := []tranInput{}

	for in, childState := range afd.Transitions[state] {
		for input := range afd.Transitions[childState] {
			if input.IsDummy() {
				children = append(children, tranInput{Input: in, State: childState})
			}
		}
	}

	return children
}

func getLowestPriorityDummy(afd *lib.AFD, state lib.AFDState) lib.AlphabetInput {
	var lowestPriority uint = math.MaxUint
	lowestDummy := lib.AlphabetInput{}
	for input := range afd.Transitions[state] {
		if input.IsDummy() {
			if input.GetDummy().Priority < lowestPriority {
				lowestPriority = input.GetDummy().Priority
				lowestDummy = input
			}
		}
	}

	if lowestPriority == math.MaxUint {
		panic("No lowest priority dummy found!")
	}
	return lowestDummy
}

func _simplifyIntoSwitch(afd *lib.AFD, state lib.AFDState, sw *afdSwitch) {
	if afd.AcceptanceStates.Contains(state) {
		return
	}

	for input, newState := range afd.Transitions[state] {
		if input.IsValue() {
			inputRune := input.GetValue().GetValue()
			_, found := sw.Transitions[state]
			if !found {
				sw.Transitions[state] = make(map[rune]afdLeafInfo)
			}

			sw.Transitions[state][inputRune] = afdLeafInfo{NewState: newState, Code: "return GIVE_NEXT"}
			_simplifyIntoSwitch(afd, newState, sw)
		}
	}

	dummyChildren := getChildrenWithDummyTransitions(afd, state)
	for _, tranInput := range dummyChildren {
		lowestPriorityDummy := getLowestPriorityDummy(afd, tranInput.State)
		code := lowestPriorityDummy.GetDummy().Code
		tranRune := tranInput.Input.GetValue().GetValue()
		sw.Transitions[state][tranRune] = afdLeafInfo{NewState: tranInput.State, Code: code}
	}
}

func WriteLexFile(filePath string, info LexFileData, afd lib.AFD) error {
	f, err := os.Create(filePath)
	if err != nil {
		panic("Error creating output file!")
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString(`
package main


// Lexer imports
import (
	"fmt"
	"os"
	"strconv"
	"strings"
)
	`)
	writer.WriteString(info.Header)
	writer.WriteString(`
const UNRECOGNIZABLE int = -1
const GIVE_NEXT int = -2

const CMD_HELP = `)
	writer.WriteRune('`')
	writer.WriteString(
		`Tokenizes a specified source file
Usage: lexer <source file>`)
	writer.WriteRune('`')
	writer.WriteString(`

type Optional[T any] struct {
	isValid bool
	value   T
}

func CreateValue[T any](val T) Optional[T] {
	return Optional[T]{value: val, isValid: true}
}

func CreateNull[T any]() Optional[T] {
	var defaultVal T
	return Optional[T]{value: defaultVal, isValid: false}
}

func (self Optional[T]) HasValue() bool {
	return self.isValid
}

func (self Optional[T]) GetValue() T {
	if !self.isValid {
		panic("Can't access not valid optional value!")
	} else {
		return self.value
	}
}

type Token struct {
	// When does this token start in the contents of the source file
	Start int
	// The type of the token that it found
	Type int
}

func (self *Token) String() string {
	b := strings.Builder{}
	b.WriteString("{ ")
	b.WriteString("Start = ")
	b.WriteString(strconv.Itoa(self.Start))
	b.WriteString(", Type = ")
	b.WriteString(strconv.Itoa(self.Type))
	b.WriteString(" }")
	return b.String()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Please supply only a source file as argument!\n")
		panic(CMD_HELP)
	}

	sourceFilePath := os.Args[1]
	sourceFileContent, err := os.ReadFile(sourceFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening the source file! %v", err)
	}

	// previousParsingResult := -1000
	// afdState := "0" // INITIAL AFD STATE!
	for i := 0; i < len(sourceFileContent); i++ {

		afdState := "0" // INITIAL AFD STATE!
		previousParsingResult := -1000
		j := 0
		for j = i; j < len(sourceFileContent); j++ {
			parsingResult := gettoken(&afdState, rune(sourceFileContent[j]))
			if parsingResult == UNRECOGNIZABLE {
				foundSomething := previousParsingResult != -1000
				if foundSomething {
					token := Token{Start: i, Type: previousParsingResult}
					fmt.Println(token.String())
					i = j - 1
					break
				} else {
					i = j
					break
				}
			} else if parsingResult != GIVE_NEXT {
				previousParsingResult = parsingResult
			}
		}
	}
}

func gettoken(state *string, input rune) int {
`)

	// TODO: Write AFD Logic into switch
	sw := simplifyIntoSwitch(&afd)
	sw.WriteTo(writer)
	writer.WriteRune('}')
	writer.WriteString(info.Footer)

	return writer.Flush()
}

func (s *afdSwitch) WriteTo(writer *bufio.Writer) {
	writer.WriteString("switch *state {\n")
	_writeTo(s, writer, s.InitialState)
	writer.WriteString(`
}
return UNRECOGNIZABLE
`)
}

func _writeTo(s *afdSwitch, w *bufio.Writer, state lib.AFDState) {
	w.WriteString("case \"")
	w.WriteString(state)
	w.WriteString(`":
	switch input {
`)

	for input, caseInfo := range s.Transitions[state] {
		w.WriteString("case '")
		w.WriteRune(input)
		w.WriteString(`':
		*state = "`)
		w.WriteString(caseInfo.NewState)
		w.WriteString("\"\n")
		w.WriteString(caseInfo.Code)
		w.WriteRune('\n')
	}
	w.WriteString("}\n")

	for _, caseInfo := range s.Transitions[state] {
		_writeTo(s, w, caseInfo.NewState)
	}
}
