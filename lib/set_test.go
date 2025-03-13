package lib

import "testing"

func TestSet(t *testing.T) {
	set := Set[string]{}

	if !set.Add("Hello") {
		t.Fatalf("Set returned false when adding `Hello` even though it isn't added!")
	}
	if set.Add("Hello") {
		t.Fatalf("Set returned true when adding `Hello` even though it was already added!")
	}
}
