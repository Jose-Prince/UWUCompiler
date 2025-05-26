package grammar

import (
	"github.com/Jose-Prince/UWUCompiler/lib"
)

type automata struct {
	nodes map[int]automataState
}

type automataState struct {
	Items       map[int]automataItem
	Productions map[string]int
}

type automataItem struct {
	Rule        GrammarRule
	DotPosition int
	Lookahead   []GrammarToken
}

func InitializeAutomata(initialRule GrammarRule, grammar Grammar) automata {
	lr1 := automata{
		nodes: make(map[int]automataState),
	}
	initialItem := automataItem{
		Rule:        initialRule,
		DotPosition: 0,
		Lookahead:   []GrammarToken{NewEndToken()},
	}

	state := automataState{
		Items:       make(map[int]automataItem),
		Productions: make(map[string]int),
	}

	state.Items[0] = initialItem

	closure(state, grammar)

	lr1.nodes[0] = state

	generateStates(0, &lr1, grammar)

	return lr1
}

func closure(state automataState, grammar Grammar) {
	items := lib.NewSet[int]()
	visited := lib.NewSet[int]()

	items.Add(0)

	itemIndex := make(map[string]int)

	for i, item := range state.Items {
		itemIndex[itemToKey(item)] = i
	}

	changed := true
	for changed {
		changed = false

		for key := range items {
			if visited.Contains(key) {
				continue
			}
			item := state.Items[key]
			if item.DotPosition < len(item.Rule.Production) && item.Rule.Production[item.DotPosition].IsNonTerminal() {
				B := item.Rule.Production[item.DotPosition]

				beta := item.Rule.Production[item.DotPosition+1:]

				lookaheadSeq := append(beta, item.Lookahead...)
				lookaheads := first(lookaheadSeq, grammar)

				for _, rule := range grammar.Rules {
					if rule.Head.Equal(&B) {
						newItem := automataItem{
							Rule:        rule,
							DotPosition: 0,
							Lookahead:   lookaheads,
						}
						var equals bool
						for _, i := range state.Items {
							if itemsEqual(i, newItem) {
								continue
							} else {
								keyStr := itemToKey(newItem)

								if idx, exists := itemIndex[keyStr]; exists && equals {
									existing := state.Items[idx]
									existing.Lookahead = unionLookaheads(existing.Lookahead, lookaheads)
									state.Items[idx] = existing
								} else {
									newID := len(state.Items)
									state.Items[newID] = newItem
									items.Add(newID)
									itemIndex[keyStr] = newID
									changed = true
								}
							}
						}

					}
				}

				visited.Add(key)
			}
		}

	}
}

func itemsEqual(a, b automataItem) bool {

	if a.Rule.Head.String() != b.Rule.Head.String() {
		return false
	}

	if len(a.Rule.Production) != len(b.Rule.Production) {
		return false
	}

	for i := range a.Rule.Production {
		if !a.Rule.Production[i].Equal(&b.Rule.Production[i]) {
			return false
		}
	}

	if len(a.Lookahead) != len(b.Lookahead) {
		return false
	}

	// Comparar conjuntos de lookahead
	setA := make(map[string]bool)
	for _, tok := range a.Lookahead {
		setA[tok.String()] = true
	}

	for _, tok := range b.Lookahead {
		if !setA[tok.String()] {
			return false
		}
	}

	return true
}

func first(sequence []GrammarToken, grammar Grammar) []GrammarToken {
	if len(sequence) == 0 {
		return []GrammarToken{}
	}
	firstToken := sequence[0]
	if firstToken.IsTerminal() {
		return []GrammarToken{firstToken}
	}

	return grammar.First(firstToken)
}

func itemToKey(item automataItem) string {
	look := ""
	for _, l := range item.Lookahead {
		look += l.String() + ","
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

func generateStates(stateID int, afd *automata, grammar Grammar) {
	state := afd.nodes[stateID]

	terminals := grammar.Terminals.ToSlice()
	nonTerminals := grammar.NonTerminals.ToSlice()

	terms := append(nonTerminals, terminals...)

	for _, nt := range terms {
		for _, val := range state.Items {
			if val.DotPosition >= len(val.Rule.Production) {
				continue
			}

			if val.Rule.Production[val.DotPosition].Equal(&nt) {
				newState := automataState{
					Items:       make(map[int]automataItem),
					Productions: make(map[string]int),
				}

				newItem := val
				newItem.DotPosition = val.DotPosition + 1
				newState.Items[0] = newItem

				closure(newState, grammar)

				for _, s := range afd.nodes {
					if !equalState(s, newState) {
						state.Productions[nt.String()] = len(afd.nodes)
						afd.nodes[len(afd.nodes)] = newState
						break
					}
				}

				break

			}
		}
	}

	// symbolToItems := make(map[string][]automataItem)

	// for _, item := range state.Items {
	// 	if item.DotPosition < len(item.Rule.Production) {
	// 		nextSymbol := item.Rule.Production[item.DotPosition]
	// 		key := nextSymbol.String()
	// 		symbolToItems[key] = append(symbolToItems[key], item)
	// 	}
	// }

	// for symbolStr, items := range symbolToItems {
	// 	newState := automataState{
	// 		Items:       make(map[int]automataItem),
	// 		Productions: make(map[string]int),
	// 	}

	// 	for _, item := range items {
	// 		newItem := automataItem{
	// 			Rule:        item.Rule,
	// 			DotPosition: item.DotPosition + 1,
	// 			Lookahead:   item.Lookahead,
	// 		}
	// 		newState.Items[len(newState.Items)] = newItem
	// 	}

	// 	closure(newState, grammar)

	// 	existingID := -1
	// 	for id, s := range afd.nodes {
	// 		if equalState(s, newState) {
	// 			existingID = id
	// 			break
	// 		}
	// 	}

	// 	var symbol GrammarToken
	// 	nonTerminals := grammar.NonTerminals.ToSlice()
	// 	terminals := grammar.Terminals.ToSlice()

	// 	for _, t := range append(nonTerminals, terminals...) {
	// 		if t.String() == symbolStr {
	// 			symbol = t
	// 			break
	// 		}
	// 	}

	// 	if existingID == -1 {
	// 		newID := len(afd.nodes)
	// 		afd.nodes[newID] = newState
	// 		afd.nodes[stateID].Productions[symbol.String()] = existingID
	// 	}
	// }
}

func equalState(a, b automataState) bool {
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
