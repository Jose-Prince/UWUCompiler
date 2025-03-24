package main

import (
	"bufio"
	"os"

	"github.com/Jose-Prince/UWULexer/lib"
)

func WriteLexFile(filePath string, info LexFileData, afd lib.AFD) {
	f, err := os.Create(filePath)
	if err != nil {
		panic("Error creating output file!")
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	writer.WriteString(info.Header)
	// TODO: Write all other Lexer logic
	writer.WriteString(info.Footer)
}
