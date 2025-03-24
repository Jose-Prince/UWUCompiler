package main

import (
	"testing"
    "reflect"
    "github.com/Jose-Prince/UWULexer/lib"
)

func TestLexParser(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		want     LexFileData // Agrega el valor esperado de tipo LexFileData
	}{
		{
			name:     "Valid file with rules",
			filePath: "examples/example.yal",
			wantErr:  false,
			want: LexFileData{
				Header: "import myToken",
				Footer: "printf(\"hola\")",
				Rule: map[string]lib.DummyInfo{
					"rule1": {Code: "some code", Priority: 1},
				},
			}, // Define el valor esperado para un archivo válido
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Llama a LexParser y guarda el valor de retorno
			got := LexParser(tt.filePath)

			// Verifica si se produjo un error
			if (got.Header == "" && got.Footer == "" && len(got.Rule) == 0) != tt.wantErr {
				t.Errorf("LexParser() error = %v, wantErr %v", got, tt.wantErr)
			}

			// Compara el valor de retorno con el valor esperado
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LexParser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveRule(t *testing.T) {
	tests := []struct {
		name      string
		rule      string
		want      string
		rules     map[string]string
	}{
		{
			name:  "Valid simple rule",
			rule:  "{delim}",
			want:  "' \t \n'",
			rules: map[string]string{"delim": "' \t \n'"},
		},
		{
			name:  "Rule with nested resolution",
			rule:  "{delim} {other}",
			want:  "' \t \n' 'more'",
			rules: map[string]string{"delim": "' \t \n'", "other": "'more'"},
		},
		{
			name:  "Rule not found",
			rule:  "{undefined}",
			want:  "{undefined}",
			rules: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRule(tt.rule, tt.rules)
			if got != tt.want {
				t.Errorf("resolveRule() = %v, want %v", got, tt.want)
			}
		})
	}
}


