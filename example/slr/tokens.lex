{
const (
	NUMBER int = iota
	PLUS
	MULT
	LPAREN
	RPAREN
)
}

let decimal_digit = [0-9]
let decimal_lit = (([1-9]{decimal_digit}*)|0)
let whitespace = ([ \t\r\n]+)

rule gettoken =
	{whitespace}			{ return IGNORE }
	| {decimal_lit}		{ return NUMBER }
	| '\('						{ return LPAREN }
	| '\)'						{ return RPAREN }
	| '\+'						{ return PLUS }
	| '\*'						{ return MULT }
