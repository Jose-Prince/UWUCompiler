{
	const (
		C int = iota
		D
	)
}

let whitespace = [ \t\n\r]+

rule gettoken =
	'c' {return C}
	| 'd' {return D}
	| {whitespace} {return UNRECOGNIZABLE}
