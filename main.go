package main

import (
	"flag"
	"fmt"
	"log"
	// "os"

	// "github.com/Jose-Prince/UWUCompiler/lib"
	"github.com/Jose-Prince/UWUCompiler/lib/grammar"
	regx "github.com/Jose-Prince/UWUCompiler/lib/regex"
	// parsertypes "github.com/Jose-Prince/UWUCompiler/parserTypes"
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

type CompilerFileInfo struct {
	LexInfo      LexFileData
	LexAFD       regx.AFD
	ParsingTable grammar.ParsingTable
}

func main() {
	params := parseProgramParams()

	fmt.Println("Lex file to use:", params.LexFilePath)
	fmt.Println("Grammar file to use:", params.GrammarFilePath)
	fmt.Println("Output file will be:", params.OutGoPath)

	lexFileData, err := LexParser(params.LexFilePath)
	if err != nil {
		panic(err)
	}
	fmt.Println("The lex file data is:", lexFileData.String())

	// Combine all regexes into a single regex
	infix := []regx.RX_Token{}
	i := 0
	ruleCount := len(lexFileData.Rules)
	for _, rule := range lexFileData.Rules {
		fmt.Printf("Converting %s...\n", rule.Regex)
		// What we want is to have something like: ((<REGEX>).(DUMMY))
		infix = append(infix, regx.CreateOperatorToken(regx.LEFT_PAREN))

		infix = append(infix, regx.CreateOperatorToken(regx.LEFT_PAREN))
		regxToTokens := DEFAULT_ALPHABET.InfixToTokens(rule.Regex)
		infix = append(infix, regxToTokens...)
		infix = append(infix, regx.CreateOperatorToken(regx.RIGHT_PAREN))
		infix = append(infix, regx.CreateOperatorToken(regx.AND))
		infix = append(infix, regx.CreateDummyToken(rule.Info))

		infix = append(infix, regx.CreateOperatorToken(regx.RIGHT_PAREN))

		if i+1 < ruleCount {
			infix = append(infix, regx.CreateOperatorToken(regx.OR))
		}
		i++
	}
	fmt.Println("The Infix expression is:\n", regx.TokenStreamToString(infix))

	postfix := DEFAULT_ALPHABET.ToPostfix(&infix)
	fmt.Println("The Postfix expression is:\n", regx.TokenStreamToString(postfix))

	// Generates BST
	bst := regx.ASTFromRegex(postfix)
	fmt.Println("The AST is:\n", bst.String())

	table := bst.ToTable()
	fmt.Println("The AST Table is:\n", table.String())

	afd := table.ToAFD()
	fmt.Println("The AFD is:", afd.String())

	// TODO Parse yal fil
	g, err := grammar.ParseYalFile(params.GrammarFilePath)
	if err != nil {
		log.Panicf("Failed to parse grammar file: %s", err)
	}

	// parsingTable := grammar.ParsingTable{
	// 	InitialNodeId: "0",
	// 	Original: grammar.Grammar{
	// 		InitialSimbol: grammar.NewNonTerminalToken("S"),
	// 		Rules: []grammar.GrammarRule{
	// 			{Head: grammar.NewNonTerminalToken("S"), Production: []grammar.GrammarToken{grammar.NewNonTerminalToken("C"), grammar.NewNonTerminalToken("C")}},
	// 			{Head: grammar.NewNonTerminalToken("C"), Production: []grammar.GrammarToken{grammar.NewTerminalToken("c"), grammar.NewNonTerminalToken("C")}},
	// 			{Head: grammar.NewNonTerminalToken("C"), Production: []grammar.GrammarToken{grammar.NewTerminalToken("d")}},
	// 		},
	// 		TokenIds: map[grammar.GrammarToken]parsertypes.GrammarToken{
	// 			grammar.NewTerminalToken("c"):    0,
	// 			grammar.NewTerminalToken("d"):    1,
	// 			grammar.NewNonTerminalToken("S"): 2,
	// 			grammar.NewNonTerminalToken("C"): 3,
	// 			grammar.NewEndToken():            4,
	// 		},
	// 		Terminals: lib.Set[grammar.GrammarToken]{
	// 			grammar.NewTerminalToken("c"): struct{}{},
	// 			grammar.NewTerminalToken("d"): struct{}{},
	// 			grammar.NewEndToken():         struct{}{},
	// 		},
	// 		NonTerminals: lib.Set[grammar.GrammarToken]{
	// 			grammar.NewNonTerminalToken("C"): struct{}{},
	// 			grammar.NewNonTerminalToken("S"): struct{}{},
	// 		},
	// 	},
	// 	GoToTable: map[grammar.AFDNodeId]map[grammar.GrammarToken]grammar.AFDNodeId{
	// 		"0": map[grammar.GrammarToken]grammar.AFDNodeId{
	// 			grammar.NewNonTerminalToken("S"): "1",
	// 			grammar.NewNonTerminalToken("C"): "2",
	// 		},
	// 		"2": map[grammar.GrammarToken]grammar.AFDNodeId{
	// 			grammar.NewNonTerminalToken("C"): "5",
	// 		},
	// 		"36": map[grammar.GrammarToken]grammar.AFDNodeId{
	// 			grammar.NewNonTerminalToken("C"): "89",
	// 		},
	// 	},
	// 	ActionTable: map[grammar.AFDNodeId]map[grammar.GrammarToken]grammar.Action{
	// 		"0": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewTerminalToken("c"): grammar.NewShiftAction("36"),
	// 			grammar.NewTerminalToken("d"): grammar.NewShiftAction("47"),
	// 		},
	// 		"1": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewEndToken(): grammar.NewAcceptAction(),
	// 		},
	// 		"2": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewTerminalToken("c"): grammar.NewShiftAction("36"),
	// 			grammar.NewTerminalToken("d"): grammar.NewShiftAction("47"),
	// 		},
	// 		"36": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewTerminalToken("c"): grammar.NewShiftAction("36"),
	// 			grammar.NewTerminalToken("d"): grammar.NewShiftAction("47"),
	// 		},
	// 		"47": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewTerminalToken("c"): grammar.NewReduceAction(2),
	// 			grammar.NewTerminalToken("d"): grammar.NewReduceAction(2),
	// 			grammar.NewEndToken():         grammar.NewReduceAction(2),
	// 		},
	// 		"5": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewEndToken(): grammar.NewReduceAction(0),
	// 		},
	// 		"89": map[grammar.GrammarToken]grammar.Action{
	// 			grammar.NewTerminalToken("c"): grammar.NewReduceAction(1),
	// 			grammar.NewTerminalToken("d"): grammar.NewReduceAction(1),
	// 			grammar.NewEndToken():         grammar.NewReduceAction(1),
	// 		},
	// 	},
	// }

	initialRule := grammar.GrammarRule{Head: grammar.NewNonTerminalToken("S'"), Production: []grammar.GrammarToken{g.InitialSimbol}}
	// extendedGrammar := grammar.Grammar{
	// 	InitialSimbol: g.InitialSimbol,
	// 	Rules:         append(g.Rules),
	// 	Terminals:     g.Terminals,
	// 	NonTerminals:  g.NonTerminals,
	// 	TokenIds:      g.TokenIds,
	// }

	fmt.Println("Creating automata...")
	// grammar.InitializeAutomata(initialRule, g)
	lalr := grammar.InitializeAutomata(initialRule, g)

	fmt.Println("Simplifying states...")
	lalr.SimplifyStates()

	//
	// fmt.Println("Generating LARLR HTML...")
	// err = GenerateHTML(lalr, "lalr_automata.html")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("LALR HTML generated")
	//
	fmt.Println("Generating Parsing table from LALR AFD...")
	parsingTable := lalr.GenerateParsingTable(&g)

	info := CompilerFileInfo{
		LexInfo:      lexFileData,
		LexAFD:       afd,
		ParsingTable: parsingTable,
	}
	fmt.Println("Writing final compiler source code...")
	err = WriteCompilerFile(params.OutGoPath, &info)
	if err != nil {
		log.Fatalf("An error ocurred writing final lexer file! %v", err)
	}
}
