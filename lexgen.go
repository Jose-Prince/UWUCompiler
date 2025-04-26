package main

import (
	"bufio"
	"math"
	"os"

	l "github.com/Jose-Prince/UWULexer/lib"
	reg "github.com/Jose-Prince/UWULexer/lib/regex"
)

type afdLeafInfo struct {
	Code     string
	NewState reg.AFDState
}

type afdSwitch struct {
	// The first key is the state
	// The second key is the input
	// The third key is the nextState and the code to write
	Transitions map[reg.AFDState]map[rune]afdLeafInfo

	InitialState     reg.AFDState
	AcceptanceStates l.Set[reg.AFDState]
}

func simplifyIntoSwitch(afd *reg.AFD) afdSwitch {
	visitedSet := l.Set[reg.AFDState]{}
	sw := afdSwitch{
		InitialState:     afd.InitialState,
		AcceptanceStates: afd.AcceptanceStates,
		Transitions:      make(map[reg.AFDState]map[rune]afdLeafInfo),
	}
	_simplifyIntoSwitch(afd, afd.InitialState, &sw, &visitedSet)
	return sw
}

type tranInput struct {
	State reg.AFDState
}

func getChildrenWithDummyTransitions(afd *reg.AFD, state reg.AFDState) l.Set[reg.AFDState] {
	children := l.Set[reg.AFDState]{}

	for _, childState := range afd.Transitions[state] {
		for input := range afd.Transitions[childState] {
			if input.IsDummy() {
				children.Add(childState)
			}
		}
	}

	return children
}

func getLowestPriorityDummy(afd *reg.AFD, state reg.AFDState) reg.AlphabetInput {
	var lowestPriority uint = math.MaxUint
	lowestDummy := reg.AlphabetInput{}
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

func _simplifyIntoSwitch(afd *reg.AFD, state reg.AFDState, sw *afdSwitch, visitedSet *l.Set[reg.AFDState]) {
	if afd.AcceptanceStates.Contains(state) || !visitedSet.Add(state) {
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
			_simplifyIntoSwitch(afd, newState, sw, visitedSet)
		}
	}

	dummyChildren := getChildrenWithDummyTransitions(afd, state)
	for tranInput := range dummyChildren {
		lowestPriorityDummy := getLowestPriorityDummy(afd, tranInput)
		code := lowestPriorityDummy.GetDummy().Code
		tranRune := rune(0)
		for input, childreState := range afd.Transitions[state] {
			if childreState == tranInput {
				tranRune = input.GetValue().GetValue()
				break
			}
		}
		sw.Transitions[state][tranRune] = afdLeafInfo{NewState: tranInput, Code: code}
	}
}

func WriteLexFile(filePath string, info LexFileData, afd reg.AFD) error {
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

		afdState := "`)
	writer.WriteString(afd.InitialState)
	writer.WriteString(`" // INITIAL AFD STATE!

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
	alreadyWrittenStates := l.Set[reg.AFDState]{}
	writer.WriteString("switch *state {\n")
	_writeTo(s, writer, s.InitialState, &alreadyWrittenStates)
	writer.WriteString(`
}
return UNRECOGNIZABLE
`)
}

func _writeTo(s *afdSwitch, w *bufio.Writer, state reg.AFDState, alreadyWrittenStates *l.Set[reg.AFDState]) {
	if !alreadyWrittenStates.Add(state) {
		return
	}

	w.WriteString("case \"")
	w.WriteString(state)
	w.WriteString(`":
	switch input {
`)

	for input, caseInfo := range s.Transitions[state] {
		w.WriteString("case '")
		switch input {
		case '\t':
			w.WriteString("\\t")
		case '\n':
			w.WriteString("\\n")
		case '\r':
			w.WriteString("\\r")
		default:
			w.WriteRune(input)
		}
		w.WriteString(`':
		*state = "`)
		w.WriteString(caseInfo.NewState)
		w.WriteString("\"\n")
		w.WriteString(caseInfo.Code)
		w.WriteRune('\n')
	}
	w.WriteString("}\n")

	for _, caseInfo := range s.Transitions[state] {
		_writeTo(s, w, caseInfo.NewState, alreadyWrittenStates)
	}
}
