package main

type LexFileData struct {
	Header string
	Footer string
	// The key represents the regex expanded to only have valid regex items
	// The value is the go code to execute when the regex matches
	Rule map[string]string
}

// Example Lex file:
// {
//     package main
// }
//
// let delim = [' ''\t''\n']
// let ws = {delim}+
// let letra = ['A'-'Z''a'-'z']
// let digito = ['0'-'9']
// let id = {letra}({letra}|{digito})*
// let numero = {digito}+(\.{digito}+)?
// let literal = \"({letra}|{digito})*\"
// let operator = '+'|'-'|'*'|'/'
// let oprel = '=='|'<='|'>='|'<'|'>'
//
// rule gettoken =
// 	  {ws}	        { continue } (* Ignora white spaces, tabs y nueva línea)
// 	| {id}          { return ID }
// 	| {numero}      { return NUM }
//     | {literal}     { return LIT }
//     | {operator}    { return OP }
//     | {oprel}       { return OPREL }
//     | '='           { return ASSIGN }
//     | '('           { return LPAREN }
//     | ')'           { return RPAREN }
//     | '{'           { return LBRACE }
//     | '}'           { return RBRACE }
//     | eof           { return nil }
//
// {
//     fmt.Println("Footer!")
// }

// El LexFileData del archivo de arriba sería:
// {
// 	Header: "package main"
// 	Footer: "fmt.Println(\"Footer!\")"
// 	Rule: {
// 		"[\t\n ]+": "continue",
// 		"[A-Za-z]([A-Za-z]|[0-9])*": "return ID"
//		...etc etc que hueva escribir todos xD
// 	}
// }
