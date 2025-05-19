package lib

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

type Set[T comparable] map[T]struct{}

// Prints a set as a string.
//
// Since a Set is internally represented as a map, the keys will be unordered!
func (self Set[T]) String() string {
	b := strings.Builder{}
	b.WriteString("[ ")

	for k := range self {
		b.WriteString(fmt.Sprint(k))
		b.WriteString(", ")
	}

	b.WriteString("]")
	return b.String()
}

func GetValuesStable[T cmp.Ordered](self Set[T]) []T {
	values := make([]T, 0, len(self))
	for k := range self {
		values = append(values, k)
	}

	slices.Sort(values)

	return values
}

// Prints a set as a string with it's keys on the same order every time!
func StableSetString[T cmp.Ordered](self Set[T]) string {
	b := strings.Builder{}
	b.WriteString("[ ")

	values := GetValuesStable(self)
	for _, k := range values {
		b.WriteString(fmt.Sprint(k))
		b.WriteString(", ")
	}

	b.WriteString("]")
	return b.String()
}

// Checks if self is equal to other.
//
// Equal means that all items in self are contained in other and no more items are in other.
func (self *Set[T]) Equals(other *Set[T]) bool {
	if len(*self) != len(*other) {
		return false
	}

	for k := range *self {
		if !other.Contains(k) {
			return false
		}
	}

	return true
}

// Checks if an element exists on the set.
//
// Returns True if the value is contained in the set.
func (self *Set[T]) Contains(val T) bool {
	_, alreadyAdded := (*self)[val]

	return alreadyAdded
}

// Adds an element to the set.
//
// Returns True if the element is new to the set,
// false otherwise.
func (self *Set[T]) Add(val T) bool {
	ref := *self
	_, alreadyAdded := ref[val]

	if !alreadyAdded {
		ref[val] = struct{}{}
	}

	return !alreadyAdded
}

// Adds all values from other into self.
func (self *Set[T]) Merge(other *Set[T]) {
	for val := range *other {
		self.Add(val)
	}
}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

func (self *Set[T]) IsEmpty() bool {
	return len(*self) == 0
}

func (self *Set[T]) Clear() {
	for k := range *self {
		delete(*self, k)
	}
}

func (self *Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(*self))
	for k := range *self {
		slice = append(slice, k)
	}

	return slice
}
