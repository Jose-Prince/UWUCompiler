package main

import (
	"math/rand"
	"testing"
)

func FuzzInfixToPostfix(f *testing.F) {
	f.Add(int64(69420))
	f.Fuzz(func(t *testing.T, seed int64) {
		source := rand.NewSource(seed)
		random := rand.New(source)

		expected := generateExpectedInfix(random)
		infix := fromTokenStreamToInfix(expected)
		result := InfixToTokens(infix)

		compareTokensStreams(t, infix, expected, result)
	})
}
