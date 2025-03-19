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
	simbol    string
}

func (b *BSTNode) IsNullable() bool {
	return b.extraProperties.nullable
}

func (b *BSTNode) IsLeaf() bool {
	return b.left == -1 && b.right == -1
}

func (b *BST) List() []*BSTNode {

	// FIXME: This operation changes the tree everytime because the values are references!
	// Should be fixed now...
	son := 0
	for {
		if b.nodes[son].left < 0 {
			break
		}

		son = b.nodes[son].left
	}

	result := []*BSTNode{}

	for {
		result = append(result, &b.nodes[son])

		father := b.nodes[son].father
		if father >= 0 {
			brother := b.nodes[father].right
			if brother >= 0 {
				result = append(result, &b.nodes[brother])
			}
		}

		if father >= 0 {
			break
		}
		son = father
	}

	return result
}

func (b *BST) Insertion(postfix []RX_Token) {
    postfix = append(postfix, CreateValueToken('#'))
    postfix = append(postfix, CreateOperatorToken(AND))

	var stack Stack[int]
	var nodes []BSTNode

	for _, v := range postfix {
		node := CreateBSTNode(v)
		nodes = append(nodes, node)
		i := len(nodes) - 1

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

				node.left = i
				nodes[left].father = i
			}
		}

		stack.Push(i)
	}
}

func ConvertTreeToTable(tree *BST, nodes []*BSTNode) []*TableRow {
	table := []*TableRow{}

	for _, node := range nodes {
		if node.IsLeaf() {
			nullable := !(node.Val.IsValue() && node.Val.GetValue().HasValue())
			firstPos := []int{node.Key}
			lastPos := []int{node.Key}

			var simbol string
			if !nullable {
				simbol = string(node.Val.GetValue().GetValue())
			}

			row := TableRow{
				nullable: nullable, firstpos: firstPos, lastpos: lastPos, simbol: simbol,
			}

			node.extraProperties = row
			table = append(table, &row)
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
				if right.IsNullable() {
					lastPos = append(lastPos, left.extraProperties.lastpos...)
				}
				table = append(table, &TableRow{nullable: nullable, firstpos: firstPos, lastpos: lastPos})

				for i := range left.extraProperties.lastpos {
					node_i := nodes[i]
					if node_i.IsLeaf() {
						node_i.extraProperties.followpos = append(node_i.extraProperties.followpos, right.extraProperties.firstpos...)
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

				table = append(table, &TableRow{nullable: nullable, firstpos: firstPos, lastpos: lastPos})

			case ZERO_OR_MANY:
				left := tree.nodes[node.left]

				nullable := true
				firstPos := left.extraProperties.firstpos

				lastPos := left.extraProperties.lastpos

				table = append(table, &TableRow{nullable: nullable, firstpos: firstPos, lastpos: lastPos})

				for i := range lastPos {
					node_i := nodes[i]
					if node_i.IsLeaf() {
						node_i.extraProperties.followpos = append(node_i.extraProperties.followpos, firstPos...)
					}
				}
			}
		}
	}

	// sets Leaf i first
	// for i, v := range nodes {
	// 	newRow := new(TableRow)

	// 	if v.Val.value != nil && v.Val.value.HasValue() {
	// 		// nullable
	// 		newRow.nullable = false

	// 		// firstpos
	// 		newRow.firtspos = append(newRow.firtspos, i)

	// 		// lastpos
	// 		newRow.lastpos = append(newRow.lastpos, i)

	// 		// simbol
	// 		newRow.simbol = string(v.Val.value.GetValue())

	// 	} else if v.Val.value != nil && !v.Val.value.HasValue() {
	// 		newRow.nullable = true
	// 	} else if *v.Val.operator == AND {
	// 		var c1 int

	// 		if nodes[i-1].Val.operator != nil {
	// 			if *nodes[i-1].Val.operator == OR {
	// 				c1 = 4
	// 			} else {
	// 				c1 = 2
	// 			}
	// 		} else {
	// 			c1 = 2
	// 		}

	// 		//nullable
	// 		newRow.nullable = table[i-c1].nullable == true && table[i-1].nullable == true

	// 		// firstpos
	// 		if table[i-2].nullable == true {
	// 			union_slice := append(table[i-c1].firtspos, table[i-1].firtspos...)
	// 			newRow.firtspos = append(newRow.firtspos, union_slice...)
	// 		} else {
	// 			newRow.firtspos = append(newRow.firtspos, table[i-c1].firtspos...)
	// 		}

	// 		// lastpos
	// 		if table[i-1].nullable == true {
	// 			union_slice := append(table[i-c1].lastpos, table[i-1].lastpos...)
	// 			newRow.lastpos = append(newRow.lastpos, union_slice...)
	// 		} else {
	// 			newRow.lastpos = append(newRow.lastpos, table[i-1].lastpos...)
	// 		}

	// 		// followpos
	// 		for _, pos := range table[i-c1].lastpos {
	// 			table[pos].followpos = append(table[pos].followpos, table[i-1].firtspos...)
	// 		}

	// 	} else if *v.Val.operator == OR {
	// 		// nullable
	// 		newRow.nullable = table[i-2].nullable == true || table[i-1].nullable == true

	// 		// firtspos
	// 		union_slice := append(table[i-2].firtspos, table[i-1].firtspos...)
	// 		newRow.firtspos = append(newRow.firtspos, union_slice...)

	// 		// lastpos
	// 		union_slice = append(table[i-2].lastpos, table[i-1].lastpos...)
	// 		newRow.lastpos = append(newRow.lastpos, union_slice...)

	// 	} else {
	// 		// nullable
	// 		newRow.nullable = true

	// 		// firstpos
	// 		newRow.firtspos = append(newRow.firtspos, table[i-1].firtspos...)

	// 		// lastpos
	// 		newRow.lastpos = append(newRow.lastpos, table[i-1].lastpos...)

	// 		// followpos
	// 		for _, pos := range newRow.lastpos {
	// 			table[pos].followpos = append(table[pos].followpos, newRow.firtspos...)
	// 		}
	// 	}

	// 	table = append(table, newRow)
	// }

	return table
}
