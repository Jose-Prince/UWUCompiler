package lib

import (
	"testing"
)

func TestSet(t *testing.T) {
	set := NewSet[string]()

	if !set.Add("Hello") {
		t.Fatalf("Set returned false when adding `Hello` even though it isn't added!")
	}
	if set.Add("Hello") {
		t.Fatalf("Set returned true when adding `Hello` even though it was already added!")
	}
}

func TestEquality(t *testing.T) {
	setA := NewSet[string]()
	setA.Add("Hello1")
	setA.Add("Hello2")

	setB := NewSet[string]()
	setB.Add("Hello1")

	if setA.Equals(&setB) {
		t.Fatalf("Set A SHOULD NOT equal Set B yet!\nSetA: %s\nSetB: %s", setA, setB)
	}

	setB.Add("Hello2")
	if !setA.Equals(&setB) {
		t.Fatalf("Set A SHOULD equal Set B!\nSetA: %s\nSetB: %s", setA, setB)
	}
}
