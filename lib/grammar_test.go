package lib

import (
	"fmt"
	"strings"
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
			NewTerminalToken("|"):        struct{}{},
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
					NewTerminalToken("|"),
					NewNonTerminalToken("S"),
					NewTerminalToken("|"),
				},
			},

			GrammarRule{
				Head: NewNonTerminalToken("P"),
				Production: []GrammarToken{
					NewTerminalToken("sentence"),
				},
			},
		},
	}
}

type StringerKey interface {
	comparable
	fmt.Stringer
}

func prettyPrintTable[K StringerKey, V fmt.Stringer](table *map[K]V) string {
	b := strings.Builder{}
	return b.String()
}

func compareSets[T StringerKey](t *testing.T, expected Set[T], actual Set[T]) {
	if len(expected) != len(actual) {
		t.Logf("Expected:\n%s", expected)
		t.Logf("Actual:\n%s", actual)
		t.Fatalf("%d != %d\nSet lengths don't match!", len(expected), len(actual))
	}

	for expectedKey := range expected {
		if !actual.Contains(expectedKey) {
			t.Logf("Expected:\n%s", expected)
			t.Logf("Actual:\n%s", actual)
			t.Fatalf("Element %s was not found in actual set!", expectedKey)
		}
	}
}

func compareTables(t *testing.T, expected *FirstFollowTable, actual *FirstFollowTable) {
	expectedTable := expected.table
	actualTable := actual.table

	if len(expectedTable) != len(actualTable) {
		t.Logf("Expected:\n%s", prettyPrintTable(&expectedTable))
		t.Logf("Actual:\n%s", prettyPrintTable(&actualTable))
		t.Fatalf("%d != %d\n expected table length is not the same as actual table length", len(expectedTable), len(actualTable))
	}

	for expectedKey, expectedValue := range expectedTable {
		actualValue, found := actualTable[expectedKey]
		if !found {
			t.Logf("Expected:\n%s", prettyPrintTable(&expectedTable))
			t.Logf("Actual:\n%s", prettyPrintTable(&actualTable))
			t.Fatalf("Key not found in actual: %s", expectedKey)
		}

		t.Logf("Checking key %s:", expectedKey)
		compareSets(t, expectedValue.First, actualValue.First)
		compareSets(t, expectedValue.Follow, actualValue.Follow)
	}
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
