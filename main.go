package main

import (
	"fmt"
	"io"
	"log"
	"os"

	regx "github.com/Jose-Prince/UWULexer/lib/regex"
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
	lexFileData := LexParser(lexFilePath)

	// Combine all regexes into a single regex
	infix := []regx.RX_Token{}
	i := 0
	keysCount := len(lexFileData.Rule)
	for regex, info := range lexFileData.Rule {
		// What we want is to have something like: ((<REGEX>).(DUMMY))
		infix = append(infix, regx.CreateOperatorToken(regx.LEFT_PAREN))

		infix = append(infix, regx.CreateOperatorToken(regx.LEFT_PAREN))
		regxToTokens := InfixToTokens(regex)
		infix = append(infix, regxToTokens...)
		infix = append(infix, regx.CreateOperatorToken(regx.RIGHT_PAREN))
		infix = append(infix, regx.CreateOperatorToken(regx.AND))
		infix = append(infix, regx.CreateDummyToken(info))

		infix = append(infix, regx.CreateOperatorToken(regx.RIGHT_PAREN))

		if i+1 < keysCount {
			infix = append(infix, regx.CreateOperatorToken(regx.OR))
		}
		i++
	}

	postfix := DEFAULT_ALPHABET.ToPostfix(&infix)
	fmt.Println("The Infix expression is:\n", regx.TokenStreamToString(infix))
	fmt.Println("The Postfix expression is:\n", regx.TokenStreamToString(postfix))

	// Generates BST
	bst := regx.ASTFromRegex(postfix)
	fmt.Println("The AST is:\n", bst.String())

	table := bst.ToTable()
	fmt.Println("The AST Table is:\n", table.String())

	afd := table.ToAFD()
	fmt.Println("The AFD is:", afd.String())

	err := WriteLexFile(outputLexerFile, lexFileData, afd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error ocurred writing final lexer file! %v", err)
		panic(err)
	}
}
