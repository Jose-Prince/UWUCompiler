package regex

import (
	"testing"

	"github.com/Jose-Prince/UWULexer/lib"
)

func validateTree(t *testing.T, tree *BST, expectedKeys *[]int, expectedVals *[]RX_Token) {
	if len(tree.nodes) != len(*expectedKeys) {
		t.Fatalf("Número incorrecto de nodos. Esperado %d, pero obtuvo %d", len(*expectedKeys), len(tree.nodes))
	}

	// Verifies each node
	for i, node := range tree.nodes {
		if node.Key != (*expectedKeys)[i] {
			t.Errorf("Nodo incorrecto en posición %d: esperado (%d) pero obtuvo (%d)",
				i, (*expectedKeys)[i], node.Key)
		}

		bothAreValue := node.Val.IsValue() && (*expectedVals)[i].IsValue()
		bothAreOperator := node.Val.IsOperator() && (*expectedVals)[i].IsOperator()
		if bothAreValue && lib.OptionalEquals(node.Val.GetValue(), (*expectedVals)[i].GetValue()) {
			t.Errorf("Nodo incorrecto en posición %d: esperado (%v) pero obtuvo (%v)",
				i, (*expectedVals)[i].GetValue(), node.Val.GetValue())

		} else if bothAreOperator && node.Val.GetOperator() == (*expectedVals)[i].GetOperator() {
			t.Errorf("Nodo incorrecto en posición %d: esperado (%d) pero obtuvo (%d)",
				i, node.Val.GetOperator(), node.Val.GetOperator())

		} else {
			t.Errorf("Nodo incorrecto en posición %d: los tipos de valor no coinciden", i)
		}

	}
}

// General Test for BST
func TestBST(t *testing.T) {
	// Node Creation
	nodes := []RX_Token{
		CreateOperatorToken(AND),
		CreateValueToken('a'),
		CreateValueToken('b'),
	}

	// Creates tree
	tree := new(BST)

	tree.Insertion(nodes)

	// Expected nodes
	expectedKeys := []int{3, 2, 1}
	expectedVals := []RX_Token{CreateValueToken('b'), CreateValueToken('a'), CreateOperatorToken(AND)}

	// Verifies total nodes
	validateTree(t, tree, &expectedKeys, &expectedVals)
}

// Test Epsilon value
func TestEpsilon(t *testing.T) {
	nodes := []RX_Token{
		CreateOperatorToken(OR),
		CreateValueToken('a'),
		CreateEpsilonToken(),
	}

	tree := new(BST)

	tree.Insertion(nodes)

	expectedKeys := []int{3, 2, 1}
	expectedVals := []RX_Token{CreateEpsilonToken(), CreateValueToken('a'), CreateOperatorToken(OR)}
	validateTree(t, tree, &expectedKeys, &expectedVals)

	table := ConvertTreeToTable(tree)

	expectedFirstPos := [][]int{{}, {1}, {1}}
	expectedLastPos := [][]int{{}, {1}, {1}}
	expectedFollowPos := [][]int{{}, {}, {}}
	expectedNullable := []bool{true, false, true}

	for i, row := range table {
		if !equalSlices(row.firstpos, expectedFirstPos[i]) {
			t.Errorf("Error en firstpos en índice %d: esperado %v, obtenido %v", i, expectedFirstPos[i], row.firstpos)
		}
		if !equalSlices(row.lastpos, expectedLastPos[i]) {
			t.Errorf("Error en lastpos en índice %d: esperado %v, obtenido %v", i, expectedLastPos[i], row.lastpos)
		}
		if !equalSlices(row.followpos, expectedFollowPos[i]) {
			t.Errorf("Error en lastpos en índice %d: esperado %v, obtenido %v", i, expectedFollowPos[i], row.followpos)
		}

		if row.nullable != expectedNullable[i] {
			t.Errorf("Error en nullable en índice %d: esperado %v, obtenido %v", i, expectedNullable[i], row.nullable)
		}
	}
}

// Class example
func TestExampleBST(t *testing.T) {
	// Node Creation
	nodes := []RX_Token{
		CreateOperatorToken(AND),
		CreateValueToken('#'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('a'),
		CreateOperatorToken(ZERO_OR_MANY),
		CreateOperatorToken(OR),
		CreateValueToken('a'),
		CreateValueToken('b'),
	}

	// Creates tree
	tree := new(BST)

	tree.Insertion(nodes)

	// Expected nodes
	expectedKeys := []int{11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	expectedVals := []RX_Token{
		CreateValueToken('b'),
		CreateValueToken('a'),
		CreateOperatorToken(OR),
		CreateOperatorToken(ZERO_OR_MANY),
		CreateValueToken('a'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('#'),
		CreateOperatorToken(AND),
	}

	validateTree(t, tree, &expectedKeys, &expectedVals)
}

func TestTable(t *testing.T) {
	nodes := []RX_Token{
		CreateOperatorToken(AND),
		CreateValueToken('a'),
		CreateValueToken('b'),
	}

	// Creates tree
	tree := new(BST)

	tree.Insertion(nodes)

	table := ConvertTreeToTable(tree)

	// Valores esperados
	expectedFirstPos := [][]int{{0}, {1}, {0}}
	expectedLastPos := [][]int{{0}, {1}, {1}}
	expectedFollowPos := [][]int{{1}, {}, {}}
	expectedNullable := []bool{false, false, false}

	// Validar la tabla generada
	for i, row := range table {
		if !equalSlices(row.firstpos, expectedFirstPos[i]) {
			t.Errorf("Error en firstpos en índice %d: esperado %v, obtenido %v", i, expectedFirstPos[i], row.firstpos)
		}
		if !equalSlices(row.lastpos, expectedLastPos[i]) {
			t.Errorf("Error en lastpos en índice %d: esperado %v, obtenido %v", i, expectedLastPos[i], row.lastpos)
		}
		if !equalSlices(row.followpos, expectedFollowPos[i]) {
			t.Errorf("Error en lastpos en índice %d: esperado %v, obtenido %v", i, expectedFollowPos[i], row.followpos)
		}

		if row.nullable != expectedNullable[i] {
			t.Errorf("Error en nullable en índice %d: esperado %v, obtenido %v", i, expectedNullable[i], row.nullable)
		}
	}
}

func TestExampleTable(t *testing.T) {
	nodes := []RX_Token{
		CreateOperatorToken(AND),
		CreateValueToken('#'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('a'),
		CreateOperatorToken(ZERO_OR_MANY),
		CreateOperatorToken(OR),
		CreateValueToken('a'),
		CreateValueToken('b'),
	}

	// Creates tree
	tree := new(BST)

	tree.Insertion(nodes)

	table := ConvertTreeToTable(tree)

	// Valores esperados
	expectedFirstPos := [][]int{{0}, {1}, {0, 1}, {0, 1}, {4}, {0, 1, 4}, {6}, {0, 1, 4}, {8}, {0, 1, 4}, {10}, {0, 1, 4}}
	expectedLastPos := [][]int{{0}, {1}, {0, 1}, {0, 1}, {4}, {4}, {6}, {6}, {8}, {8}, {10}, {10}}
	expectedNullable := []bool{false, false, false, true, false, false, false, false, false, false, false, false}

	// Validar la tabla generada
	for i, row := range table {
		if !equalSlices(row.firstpos, expectedFirstPos[i]) {
			t.Errorf("Error en firstpos en índice %d: esperado %v, obtenido %v", i, expectedFirstPos[i], row.firstpos)
		}
		if !equalSlices(row.lastpos, expectedLastPos[i]) {
			t.Errorf("Error en lastpos en índice %d: esperado %v, obtenido %v", i, expectedLastPos[i], row.lastpos)
		}
		if row.nullable != expectedNullable[i] {
			t.Errorf("Error en nullable en índice %d: esperado %v, obtenido %v", i, expectedNullable[i], row.nullable)
		}
	}
}

func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// func TestBST_Insertion(t *testing.T) {
// 	tests := []struct {
// 		name string // description of this test case
// 		// Named input parameters for target function.
// 		postfix []RX_Token
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// TODO: construct the receiver type.
// 			var b BST
// 			b.Insertion(tt.postfix)
// 		})
// 	}
// }
