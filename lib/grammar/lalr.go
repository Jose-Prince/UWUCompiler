package grammar

import (
	"fmt"
	"sort"

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

	generateStates(&lr1, grammar)

	return lr1
}

func closure(state automataState, grammar Grammar) {
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

			newItem := automataItem{
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

func itemToKeyWithoutLookAhead(item automataItem) string {
	return item.Rule.ToString() + "|" + fmt.Sprintf("%d", item.DotPosition)
}

func itemToKey(item automataItem) string {
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

func generateStates(afd *automata, grammar Grammar) {
	visited := lib.NewSet[int]()
	changed := true

	for changed {
		changed = false

		terminals := grammar.Terminals.ToSlice()
		nonTerminals := grammar.NonTerminals.ToSlice()

		terms := append(nonTerminals, terminals...)

		for k := range afd.nodes {
			state := afd.nodes[k]

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
