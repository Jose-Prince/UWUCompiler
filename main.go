package main

import (
	"fmt"
	"io"
	"log"
	"os"

	regx "github.com/Jose-Prince/UWUCompiler/lib/regex"
)

const CMD_HELP string = `
UWUCompiler is a compiler generator similar to Yalex and Lex.
Usage:
	UWUCompiler <lexfile> <grammarfile> [outputFileToWriteCodeTo]

Example:
	- To produce the compiler code inside a file named MyCompiler.go
	UWUCompiler ./input.lex ./grammar.yal MyCompiler.go
`

func main() {
	// Disable loggin
	log.SetOutput(io.Discard)

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Invalid use!\n")
		panic(CMD_HELP)
	}

	lexFilePath := os.Args[1]
	fmt.Println("Lex file to use:", lexFilePath)

	grammarFilePath := os.Args[2]
	fmt.Println("Grammar file to use:", grammarFilePath)

	outputLexerFile := "out_compiler.go"
	if len(os.Args) == 3 {
		outputLexerFile = os.Args[2]
	}
	fmt.Println("Output file will be:", outputLexerFile)

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
