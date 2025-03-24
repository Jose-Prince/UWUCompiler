package lib

import "testing"

func TestConvertFromTableToAFD(t *testing.T) {
	// Definir tabla de ejemplo
	table := []*TableRow{
		{false, []int{0}, []int{0}, []int{1}, 'a'},
		{false, []int{1}, []int{1}, []int{}, 'b'},
		{false, []int{0}, []int{1}, []int{}, '\x00'},
	}

	afd := ConvertFromTableToAFD(table)

	// Verificar estado inicial
	expectedInitial := "0"
	if afd.InitialState != expectedInitial {
		t.Errorf("Expected initial state %s, got %s", expectedInitial, afd.InitialState)
	}

	// Verificar transiciones
	expectedTransitions := map[string]map[string]string{
		"0": {"a": "1"},
	}

	for state, transitions := range expectedTransitions {
		for input, expectedNextState := range transitions {
			if afd.Transitions[state][input] != expectedNextState {
				t.Errorf("Expected transition (%s, %s) -> %s, got %s",
					state, input, expectedNextState, afd.Transitions[state][input])
			}
		}
	}

	// Verificar estados de aceptación
	expectedFinalState := "1"
	if !afd.AcceptanceStates.Contains(expectedFinalState) {
		t.Errorf("Expected final state %s to be in AcceptanceStates", expectedFinalState)
	}
}

// Test AFD derivation
func TestDerivation(t *testing.T) {
	afd := AFD{
		InitialState: "1",
		Transitions: map[string]map[string]string{
			"1": {"a": "2"},
		},
		AcceptanceStates: Set[string]{"2": struct{}{}},
	}

	str_example := "a"

	result := afd.derivation(str_example)

	if !result {
		t.Errorf("Final state not in acceptance state")
	}
}

// func TestConvertFromTableToAFD(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		table []*TableRow
// 		want  *AFD
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := ConvertFromTableToAFD(tt.table)
// 			// TODO: update the condition below to compare got with tt.want.
// 			if true {
// 				t.Errorf("ConvertFromTableToAFD() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
