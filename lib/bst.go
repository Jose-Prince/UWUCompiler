package lib

type BSTNode struct {
	Key int
	Val RX_Token

	father int

	left  int
	right int

	extraProperties TableRow
}

func CreateBSTNode(val RX_Token) BSTNode {
	return BSTNode{
		Val: val,

		father: -1,
		left:   -1,
		right:  -1,
	}
}

type BST struct {
	nodes []BSTNode
}

func (b BSTNode) Copy() BSTNode {
	var other BSTNode
	other.Key = b.Key
	other.Val = b.Val
	other.father = b.father
	other.left = b.left
	other.right = b.right
	other.extraProperties = b.extraProperties

	return other
}

type TableRow struct {
	nullable  bool
	firstpos  []int
	lastpos   []int
	followpos []int
	simbol    rune
}

func (b *BSTNode) IsNullable() bool {
	return b.extraProperties.nullable
}

func (b *BSTNode) IsLeaf() bool {
	return b.left == -1 && b.right == -1
}

func (b *BST) Insertion(postfix []RX_Token) {
	postfix = append(postfix, CreateValueToken('#'))
	postfix = append(postfix, CreateOperatorToken(AND))

	var stack Stack[int]
	var nodes []BSTNode

	for _, v := range postfix {
		node := CreateBSTNode(v)
		i := len(nodes)

		if v.IsOperator() {
			op := v.GetOperator()
			switch op {
			case AND, OR:
				right := stack.Pop().GetValue()
				left := stack.Pop().GetValue()

				node.left = left
				node.right = right

				nodes[left].father = i
				nodes[right].father = i

			case ZERO_OR_MANY:
				left := stack.Pop().GetValue()

				node.left = left
				nodes[left].father = i
			}
		}

		stack.Push(i)
		nodes = append(nodes, node)
	}

	b.nodes = nodes
}

func ConvertTreeToTable(tree *BST) []*TableRow {
	table := []*TableRow{}
	andToken := CreateOperatorToken(AND)
	zeroToken := CreateOperatorToken(ZERO_OR_MANY)

	for i, node := range tree.nodes {
		if node.IsLeaf() {
			nullable := !(node.Val.IsValue() && node.Val.GetValue().HasValue())
			firstPos := []int{}
			lastPos := []int{}

			if node.Val.IsDummy() || node.Val.GetValue().HasValue() {
				firstPos = append(firstPos, i)
				lastPos = append(lastPos, i)
			} else {
				for j := i - 1; j >= 0; j-- {
					if tree.nodes[j].Val.IsOperator() && (tree.nodes[j].Val.Equals(&andToken) || tree.nodes[j].Val.Equals(&zeroToken)) {
						lastPos = append(lastPos, tree.nodes[j].extraProperties.lastpos...)
						break
					}
				}
			}

			var simbol rune
			// if node.Val.IsDummy() {
			// 	simbol = ''
			// }

			if !nullable && !node.Val.IsDummy() {
				simbol = node.Val.GetValue().GetValue()
			}

			row := TableRow{
				nullable: nullable, firstpos: firstPos, lastpos: lastPos, simbol: simbol,
			}

			tree.nodes[i].extraProperties = row
		} else {
			op := node.Val.GetOperator()
			switch op {
			case AND:
				left := tree.nodes[node.left]
				right := tree.nodes[node.right]

				nullable := left.IsNullable() && right.IsNullable()
				firstPos := left.extraProperties.firstpos
				if left.IsNullable() {
					firstPos = append(firstPos, right.extraProperties.firstpos...)
				}

				lastPos := right.extraProperties.lastpos
				if left.IsNullable() {
					lastPos = append(lastPos, left.extraProperties.lastpos...)
				}

				row := TableRow{
					nullable: nullable, firstpos: firstPos, lastpos: lastPos, simbol: '\x00',
				}
				tree.nodes[i].extraProperties = row
				for _, i := range left.extraProperties.lastpos {
					node_i := tree.nodes[i]
					if node_i.IsLeaf() {
						tree.nodes[i].extraProperties.followpos = append(tree.nodes[i].extraProperties.followpos, right.extraProperties.firstpos...)
					}
				}

			case OR:
				left := tree.nodes[node.left]
				right := tree.nodes[node.right]

				nullable := left.IsNullable() || right.IsNullable()
				firstPos := left.extraProperties.firstpos
				firstPos = append(firstPos, right.extraProperties.firstpos...)

				lastPos := right.extraProperties.lastpos
				lastPos = append(lastPos, left.extraProperties.lastpos...)

				row := TableRow{
					nullable: nullable, firstpos: firstPos, lastpos: lastPos, simbol: '\x00',
				}
				tree.nodes[i].extraProperties = row

			case ZERO_OR_MANY:
				left := tree.nodes[node.left]

				nullable := true
				firstPos := left.extraProperties.firstpos

				lastPos := left.extraProperties.lastpos

				for j := i - 1; j >= 0; j-- {
					if tree.nodes[j].Val.Equals(&andToken) && j != i-1 {
						lastPos = append(lastPos, tree.nodes[j].extraProperties.lastpos...)
						break
					}
				}
				row := TableRow{
					nullable: nullable, firstpos: firstPos, lastpos: lastPos, simbol: '\x00',
				}
				tree.nodes[i].extraProperties = row

				for _, i := range lastPos {
					node_i := tree.nodes[i]
					if node_i.IsLeaf() {
						tree.nodes[i].extraProperties.followpos = append(tree.nodes[i].extraProperties.followpos, firstPos...)
					}
				}
			}
		}
	}

	for _, n := range tree.nodes {
		row := n.extraProperties

		table = append(table, &row)
	}

	return table
}
