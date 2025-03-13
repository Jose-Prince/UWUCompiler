package lib

import "testing"

func TestExpressionStack(t *testing.T) {
	stack := ExprStack{}
	stack.Push("ab")
	stack.Push("c")

	stack.AppendTop("Hola")

	if val := stack.Pop().GetValue(); val != "cHola" {
		t.Fatalf("Popped value is not the same! `%s` != `cHola`", val)
	}

	if val := stack.Pop().GetValue(); val != "abHola" {
		t.Fatalf("Popped value is not the same! `%s` != `abHola`", val)
	}

}
