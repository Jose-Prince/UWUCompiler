package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	regx "github.com/Jose-Prince/UWUCompiler/lib/regex"
)

type programParams struct {
	LexFilePath     string
	GrammarFilePath string
	OutGoPath       string
}

func parseProgramParams() programParams {
	params := programParams{}

	flag.StringVar(&params.LexFilePath, "lexPath", "tokens.lex", "The path to the .lex file with the tokens definitions!")
	flag.StringVar(&params.GrammarFilePath, "grammarPath", "grammar.yal", "The path to the .yal file with the grammar definition!")
	flag.StringVar(&params.OutGoPath, "outPath", "out.go", "The path where the generated code should be outputted!")

	flag.Parse()
	return params
}

func main() {
	// Disable loggin
	log.SetOutput(io.Discard)

	params := parseProgramParams()

	fmt.Println("Lex file to use:", params.LexFilePath)
	fmt.Println("Grammar file to use:", params.GrammarFilePath)
	fmt.Println("Output file will be:", params.OutGoPath)

	lexFileData, err := LexParser(params.LexFilePath)
	if err != nil {
		panic(err)
	}

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

	err = WriteLexFile(params.OutGoPath, lexFileData, afd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error ocurred writing final lexer file! %v", err)
		panic(err)
	}
}
