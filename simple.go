package main

type TokenType = int

const (
	EOL TokenType = iota
)

// AFD completo con la combinación de las regexes de la rule: [ \t\r] y \n
// const AFD = AFD{}

func gettoken(contents string) {
	// Usa el AFD para identificar los patrones de los tokens.
	// En caso hace match ejecuta el código dentro de {} en cada definición de token
}

// Comentario extra!
