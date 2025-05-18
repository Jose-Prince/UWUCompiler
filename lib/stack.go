package lib

type Stack[T any] []T

func NewStack[T any]() Stack[T] {
	return Stack[T]{}
}

func (self *Stack[T]) Empty() bool {
	return len(*self) == 0
}

func (self *Stack[T]) Peek() Optional[T] {
	idx := len(*self) - 1
	ref := *self

	if idx < 0 {
		return CreateNull[T]()
	}

	return CreateValue(ref[idx])
}

func (self *Stack[T]) Push(val T) *Stack[T] {
	*self = append(*self, val)
	return self
}

func (self *Stack[T]) Pop() Optional[T] {
	ref := *self
	length := len(ref)
	idx := length - 1

	if idx < 0 {
		return CreateNull[T]()
	}

	val := ref[idx]
	*self = ref[:idx]

	return CreateValue(val)
}
