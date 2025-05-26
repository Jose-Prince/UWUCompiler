package parsertypes

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/Jose-Prince/UWUCompiler/lib"
)

type AFDNodeId = string
type GrammarToken = int

// An action can be either:
// * a shift action
// * a reduce action
// * an accept state
type Action struct {
	// Shift to another AFDNode
	Shift lib.Optional[AFDNodeId]
	// Reduce according to a production in the original grammar
	Reduce lib.Optional[int]
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

type EpsilonString = lib.Optional[string]

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

func (self *Stack[T]) Peek() lib.Optional[T] {
	idx := len(*self) - 1
	ref := *self

	if idx < 0 {
		return lib.CreateNull[T]()
	}

	return lib.CreateValue(ref[idx])
}

func (self *Stack[T]) Push(val T) *Stack[T] {
	*self = append(*self, val)
	return self
}

func (self *Stack[T]) Pop() lib.Optional[T] {
	ref := *self
	length := len(ref)
	idx := length - 1

	if idx < 0 {
		return lib.CreateNull[T]()
	}

	val := ref[idx]
	*self = ref[:idx]

	return lib.CreateValue(val)
}

type ParseItem struct {
	Token  lib.Optional[GrammarToken]
	NodeId lib.Optional[AFDNodeId]
}

func CreateTokenItem(token GrammarToken) ParseItem {
	return ParseItem{
		Token:  lib.CreateValue(token),
		NodeId: lib.CreateNull[AFDNodeId](),
	}
}

func CreateNodeItem(nodeId AFDNodeId) ParseItem {
	return ParseItem{
		Token:  lib.CreateNull[GrammarToken](),
		NodeId: lib.CreateValue(nodeId),
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
