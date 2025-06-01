package grammar

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/Jose-Prince/UWUCompiler/lib"
)

type Automata struct {
	Nodes map[int]AutomataState
}

type AutomataState struct {
	Items       map[int]AutomataItem
	Productions map[string]int
	Initial     bool
	Accept      bool
}

type AutomataItem struct {
	Rule        GrammarRule
	DotPosition int
	Lookahead   []GrammarToken
}

func InitializeAutomata(initialRule GrammarRule, grammar Grammar) Automata {
	lr1 := Automata{
		Nodes: make(map[int]AutomataState),
	}
	initialItem := AutomataItem{
		Rule:        initialRule,
		DotPosition: 0,
		Lookahead:   []GrammarToken{NewEndToken()},
	}

	state := AutomataState{
		Items:       make(map[int]AutomataItem),
		Productions: make(map[string]int),
		Initial:     true,
		Accept:      false,
	}

	state.Items[0] = initialItem

	closure(state, grammar)

	lr1.Nodes[0] = state

	generateStates(&lr1, grammar)

	return lr1
}

func closure(state AutomataState, grammar Grammar) {
	workList := make([]int, 0)

	for i := range state.Items {
		workList = append(workList, i)
	}

	itemIndex := make(map[string]int)
	for i, item := range state.Items {
		key := itemToKey(item)
		itemIndex[key] = i
	}

	for len(workList) > 0 {
		currentIdx := workList[0]
		workList = workList[1:]

		item := state.Items[currentIdx]

		if item.DotPosition >= len(item.Rule.Production) {
			continue
		}

		currentSymbol := item.Rule.Production[item.DotPosition]
		if !currentSymbol.IsNonTerminal() {
			continue
		}

		beta := make([]GrammarToken, 0)
		if item.DotPosition+1 < len(item.Rule.Production) {
			beta = item.Rule.Production[item.DotPosition+1:]
		}

		betaAlpha := append(beta, item.Lookahead...)
		firstSet := first(betaAlpha, grammar)

		for _, rule := range grammar.Rules {
			if !rule.Head.Equal(&currentSymbol) {
				continue
			}

			newItem := AutomataItem{
				Rule:        rule,
				DotPosition: 0,
				Lookahead:   firstSet,
			}

			key := itemToKeyWithoutLookAhead(newItem)

			if existingIdx, exists := itemIndex[key]; exists {
				existing := state.Items[existingIdx]
				originalSize := len(existing.Lookahead)
				existing.Lookahead = unionLookaheads(existing.Lookahead, firstSet)

				if len(existing.Lookahead) != originalSize {
					state.Items[existingIdx] = existing

					found := false
					for _, idx := range workList {
						if idx == existingIdx {
							found = true
							break
						}
					}
					if !found {
						workList = append(workList, existingIdx)
					}
				}
			} else {
				newIdx := len(state.Items)
				state.Items[newIdx] = newItem
				itemIndex[key] = newIdx
				workList = append(workList, newIdx)
			}
		}
	}
}

func first(sequence []GrammarToken, grammar Grammar) []GrammarToken {
	table := NewFirstFollowTable()
	GetFirsts(&grammar, &table)

	if len(sequence) == 0 {
		return []GrammarToken{}
	}
	firstToken := sequence[0]
	if firstToken.IsTerminal() {
		return []GrammarToken{firstToken}
	}

	return table.table[firstToken].First.ToSlice_()
}

func itemToKeyWithoutLookAhead(item AutomataItem) string {
	return item.Rule.ToString() + "|" + fmt.Sprintf("%d", item.DotPosition)
}

func itemToKey(item AutomataItem) string {
	lookStrs := make([]string, len(item.Lookahead))
	for i, l := range item.Lookahead {
		lookStrs[i] = l.String()
	}
	sort.Strings(lookStrs) // Ordenar alfabéticamente

	look := ""
	for _, l := range lookStrs {
		look += l + ","
	}
	return item.Rule.ToString() + "|" + string(rune(item.DotPosition)) + "|" + look
}

func unionLookaheads(a, b []GrammarToken) []GrammarToken {
	seen := make(map[string]bool)
	union := make([]GrammarToken, 0)

	for _, tok := range a {
		key := tok.String()
		if !seen[key] {
			seen[key] = true
			union = append(union, tok)
		}
	}

	for _, tok := range b {
		key := tok.String()
		if !seen[key] {
			seen[key] = true
			union = append(union, tok)
		}
	}

	return union
}

func generateStates(afd *Automata, grammar Grammar) {
	visited := lib.NewSet[int]()
	changed := true

	initialToken := NewNonTerminalToken("S'")

	for changed {
		changed = false

		terminals := grammar.Terminals.ToSlice()
		nonTerminals := grammar.NonTerminals.ToSlice()

		terms := append(nonTerminals, terminals...)

		for k := range afd.Nodes {
			state := afd.Nodes[k]

			for _, nt := range terms {
				for _, val := range state.Items {
					if val.DotPosition >= len(val.Rule.Production) {
						continue
					}

					if val.Rule.Production[val.DotPosition].Equal(&nt) {
						newState := AutomataState{
							Items:       make(map[int]AutomataItem),
							Productions: make(map[string]int),
							Initial:     false,
							Accept:      false,
						}

						newItem := val
						newItem.DotPosition = val.DotPosition + 1
						newState.Items[0] = newItem

						closure(newState, grammar)

						for _, item := range newState.Items {
							if item.Rule.Head.Equal(&initialToken) && item.DotPosition == len(item.Rule.Production) && item.Lookahead[0].IsEnd {
								newState.Accept = true
								break
							}
						}

						for _, s := range afd.Nodes {
							if !equalState(s, newState) {
								state.Productions[nt.String()] = len(afd.Nodes)
								afd.Nodes[len(afd.Nodes)] = newState
								visited.Add(k)
								break
							}
						}

						break

					}
				}
			}
		}
	}
}

func equalState(a, b AutomataState) bool {
	if len(a.Items) != len(b.Items) {
		return false
	}

	for _, itemA := range a.Items {
		found := false
		for _, itemB := range b.Items {
			if itemA.Rule.EqualRule(&itemB.Rule) &&
				itemA.DotPosition == itemB.DotPosition &&
				equalLookahead(itemA.Lookahead, itemB.Lookahead) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func equalLookahead(a, b []GrammarToken) bool {
	if len(a) != len(b) {
		return false
	}
	count := make(map[string]int)
	for _, tok := range a {
		count[tok.String()]++
	}
	for _, tok := range b {
		if count[tok.String()] == 0 {
			return false
		}
		count[tok.String()]--
	}
	return true
}

func (a *Automata) SimplifyStates() {
	coreMap := make(map[string][]int)

	for stateIdx, state := range a.Nodes {
		core := getCoreKey(state)
		coreMap[core] = append(coreMap[core], stateIdx)
	}

	newNodes := make(map[int]AutomataState)
	stateMapping := make(map[int]int)

	newIdx := 0

	for _, group := range coreMap {
		mergedItems := make(map[string]AutomataItem)
		initial := false
		accept := false

		for _, idx := range group {
			origState := a.Nodes[idx]

			if origState.Initial {
				initial = true
			}
			if origState.Accept {
				accept = true
			}

			for _, item := range a.Nodes[idx].Items {
				key := itemToKeyWithoutLookAhead(item)
				if existing, ok := mergedItems[key]; ok {
					existing.Lookahead = unionLookaheads(existing.Lookahead, item.Lookahead)
					mergedItems[key] = existing
				} else {
					mergedItems[key] = item
				}
			}
		}

		itemMap := make(map[int]AutomataItem)
		i := 0
		for _, v := range mergedItems {
			itemMap[i] = v
			i++
		}

		newNodes[newIdx] = AutomataState{
			Items:       itemMap,
			Productions: make(map[string]int),
			Initial:     initial,
			Accept:      accept,
		}

		for _, oldIdx := range group {
			stateMapping[oldIdx] = newIdx
		}
		newIdx++
	}

	for oldIdx, oldState := range a.Nodes {
		newIdx := stateMapping[oldIdx]
		for symbol, target := range oldState.Productions {
			newTarget := stateMapping[target]
			newNodes[newIdx].Productions[symbol] = newTarget
		}
	}

	for stateIdx, st := range newNodes {
		for itemIdx, item := range st.Items {
			if item.DotPosition >= len(item.Rule.Production) {
				item.Lookahead = append(item.Lookahead, NewEndToken())
				st.Items[itemIdx] = item
			}
		}

		newNodes[stateIdx] = st
	}

	a.Nodes = newNodes
}

func getCoreKey(state AutomataState) string {
	coreItems := make([]string, 0)
	for _, item := range state.Items {
		coreItems = append(coreItems, itemToKeyWithoutLookAhead(item))
	}
	sort.Strings(coreItems)
	return fmt.Sprintf("%v", coreItems)
}

func (lalr *Automata) GenerateParsingTable(grammar *Grammar) ParsingTable {
	table := ParsingTable{
		ActionTable:   make(map[AFDNodeId]map[GrammarToken]Action),
		GoToTable:     make(map[AFDNodeId]map[GrammarToken]AFDNodeId),
		Original:      *grammar,
		InitialNodeId: lalr.findInitialState(),
	}

	for stateID, state := range lalr.Nodes {
		stateKey := fmt.Sprintf("%d", stateID)

		if _, exists := table.ActionTable[stateKey]; !exists {
			table.ActionTable[stateKey] = make(map[GrammarToken]Action)
		}

		if _, exists := table.GoToTable[stateKey]; !exists {
			table.GoToTable[stateKey] = make(map[GrammarToken]AFDNodeId)
		}

		for _, item := range state.Items {
			if item.DotPosition == len(item.Rule.Production) {
				for _, lookahead := range item.Lookahead {
					if item.Rule.Head.Equal(&grammar.InitialSimbol) && lookahead.IsEnd {

						table.ActionTable[stateKey][item.Lookahead[0]] = Action{
							Shift:  lib.CreateNull[AFDNodeId](),
							Reduce: lib.CreateNull[int](),
							Accept: true,
						}

					} else {
						ruleIndex := -1
						for i, rule := range grammar.Rules {
							if rule.EqualRule(&item.Rule) {
								ruleIndex = i
								break
							}
						}
						if ruleIndex != -1 {
							table.ActionTable[stateKey][lookahead] = NewReduceAction(ruleIndex)
						}
					}
				}
			}
		}

		for symbol, targetStateID := range state.Productions {
			targetKey := fmt.Sprintf("%d", targetStateID)

			symbolToken := grammar.GetTokenByString(symbol)

			if symbolToken.IsTerminal() {
				table.ActionTable[stateKey][symbolToken] = NewShiftAction(targetKey)
			} else {
				table.GoToTable[stateKey][symbolToken] = targetKey
			}
		}
	}

	return table
}

func (lalr *Automata) findInitialState() string {
	for key, state := range lalr.Nodes {
		if state.Initial {
			keyStr := strconv.Itoa(key)
			return keyStr
		}
	}

	return "0"
}
