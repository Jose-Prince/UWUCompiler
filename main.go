package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Jose-Prince/UWULexer/lib"
)

const CMD_HELP string = `
UWULexer is a Lexer generator similar to Yalex and Lex.
Usage:
	UWULexer <lexfile> [outputFileToWriteCodeTo]

Example:
	- To produce the Lexer code inside a file named MyLexer.go
	UWULexer ./input.lex MyLexer.go
`

func main() {
	// Disable loggin
	log.SetOutput(io.Discard)

	if len(os.Args) != 2 || len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Please ONLY supply a lex file!")
		panic(CMD_HELP)
	}

	lexFilePath := os.Args[1]
	fmt.Println("Parsing file:", lexFilePath)

	outputLexerFile := "out_lexer.go"
	if len(os.Args) == 3 {
		outputLexerFile = os.Args[2]
	}
	fmt.Println("Output file will be:", outputLexerFile)

	// TODO: Parse lex file instead of using a default values
	var lexFileData LexFileData
	lexFileData = LexFileData{
		Rule: map[string]lib.DummyInfo{
			"abc":     lib.DummyInfo{Code: "fmt.Println(\"Hello!\")", Priority: 1, Regex: "abc"},
			"(abc)|c": lib.DummyInfo{Code: "fmt.Println(\"Goodbye!\")", Priority: 2, Regex: "(abc)|c"},
		},
	}

	// Combine all regexes into a single regex
	infix := []lib.RX_Token{}
	i := 0
	keysCount := len(lexFileData.Rule)
	for regex, info := range lexFileData.Rule {
		// What we want is to have something like: ((<REGEX>).(DUMMY))
		infix = append(infix, lib.CreateOperatorToken(lib.LEFT_PAREN))

		infix = append(infix, lib.CreateOperatorToken(lib.LEFT_PAREN))
		regexToTokens := InfixToTokens(regex)
		infix = append(infix, regexToTokens...)
		infix = append(infix, lib.CreateOperatorToken(lib.RIGHT_PAREN))
		infix = append(infix, lib.CreateOperatorToken(lib.AND))
		infix = append(infix, lib.CreateDummyToken(info))

		infix = append(infix, lib.CreateOperatorToken(lib.RIGHT_PAREN))

		if i+1 < keysCount {
			infix = append(infix, lib.CreateOperatorToken(lib.OR))
		}
		i++
	}

	// TODO: Regex to AFD
	postfix := DEFAULT_ALPHABET.ToPostfix(&infix)
	fmt.Println("The Postfix expression is:", lib.TokenStreamToString(postfix))
	// ...do other conversions
	afd := lib.AFD{InitialState: "0",
		AcceptanceStates: lib.Set[lib.AFDState]{"f": struct{}{}},
		Transitions: map[lib.AFDState]map[lib.AlphabetInput]lib.AFDState{
			"0": {
				lib.CreateValueToken('a'): "1",
				lib.CreateValueToken('c'): "4",
			},
			"1": {
				lib.CreateValueToken('b'): "2",
			},
			"2": {
				lib.CreateValueToken('c'): "3",
			},
			"3": {
				lib.CreateDummyToken(lexFileData.Rule["abc"]):     "f",
				lib.CreateDummyToken(lexFileData.Rule["(abc)|c"]): "f",
			},
			"4": {
				lib.CreateDummyToken(lexFileData.Rule["(abc)|c"]): "f",
			},
		}}

	fmt.Println("The AFD is:", afd.String())

	// TODO: Generate AFD simulator (lexer)

	WriteLexFile(outputLexerFile, lexFileData, afd)
}
