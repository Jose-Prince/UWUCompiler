{
const (
	LET int = iota
	IF
	WHILE
	ASSIGN
	PLUS
	MINUS
	MULT
	DIV
	GT
	LT
	EQ
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	ID
	NUMBER
)
}

let letter = ([a-z]|_|[A-Z])
let decimal_digit = [0-9]

let identifier = ({letter}({letter}|{decimal_digit})*)
let decimal_lit = (([1-9]{decimal_digit}*)|0)
let whitespace = ([ \t\r\n]+)

rule gettoken =
	{whitespace}	{}
	| 'let'				{ return LET }
	| 'if'				{ return IF }
	| 'while'			{ return WHILE }
	| ':='				{ return ASSIGN }
	| '\+'				{ return PLUS }
	| '-'					{ return MINUS }
	| '\*'				{ return MULT }
	| '/'					{ return DIV }
	| '>'					{ return GT }
	| '<'					{ return LT }
	| '=='					{ return EQ }
	| '\('					{ return LPAREN }
	| '\)'					{ return RPAREN }
	| '\{'					{ return LBRACE }
	| '\}'					{ return RBRACE }
	| {decimal_lit}	{ return NUMBER }
	| {identifier}	{ return ID }
