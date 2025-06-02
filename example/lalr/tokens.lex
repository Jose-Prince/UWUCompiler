{
	const (
		C int = iota
		D
	)
}

let whitespace      = ([ \t\r\n]+)

rule gettoken =
	{whitespace} {return IGNORE}
	| 'c' {return C}
	| 'd' {return D}
