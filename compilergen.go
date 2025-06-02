package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	l "github.com/Jose-Prince/UWUCompiler/lib"
	"github.com/Jose-Prince/UWUCompiler/lib/grammar"
	reg "github.com/Jose-Prince/UWUCompiler/lib/regex"
	parsertypes "github.com/Jose-Prince/UWUCompiler/parserTypes"
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

func getChildrenWithDummyTransitions(afd *reg.AFD, state reg.AFDState) l.Set[reg.AFDState] {
	children := l.Set[reg.AFDState]{}

	for stateInput, childState := range afd.Transitions[state] {
		if stateInput.IsDummy() {
			continue
		}

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

	childrenWithDummy := getChildrenWithDummyTransitions(afd, state)
	for childState := range childrenWithDummy {
		lowestPriorityDummy := getLowestPriorityDummy(afd, childState)
		code := lowestPriorityDummy.GetDummy().Code
		tranRune := rune(0)
		for input, childrenState := range afd.Transitions[state] {
			if childrenState == childState {
				tranRune = input.GetValue().GetValue()
				// if afd.StateHasOnlyDummyTransitions(childState) {
				// 	sw.Transitions[state][tranRune] = afdLeafInfo{NewState: state, Code: code}
				// } else {
				// 	sw.Transitions[state][tranRune] = afdLeafInfo{NewState: childState, Code: code}
				// }
				sw.Transitions[state][tranRune] = afdLeafInfo{NewState: childState, Code: code}
			}
		}
	}
}

func WriteCompilerFile(filePath string, info *CompilerFileInfo) error {
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
	"cmp"
	"slices"
)
	`)
	writer.WriteString(info.LexInfo.Header)
	writer.WriteString(`
const END_TOKEN_TYPE =`)
	endTk := grammar.NewEndToken()
	writer.WriteString(strconv.FormatInt(int64(info.ParsingTable.Original.TokenToParserType(&endTk)), 10))
	writer.WriteString(`
const UNRECOGNIZABLE int = -1
const GIVE_NEXT int = -2
const INVALID_TRANSITION int = -3

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

type AFDNodeId = string
type GrammarToken = int

// An action can be either:
// * a shift action
// * a reduce action
// * an accept state
type Action struct {
	// Shift to another AFDNode
	Shift Optional[AFDNodeId]
	// Reduce according to a production in the original grammar
	Reduce Optional[int]
	Accept bool
}

func (s Action) IsShift() bool {
	return s.Shift.HasValue()
}

func (s Action) GetShift() AFDNodeId {
	if !s.IsShift() {
		panic("Can't get shift of an action that isn't a shift!")
	}

	return s.Shift.GetValue()
}

func (s Action) IsReduce() bool {
	return s.Reduce.HasValue()
}

func (s Action) GetReduce() int {
	if !s.IsReduce() {
		panic("Can't get reduce of an action that isn't a reduce!")
	}

	return s.Reduce.GetValue()
}

type EpsilonString = Optional[string]

type GrammarRule struct {
	Head       GrammarToken
	Production []GrammarToken
}

type Set[T comparable] map[T]struct{}

// Prints a set as a string.
//
// Since a Set is internally represented as a map, the keys will be unordered!
func (self Set[T]) String() string {
	b := strings.Builder{}
	b.WriteString("[ ")

	for k := range self {
		b.WriteString(fmt.Sprint(k))
		b.WriteString(", ")
	}

	b.WriteString("]")
	return b.String()
}

func GetValuesStable[T cmp.Ordered](self Set[T]) []T {
	values := make([]T, 0, len(self))
	for k := range self {
		values = append(values, k)
	}

	slices.Sort(values)

	return values
}

// Prints a set as a string with it's keys on the same order every time!
func StableSetString[T cmp.Ordered](self Set[T]) string {
	b := strings.Builder{}
	b.WriteString("[ ")

	values := GetValuesStable(self)
	for _, k := range values {
		b.WriteString(fmt.Sprint(k))
		b.WriteString(", ")
	}

	b.WriteString("]")
	return b.String()
}

// Checks if self is equal to other.
//
// Equal means that all items in self are contained in other and no more items are in other.
func (self *Set[T]) Equals(other *Set[T]) bool {
	if len(*self) != len(*other) {
		return false
	}

	for k := range *self {
		if !other.Contains(k) {
			return false
		}
	}

	return true
}

// Checks if an element exists on the set.
//
// Returns True if the value is contained in the set.
func (self *Set[T]) Contains(val T) bool {
	_, alreadyAdded := (*self)[val]

	return alreadyAdded
}

// Adds an element to the set.
//
// Returns True if the element is new to the set,
// false otherwise.
func (self *Set[T]) Add(val T) bool {
	ref := *self
	_, alreadyAdded := ref[val]

	if !alreadyAdded {
		ref[val] = struct{}{}
	}

	return !alreadyAdded
}

// Adds all values from other into self.
func (self *Set[T]) Merge(other *Set[T]) {
	for val := range *other {
		self.Add(val)
	}
}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (self *Set[T]) IsEmpty() bool {
	return len(*self) == 0
}

func (self *Set[T]) Clear() {
	for k := range *self {
		delete(*self, k)
	}
}

func (self *Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(*self))
	for k := range *self {
		slice = append(slice, k)
	}

	return slice
}

type Grammar struct {
	InitialSimbol GrammarToken
	Rules         []GrammarRule
	Terminals     Set[GrammarToken]
	NonTerminals  Set[GrammarToken]
}

type ParsingTable struct {
	// The Action table contains all the reduce and shifts of the parsing table.
	ActionTable map[AFDNodeId]map[GrammarToken]Action
	// The GoTo table contains all the nonterminal tokens and what transitions to make of them.
	GoToTable map[AFDNodeId]map[GrammarToken]AFDNodeId
	// The original grammar, IT MUST NOT BE EXPANDED!
	Original Grammar

	// The node to start parsing from
	InitialNodeId AFDNodeId
}

type Stack[T any] []T

func NewStack[T any]() Stack[T] {
	return Stack[T]{}
}

func (self *Stack[T]) Empty() bool {
	return len(*self) == 0
}

func (self *Stack[T]) Peek() Optional[T] {
	idx := len(*self) - 1
	ref := *self

	if idx < 0 {
		return CreateNull[T]()
	}

	return CreateValue(ref[idx])
}

func (self *Stack[T]) Push(val T) *Stack[T] {
	*self = append(*self, val)
	return self
}

func (self *Stack[T]) Pop() Optional[T] {
	ref := *self
	length := len(ref)
	idx := length - 1

	if idx < 0 {
		return CreateNull[T]()
	}

	val := ref[idx]
	*self = ref[:idx]

	return CreateValue(val)
}

type ParseItem struct {
	Token  Optional[GrammarToken]
	NodeId Optional[AFDNodeId]
}

func CreateTokenItem(token GrammarToken) ParseItem {
	return ParseItem{
		Token:  CreateValue(token),
		NodeId: CreateNull[AFDNodeId](),
	}
}

func CreateNodeItem(nodeId AFDNodeId) ParseItem {
	return ParseItem{
		Token:  CreateNull[GrammarToken](),
		NodeId: CreateValue(nodeId),
	}
}

func (item ParseItem) IsToken() bool {
	return item.Token.HasValue()
}

func (item ParseItem) IsNodeId() bool {
	return item.NodeId.HasValue()
}

func (item ParseItem) GetNodeId() AFDNodeId {
	if !item.IsNodeId() {
		panic("Invalid acces! Can't get node id from invalid item")
	}

	return item.NodeId.GetValue()
}

func (item ParseItem) GetToken() GrammarToken {
	if !item.IsToken() {
		panic("Invalid acces! Can't get node id from invalid item")
	}

	return item.Token.GetValue()
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

	tokens := make([]Token, 0, 1000)

	for i := 0; i < len(sourceFileContent); i++ {
		afdState := "`)
	writer.WriteString(info.LexAFD.InitialState)
	writer.WriteString(`" // INITIAL AFD STATE!

		previousParsingResult := -1000
		// previousParsingResult := UNRECOGNIZABLE
		j := 0
		for j = i; j < len(sourceFileContent); j++ {
			parsingResult := gettoken(&afdState, rune(sourceFileContent[j]))
			if parsingResult == INVALID_TRANSITION {
				line, col := getLineAndCol(sourceFileContent, j)
				start := min(i, len(tokens), 3)
				panic(fmt.Sprintf(`)
	writer.WriteString("`")
	writer.WriteString(`
SYNTAX ERROR: Unexpected character (%c)
==============================================
ON (%s:%d:%d)
%s`)
	writer.WriteString("`, sourceFileContent[j], sourceFilePath, line, col, sourceFileContent[start:j+2]))")

	writer.WriteString(`} else if parsingResult == UNRECOGNIZABLE {
				foundSomething := previousParsingResult != -1000
				if foundSomething {
					token := Token{Start: i, Type: previousParsingResult}
					tokens = append(tokens, token)
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

	tokens = append(tokens, Token {Start: len(sourceFileContent), Type: END_TOKEN_TYPE})

	`)

	writer.WriteString("table := ")
	var transformedTable parsertypes.ParsingTable = info.ParsingTable.ToParserTable()
	writer.WriteString(removeModulesFromStaticType(fmt.Sprintf("%#v", transformedTable)))

	writer.WriteString(`

	isAccepted := false
	stack := Stack[ParseItem]{}
	stack = append(stack, CreateNodeItem(table.InitialNodeId))
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if val := stack.Peek(); !val.HasValue() {
			panic("Invalid parsing state! Stack is empty!")
		}

		item := stack.Peek().GetValue()
		if !item.IsNodeId() {
			panic("Invalid parsing state! The item on the stack is not a NodeID!")
		}

		nodeId := item.GetNodeId()

		// go check the action table
		isTerminal := table.Original.Terminals.Contains(token.Type)
		if isTerminal {
			action := table.ActionTable[nodeId][token.Type]
			if action.Accept {
				isAccepted = true
				break
			} else if action.IsShift() {
				stack.Push(CreateTokenItem(token.Type))
				stack.Push(CreateNodeItem(action.GetShift()))

			} else if action.IsReduce() {
				idx := action.GetReduce()
				// rule := table.Original.Rules[idx]
				productionsCopy := make([]GrammarToken, len(table.Original.Rules[idx].Production))
				copy(productionsCopy, table.Original.Rules[idx].Production)

				for len(productionsCopy) > 0 {
					reduceItem := stack.Pop()
					if !reduceItem.HasValue() {
						panic("Invalid parsing state! The stack is empty, can't keep up reducing!")
					}

					{
						reduceItem := reduceItem.GetValue()
						if reduceItem.IsToken() {
							itemIdx := -1
							for j, prodToken := range productionsCopy {
								if prodToken == reduceItem.Token.GetValue() {
									itemIdx = j
									break
								}
							}

							if itemIdx == -1 {
								panic("Token not found in reduce production!")
							}
							productionsCopy = slices.Delete(productionsCopy, itemIdx, itemIdx+1)
						}

					}
				}
				stack.Push(CreateTokenItem(table.Original.Rules[idx].Head))

				// Now we execute the follow
				nonTerminalToken := stack.Pop().GetValue()
				gotoNodeId := stack.Pop().GetValue()

				newNodeId := table.GoToTable[gotoNodeId.GetNodeId()][nonTerminalToken.GetToken()]
				stack.Push(gotoNodeId)
				stack.Push(nonTerminalToken)
				stack.Push(CreateNodeItem(newNodeId))
				i--
			}
		} else {
			panic("Token should always be a terminal!")
		}
	}

	if isAccepted {
		fmt.Println("The input is accepted!")
	} else {
		fmt.Println("The input can't be accepted!")
	}
}

func getLineAndCol(contents []byte, idx int) (int, int) {
	line := 0
	col := 0

	for i := range idx {
		if contents[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	return line+1, col+1
}

func gettoken(state *string, input rune) int {
`)

	sw := simplifyIntoSwitch(&info.LexAFD)
	sw.WriteTo(writer)
	writer.WriteRune('}')
	writer.WriteString(info.LexInfo.Footer)

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
		case '\'':
			w.WriteString("\\'")
		case '\\':
			w.WriteString("\\\\")
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
	if len(s.Transitions[state]) > 0 {
		w.WriteString(`default:
	return INVALID_TRANSITION
`)
	}
	w.WriteString("}\n")

	for _, caseInfo := range s.Transitions[state] {
		_writeTo(s, w, caseInfo.NewState, alreadyWrittenStates)
	}
}

func removeModulesFromStaticType(t string) string {
	reg, err := regexp.Compile(`([A-Za-z\/-]+)\.`)
	if err != nil {
		panic(err)
	}

	replaced := reg.ReplaceAll([]byte(t), []byte{})
	return string(replaced)
}
