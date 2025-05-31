package grammar

import (
	"testing"

	"github.com/Jose-Prince/UWUCompiler/lib"
)

func buildSimpleGrammar() *Grammar {
	E := NewNonTerminalToken("E")
	T := NewNonTerminalToken("T")
	plus := NewTerminalToken("+")
	id := NewTerminalToken("id")

	terminals := lib.NewSet[GrammarToken]()
	terminals.Add(plus)
	terminals.Add(id)

	nonTerminals := lib.NewSet[GrammarToken]()
	nonTerminals.Add(T)
	terminals.Add(E)

	return &Grammar{
		InitialSimbol: E,
		Rules: []GrammarRule{
			{Head: E, Production: []GrammarToken{E, plus, T}},
			{Head: E, Production: []GrammarToken{T}},
			{Head: T, Production: []GrammarToken{id}},
		},
		Terminals:    terminals,
		NonTerminals: nonTerminals,
	}
}

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
		if _, found := state0.Items[key]; !found {
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
		if _, found := state0.Items[key]; !found {
			t.Errorf("No se encontró el item esperado: %s", expected.Rule.ToString())
		}
	}
}

func TestGeneratedStatesFromInitial(t *testing.T) {
	// Tokens
	S_ := NewNonTerminalToken("S'")
	S := NewNonTerminalToken("S")
	A := NewNonTerminalToken("A")
	a := NewTerminalToken("a")
	b := NewTerminalToken("b")
	end := NewEndToken()

	// Reglas
	rules := []GrammarRule{
		{Head: S_, Production: []GrammarToken{S}},   // S'→ S
		{Head: S, Production: []GrammarToken{A}},    // S → A
		{Head: A, Production: []GrammarToken{a, A}}, // A → a A
		{Head: A, Production: []GrammarToken{b}},    // A → b
	}
	extended := GrammarRule{
		Head:       NewNonTerminalToken("S'"),
		Production: []GrammarToken{S},
	}

	// Gramática
	grammar := Grammar{
		InitialSimbol: S,
		Rules:         rules,
		Terminals:     lib.NewSet[GrammarToken](),
		NonTerminals:  lib.NewSet[GrammarToken](),
	}
	grammar.Terminals.Add(a)
	grammar.Terminals.Add(b)
	grammar.NonTerminals.Add(S)
	grammar.NonTerminals.Add(A)

	// Generar autómata real
	auto := InitializeAutomata(extended, grammar)

	// Crear estado esperado 1 (después de transición sobre S)
	expectedItemsS := map[int]automataItem{
		0: {
			Rule:        rules[0], // A -> a A
			DotPosition: 1,
			Lookahead:   []GrammarToken{end},
		},
	}

	expectedItemsA := map[int]automataItem{
		0: {
			Rule:        rules[1], // A -> a A
			DotPosition: 1,
			Lookahead:   []GrammarToken{end},
		},
	}

	expectedItemsa := map[int]automataItem{
		0: {
			Rule:        rules[2], // A -> a A
			DotPosition: 1,
			Lookahead:   []GrammarToken{end},
		},
		1: {
			Rule:        rules[3], // A -> a A
			DotPosition: 0,
			Lookahead:   []GrammarToken{end},
		},
	}

	expectedItemsb := map[int]automataItem{
		0: {
			Rule:        rules[3], // A -> a A
			DotPosition: 1,
			Lookahead:   []GrammarToken{end},
		},
	}

	// Verificar si el estado generado contiene exactamente esos ítems
	initial := auto.nodes[0]
	targetStateID, ok := initial.Productions[S.String()]
	if !ok {
		t.Fatalf("No se encontró transición sobre símbolo %s", S.String())
	}

	targetState := auto.nodes[targetStateID]
	if !compareStateItems(targetState.Items, expectedItemsS) {
		t.Errorf("Los ítems del estado %d no coinciden con los esperados", targetStateID)
	}
	targetState = auto.nodes[targetStateID+1]
	if !compareStateItems(targetState.Items, expectedItemsA) {
		t.Errorf("Los ítems del estado %d no coinciden con los esperados", targetStateID)
	}
	targetState = auto.nodes[targetStateID+2]
	if !compareStateItems(targetState.Items, expectedItemsa) {
		t.Errorf("Los ítems del estado %d no coinciden con los esperados", targetStateID)
	}
	targetState = auto.nodes[targetStateID+3]
	if !compareStateItems(targetState.Items, expectedItemsb) {
		t.Errorf("Los ítems del estado %d no coinciden con los esperados", targetStateID)
	}
}

func compareStateItems(actual, expected map[int]automataItem) bool {
	if len(actual) != len(expected) {
		return false
	}

	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Rule.EqualRule(&e.Rule) &&
				a.DotPosition == e.DotPosition {
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

func TestClosure(t *testing.T) {
	// Tokens
	E := NewNonTerminalToken("E")
	T := NewNonTerminalToken("T")
	plus := NewTerminalToken("+")
	id := NewTerminalToken("id")
	end := NewEndToken()

	// Gramática:
	// E → E + T
	// E → T
	// T → id

	rules := []GrammarRule{
		{Head: E, Production: []GrammarToken{E, plus, T}},
		{Head: E, Production: []GrammarToken{T}},
		{Head: T, Production: []GrammarToken{id}},
	}

	// Gramática
	grammar := Grammar{
		InitialSimbol: E,
		Rules:         rules,
		Terminals:     lib.NewSet[GrammarToken](),
		NonTerminals:  lib.NewSet[GrammarToken](),
	}
	grammar.Terminals.Add(plus)
	grammar.Terminals.Add(id)
	grammar.NonTerminals.Add(E)
	grammar.NonTerminals.Add(T)

	// Item inicial: E → .E + T, $
	initialItem := automataItem{
		Rule:        GrammarRule{Head: NewNonTerminalToken("S'"), Production: []GrammarToken{E}},
		DotPosition: 0,
		Lookahead:   []GrammarToken{end},
	}

	initialItems := map[int]automataItem{
		0: initialItem,
	}

	instialState := automataState{
		Items:       initialItems,
		Productions: make(map[string]int),
	}

	// Ejecutar closure
	closure(instialState, grammar)

	// Esperamos ver los siguientes ítems:
	expected := []automataItem{
		initialItem,
		{
			Rule:        rules[0], // E → . E + T
			DotPosition: 0,
			Lookahead:   []GrammarToken{plus},
		},
		{
			Rule:        rules[1], // E → . T
			DotPosition: 0,
			Lookahead:   []GrammarToken{plus},
		},
		{
			Rule:        rules[2], // T → . id
			DotPosition: 0,
			Lookahead:   []GrammarToken{plus},
		},
	}

	// Validación
	for _, exp := range expected {
		found := false
		for _, item := range instialState.Items {
			if item.Rule.EqualRule(&exp.Rule) &&
				item.DotPosition == exp.DotPosition &&
				lookaheadsEqual(item.Lookahead, exp.Lookahead) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("No se encontró el item esperado: %v", exp.Rule.ToString())
		}
	}
}

func lookaheadsEqual(a, b []GrammarToken) bool {
	if a[0].IsEnd && b[0].IsEnd {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]bool)
	for _, tok := range a {
		m[tok.String()] = true
	}
	for _, tok := range b {
		if !m[tok.String()] {
			return false
		}
	}
	return true
}

func TestConvertToLALR(t *testing.T) {
	// Gramática simple:
	// S → CC
	// C → cC | d

	S := NewNonTerminalToken("S")
	C := NewNonTerminalToken("C")
	c := NewTerminalToken("c")
	d := NewTerminalToken("d")

	rules := []GrammarRule{
		{Head: S, Production: []GrammarToken{C, C}},
		{Head: C, Production: []GrammarToken{c, C}},
		{Head: C, Production: []GrammarToken{d}},
	}

	nonTerminals := lib.NewSet[GrammarToken]()
	nonTerminals.Add(S)
	nonTerminals.Add(C)

	terminals := lib.NewSet[GrammarToken]()
	terminals.Add(c)
	terminals.Add(d)

	g := Grammar{
		Rules:         rules,
		Terminals:     terminals,
		NonTerminals:  nonTerminals,
		InitialSimbol: S,
	}

	initialRule := GrammarRule{Head: NewNonTerminalToken("S'"), Production: []GrammarToken{S}}

	// Construir autómata LR(1)
	lr1 := InitializeAutomata(initialRule, g)
	lalr := lr1

	lr1StateCount := len(lr1.nodes)
	t.Logf("Estados LR(1): %d", lr1StateCount)

	// Convertir a LALR
	lalr.simplifyStates()

	lalrStateCount := len(lalr.nodes)
	t.Logf("Estados LALR: %d", lalrStateCount)

	if lalrStateCount >= lr1StateCount {
		t.Errorf("Esperábamos menos estados en LALR que en LR(1), pero LALR tiene %d y LR(1) tiene %d", lalrStateCount, lr1StateCount)
	}
}

func TestGenerateParsingTable(t *testing.T) {
	// Gramática simple:
	// S → CC
	// C → cC | d

	S := NewNonTerminalToken("S")
	C := NewNonTerminalToken("C")
	c := NewTerminalToken("c")
	d := NewTerminalToken("d")

	rules := []GrammarRule{
		{Head: S, Production: []GrammarToken{C, C}},
		{Head: C, Production: []GrammarToken{c, C}},
		{Head: C, Production: []GrammarToken{d}},
	}

	nonTerminals := lib.NewSet[GrammarToken]()
	nonTerminals.Add(S)
	nonTerminals.Add(C)

	terminals := lib.NewSet[GrammarToken]()
	terminals.Add(c)
	terminals.Add(d)

	g := Grammar{
		Rules:         rules,
		Terminals:     terminals,
		NonTerminals:  nonTerminals,
		InitialSimbol: S,
	}

	initialRule := GrammarRule{Head: NewNonTerminalToken("S'"), Production: []GrammarToken{S}}

	// Construir autómata LR(1)
	lr1 := InitializeAutomata(initialRule, g)
	lalr := lr1
	lalr.simplifyStates()

	parsingTable := lalr.generateParsingTable(&g)

	foundShift := false
	foundReduce := false
	foundAccept := false

	for stateID, actions := range parsingTable.ActionTable {
		for symbol, action := range actions {
			t.Logf("State %v on symbol %v => Action: %+v", stateID, symbol.String(), action)

			if action.Accept {
				foundAccept = true
			}
			if action.Shift.HasValue() {
				foundShift = true
			}
			if action.Reduce.HasValue() {
				foundReduce = true
			}
		}
	}

	if !foundAccept {
		t.Error("No se encontró ninguna acción de aceptación (Accept)")
	}
	if !foundShift {
		t.Error("No se encontró ninguna acción Shift en la tabla de parseo")
	}
	if !foundReduce {
		t.Error("No se encontró ninguna acción Reduce en la tabla de parseo")
	}
}
