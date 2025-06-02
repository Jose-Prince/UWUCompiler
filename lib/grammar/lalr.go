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

	conflicts := make(map[string][]string)

	for stateID, state := range lalr.Nodes {
		stateKey := fmt.Sprintf("%d", stateID)

		if _, exists := table.ActionTable[stateKey]; !exists {
			table.ActionTable[stateKey] = make(map[GrammarToken]Action)
		}

		if _, exists := table.GoToTable[stateKey]; !exists {
			table.GoToTable[stateKey] = make(map[GrammarToken]AFDNodeId)
		}

		lalr.processReduceActions(state, stateKey, grammar, &table, conflicts)

		lalr.processShiftAndGotoActions(state, stateKey, grammar, &table, conflicts)
	}

	if len(conflicts) > 0 {
		fmt.Println("Parsing conflixts detected:")
		for state, conflictList := range conflicts {
			fmt.Printf("State %s:\n", state)
			for _, conflict := range conflictList {
				fmt.Printf("	%s\n", conflict)
			}
		}
	}

	return table
}

func (lalr *Automata) processReduceActions(state AutomataState, stateKey string, grammar *Grammar, table *ParsingTable, conflicts map[string][]string) {
	for _, item := range state.Items {
		if item.DotPosition == len(item.Rule.Production) {
			if lalr.isAcceptItem(item, grammar) {
				for _, lookahead := range item.Lookahead {
					if lookahead.IsEnd {
						lalr.setAction(table, stateKey, lookahead, NewAcceptAction(), conflicts)
					}
				}
			} else {
				ruleIndex := lalr.findRuleIndex(item.Rule, grammar)
				if ruleIndex != -1 {
					for _, lookahead := range item.Lookahead {
						if !lookahead.IsEnd || !lalr.isAcceptItem(item, grammar) {
							lalr.setAction(table, stateKey, lookahead, NewReduceAction(ruleIndex), conflicts)
						}
					}
				}
			}
		}
	}
}

func (lalr *Automata) processShiftAndGotoActions(state AutomataState, stateKey string, grammar *Grammar, table *ParsingTable, conflicts map[string][]string) {
	for symbol, targetStateID := range state.Productions {
		targetKey := fmt.Sprintf("%d", targetStateID)
		symbolToken := grammar.GetTokenByString(symbol)

		if symbolToken.IsTerminal() && !symbolToken.IsEnd {
			lalr.setAction(table, stateKey, symbolToken, NewShiftAction(targetKey), conflicts)
		} else if symbolToken.IsNonTerminal() {
			table.GoToTable[stateKey][symbolToken] = targetKey
		}
	}
}

func (lalr *Automata) isAcceptItem(item AutomataItem, grammar *Grammar) bool {
	initialToken := NewNonTerminalToken("S'")
	if item.Rule.Head.Equal(&initialToken) || (len(item.Rule.Production) == 1 && item.Rule.Production[0].Equal(&grammar.InitialSimbol)) {
		return true
	}
	return false
}

func (lalr *Automata) findRuleIndex(rule GrammarRule, grammar *Grammar) int {
	for i, grammarRule := range grammar.Rules {
		if rule.EqualRule(&grammarRule) {
			return i
		}
	}

	return -1
}

func (lalr *Automata) setAction(table *ParsingTable, stateKey string, symbol GrammarToken, newAction Action, conflicts map[string][]string) {
	if existingAction, exists := table.ActionTable[stateKey][symbol]; exists {
		conflictType := lalr.resolveConflict(existingAction, newAction, symbol)
		if conflictType != "" {
			conflictMsg := fmt.Sprintf("%s conflict on symbol %s: existing=%s, new=%s",
				conflictType, symbol.String(), actionToString(existingAction), actionToString(newAction))
			conflicts[stateKey] = append(conflicts[stateKey], conflictMsg)

			resolvedAction := lalr.applyConflictResolution(existingAction, newAction, symbol)
			table.ActionTable[stateKey][symbol] = resolvedAction
		}
	} else {
		table.ActionTable[stateKey][symbol] = newAction
	}
}

func (lalr *Automata) resolveConflict(existing, new Action, symbol GrammarToken) string {
	if existing.Shift.HasValue() && new.Shift.HasValue() {
		return "shift-shift"
	}
	if existing.Reduce.HasValue() && new.Reduce.HasValue() {
		return "reduce-reduce"
	}
	if (existing.Shift.HasValue() && new.Reduce.HasValue()) ||
		(existing.Reduce.HasValue() && new.Shift.HasValue()) {
		return "shift-reduce"
	}
	return ""
}

func (lalr *Automata) applyConflictResolution(existing, new Action, symbol GrammarToken) Action {
	if existing.Shift.HasValue() && new.Reduce.HasValue() {
		return existing
	}
	if existing.Reduce.HasValue() && new.Shift.HasValue() {
		return new
	}

	if existing.Reduce.HasValue() && new.Reduce.HasValue() {
		if existing.Reduce.GetValue() < new.Reduce.GetValue() {
			return existing
		}
		return new
	}

	return existing
}

func actionToString(action Action) string {
	if action.Accept {
		return "accept"
	}
	if action.Shift.HasValue() {
		return fmt.Sprintf("shift(%s)", action.Shift.GetValue())
	}
	if action.Reduce.HasValue() {
		return fmt.Sprintf("reduce(%d)", action.Reduce.GetValue())
	}
	return "unknown"
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
