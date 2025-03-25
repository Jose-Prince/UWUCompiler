package main

import (
	"bufio"
	"os"

	"github.com/Jose-Prince/UWULexer/lib"
)

func WriteLexFile(filePath string, info LexFileData, afd lib.AFD) error {
	f, err := os.Create(filePath)
	if err != nil {
		panic("Error creating output file!")
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString(`
package lexer


// Lexer imports
import (
	"fmt"
	"os"
	"strconv"
	"strings"
)
	`)
	writer.WriteString(info.Header)
	writer.WriteString(`
const UNRECOGNIZABLE int = -1
const GIVE_NEXT int = -2

const CMD_HELP = `)
	writer.WriteRune('`')
	writer.WriteString(
		`Tokenizes a specified source file
Usage: lexer <source file>`)
	writer.WriteRune('`')
	writer.WriteString(`

type Optional[T any] struct {
	isValid bool
	value   T
}

func CreateValue[T any](val T) Optional[T] {
	return Optional[T]{value: val, isValid: true}
}

func CreateNull[T any]() Optional[T] {
	var defaultVal T
	return Optional[T]{value: defaultVal, isValid: false}
}

func (self Optional[T]) HasValue() bool {
	return self.isValid
}

func (self Optional[T]) GetValue() T {
	if !self.isValid {
		panic("Can't access not valid optional value!")
	} else {
		return self.value
	}
}

type Token struct {
	// When does this token start in the contents of the source file
	Start int
	// The type of the token that it found
	Type int
}

func (self *Token) String() string {
	b := strings.Builder{}
	b.WriteString("{ ")
	b.WriteString("Start = ")
	b.WriteString(strconv.Itoa(self.Start))
	b.WriteString(", Type = ")
	b.WriteString(strconv.Itoa(self.Type))
	b.WriteString(" }")
	return b.String()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Please supply only a source file as argument!\n")
		panic(CMD_HELP)
	}

	sourceFilePath := os.Args[1]
	sourceFileContent, err := os.ReadFile(sourceFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening the source file! %v", err)
	}

	// previousParsingResult := -1000
	// afdState := "0" // INITIAL AFD STATE!
	for i := 0; i < len(sourceFileContent); i++ {

		afdState := "0" // INITIAL AFD STATE!
		previousParsingResult := -1000
		j := 0
		for j = i; j < len(sourceFileContent); j++ {
			parsingResult := gettoken(&afdState, rune(sourceFileContent[j]))
			if parsingResult == UNRECOGNIZABLE {
				foundSomething := previousParsingResult != -1000
				if foundSomething {
					token := Token{Start: i, Type: previousParsingResult}
					fmt.Println(token.String())
					i = j - 1
					break
				} else {
					i = j
					break
				}
			} else if parsingResult != GIVE_NEXT {
				previousParsingResult = parsingResult
			}
		}
	}
}

func gettoken(state *string, input rune) int {
`)

	// TODO: Write AFD Logic into switch
	writer.WriteRune('}')
	writer.WriteString(info.Footer)

	return writer.Flush()
}
