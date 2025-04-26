package lib

import "strings"

type ExprStackItem = []RX_Token

func ExprStackItem_ToString(self *ExprStackItem) string {
	b := strings.Builder{}

	b.WriteString("[ ")
	for i, elm := range *self {
		b.WriteString(elm.String())

		if i+1 < len(*self) {
			b.WriteString(", ")
		}
	}
	b.WriteString(" ]")

	return b.String()
}

type ExprStack []ExprStackItem

func (self *ExprStack) Push(tokens ExprStackItem) {
	if !self.IsEmpty() {
		for _, token := range tokens {
			self.AppendTop(token)
		}
	}

	*self = append(*self, tokens)
}

func (self *ExprStack) IsEmpty() bool {
	ref := *self
	length := len(ref)

	return length == 0
}

func (self *ExprStack) Peek() Optional[ExprStackItem] {
	ref := *self
	length := len(ref)

	if length == 0 {
		return CreateNull[ExprStackItem]()
	}

	val := ref[length-1]
	return CreateValue(val)

}

func (self *ExprStack) Pop() Optional[ExprStackItem] {
	length := len(*self)

	if length == 0 {
		return CreateNull[ExprStackItem]()
	}

	val := (*self)[length-1]
	(*self) = (*self)[:length-1]

	return CreateValue(val)
}

func (self *ExprStack) AppendTop(token RX_Token) {
	if self.IsEmpty() {
		self.Push([]RX_Token{token})
	}

	for i, val := range *self {
		(*self)[i] = append(val, token)
	}
}
