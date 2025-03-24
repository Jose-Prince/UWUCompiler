package lib

import (
	"fmt"
	"strings"
	"testing"
)

func printSideBySide(t *testing.T, markedIdx int, expected []RX_Token, result []RX_Token) {
	maxLength := max(len(expected), len(result))
	header1 := "EXPECTED:"
	header2 := "VALUE:"
	header3 := "IDX:"
	maxLeftLength := 20
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("\n%-*s%-*s%s\n", maxLeftLength, header1, maxLeftLength, header2, header3))

	for i := range maxLength {

		if i == markedIdx {
			b.WriteString("\033[31m")
		}

		if i < len(expected)-1 {
			elem := expected[i].String()
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, elem))
		} else {
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, "<N/A>"))
		}

		if i < len(result)-1 {
			elem := result[i].String()
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, elem))
		} else {
			b.WriteString(fmt.Sprintf("%-*s", maxLeftLength, "<N/A>"))
		}

		b.WriteString(fmt.Sprintf("%-*d", maxLeftLength, i))
		b.WriteString("\033[0m")
		b.WriteRune('\n')
	}

	t.Log(b.String())
}

func areEqual(t *testing.T, expected []RX_Token, result []RX_Token) bool {
	for i, elem := range expected {
		if i >= len(result) {
			t.Logf("EXPECTED (%s) != RESULT: (< No value on idx >) IDX: %d", elem.String(), i)
			printSideBySide(t, i, expected, result)
			return false
		}

		resultElem := result[i]

		if !elem.Equals(&resultElem) {
			t.Logf("EXPECTED (%s) != RESULT: (%s) IDX: %d", elem.String(), resultElem.String(), i)
			printSideBySide(t, i, expected, result)
			return false
		}
	}

	return true
}

func TestExpressionStack(t *testing.T) {
	stack := ExprStack{}
	levelA := []RX_Token{}
	levelA = append(levelA, CreateValueToken('a'))
	levelA = append(levelA, CreateValueToken('b'))

	levelB := []RX_Token{}
	levelB = append(levelB, CreateValueToken('c'))

	stack.Push(levelA)
	stack.Push(levelB)

	stack.AppendTop(CreateValueToken('0'))

	expectedA := []RX_Token{}
	expectedA = append(expectedA, CreateValueToken('a'))
	expectedA = append(expectedA, CreateValueToken('b'))
	expectedA = append(expectedA, CreateValueToken('0'))

	expectedB := []RX_Token{}
	expectedB = append(expectedB, CreateValueToken('c'))
	expectedB = append(expectedB, CreateValueToken('0'))

	if val := stack.Pop().GetValue(); !areEqual(t, expectedB, val) {
		t.Fatalf("Popped value is not the same!")
	}

	if val := stack.Pop().GetValue(); !areEqual(t, expectedA, val) {
		t.Fatalf("Popped value is not the same!")
	}

}
