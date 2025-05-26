package grammar

import (
	"fmt"

	"github.com/Jose-Prince/UWUCompiler/lib"
	parsertypes "github.com/Jose-Prince/UWUCompiler/parserTypes"
)

type AFDNodeId = string

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

func NewShiftAction(id AFDNodeId) Action {
	return Action{
		Shift:  lib.CreateValue(id),
		Reduce: lib.CreateNull[int](),
		Accept: false,
	}
}

func NewReduceAction(idx int) Action {
	return Action{
		Shift:  lib.CreateNull[AFDNodeId](),
		Reduce: lib.CreateValue(idx),
		Accept: false,
	}
}

func NewAcceptAction() Action {
	return Action{
		Shift:  lib.CreateNull[AFDNodeId](),
		Reduce: lib.CreateNull[int](),
		Accept: true,
	}
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

func (g *Grammar) TokenToParserType(token *GrammarToken) parsertypes.GrammarToken {
	id, found := g.TokenIds[*token]
	if !found {
		panic(fmt.Sprintf("Token %s was not found on grammar ids from %v", *token, *g))
	}

	return id
}

func convertGrammar(g *Grammar) parsertypes.Grammar {
	grammar := parsertypes.Grammar{}
	grammar.InitialSimbol = g.TokenToParserType(&g.InitialSimbol)

	grammar.Terminals = parsertypes.NewSet[parsertypes.GrammarToken]()
	for term := range g.Terminals {
		grammar.Terminals.Add(g.TokenToParserType(&term))
	}

	grammar.NonTerminals = parsertypes.NewSet[parsertypes.GrammarToken]()
	for term := range g.Terminals {
		grammar.NonTerminals.Add(g.TokenToParserType(&term))
	}

	grammar.Rules = make([]parsertypes.GrammarRule, 0, len(g.Rules))
	for _, rule := range g.Rules {
		parserRule := parsertypes.GrammarRule{
			Head:       g.TokenToParserType(&rule.Head),
			Production: make([]parsertypes.GrammarToken, 0, len(rule.Production)),
		}
		for _, productionT := range rule.Production {
			parserRule.Production = append(parserRule.Production, g.TokenToParserType(&productionT))
		}

		grammar.Rules = append(grammar.Rules, parserRule)
	}
	return grammar
}

func (a Action) ToParserType() parsertypes.Action {
	action := parsertypes.Action{
		Shift:  a.Shift,
		Reduce: a.Reduce,
		Accept: a.Accept,
	}

	return action
}

func (s *ParsingTable) ToParserTable() parsertypes.ParsingTable {
	table := parsertypes.ParsingTable{
		InitialNodeId: s.InitialNodeId,
		Original:      convertGrammar(&s.Original),
		ActionTable:   make(map[parsertypes.AFDNodeId]map[parsertypes.GrammarToken]parsertypes.Action),
		GoToTable:     make(map[parsertypes.AFDNodeId]map[parsertypes.GrammarToken]parsertypes.AFDNodeId),
	}

	for nodeId, row := range s.ActionTable {
		for token, action := range row {
			if _, found := table.ActionTable[nodeId]; !found {
				table.ActionTable[nodeId] = make(map[parsertypes.GrammarToken]parsertypes.Action)
			}

			transformedTk := s.Original.TokenToParserType(&token)
			transformedAction := action.ToParserType()
			table.ActionTable[nodeId][transformedTk] = transformedAction
		}
	}

	for nodeId, row := range s.GoToTable {
		for token, newNodeId := range row {
			if _, found := table.GoToTable[nodeId]; !found {
				table.GoToTable[nodeId] = make(map[parsertypes.GrammarToken]parsertypes.AFDNodeId)
			}

			transformedTk := s.Original.TokenToParserType(&token)
			table.GoToTable[nodeId][transformedTk] = newNodeId
		}
	}

	return table
}
