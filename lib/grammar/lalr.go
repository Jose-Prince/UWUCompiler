package grammar

import "github.com/Jose-Prince/UWULexer/lib"

type automata struct {
	nodes map[int]automataState
}

type automataState struct {
	States      map[int]automataItem
	Productions map[GrammarToken]int
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
		States:      make(map[int]automataItem),
		Productions: make(map[GrammarToken]int),
	}

	state.States[0] = initialItem

	closure(state, grammar)

	lr1.nodes[0] = state

	return lr1
}

func closure(state automataState, grammar Grammar) {
	items := lib.NewSet[int]()
	visited := lib.NewSet[int]()

	items.Add(0)

	itemIndex := make(map[string]int)

	for i, item := range state.States {
		itemIndex[itemToKey(item)] = i
	}

	changed := true
	for changed {
		changed = false

		for key := range items {
			if visited.Contains(key) {
				continue
			}
			item := state.States[key]
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

						keyStr := itemToKey(newItem)

						if idx, exists := itemIndex[keyStr]; exists {
							existing := state.States[idx]
							existing.Lookahead = unionLookaheads(existing.Lookahead, lookaheads)
							state.States[idx] = existing
						} else {
							newID := len(state.States)
							state.States[newID] = newItem
							items.Add(newID)
							itemIndex[keyStr] = newID
							changed = true

						}
					}
				}

				visited.Add(key)
			}
		}

	}
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
