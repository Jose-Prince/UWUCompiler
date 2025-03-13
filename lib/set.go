package lib

type Set[T comparable] map[T]struct{}

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

func NewSet[T comparable]() Set[T] {
    return make(Set[T])
}
