package lib

import (
	"fmt"
	"reflect"
)

type Optional[T any] struct {
	isValid bool
	value   T
}

func OptionalEquals[T comparable](a Optional[T], b Optional[T]) bool {
	if a.HasValue() && b.HasValue() {
		return a.GetValue() == b.GetValue()
	}

	return !a.HasValue() && !b.HasValue()
}

func CreateValue[T any](val T) Optional[T] {
	return Optional[T]{value: val, isValid: true}
}

func CreateNull[T any]() Optional[T] {
	var defaultVal T
	return Optional[T]{value: defaultVal, isValid: false}
}

func (self Optional[T]) HasValue() bool {
	return self.isValid
}

func (self Optional[T]) GetValue() T {
	if !self.isValid {
		panic("Can't access not valid optional value!")
	} else {
		return self.value
	}
}

func (self *Optional[T]) Equals(other *Optional[T]) bool {
	noneAreValid := !self.isValid && !other.isValid
	if noneAreValid {
		return true
	}

	return self.isValid == other.isValid && reflect.DeepEqual(self.value, other.value)
}

func (self *Optional[T]) ToString() string {
	if !self.isValid {
		return "nil"
	}

	return fmt.Sprintf("%#v", self.GetValue())
}
