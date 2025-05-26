package grammar

import "github.com/Jose-Prince/UWUCompiler/lib"

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

type ParsingTable struct {
	// The Action table contains all the reduce and shifts of the parsing table.
	ActionTable map[AFDNodeId]map[GrammarToken]Action
	// The GoTo table contains all the nonterminal tokens and what transitions to make of them.
	GoToTable map[AFDNodeId]map[GrammarToken]AFDNodeId
	// The original grammar, IT MUST NOT BE EXPANDED!
	Original Grammar
}
