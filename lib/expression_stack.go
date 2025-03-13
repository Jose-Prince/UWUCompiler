package lib

import "strings"

type ExprStack []*string

func (self *ExprStack) Push(expression string) {
	*self = append(*self, &expression)
}

func (self *ExprStack) IsEmpty() bool {
	ref := *self
	length := len(ref)

	return length == 0
}

func (self *ExprStack) Peek() Optional[string] {
	ref := *self
	length := len(ref)

	if length == 0 {
		return CreateNull[string]()
	}

	val := ref[length-1]
	return CreateValue(*val)

}

func (self *ExprStack) Pop() Optional[string] {
	ref := *self
	length := len(ref)

	if length == 0 {
		return CreateNull[string]()
	}

	val := ref[length-1]
	*self = ref[:length-1]

	return CreateValue(*val)
}

func (self *ExprStack) AppendTop(char string) {
	if self.IsEmpty() {
		self.Push(char)
	} else {
		for _, val := range *self {
			*val = strings.Join([]string{*val, char}, "")
		}
	}
}
