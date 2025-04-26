package lib

import (
	"testing"
)

func createExampleGrammar() Grammar {
	return Grammar{
		InitialSimbol: NewNonTerminalToken("S"),
		NonTerminals: Set[GrammarToken]{
			NewNonTerminalToken("S"): struct{}{},
			NewNonTerminalToken("P"): struct{}{},
			NewNonTerminalToken("Q"): struct{}{},
		},
		Terminals: Set[GrammarToken]{
			NewTerminalToken("∨"):        struct{}{},
			NewTerminalToken("∧"):        struct{}{},
			NewTerminalToken("["):        struct{}{},
            NewTerminalToken("]"):        struct{}{},
			NewTerminalToken("sentence"): struct{}{},
		},
		Rules: []GrammarRule{
			GrammarRule{
				Head: NewNonTerminalToken("S"),
				Production: []GrammarToken{
					NewNonTerminalToken("S"),
					NewTerminalToken("∧"),
					NewNonTerminalToken("P"),
				},
			},

			GrammarRule{
				Head: NewNonTerminalToken("S"),
				Production: []GrammarToken{
					NewNonTerminalToken("P"),
				},
			},

			GrammarRule{
				Head: NewNonTerminalToken("P"),
				Production: []GrammarToken{
					NewNonTerminalToken("S"),
					NewTerminalToken("∨"),
					NewNonTerminalToken("P"),
				},
			},

			GrammarRule{
				Head: NewNonTerminalToken("P"),
				Production: []GrammarToken{
					NewNonTerminalToken("Q"),
				},
			},

			GrammarRule{
				Head: NewNonTerminalToken("Q"),
				Production: []GrammarToken{
					NewTerminalToken("["),
					NewNonTerminalToken("S"),
					NewTerminalToken("]"),
				},
			},

			GrammarRule{
				Head: NewNonTerminalToken("Q"),
				Production: []GrammarToken{
					NewTerminalToken("sentence"),
				},
			},
		},
	}
}

func compareTables(t *testing.T, expected *FirstFollowTable, actual *FirstFollowTable) {
	panic("TODO")
}

func TestGetFollows(t *testing.T) {
	grammar := createExampleGrammar()
	table := FirstFollowTable{}
	expectedTable := FirstFollowTable{
		table: map[GrammarToken]FirstFollowRow{
			NewNonTerminalToken("S"): FirstFollowRow{
				First: Set[GrammarToken]{
					NewTerminalToken("sentence"): struct{}{},
					NewTerminalToken("|"):        struct{}{},
				},
			},

			NewNonTerminalToken("P"): FirstFollowRow{
				First: Set[GrammarToken]{
					NewTerminalToken("sentence"): struct{}{},
					NewTerminalToken("|"):        struct{}{},
				},
			},

			NewNonTerminalToken("Q"): FirstFollowRow{
				First: Set[GrammarToken]{
					NewTerminalToken("sentence"): struct{}{},
					NewTerminalToken("|"):        struct{}{},
				},
			},
		},
	}

	GetFirsts(&grammar, &table)

	compareTables(t, &expectedTable, &table)
}
