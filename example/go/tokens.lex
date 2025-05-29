{

// Token constants
const (
	// Special tokens
	COMMENT int = iota

	// Identifiers and literals
	IDENT
	INT
	FLOAT
	IMAG
	CHAR
	STRING

	// Keywords
	BREAK
	CASE
	CHAN
	CONST
	CONTINUE
	DEFAULT
	DEFER
	ELSE
	FALLTHROUGH
	FOR
	FUNC
	GO
	GOTO
	IF
	IMPORT
	INTERFACE
	MAP
	PACKAGE
	RANGE
	RETURN
	SELECT
	STRUCT
	SWITCH
	TYPE
	VAR

	// Operators and delimiters
	ADD      // +
	SUB      // -
	MUL      // *
	QUO      // /
	REM      // %
	AND      // &
	OR       // |
	XOR      // ^
	SHL      // <<
	SHR      // >>
	AND_NOT  // &^

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=
	AND_ASSIGN // &=
	OR_ASSIGN  // |=
	XOR_ASSIGN // ^=
	SHL_ASSIGN // <<=
	SHR_ASSIGN // >>=
	AND_NOT_ASSIGN // &^=

	LAND  // &&
	LOR   // ||
	ARROW // <-
	INC   // ++
	DEC   // --

	EQL    // ==
	LSS    // <
	GTR    // >
	ASSIGN // =
	NOT    // !

	NEQ      // !=
	LEQ      // <=
	GEQ      // >=
	DEFINE   // :=
	ELLIPSIS // ...

	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,
	PERIOD // .

	RPAREN    // )
	RBRACK    // ]
	RBRACE    // }
	SEMICOLON // ;
	COLON     // :
)
}

(* Character classes *)
let letter          = ([a-zA-Z_])
let decimal_digit   = ([0-9])
let octal_digit     = ([0-7])
let hex_digit       = ([0-9A-Fa-f])

(* Number patterns *)
let decimal_lit     = (([1-9][0-9]*)|0)
let octal_lit       = (0[0-7]*)
let hex_lit         = (0[xX][0-9A-Fa-f]+)
let float_lit = ({decimal_lit}\.{decimal_lit}?)
(* let float_lit       = {decimal_lit}\.{decimal_lit}?([eE][+-]?{decimal_lit})?|\.{decimal_lit}([eE][+-]?{decimal_lit})?|{decimal_lit}[eE][+-]?{decimal_lit} *)
let imaginary_lit   = ({decimal_lit}|{float_lit})i

(* String and character patterns *)
let unicode_char    = ([^\n\r\\"])
let char_lit        = ('{unicode_char}')
let string_lit      = (("{unicode_char}*")|(`[^`]*`))

(* Identifier pattern *)
let identifier      = ({letter}({letter}|{decimal_digit})*)

(* Whitespace and comments *)
let whitespace      = ([ \t\r\n]+)
let line_comment    = (//[^\n\r]*)
(* let block_comment   = /\*([^*]|\*+[^*/])*\*+/ *)

rule gettoken =
	{whitespace}        { /* ignore whitespace */ }
	| {line_comment}      { return COMMENT }
	| {block_comment}     { return COMMENT }

(* Keywords *)
	| 'break'             { return BREAK }
	| 'case'              { return CASE }
	| 'chan'              { return CHAN }
	| 'const'             { return CONST }
	| 'continue'          { return CONTINUE }
	| 'default'           { return DEFAULT }
	| 'defer'             { return DEFER }
	| 'else'              { return ELSE }
	| 'fallthrough'       { return FALLTHROUGH }
	| 'for'               { return FOR }
	| 'func'              { return FUNC }
	| 'go'                { return GO }
	| 'goto'              { return GOTO }
	| 'if'                { return IF }
	| 'import'            { return IMPORT }
	| 'interface'         { return INTERFACE }
	| 'map'               { return MAP }
	| 'package'           { return PACKAGE }
	| 'range'             { return RANGE }
	| 'return'            { return RETURN }
	| 'select'            { return SELECT }
	| 'struct'            { return STRUCT }
	| 'switch'            { return SWITCH }
	| 'type'              { return TYPE }
	| 'var'               { return VAR }

(* Operators - must be before single character operators *)
	| '\+='                { return ADD_ASSIGN }
	| '-='                { return SUB_ASSIGN }
	| '\*='                { return MUL_ASSIGN }
	| '/='                { return QUO_ASSIGN }
	| '%='                { return REM_ASSIGN }
	| '&='                { return AND_ASSIGN }
	| '\|='                { return OR_ASSIGN }
	| '^='                { return XOR_ASSIGN }
	| '<<='               { return SHL_ASSIGN }
	| '>>='               { return SHR_ASSIGN }
	| '&^='               { return AND_NOT_ASSIGN }
	| '<<'                { return SHL }
	| '>>'                { return SHR }
	| '&^'                { return AND_NOT }
	| '&&'                { return LAND }
	| '\|\|'                { return LOR }
	| '<-'                { return ARROW }
	| '\+\+'                { return INC }
	| '--'                { return DEC }
	| '=='                { return EQL }
	| '!='                { return NEQ }
	| '<='                { return LEQ }
	| '>='                { return GEQ }
	| ':='                { return DEFINE }
	| '...'               { return ELLIPSIS }

(* Single character operators and delimiters *)
	| '\+'                 { return ADD }
	| '-'                 { return SUB }
	| '\*'                 { return MUL }
	| '/'                 { return QUO }
	| '%'                 { return REM }
	| '&'                 { return AND }
	| '\|'                 { return OR }
	| '^'                 { return XOR }
	| '<'                 { return LSS }
	| '>'                 { return GTR }
	| '='                 { return ASSIGN }
	| '!'                 { return NOT }
	| '\('                 { return LPAREN }
	| '\['                 { return LBRACK }
	| '{'                 { return LBRACE }
	| ','                 { return COMMA }
	| '\.'                 { return PERIOD }
	| '\)'                 { return RPAREN }
	| '\]'                 { return RBRACK }
	| '}'                 { return RBRACE }
	| ';'                 { return SEMICOLON }
	| ':'                 { return COLON }

(* Literals *)
	| {imaginary_lit}     { return IMAG }
	| {float_lit}         { return FLOAT }
	| {decimal_lit}       { return INT }
	| {octal_lit}         { return INT }
	| {hex_lit}           { return INT }
	| {char_lit}          { return CHAR }
	| {string_lit}        { return STRING }

(* Identifiers *)
	| {identifier}        { return IDENT }
