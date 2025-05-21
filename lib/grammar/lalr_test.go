package grammar

import (
	"testing"

	"github.com/Jose-Prince/UWULexer/lib"
)

func TestInitializeAutomata(t *testing.T) {
	// Definimos los tokens
	S := NewNonTerminalToken("S")
	A := NewNonTerminalToken("A")
	a := NewTerminalToken("a")
	b := NewTerminalToken("b")

	// Gramática:
	// S → A
	// A → a A | b

	rules := []GrammarRule{
		{
			Head:       S,
			Production: []GrammarToken{A},
		},
		{
			Head:       A,
			Production: []GrammarToken{a, A},
		},
		{
			Head:       A,
			Production: []GrammarToken{b},
		},
	}

	grammar := Grammar{
		InitialSimbol: S,
		Rules:         rules,
		Terminals:     lib.NewSet[GrammarToken](),
		NonTerminals:  lib.NewSet[GrammarToken](),
	}

	// Poblamos los sets de terminales y no terminales
	grammar.Terminals.Add(a)
	grammar.Terminals.Add(b)
	grammar.NonTerminals.Add(S)
	grammar.NonTerminals.Add(A)

	// Definimos la regla inicial extendida: S' → . S, $
	extendedRule := GrammarRule{
		Head:       NewNonTerminalToken("S'"),
		Production: []GrammarToken{S},
	}

	// Inicializamos el autómata
	auto := InitializeAutomata(extendedRule, grammar)

	// Obtenemos el estado inicial
	state0, exists := auto.nodes[0]
	if !exists {
		t.Fatalf("No se encontró el estado inicial")
	}

	// Verificamos que existan las producciones esperadas en el estado inicial
	expectedItems := []automataItem{
		{
			Rule:        extendedRule,
			DotPosition: 0,
			Lookahead:   []GrammarToken{NewEndToken()},
		},
		{
			Rule:        rules[0], // S → . A
			DotPosition: 0,
			Lookahead:   []GrammarToken{NewEndToken()},
		},
		{
			Rule:        rules[1], // A → . a A
			DotPosition: 0,
			Lookahead:   []GrammarToken{NewEndToken()},
		},
		{
			Rule:        rules[2], // A → . b
			DotPosition: 0,
			Lookahead:   []GrammarToken{NewEndToken()},
		},
	}

	for key, expected := range expectedItems {
		if _, found := state0.States[key]; !found {
			t.Errorf("No se encontró el item esperado: %s", expected.Rule.ToString())
		}
	}
}

func TestInitializeAutomataBrolo(t *testing.T) {
	// Definimos los tokens
	S := NewNonTerminalToken("S")
	C := NewNonTerminalToken("C")
	c := NewTerminalToken("c")
	d := NewTerminalToken("d")

	// Gramática:
	// S → C C
	// C → c C | d

	rules := []GrammarRule{
		{
			Head:       S,
			Production: []GrammarToken{C, C},
		},
		{
			Head:       C,
			Production: []GrammarToken{c, C},
		},
		{
			Head:       C,
			Production: []GrammarToken{d},
		},
	}

	grammar := Grammar{
		InitialSimbol: S,
		Rules:         rules,
		Terminals:     lib.NewSet[GrammarToken](),
		NonTerminals:  lib.NewSet[GrammarToken](),
	}

	// Poblamos los sets de terminales y no terminales
	grammar.Terminals.Add(c)
	grammar.Terminals.Add(d)
	grammar.NonTerminals.Add(S)
	grammar.NonTerminals.Add(C)

	// Definimos la regla inicial extendida: S' → . S, $
	extendedRule := GrammarRule{
		Head:       NewNonTerminalToken("S'"),
		Production: []GrammarToken{S},
	}

	// Inicializamos el autómata
	auto := InitializeAutomata(extendedRule, grammar)

	// Obtenemos el estado inicial
	state0, exists := auto.nodes[0]
	if !exists {
		t.Fatalf("No se encontró el estado inicial")
	}

	// Verificamos que existan las producciones esperadas en el estado inicial
	expectedItems := []automataItem{
		{
			Rule:        extendedRule,
			DotPosition: 0,
			Lookahead:   []GrammarToken{NewEndToken()},
		},
		{
			Rule:        rules[0], // S → . A
			DotPosition: 0,
			Lookahead:   []GrammarToken{NewEndToken()},
		},
		{
			Rule:        rules[1], // A → . a A
			DotPosition: 0,
			Lookahead:   []GrammarToken{c, d},
		},
		{
			Rule:        rules[2], // A → . b
			DotPosition: 0,
			Lookahead:   []GrammarToken{c, d},
		},
	}

	for key, expected := range expectedItems {
		if _, found := state0.States[key]; !found {
			t.Errorf("No se encontró el item esperado: %s", expected.Rule.ToString())
		}
	}
}
