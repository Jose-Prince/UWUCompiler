package regex

import (
	"math"
	"strconv"
	"strings"

	"github.com/Jose-Prince/UWULexer/lib"
)

type ASTNode struct {
	Val RX_Token

	father int

	left  int
	right int

	extraProperties TableRow
}

func (s ASTNode) String() string {
	b := strings.Builder{}
	b.WriteString("{ ")
	b.WriteString(s.Val.String())
	b.WriteString(" }")
	return b.String()
}

func NewASTNode(val RX_Token) ASTNode {
	return ASTNode{
		Val: val,

		father: -1,
		left:   -1,
		right:  -1,

		extraProperties: NewTableRow(),
	}
}

type AST struct {
	nodes       []ASTNode
	RootIdx     int
	AcceptedIdx int
}

func bstTreeToString(s *AST, current int, b *strings.Builder, level uint) {
	if current == -1 {
		return
	}

	for range level {
		b.WriteString("  ")
	}
	b.WriteString(s.nodes[current].String())
	b.WriteRune('\n')

	left := s.nodes[current].left
	bstTreeToString(s, left, b, level+1)

	right := s.nodes[current].right
	bstTreeToString(s, right, b, level+1)
}

func (s AST) String() string {
	b := strings.Builder{}
	bstTreeToString(&s, s.RootIdx, &b, 0)
	return b.String()
}

func (b ASTNode) Copy() ASTNode {
	var other ASTNode
	other.Val = b.Val
	other.father = b.father
	other.left = b.left
	other.right = b.right
	other.extraProperties = b.extraProperties

	return other
}

type TableRow struct {
	nullable   bool
	firstpos   lib.Set[int]
	lastpos    lib.Set[int]
	followpos  lib.Set[int]
	simbol     rune
	token      RX_Token
	acceptance bool
}

func NewTableRow() TableRow {
	return TableRow{
		firstpos:  lib.NewSet[int](),
		lastpos:   lib.NewSet[int](),
		followpos: lib.NewSet[int](),
	}
}

func (s TableRow) Equals(other *TableRow) bool {
	return s.nullable == other.nullable &&
		s.simbol == other.simbol &&
		s.acceptance == other.acceptance &&
		s.token.Equals(&other.token) &&
		s.firstpos.Equals(&other.firstpos) &&
		s.lastpos.Equals(&other.lastpos) &&
		s.followpos.Equals(&other.followpos)
}

func (s TableRow) String() string {
	b := strings.Builder{}
	b.WriteString("{ ")

	b.WriteString("simbol = ")
	b.WriteRune(s.simbol)

	b.WriteString(", acceptance = ")
	if s.acceptance {
		b.WriteRune('T')
	} else {
		b.WriteRune('F')
	}

	b.WriteString(", nullable = ")
	if s.nullable {
		b.WriteRune('T')
	} else {
		b.WriteRune('F')
	}

	b.WriteString(", firstPos = ")
	b.WriteString(lib.StableSetString(s.firstpos))
	b.WriteString(", lasPos = ")
	b.WriteString(lib.StableSetString(s.lastpos))
	b.WriteString(", followPos = ")
	b.WriteString(lib.StableSetString(s.followpos))

	b.WriteString(", tk = ")
	b.WriteString(s.token.String())

	b.WriteString(" }")
	return b.String()
}

func (b *ASTNode) IsNullable() bool {
	return b.extraProperties.nullable
}

func (b *ASTNode) IsLeaf() bool {
	return b.left == -1 && b.right == -1
}

func ASTFromRegex(postfix []RX_Token) *AST {
	b := new(AST)
	postfix = append(postfix, CreateValueToken('#'))
	postfix = append(postfix, CreateOperatorToken(AND))

	stack := lib.NewStack[int]()

	for _, v := range postfix {
		node := NewASTNode(v)
		i := len(b.nodes)

		if v.IsOperator() {
			op := v.GetOperator()
			switch op {
			case AND, OR:
				right := stack.Pop().GetValue()
				left := stack.Pop().GetValue()

				node.left = left
				node.right = right

				b.nodes[left].father = i
				b.nodes[right].father = i

			case ZERO_OR_MANY:
				left := stack.Pop().GetValue()

				node.left = left
				b.nodes[left].father = i
			}
		}

		stack.Push(i)
		b.nodes = append(b.nodes, node)
	}

	b.RootIdx = len(b.nodes) - 1
	b.AcceptedIdx = len(b.nodes) - 2
	return b
}

type ASTTable []TableRow

func (s ASTTable) String() string {
	b := strings.Builder{}

	MAX_DIGITS := 3
	for i, row := range s {
		b.WriteString(strconv.FormatInt(int64(i), 10))

		rightPadding := max(0, MAX_DIGITS-1-int(math.Floor(math.Log10(float64(i)))))
		if i == 0 {
			rightPadding = MAX_DIGITS - 1
		}
		for range rightPadding {
			b.WriteString(" ")
		}

		b.WriteString(": ")
		b.WriteString(row.String())
		b.WriteRune('\n')
	}

	return b.String()
}

func (tree *AST) ToTable() ASTTable {
	// Compute first and last pos of all nodes...
	for i, node := range tree.nodes {
		if node.IsLeaf() {
			nullable := node.Val.IsValue() && !node.Val.GetValue().HasValue()
			firstPos := lib.NewSet[int]()
			lastPos := lib.NewSet[int]()
			simbol := '\x00'

			if !nullable {
				firstPos.Add(i)
				lastPos.Add(i)
				simbol = node.Val.GetValue().GetValue()
			}

			row := NewTableRow()
			row.nullable = nullable
			row.firstpos = firstPos
			row.lastpos = lastPos
			row.simbol = simbol
			row.token = node.Val
			row.acceptance = i == tree.AcceptedIdx

			tree.nodes[i].extraProperties = row
		} else {
			simbol := '\x00'

			op := node.Val.GetOperator()
			switch op {

			case OR:
				left := tree.nodes[node.left]
				right := tree.nodes[node.right]

				nullable := left.IsNullable() || right.IsNullable()
				firstPos := lib.NewSet[int]()
				lastPos := lib.NewSet[int]()

				firstPos.Merge(&right.extraProperties.firstpos)
				firstPos.Merge(&left.extraProperties.firstpos)

				lastPos.Merge(&right.extraProperties.lastpos)
				lastPos.Merge(&left.extraProperties.lastpos)

				row := NewTableRow()
				row.nullable = nullable
				row.firstpos = firstPos
				row.lastpos = lastPos
				row.simbol = simbol
				row.token = node.Val
				tree.nodes[i].extraProperties = row

			case AND:
				left := tree.nodes[node.left]
				right := tree.nodes[node.right]

				nullable := left.IsNullable() && right.IsNullable()
				firstPos := lib.NewSet[int]()
				firstPos.Merge(&left.extraProperties.firstpos)
				if left.IsNullable() {
					firstPos.Merge(&right.extraProperties.firstpos)
				}

				lastPos := lib.NewSet[int]()
				lastPos.Merge(&right.extraProperties.lastpos)
				if right.IsNullable() {
					lastPos.Merge(&left.extraProperties.lastpos)
				}

				row := NewTableRow()
				row.nullable = nullable
				row.firstpos = firstPos
				row.lastpos = lastPos
				row.simbol = simbol
				row.token = node.Val
				tree.nodes[i].extraProperties = row

			case ZERO_OR_MANY:
				left := tree.nodes[node.left]

				nullable := true
				firstPos := left.extraProperties.firstpos
				lastPos := left.extraProperties.lastpos

				row := NewTableRow()
				row.nullable = nullable
				row.firstpos = firstPos
				row.lastpos = lastPos
				row.simbol = simbol
				row.token = node.Val

				tree.nodes[i].extraProperties = row
			}
		}
	}

	// Compute followpos of all nodes...
	for _, node := range tree.nodes {
		if !node.Val.IsOperator() {
			continue
		}

		op := node.Val.GetOperator()
		switch op {
		case AND:
			left := tree.nodes[node.left]
			right := tree.nodes[node.right]

			for i := range left.extraProperties.lastpos {
				tree.nodes[i].extraProperties.followpos.Merge(&right.extraProperties.firstpos)
			}
		case ZERO_OR_MANY:
			for i := range node.extraProperties.lastpos {
				tree.nodes[i].extraProperties.followpos.Merge(&node.extraProperties.firstpos)
			}
		}
	}

	table := []TableRow{}
	for _, n := range tree.nodes {
		row := n.extraProperties

		table = append(table, row)
	}

	return table
}
