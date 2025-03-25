package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	TOKENA int = iota
	TOKENB
)

const UNRECOGNIZABLE int = -1
const GIVE_NEXT int = -2

const CMD_HELP = `
Tokenizes a specified source file
Usage: lexer <source file>
`

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

type FileReader struct {
	peekedRune Optional[rune]
	reader     *bufio.Reader
}

func NewFileReader(reader *bufio.Reader) FileReader {
	return FileReader{
		reader:     reader,
		peekedRune: CreateNull[rune](),
	}
}

func (self *FileReader) ReadRune() (rune, error) {
	if self.peekedRune.HasValue() {
		val := self.peekedRune.GetValue()
		self.peekedRune = CreateNull[rune]()
		return val, nil
	}

	rune, _, err := self.reader.ReadRune()
	return rune, err
}

func (self *FileReader) PeekRune() (Optional[rune], error) {
	if self.peekedRune.HasValue() {
		return self.peekedRune, nil
	}

	rune, _, err := self.reader.ReadRune()
	if err != nil {
		self.peekedRune = CreateValue(rune)
	}
	return self.peekedRune, err
}

type Token struct {
	// When does this token start in the contents of the source file
	Start int
	// The type of the token that it found
	Type int
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
					fmt.Println(Token{Start: i, Type: previousParsingResult})
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
	// Usa el AFD para identificar los patrones de los tokens.
	// En caso hace match ejecuta el código dentro de {} en cada definición de token

	switch *state {
	case "0":
		switch input {
		case 'a':
			*state = "1"
			return GIVE_NEXT
		case 'c':
			*state = "4"
			return TOKENB
		default:
			return UNRECOGNIZABLE
		}

	case "1":
		switch input {
		case 'b':
			*state = "2"
			return GIVE_NEXT
		default:
			return UNRECOGNIZABLE
		}

	case "2":
		switch input {
		case 'c':
			*state = "3"
			return TOKENA
		default:
			return UNRECOGNIZABLE
		}
	default:
		return UNRECOGNIZABLE
	}
}

// Comentario extra!
