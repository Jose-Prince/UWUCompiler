{
	const (
		C int = iota
		D
	)
}

rule gettoken =
	'c' {return C}
	| 'd' {return D}
