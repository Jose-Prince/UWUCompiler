package regex

import (
	"errors"
	"fmt"
	"testing"
)

func _validateTree(t *testing.T, expected *BST, expectedCurrent int, result *BST, resultCurrent int, level int) error {
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

func validateTree(t *testing.T, expected *BST, actual *BST) error {
	if len(expected.nodes) != len(actual.nodes) {
		t.Errorf("Tree nodes don't match! %d != %d", len(expected.nodes), len(actual.nodes))
	}

	return _validateTree(t, expected, expected.RootIdx, actual, actual.RootIdx, 1)
}

func createLeftChild(b *BST, father int, value RX_Token) int {
	node := CreateBSTNode(value)
	node.father = father

	insertedIdx := len(b.nodes)
	b.nodes[father].left = insertedIdx
	b.nodes = append(b.nodes, node)

	return insertedIdx
}

func createRightChild(b *BST, father int, value RX_Token) int {
	node := CreateBSTNode(value)
	node.father = father

	insertedIdx := len(b.nodes)
	b.nodes[father].right = insertedIdx
	b.nodes = append(b.nodes, node)

	return insertedIdx
}

func CreateCanvasExampleTree() *BST {
	b := new(BST)
	root := CreateBSTNode(CreateOperatorToken(AND))
	b.nodes = append(b.nodes, root)
	b.RootIdx = 0

	createRightChild(b, 0, CreateValueToken('#'))
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

func TestCanvasExample(t *testing.T) {
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
	tree := BSTFromRegexStream(regexStream)

	expectedTree := CreateCanvasExampleTree()

	err := validateTree(t, expectedTree, tree)
	if err != nil {
		t.Fatal(err.Error())
	}
}
