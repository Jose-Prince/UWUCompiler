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

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please ONLY supply a lex file!\n")
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
    lexFileData = LexParser(lexFilePath)

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

    // Generates BST
    bst := new(lib.BST)
    bst.Insertion(postfix)

    // Creates tables with nodes from tree
    table := lib.ConvertTreeToTable(bst)
    
    afd := new(lib.AFD)
    afd = lib.ConvertFromTableToAFD(table)

	fmt.Println("The AFD is:", afd.String())

	// TODO: Generate AFD simulator (lexer)

	err := WriteLexFile(outputLexerFile, lexFileData, *afd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error ocurred writing final lexer file! %v", err)
		panic(err)
	}
}
