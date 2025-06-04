package grammar

import (
	// "fmt"
	// "sort"
	// "strconv"

	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Jose-Prince/UWUCompiler/lib"
)

type AutomataStateIndex = string
type AlphabetInput = GrammarToken

type Automata struct {
	InitialState     AutomataStateIndex
	Transitions      map[AutomataStateIndex]map[AlphabetInput]AutomataStateIndex
	AcceptanceStates lib.Set[AutomataStateIndex]

	Nodes map[AutomataStateIndex]AutomataState
}

func (auto *Automata) FindIndexOfState(state *AutomataState) (AutomataStateIndex, bool) {
	finalIdx := ""
	foundSomething := false

NodeLoop:
	for idx, st := range auto.Nodes {
		if len(st.Items) != len(state.Items) {
			continue
		}

		for i, rule := range st.Items {
			stateIRule := state.Items[i]

			if !rule.Equals(&stateIRule) {
				continue NodeLoop
			}
		}

		finalIdx = idx
		foundSomething = true
		break
	}

	return finalIdx, foundSomething
}

type AutomataState struct {
	Items []AutomataItem
}

func (state *AutomataState) EQ_WithoutLookAhead(other *AutomataState) bool {
	if len(state.Items) != len(other.Items) {
		return false
	}

	for i, r := range state.Items {
		other_r := other.Items[i]
		if !r.EqualsWithoutLookahead(&other_r) {
			return false
		}
	}
	return true
}

type AutomataItem struct {
	Head       GrammarToken
	Production []GrammarToken
	Dot        int
	Lookahead  lib.Set[GrammarToken]
}

func (item *AutomataItem) DotIsAtEnd() bool {
	return item.Dot >= len(item.Production)
}

func (rule *AutomataItem) EqualsWithoutLookahead(other *AutomataItem) bool {
	if len(rule.Production) != len(other.Production) {
		return false
	}

	if !rule.Head.Equal(&other.Head) || rule.Dot != rule.Dot {
		return false
	}

	for i, prod := range rule.Production {
		other_i := other.Production[i]
		if !prod.Equal(&other_i) {
			return false
		}
	}
	return true
}

func (rule *AutomataItem) Equals(other *AutomataItem) bool {
	if len(rule.Production) != len(other.Production) {
		return false
	}

	if !rule.Head.Equal(&other.Head) || rule.Dot != other.Dot {
		return false
	}

	for i := range other.Lookahead {
		rule.Lookahead.Add(i)
	}

	for i, prod := range rule.Production {
		other_i := other.Production[i]
		if !prod.Equal(&other_i) {
			return false
		}
	}
	return true
}

func InitializeAutomata(initialRule GrammarRule, grammar Grammar) Automata {
	firsts := NewFirstFollowTable()
	GetFirsts(&grammar, &firsts)

	lr1 := Automata{
		InitialState:     "",
		Transitions:      make(map[AutomataStateIndex]map[AlphabetInput]AutomataStateIndex),
		AcceptanceStates: lib.NewSet[AutomataStateIndex](),
		Nodes:            make(map[AutomataStateIndex]AutomataState),
	}

	// S' -> .Expr , $
	lookAhead := lib.NewSet[GrammarToken]()
	lookAhead.Add(NewEndToken())
	initRule := AutomataItem{
		Head:       initialRule.Head,
		Production: initialRule.Production,
		Dot:        0,
		Lookahead:  lookAhead,
	}

	state := AutomataState{
		Items: []AutomataItem{initRule},
	}

	closure(&state, &grammar, &firsts)

	lr1.InitialState = "0"
	lr1.Nodes[lr1.InitialState] = state

	alreadyVisited := lib.NewSet[AutomataStateIndex]()
	queue := lib.NewQueue[AutomataStateIndex]()
	queue.Enqueue(lr1.InitialState)
	for !queue.IsEmpty() {
		currentState, _ := queue.Dequeue()
		generateStates(currentState, &lr1, &grammar, &firsts, queue, &alreadyVisited)
	}

	return lr1
}

type GrammarTokenRulePair struct {
	Token GrammarToken
	Rule  AutomataItem
}

func (item *AutomataItem) ExpansionKey() string {
	b := strings.Builder{}

	b.WriteString(item.Head.String())
	// b.WriteString(" -> [ ")
	// for _, prod := range item.Production {
	// 	b.WriteString(prod.String())
	// 	b.WriteString(" ")
	// }
	// b.WriteString("] ")
	// b.WriteString(strconv.FormatInt(int64(item.Dot), 10))

	b.WriteString(" LK: { ")
	for k := range item.Lookahead {
		b.WriteString(k.String())
		b.WriteString(" ")
	}
	b.WriteString("} ")

	return b.String()
}

func (item *AutomataItem) ToUniqueString() string {
	b := strings.Builder{}

	b.WriteString(item.Head.String())
	b.WriteString(" -> [ ")
	for _, prod := range item.Production {
		b.WriteString(prod.String())
		b.WriteString(" ")
	}
	b.WriteString("] ")
	b.WriteString(strconv.FormatInt(int64(item.Dot), 10))

	b.WriteString(" LK: { ")

	lkSlices := make([]string, 0, len(item.Lookahead))
	for lk := range item.Lookahead {
		lkSlices = append(lkSlices, lk.String())
	}
	slices.Sort(lkSlices)

	for _, k := range lkSlices {
		b.WriteString(k)
		b.WriteString(" ")
	}
	b.WriteString("} ")

	return b.String()
}

func closure(
	state *AutomataState,
	grammar *Grammar,
	firsts *FirstFollowTable,
) {

	alreadyComputedItems := lib.NewSet[string]()
	queueOfRules := lib.NewQueue[AutomataItem]()
	for _, v := range state.Items {
		alreadyComputedItems.Add(v.ToUniqueString())
		queueOfRules.Enqueue(v)
	}

	// A → α • X β c
	// S -> lkajsdlfkjasd •  X ( asdf ) , $
	// S -> lkajsdlfkjasd • 	X , $

	// From https://ocw.mit.edu/courses/6-035-computer-language-engineering-spring-2010/c86c6ebce6973a6f8441f200a3b34fbd_MIT6_035S10_lec03b.pdf
	// repeat
	// 	for all items [A → α • X β c] in I
	// 		for any production X → γ
	// 			for any d ∈ First(βc)
	// 				I = I ∪ { [X → • γ d] }
	// until I does not change

	addedNewItem := true
	for addedNewItem {
		addedNewItem = false

		for _, item := range state.Items {
			if item.Dot >= len(item.Production) {
				continue
			}
			dotToken := item.Production[item.Dot]

			for _, prod := range grammar.Rules {
				if !prod.Head.Equal(&dotToken) {
					continue
				}

				lookAhead := lib.NewSet[GrammarToken]()
				if item.Dot+1 < len(item.Production) {
					merge := firsts.table[item.Production[item.Dot+1]]
					lookAhead.Merge(&merge.First)
				} else {
					lookAhead.Merge(&item.Lookahead)
				}

				newItem := AutomataItem{
					Head:       dotToken,
					Production: prod.Production,
					Dot:        0,
					Lookahead:  lookAhead,
				}

				if alreadyComputedItems.Add(newItem.ToUniqueString()) {
					addedNewItem = true
					state.Items = append(state.Items, newItem)
				}
			}
		}
	}

	// nextTokens := []GrammarTokenRulePair{}
	// for _, rule := range grammar.Rules {
	// 	if rule.Head.Equal(&token) {
	// 		dotToken := rule.Production[0]
	//
	// 		lookAhead := lib.NewSet[GrammarToken]()
	// 		if 1 < len(rule.Production) {
	// 			firsts := firsts.table[rule.Production[1]].First
	// 			lookAhead.Merge(&firsts)
	// 		}
	// 		if initRule.Dot+1 < len(initRule.Production) {
	// 			tk := initRule.Production[initRule.Dot+1]
	// 			firsts := firsts.table[tk].First
	// 			lookAhead.Merge(&firsts)
	// 		} else if initRule.Lookahead.IsEmpty() {
	// 			lookAhead.Add(NewEndToken())
	// 		} else {
	// 			lookAhead.Merge(&initRule.Lookahead)
	// 		}
	//
	// 		newRule := AutomataRule{
	// 			Head:       rule.Head,
	// 			Production: rule.Production,
	// 			Dot:        0,
	// 			Lookahead:  lookAhead,
	// 		}
	//
	// 		if !alreadyComputed.Contains(dotToken) && dotToken.IsNonTerminal() {
	// 			nextTokens = append(nextTokens, GrammarTokenRulePair{
	// 				Token: dotToken,
	// 				Rule:  newRule,
	// 			})
	// 		}
	// 		state.Rules = append(state.Rules, newRule)
	// 	}
	// }

	// for _, pair := range nextTokens {
	// 	closure(pair.Token, &pair.Rule, state, grammar, firsts, alreadyComputed)
	// }
}

func goTo(
	state *AutomataState,
	token GrammarToken,
	grammar *Grammar,
	firsts *FirstFollowTable,
) AutomataState {
	// https://ocw.mit.edu/courses/6-035-computer-language-engineering-spring-2010/c86c6ebce6973a6f8441f200a3b34fbd_MIT6_035S10_lec03b.pdf
	// Goto(I, X)
	// 	J = { }
	// 	for any item [A → α • X β c] in I
	// 		J = J ∪ {[A → α X • β c] }
	// 	return Closure(J)

	newState := AutomataState{
		Items: make([]AutomataItem, 0, len(state.Items)),
	}
	alreadyAddedItems := lib.NewSet[string]()

	for _, item := range state.Items {
		if item.DotIsAtEnd() {
			continue
		}
		dotToken := item.Production[item.Dot]

		if !dotToken.Equal(&token) {
			continue
		}

		transformedItem := AutomataItem{
			Head:       item.Head,
			Production: item.Production,
			Dot:        item.Dot + 1,
			Lookahead:  item.Lookahead.Copy(),
		}
		if alreadyAddedItems.Add(transformedItem.ToUniqueString()) {
			newState.Items = append(newState.Items, transformedItem)
		}
	}

	closure(&newState, grammar, firsts)
	return newState
}

type InputStateRulesPair struct {
	Input GrammarToken
	Rules []AutomataItem
}

func generateStates(
	currentIdx AutomataStateIndex,
	automata *Automata,
	grammar *Grammar,
	firsts *FirstFollowTable,
	queue *lib.Queue[AutomataStateIndex],
	alreadyVisited *lib.Set[AutomataStateIndex],
) {
	if !alreadyVisited.Add(currentIdx) {
		return
	}
	currentState := automata.Nodes[currentIdx]

	for _, item := range currentState.Items {
		if item.DotIsAtEnd() {
			continue
		}
		dotToken := item.Production[item.Dot]

		newState := goTo(&currentState, dotToken, grammar, firsts)
		idx, found := automata.FindIndexOfState(&newState)
		if !found {
			idx = strconv.FormatInt(int64(len(automata.Nodes)), 10)
			automata.Nodes[idx] = newState
			queue.Enqueue(idx)
		}

		if _, found := automata.Transitions[currentIdx]; !found {
			automata.Transitions[currentIdx] = make(map[AlphabetInput]AutomataStateIndex)
		}
		// if _, exists := automata.Transitions[currentIdx][dotToken]; !exists {
		automata.Transitions[currentIdx][dotToken] = idx
		// } else {
		// 	fmt.Printf("State transition collision, taking oldest...")
		// 	// panic("State transition collision!")
		// }
	}

	// alreadyScannedInputs := lib.NewSet[GrammarToken]()
	// transitionInputs := make([]InputStateRulesPair, 0, len(currentState.Rules))
	// for _, rule := range currentState.Rules {
	// 	if rule.Dot >= len(rule.Production) {
	// 		continue
	// 	}
	//
	// 	dotOfRule := rule.Production[rule.Dot]
	// 	if !alreadyScannedInputs.Add(dotOfRule) {
	// 		continue
	// 	}
	//
	// 	pair := InputStateRulesPair{
	// 		Input: dotOfRule,
	// 		Rules: []AutomataRule{},
	// 	}
	// 	for _, rr := range currentState.Rules {
	// 		if rr.Dot >= len(rr.Production) {
	// 			continue
	// 		}
	//
	// 		dotOfRule2 := rr.Production[rr.Dot]
	// 		if dotOfRule.Equal(&dotOfRule2) {
	// 			pair.Rules = append(pair.Rules, rr)
	// 		}
	// 	}
	//
	// 	transitionInputs = append(transitionInputs, pair)
	// }
	//
	// for _, pair := range transitionInputs {
	// 	transitionToken := pair.Input
	// 	newState := AutomataState{
	// 		Rules: make([]AutomataRule, 0, len(pair.Rules)),
	// 	}
	//
	// 	for _, rule := range pair.Rules {
	// 		if rule.Dot >= len(rule.Production) {
	// 			continue // The dot has reached the end
	// 		}
	//
	// 		newInitialRule := AutomataRule{
	// 			Head:       rule.Head,
	// 			Production: rule.Production,
	// 			Dot:        rule.Dot + 1,
	// 			Lookahead:  rule.Lookahead,
	// 		}
	// 		newState.Rules = append(newState.Rules, newInitialRule)
	//
	// 		if newInitialRule.Dot < len(newInitialRule.Production) {
	// 			closureToken := newInitialRule.Production[newInitialRule.Dot]
	// 			if closureToken.IsNonTerminal() {
	// 				set := lib.NewSet[GrammarToken]()
	// 				closure(closureToken, &newInitialRule, &newState, grammar, firsts, &set)
	// 			}
	// 		}
	// 	}
	//
	// 	idx := automata.FindIndexOfState(&newState)
	// 	newStateNotDefined := idx == ""
	// 	if newStateNotDefined {
	// 		idx = strconv.FormatInt(int64(len(automata.Nodes)), 10)
	// 		automata.Nodes[idx] = newState
	// 		queue.Enqueue(idx)
	// 	}
	//
	// 	if _, found := automata.Transitions[currentIdx]; !found {
	// 		automata.Transitions[currentIdx] = make(map[AlphabetInput]AutomataStateIndex)
	// 	}
	// 	if _, exists := automata.Transitions[currentIdx][transitionToken]; !exists {
	// 		automata.Transitions[currentIdx][transitionToken] = idx
	// 	}
	// }
}

func (auto *Automata) SimplifyStates() {

	simplifiedSomething := false

outerLoop:
	for i, state := range auto.Nodes {
		for j, other := range auto.Nodes {
			if i == j {
				continue
			}

			if state.EQ_WithoutLookAhead(&other) {
				simplifiedSomething = true
				rules := state.Items
				for i := range rules {
					other_i := other.Items[i]
					rules[i].Lookahead.Merge(&other_i.Lookahead)
				}
				newState := AutomataState{
					Items: rules,
				}
				newStateId := fmt.Sprintf("%s-%s", i, j)

				delete(auto.Nodes, i)
				delete(auto.Nodes, j)

				if i == auto.InitialState || j == auto.InitialState {
					auto.InitialState = newStateId
				}
				auto.Nodes[newStateId] = newState

				if _, found := auto.Transitions[newStateId]; !found {
					auto.Transitions[newStateId] = make(map[AlphabetInput]AutomataStateIndex)
				}

				for inputState := range auto.Transitions {
					matchedInputState := false
					for input, outState := range auto.Transitions[inputState] {
						if outState == i || outState == j {
							auto.Transitions[inputState][input] = newStateId
						}

						if inputState == i || inputState == j {
							auto.Transitions[newStateId][input] = outState
							matchedInputState = true
						}
					}

					if matchedInputState {
						delete(auto.Transitions, inputState)
					}
				}

				continue outerLoop
			}
		}
	}

	if simplifiedSomething {
		auto.SimplifyStates()
	}
}

func (auto *Automata) GenerateParsingTable(grammar *Grammar) ParsingTable {
	table := ParsingTable{
		ActionTable:   make(map[AFDNodeId]map[GrammarToken]Action),
		GoToTable:     make(map[AFDNodeId]map[GrammarToken]AFDNodeId),
		Original:      *grammar,
		InitialNodeId: auto.InitialState,
	}

	initialDefaultToken := NewNonTerminalToken("S'")
	for nodeId, state := range auto.Nodes {
		if _, found := table.ActionTable[nodeId]; !found {
			table.ActionTable[nodeId] = make(map[GrammarToken]Action)
		}
		if _, found := table.GoToTable[nodeId]; !found {
			table.GoToTable[nodeId] = make(map[GrammarToken]AFDNodeId)
		}

		for input, outNodeId := range auto.Transitions[nodeId] {
			if input.IsNonTerminal() {
				table.GoToTable[nodeId][input] = outNodeId
			} else {
				table.ActionTable[nodeId][input] = NewShiftAction(outNodeId)
			}
		}

		for _, rule := range state.Items {
			if len(rule.Lookahead) <= 0 {
				continue
			}

			if rule.Head.Equal(&initialDefaultToken) {
				continue
			}

			if rule.Dot < len(rule.Production) {
				continue
			}

			ruleId := grammar.FindIndexOfRule(&rule)
			if ruleId == -1 {
				panic(fmt.Sprintf("Failed to find rule: %#v\non grammar %s", rule, grammar))
			}
			for input := range rule.Lookahead {
				table.ActionTable[nodeId][input] = NewReduceAction(ruleId)
			}
		}
	}

	// Find Accept
	acceptNodeId := auto.Transitions[auto.InitialState][grammar.InitialSimbol]
	if _, found := table.ActionTable[acceptNodeId]; !found {
		table.ActionTable[acceptNodeId] = make(map[GrammarToken]Action)
	}
	table.ActionTable[acceptNodeId][NewEndToken()] = NewAcceptAction()

	return table
}

// func getTransitions(state AutomataState, grammar Grammar) map[string][]AutomataItem {
// 	transitions := make(map[string][]AutomataItem)
//
// 	for _, item := range state.Items {
// 		if item.DotPosition < len(item.Rule.Production) {
// 			symbol := item.Rule.Production[item.DotPosition]
// 			symbolStr := symbol.String()
//
// 			// Create new item with dot moved forward
// 			newItem := AutomataItem{
// 				Rule:        item.Rule,
// 				DotPosition: item.DotPosition + 1,
// 				Lookahead:   item.Lookahead,
// 			}
//
// 			transitions[symbolStr] = append(transitions[symbolStr], newItem)
// 		}
// 	}
//
// 	return transitions
// }
//
// func createNewState(items []AutomataItem, grammar Grammar) AutomataState {
// 	newState := AutomataState{
// 		Items:       make(map[int]AutomataItem),
// 		Productions: make(map[string]int),
// 		Initial:     false,
// 		Accept:      false,
// 	}
//
// 	// Add kernel items
// 	for i, item := range items {
// 		newState.Items[i] = item
// 	}
//
// 	// Compute closure
// 	closure(newState, grammar)
//
// 	// Check if this is an accept state
// 	initialToken := NewNonTerminalToken("S'")
// 	for _, item := range newState.Items {
// 		if item.Rule.Head.Equal(&initialToken) &&
// 			item.DotPosition == len(item.Rule.Production) {
// 			for _, lookahead := range item.Lookahead {
// 				if lookahead.IsEnd {
// 					newState.Accept = true
// 					break
// 				}
// 			}
// 		}
// 	}
//
// 	return newState
// }
//
// func findEquivalentState(afd *Automata, newState AutomataState) int {
// 	for stateId, existingState := range afd.Nodes {
// 		if equalState(existingState, newState) {
// 			return stateId
// 		}
// 	}
// 	return -1
// }
//
// func equalState(a, b AutomataState) bool {
// 	if len(a.Items) != len(b.Items) {
// 		return false
// 	}
//
// 	for _, itemA := range a.Items {
// 		found := false
// 		for _, itemB := range b.Items {
// 			if itemA.Rule.EqualRule(&itemB.Rule) &&
// 				itemA.DotPosition == itemB.DotPosition &&
// 				equalLookahead(itemA.Lookahead, itemB.Lookahead) {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			return false
// 		}
// 	}
//
// 	return true
// }
//
// func equalLookahead(a, b []GrammarToken) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}
// 	count := make(map[string]int)
// 	for _, tok := range a {
// 		count[tok.String()]++
// 	}
// 	for _, tok := range b {
// 		if count[tok.String()] == 0 {
// 			return false
// 		}
// 		count[tok.String()]--
// 	}
// 	return true
// }
//
// func (a *Automata) SimplifyStates() {
// 	coreMap := make(map[string][]int)
//
// 	for stateIdx, state := range a.Nodes {
// 		core := getCoreKey(state)
// 		coreMap[core] = append(coreMap[core], stateIdx)
// 	}
//
// 	newNodes := make(map[int]AutomataState)
// 	stateMapping := make(map[int]int)
//
// 	newIdx := 0
//
// 	for _, group := range coreMap {
// 		mergedItems := make(map[string]AutomataItem)
// 		initial := false
// 		accept := false
//
// 		for _, idx := range group {
// 			origState := a.Nodes[idx]
//
// 			if origState.Initial {
// 				initial = true
// 			}
// 			if origState.Accept {
// 				accept = true
// 			}
//
// 			for _, item := range a.Nodes[idx].Items {
// 				key := itemToKeyWithoutLookAhead(item)
// 				if existing, ok := mergedItems[key]; ok {
// 					existing.Lookahead = unionLookaheads(existing.Lookahead, item.Lookahead)
// 					mergedItems[key] = existing
// 				} else {
// 					mergedItems[key] = item
// 				}
// 			}
// 		}
//
// 		itemMap := make(map[int]AutomataItem)
// 		i := 0
// 		for _, v := range mergedItems {
// 			itemMap[i] = v
// 			i++
// 		}
//
// 		newNodes[newIdx] = AutomataState{
// 			Items:       itemMap,
// 			Productions: make(map[string]int),
// 			Initial:     initial,
// 			Accept:      accept,
// 		}
//
// 		for _, oldIdx := range group {
// 			stateMapping[oldIdx] = newIdx
// 		}
// 		newIdx++
// 	}
//
// 	for oldIdx, oldState := range a.Nodes {
// 		newIdx := stateMapping[oldIdx]
// 		for symbol, target := range oldState.Productions {
// 			newTarget := stateMapping[target]
// 			newNodes[newIdx].Productions[symbol] = newTarget
// 		}
// 	}
//
// 	for stateIdx, st := range newNodes {
// 		for itemIdx, item := range st.Items {
// 			if item.DotPosition >= len(item.Rule.Production) {
// 				item.Lookahead = append(item.Lookahead, NewEndToken())
// 				st.Items[itemIdx] = item
// 			}
// 		}
//
// 		newNodes[stateIdx] = st
// 	}
//
// 	a.Nodes = newNodes
// }
//
// func getCoreKey(state AutomataState) string {
// 	coreItems := make([]string, 0)
// 	for _, item := range state.Items {
// 		coreItems = append(coreItems, itemToKeyWithoutLookAhead(item))
// 	}
// 	sort.Strings(coreItems)
// 	return fmt.Sprintf("%v", coreItems)
// }
//
// func (lalr *Automata) GenerateParsingTable(grammar *Grammar) ParsingTable {
// 	table := ParsingTable{
// 		ActionTable:   make(map[AFDNodeId]map[GrammarToken]Action),
// 		GoToTable:     make(map[AFDNodeId]map[GrammarToken]AFDNodeId),
// 		Original:      *grammar,
// 		InitialNodeId: lalr.findInitialState(),
// 	}
//
// 	conflicts := make(map[string][]string)
//
// 	for stateID, state := range lalr.Nodes {
// 		stateKey := fmt.Sprintf("%d", stateID)
//
// 		if _, exists := table.ActionTable[stateKey]; !exists {
// 			table.ActionTable[stateKey] = make(map[GrammarToken]Action)
// 		}
//
// 		if _, exists := table.GoToTable[stateKey]; !exists {
// 			table.GoToTable[stateKey] = make(map[GrammarToken]AFDNodeId)
// 		}
//
// 		lalr.processReduceActions(state, stateKey, grammar, &table, conflicts)
//
// 		lalr.processShiftAndGotoActions(state, stateKey, grammar, &table, conflicts)
// 	}
//
// 	if len(conflicts) > 0 {
// 		fmt.Println("Parsing conflixts detected:")
// 		for state, conflictList := range conflicts {
// 			fmt.Printf("State %s:\n", state)
// 			for _, conflict := range conflictList {
// 				fmt.Printf("	%s\n", conflict)
// 			}
// 		}
// 	}
//
// 	return table
// }
//
// func (lalr *Automata) processReduceActions(state AutomataState, stateKey string, grammar *Grammar, table *ParsingTable, conflicts map[string][]string) {
// 	for _, item := range state.Items {
// 		if item.DotPosition == len(item.Rule.Production) {
// 			if lalr.isAcceptItem(item, grammar) {
// 				for _, lookahead := range item.Lookahead {
// 					if lookahead.IsEnd {
// 						lalr.setAction(table, stateKey, lookahead, NewAcceptAction(), conflicts)
// 					}
// 				}
// 			} else {
// 				ruleIndex := lalr.findRuleIndex(item.Rule, grammar)
// 				if ruleIndex != -1 {
// 					for _, lookahead := range item.Lookahead {
// 						if !lookahead.IsEnd || !lalr.isAcceptItem(item, grammar) {
// 							lalr.setAction(table, stateKey, lookahead, NewReduceAction(ruleIndex), conflicts)
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }
//
// func (lalr *Automata) processShiftAndGotoActions(state AutomataState, stateKey string, grammar *Grammar, table *ParsingTable, conflicts map[string][]string) {
// 	for symbol, targetStateID := range state.Productions {
// 		targetKey := fmt.Sprintf("%d", targetStateID)
// 		symbolToken := grammar.GetTokenByString(symbol)
//
// 		if symbolToken.IsTerminal() && !symbolToken.IsEnd {
// 			lalr.setAction(table, stateKey, symbolToken, NewShiftAction(targetKey), conflicts)
// 		} else if symbolToken.IsNonTerminal() {
// 			table.GoToTable[stateKey][symbolToken] = targetKey
// 		}
// 	}
// }
//
// func (lalr *Automata) isAcceptItem(item AutomataItem, grammar *Grammar) bool {
// 	initialToken := NewNonTerminalToken("S'")
// 	if item.Rule.Head.Equal(&initialToken) || (len(item.Rule.Production) == 1 && item.Rule.Production[0].Equal(&grammar.InitialSimbol)) {
// 		return true
// 	}
// 	return false
// }
//
// func (lalr *Automata) findRuleIndex(rule GrammarRule, grammar *Grammar) int {
// 	for i, grammarRule := range grammar.Rules {
// 		if rule.EqualRule(&grammarRule) {
// 			return i
// 		}
// 	}
//
// 	return -1
// }
//
// func (lalr *Automata) setAction(table *ParsingTable, stateKey string, symbol GrammarToken, newAction Action, conflicts map[string][]string) {
// 	if existingAction, exists := table.ActionTable[stateKey][symbol]; exists {
// 		conflictType := lalr.resolveConflict(existingAction, newAction)
// 		if conflictType != "" {
// 			conflictMsg := fmt.Sprintf("%s conflict on symbol %s: existing=%s, new=%s",
// 				conflictType, symbol.String(), actionToString(existingAction), actionToString(newAction))
// 			conflicts[stateKey] = append(conflicts[stateKey], conflictMsg)
//
// 			resolvedAction := lalr.applyConflictResolution(existingAction, newAction)
// 			table.ActionTable[stateKey][symbol] = resolvedAction
// 		}
// 	} else {
// 		table.ActionTable[stateKey][symbol] = newAction
// 	}
// }
//
// func (lalr *Automata) resolveConflict(existing, new Action) string {
// 	if existing.Shift.HasValue() && new.Shift.HasValue() {
// 		return "shift-shift"
// 	}
// 	if existing.Reduce.HasValue() && new.Reduce.HasValue() {
// 		return "reduce-reduce"
// 	}
// 	if (existing.Shift.HasValue() && new.Reduce.HasValue()) ||
// 		(existing.Reduce.HasValue() && new.Shift.HasValue()) {
// 		return "shift-reduce"
// 	}
// 	return ""
// }
//
// func (lalr *Automata) applyConflictResolution(existing, new Action) Action {
// 	if existing.Shift.HasValue() && new.Reduce.HasValue() {
// 		return existing
// 	}
// 	if existing.Reduce.HasValue() && new.Shift.HasValue() {
// 		return new
// 	}
//
// 	if existing.Reduce.HasValue() && new.Reduce.HasValue() {
// 		if existing.Reduce.GetValue() < new.Reduce.GetValue() {
// 			return existing
// 		}
// 		return new
// 	}
//
// 	return existing
// }
//
// func actionToString(action Action) string {
// 	if action.Accept {
// 		return "accept"
// 	}
// 	if action.Shift.HasValue() {
// 		return fmt.Sprintf("shift(%s)", action.Shift.GetValue())
// 	}
// 	if action.Reduce.HasValue() {
// 		return fmt.Sprintf("reduce(%d)", action.Reduce.GetValue())
// 	}
// 	return "unknown"
// }
//
// func (lalr *Automata) findInitialState() string {
// 	for key, state := range lalr.Nodes {
// 		if state.Initial {
// 			keyStr := strconv.Itoa(key)
// 			return keyStr
// 		}
// 	}
//
// 	return "0"
// }
