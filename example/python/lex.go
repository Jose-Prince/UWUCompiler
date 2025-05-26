package main

// Lexer imports
import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	NUMBER int = iota
	PLUS
	MINUS
	TIMES
	DIV
	LPAREN
	RPAREN
)

const UNRECOGNIZABLE int = -1
const GIVE_NEXT int = -2

const CMD_HELP = `Tokenizes a specified source file
Usage: lexer <source file>`

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
		afdState := "[ 0, 3, 4, 6, 8, 10, 12, 14, 16, 18, 20, 46, 50, 54, 58, 62, ]" // INITIAL AFD STATE!

		previousParsingResult := -1000
		j := 0
		for j = i; j < len(sourceFileContent); j++ {
			parsingResult := gettoken(&afdState, rune(sourceFileContent[j]))
			if parsingResult == UNRECOGNIZABLE {
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

	table := ParsingTable{}

	stack := Stack[ParseItem]{}
	stack = append(stack, CreateNodeItem(table.InitialNodeId))
	for _, token := range tokens {
		if val := stack.Peek(); !val.HasValue() {
			panic("Invalid parsing state! Stack is empty!")
		}

		item := stack.Pop().GetValue()
		if !item.IsNodeId() {
			panic("Invalid parsing state! The item on the stack is not a NodeID!")
		}

		nodeId := item.GetNodeId()

		// go check the action table
		isTerminal := table.Original.Terminals.Contains(token.Type)
		if isTerminal {
			action := table.ActionTable[nodeId][token.Type]
			if action.Accept {
				fmt.Println("ACCEPTED")
			} else if action.IsShift() {
				stack.Push(CreateTokenItem(token.Type))
				stack.Push(CreateNodeItem(action.GetShift()))

			} else if action.IsReduce() {
				idx := action.GetReduce()
				rule := table.Original.Rules[idx]
				productionsCopy := make([]GrammarToken, 0, len(rule.Production))
				copy(productionsCopy, rule.Production)

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
				stack.Push(CreateTokenItem(rule.Head))

				// Now we execute the follow
				nonTerminalToken := stack.Pop().GetValue()
				gotoNodeId := stack.Pop().GetValue()

				newNodeId := table.GoToTable[gotoNodeId.GetNodeId()][nonTerminalToken.GetToken()]
				stack.Push(gotoNodeId)
				stack.Push(nonTerminalToken)
				stack.Push(CreateNodeItem(newNodeId))
			}
		} else {
			panic("Token should always be a terminal!")
		}
	}
}

func gettoken(state *string, input rune) int {
	switch *state {
	case "[ 0, 3, 4, 6, 8, 10, 12, 14, 16, 18, 20, 46, 50, 54, 58, 62, ]":
		switch input {
		case '4':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '8':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '*':
			*state = "[ 55, ]"
			return TIMES
		case '2':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case ')':
			*state = "[ 1, ]"
			return RPAREN
		case '1':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '0':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '9':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '5':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '/':
			*state = "[ 59, ]"
			return DIV
		case '(':
			*state = "[ 63, ]"
			return LPAREN
		case '7':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '6':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '3':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '-':
			*state = "[ 51, ]"
			return MINUS
		case '+':
			*state = "[ 47, ]"
			return PLUS
		}
	case "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]":
		switch input {
		case '2':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '0':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '4':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '8':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '9':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '5':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '3':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '1':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '6':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		case '7':
			*state = "[ 22, 23, 25, 27, 29, 31, 33, 35, 37, 39, 43, ]"
			return NUMBER
		}
	case "[ 51, ]":
		switch input {
		}
	case "[ 47, ]":
		switch input {
		}
	case "[ 55, ]":
		switch input {
		}
	case "[ 1, ]":
		switch input {
		}
	case "[ 59, ]":
		switch input {
		}
	case "[ 63, ]":
		switch input {
		}

	}
	return UNRECOGNIZABLE
}
