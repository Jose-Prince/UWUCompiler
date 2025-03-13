package lib

import (
	"fmt"
	"strconv"
	"strings"
)

type AFDState = string
type AlphabetInput = string

type TransitionInput struct {
	State AFDState
	Input AlphabetInput
}

type AFD struct {
	InitialState AFDState
	// A dictionary that contains a bunch of states.
	// Each AFD state has another dictionary associated with it.
	// Each key on this second dictionary represents an input from the alphabet,
	// and the value is the new State the automata should transition.
	Transitions      map[AFDState]map[AlphabetInput]AFDState
	AcceptanceStates Set[AFDState]
}

type AFDPairType int

func (self *AFD) GetAllStates() []AFDState {
	out := []AFDState{}

	for state := range self.Transitions {
		out = append(out, state)
	}

	return out
}

func (self *AFD) IsAccepted(state *AFDState) bool {
	_, found := self.AcceptanceStates[*state]
	return found
}

func (self *AFD) MarkIfDistinguishable(aState *AFDState, bState *AFDState, table *AFDStateTable[AFDPairType]) AFDPairType {
	afd := *self

	if pairType, found := table.Get(aState, bState); found {
		return pairType
	}

	aTransitions := afd.Transitions[*aState]
	bTransitions := afd.Transitions[*bState]

	if len(aTransitions) != len(bTransitions) {
		msg := fmt.Sprintf("Supplied AFD is not really an AFD! The transitions length of these states didn't match\nAFD: %#v\nState A:%s\nState B: %s", afd, *aState, *bState)
		panic(msg)
	}

	if *aState == *bState {
		table.AddOrUpdate(*aState, *bState, EQUIVALENT)
		return EQUIVALENT
	}

	if afd.IsAccepted(aState) && !afd.IsAccepted(bState) ||
		afd.IsAccepted(bState) && !afd.IsAccepted(aState) {
		table.AddOrUpdate(*aState, *bState, DISTINCT)
		return DISTINCT
	}

	for input, aOutState := range aTransitions {
		bOutState, foundbOutState := bTransitions[input]
		if !foundbOutState {
			msg := fmt.Sprintf("B state doesn't contains the same input transition as A state!\nState B: %s\nState A: %s\nAFD: %#v", *bState, *aState, afd)
			panic(msg)
		}

		if aOutState == *aState && bOutState == *bState ||
			bOutState == *aState && aOutState == *bState {
			continue
		}

		derivedType := self.MarkIfDistinguishable(&aOutState, &bOutState, table)
		if DISTINCT == derivedType {
			table.AddOrUpdate(*aState, *bState, DISTINCT)
			return DISTINCT
		}
	}

	if pairType, found := table.Get(aState, bState); found {
		return pairType
	} else {
		table.AddOrUpdate(*aState, *bState, EQUIVALENT)
		return EQUIVALENT
	}
}

const (
	DISTINCT AFDPairType = iota
	EQUIVALENT
)

type AFDStateTable[T any] map[AFDState]map[AFDState]T

func (self *AFDStateTable[T]) PairAlreadyExists(a *AFDState, b *AFDState) bool {
	s := *self
	_, topLevelAFound := s[*a]
	_, topLevelBFound := s[*b]

	if !topLevelAFound || !topLevelBFound {
		return false
	}

	if topLevelAFound {
		if _, bFound := s[*a][*b]; bFound {
			return true
		}
	}

	if topLevelBFound {
		if _, aFound := s[*b][*a]; aFound {
			return true
		}
	}

	return false
}

func (self *AFDStateTable[T]) AddIfNotExists(a AFDState, b AFDState, stateType T) {
	s := *self

	_, aFound := s[a]
	if !aFound {
		s[a] = make(map[AFDState]T)
	}

	_, bFound := s[b]
	if !bFound {
		s[a][b] = stateType
	}

	_, bFound = s[b]
	if !bFound {
		s[b] = make(map[AFDState]T)
	}

	_, aFound = s[a]
	if !aFound {
		s[b][a] = stateType
	}
}

func (self *AFDStateTable[T]) AddOrUpdate(a AFDState, b AFDState, stateType T) {
	s := *self

	if _, found := s[a]; !found {
		s[a] = map[AFDState]T{}
	}
	s[a][b] = stateType

	if _, found := s[b]; !found {
		s[b] = map[AFDState]T{}
	}
	s[b][a] = stateType
}

func (self *AFDStateTable[T]) Get(a *AFDState, b *AFDState) (T, bool) {
	var defaultPairType T

	if !self.PairAlreadyExists(a, b) {
		return defaultPairType, false
	}

	if pairType, found := (*self)[*a][*b]; found {
		return pairType, true
	}

	if pairType, found := (*self)[*b][*a]; found {
		return pairType, true
	}

	return defaultPairType, false
}

func ConvertFromTableToAFD(table []*TableRow) *AFD {
	afd := &AFD{
		Transitions:      make(map[AFDState]map[AlphabetInput]AFDState),
		AcceptanceStates: NewSet[string](),
	}

	alphabet := NewSet[string]()

	// recognizes the alphabet of the afd
	for i := range table {
		if table[i].simbol != "" && table[i].simbol != "#" {
			alphabet.Add(table[i].simbol)
		}
	}

	// Creates trap state
	trapState := "TRAP"
	afd.Transitions[trapState] = make(map[AlphabetInput]AFDState)
	for value := range alphabet {
		afd.Transitions[trapState][value] = trapState
	}

	// Set AFD initial state
	afd.InitialState = convertSliceIntToString(table[len(table)-1].firstpos)

	var states Queue[string]
	states.Enqueue(afd.InitialState)

	visited := NewSet[string]()
	visited.Add(afd.InitialState)

	// Determines transitions for AFD
	for !states.IsEmpty() {
		currentState, _ := states.Dequeue()

		for value := range alphabet {
			var nextState []int
			nextStateSet := make(map[int]bool)
			for _, strIndex := range strings.Split(currentState, ",") {
				if strIndex != "" {
					index, err := strconv.Atoi(strIndex)
					if err != nil {
						fmt.Println("Error converting to string:", err)
						continue
					}

					if table[index].simbol == value {
						for _, pos := range table[index].followpos {
							nextStateSet[pos] = true
						}
					}
				}
			}

			nextState = make([]int, 0, len(nextStateSet))
			for key := range nextStateSet {
				nextState = append(nextState, key)
			}

			strNextState := convertSliceIntToString(nextState)
			if strNextState == "" {
				strNextState = trapState
			}

			if _, exists := afd.Transitions[currentState]; !exists {
				afd.Transitions[currentState] = make(map[AlphabetInput]AFDState)
			}

			afd.Transitions[currentState][value] = strNextState

			if strNextState != trapState && !visited.Contains(strNextState) {
				visited.Add(strNextState)
				states.Enqueue(strNextState)
			}
		}

	}

	// Determines final states
	finalNode := len(table) - 2

	for i := range visited {
		if strings.Contains(i, fmt.Sprintf("%d", finalNode)) {
			afd.AcceptanceStates.Add(i)
		}
	}

	return afd
}

func (self *AFD) Derivation(w string) bool {
	state := self.InitialState
	for _, ch := range w {
		state = self.Transitions[state][string(ch)]
	}

	return self.AcceptanceStates.Contains(state)
}

func convertSliceIntToString(slice []int) string {
	var sb strings.Builder
	for _, i := range slice {
		sb.WriteString(fmt.Sprintf("%d,", i))
	}

	return sb.String()
}

func stringToIntSlice(str string) []int {
	var intSlice []int
	for _, s := range str {
		num, err := strconv.Atoi(string(s))
		if err != nil {
			return []int{}
		}
		intSlice = append(intSlice, num)
	}
	return intSlice
}
