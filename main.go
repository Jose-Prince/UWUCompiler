package main

import (
	"fmt"
	"os"

	"github.com/Jose-Prince/UWULexer/lib"
)

func main() {
	if len(os.Args) != 2 {
		panic("Please ONLY supply a lex file!")
	}

	lexFilePath := os.Args[1]
	fmt.Println("Parsing file:", lexFilePath)

	// TODO: Parse lex file instead of using a default values
	var lexFileData LexFileData
	lexFileData = LexFileData{
		Rule: map[string]string{
			"a|b": "fmt.Println(\"Hello!\")",
			"db":  "fmt.Println(\"Goodbye!\")",
		},
	}

	// Combine all regexes into a single regex
	infix := []lib.RX_Token{}
	i := 0
	keysCount := len(lexFileData.Rule)
	for regex, code := range lexFileData.Rule {
		// What we want is to have something like: ((<REGEX>).(DUMMY))
		infix = append(infix, lib.CreateOperatorToken(lib.LEFT_PAREN))

		infix = append(infix, lib.CreateOperatorToken(lib.LEFT_PAREN))
		regexToTokens := InfixToTokens(regex)
		infix = append(infix, regexToTokens...)
		infix = append(infix, lib.CreateOperatorToken(lib.RIGHT_PAREN))
		infix = append(infix, lib.CreateOperatorToken(lib.AND))
		infix = append(infix, lib.CreateDummyToken(lib.DummyInfo{Code: code}))

		infix = append(infix, lib.CreateOperatorToken(lib.RIGHT_PAREN))

		if i+1 < keysCount {
			infix = append(infix, lib.CreateOperatorToken(lib.OR))
		}
		i++
	}

	// TODO: Regex to AFD
	postfix := DEFAULT_ALPHABET.ToPostfix(&infix)
	fmt.Println("The Postfix expression is:", postfix)

	// TODO: Generate AFD simulator (lexer)
}
