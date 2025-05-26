package regex

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Jose-Prince/UWUCompiler/lib"
)

func _validateTree(t *testing.T, expected *AST, expectedCurrent int, result *AST, resultCurrent int, level int) error {
	bothDontExist := resultCurrent == -1 && expectedCurrent == -1
	if bothDontExist {
		return nil
	}

	expectedExistsButResultDoesnt := resultCurrent == -1 && expectedCurrent != resultCurrent
	if expectedExistsButResultDoesnt {
		return errors.New(fmt.Sprintf(`Expected:
%s
Actual:
%s
Result tree doesn't have node %s on level %d!`,
			expected.String(),
			result.String(),
			expected.nodes[expectedCurrent],
			level,
		))
	}

	resultExistsButExpectedDoesnt := expectedCurrent == -1 && expectedCurrent != resultCurrent
	if resultExistsButExpectedDoesnt {
		return errors.New(fmt.Sprintf(`Expected:
%s
Actual:
%s
Result tree has extra node %s on level %d!`,
			expected.String(),
			result.String(),
			result.nodes[resultCurrent],
			level,
		))
	}

	expectedNode := expected.nodes[expectedCurrent]
	actualNode := result.nodes[resultCurrent]
	if !expectedNode.Val.Equals(&actualNode.Val) {
		return errors.New(fmt.Sprintf(`Expected:
%s
Actual:
%s
Nodes on level %d don't match! %s != %s`,
			expected.String(),
			result.String(),
			level,
			expectedNode.String(),
			actualNode.String(),
		))
	}

	err := _validateTree(t, expected, expectedNode.left, result, actualNode.left, level+1)
	if err != nil {
		return err
	}

	err = _validateTree(t, expected, expectedNode.right, result, actualNode.right, level+1)
	if err != nil {
		return err
	}

	return nil
}

func validateTree(t *testing.T, expected *AST, actual *AST) error {
	if len(expected.nodes) != len(actual.nodes) {
		t.Errorf("Tree nodes don't match! %d != %d", len(expected.nodes), len(actual.nodes))
	}

	return _validateTree(t, expected, expected.RootIdx, actual, actual.RootIdx, 1)
}

func validateTable(t *testing.T, expected ASTTable, result ASTTable) error {
	if len(expected.Rows) != len(result.Rows) {
		t.Errorf("Tables lengths don't match! %d != %d", len(expected.Rows), len(result.Rows))
	}

	for i, expRow := range expected.Rows {
		resRow := result.Rows[i]

		if !expRow.Equals(&resRow) {
			return errors.New(fmt.Sprintf(`Expected:
%s
Result:
%s
Rows at index %d don't match!
%s
!=
%s`,
				expected,
				result,
				i,
				expRow,
				resRow,
			))
		}
	}

	return nil
}

func createLeftChild(b *AST, father int, value RX_Token) int {
	node := NewASTNode(value)
	node.father = father

	insertedIdx := len(b.nodes)
	b.nodes[father].left = insertedIdx
	b.nodes = append(b.nodes, node)

	return insertedIdx
}

func createRightChild(b *AST, father int, value RX_Token) int {
	node := NewASTNode(value)
	node.father = father

	insertedIdx := len(b.nodes)
	b.nodes[father].right = insertedIdx
	b.nodes = append(b.nodes, node)

	return insertedIdx
}

func CreateCanvasExampleTree() *AST {
	b := new(AST)
	root := NewASTNode(CreateOperatorToken(AND))
	b.nodes = append(b.nodes, root)
	b.RootIdx = 0

	rightTree := createRightChild(b, 0, CreateValueToken('#'))
	b.nodes[rightTree].extraProperties.acceptance = true
	b.AcceptedIdx = rightTree
	leftTree := createLeftChild(b, 0, CreateOperatorToken(AND))

	createRightChild(b, leftTree, CreateValueToken('b'))
	leftTree = createLeftChild(b, leftTree, CreateOperatorToken(AND))

	createRightChild(b, leftTree, CreateValueToken('b'))
	leftTree = createLeftChild(b, leftTree, CreateOperatorToken(AND))

	createRightChild(b, leftTree, CreateValueToken('a'))
	leftTree = createLeftChild(b, leftTree, CreateOperatorToken(ZERO_OR_MANY))

	leftTree = createLeftChild(b, leftTree, CreateOperatorToken(OR))
	createRightChild(b, leftTree, CreateValueToken('b'))
	createLeftChild(b, leftTree, CreateValueToken('a'))

	return b
}

func CreateCanvasExampleTable() ASTTable {
	return ASTTable{
		RootRow:       11,
		AcceptanceRow: 10,
		Rows: []TableRow{
			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{0: struct{}{}},
				lastpos:   lib.Set[int]{0: struct{}{}},
				followpos: lib.Set[int]{0: struct{}{}, 1: struct{}{}, 4: struct{}{}},
				simbol:    'a',
				token:     CreateValueToken('a'),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{1: struct{}{}},
				lastpos:   lib.Set[int]{1: struct{}{}},
				followpos: lib.Set[int]{0: struct{}{}, 1: struct{}{}, 4: struct{}{}},
				simbol:    'b',
				token:     CreateValueToken('b'),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{0: struct{}{}, 1: struct{}{}},
				lastpos:   lib.Set[int]{0: struct{}{}, 1: struct{}{}},
				followpos: lib.Set[int]{},
				simbol:    '\x00',
				token:     CreateOperatorToken(OR),
			},
			TableRow{
				nullable:  true,
				firstpos:  lib.Set[int]{0: struct{}{}, 1: struct{}{}},
				lastpos:   lib.Set[int]{0: struct{}{}, 1: struct{}{}},
				followpos: lib.Set[int]{},
				simbol:    '\x00',
				token:     CreateOperatorToken(ZERO_OR_MANY),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{4: struct{}{}},
				lastpos:   lib.Set[int]{4: struct{}{}},
				followpos: lib.Set[int]{6: struct{}{}},
				simbol:    'a',
				token:     CreateValueToken('a'),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{0: struct{}{}, 1: struct{}{}, 4: struct{}{}},
				lastpos:   lib.Set[int]{4: struct{}{}},
				followpos: lib.Set[int]{},
				simbol:    '\x00',
				token:     CreateOperatorToken(AND),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{6: struct{}{}},
				lastpos:   lib.Set[int]{6: struct{}{}},
				followpos: lib.Set[int]{8: struct{}{}},
				simbol:    'b',
				token:     CreateValueToken('b'),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{0: struct{}{}, 1: struct{}{}, 4: struct{}{}},
				lastpos:   lib.Set[int]{6: struct{}{}},
				followpos: lib.Set[int]{},
				simbol:    '\x00',
				token:     CreateOperatorToken(AND),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{8: struct{}{}},
				lastpos:   lib.Set[int]{8: struct{}{}},
				followpos: lib.Set[int]{10: struct{}{}},
				simbol:    'b',
				token:     CreateValueToken('b'),
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{0: struct{}{}, 1: struct{}{}, 4: struct{}{}},
				lastpos:   lib.Set[int]{8: struct{}{}},
				followpos: lib.Set[int]{},
				simbol:    '\x00',
				token:     CreateOperatorToken(AND),
			},

			TableRow{
				nullable:   false,
				firstpos:   lib.Set[int]{10: struct{}{}},
				lastpos:    lib.Set[int]{10: struct{}{}},
				followpos:  lib.Set[int]{},
				simbol:     '#',
				token:      CreateValueToken('#'),
				acceptance: true,
			},

			TableRow{
				nullable:  false,
				firstpos:  lib.Set[int]{0: struct{}{}, 1: struct{}{}, 4: struct{}{}},
				lastpos:   lib.Set[int]{10: struct{}{}},
				followpos: lib.Set[int]{},
				simbol:    '\x00',
				token:     CreateOperatorToken(AND),
			},
		},
	}
}

func TestASTCanvasExample(t *testing.T) {
	regexStream := []RX_Token{
		CreateValueToken('a'),
		CreateValueToken('b'),
		CreateOperatorToken(OR),
		CreateOperatorToken(ZERO_OR_MANY),
		CreateValueToken('a'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
		CreateValueToken('b'),
		CreateOperatorToken(AND),
	}
	tree := ASTFromRegex(regexStream)
	expectedTree := CreateCanvasExampleTree()

	err := validateTree(t, expectedTree, tree)
	if err != nil {
		t.Fatal(err.Error())
	}

	expectedTable := CreateCanvasExampleTable()
	table := tree.ToTable()

	err = validateTable(t, expectedTable, table)
	if err != nil {
		t.Fatal(err.Error())
	}
}
