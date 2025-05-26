package main

import (
	"reflect"
	"strings"
	"testing"

	reg "github.com/Jose-Prince/UWUCompiler/lib/regex"
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
			filePath: "example/simple/grammar.yal",
			wantErr:  false,
			want: LexFileData{
				Header: `const (
	TOKENA int = iota
	TOKENB
)`,
				Rule: map[string]reg.DummyInfo{
					"[ \\t\\n]": {Regex: "[ \\t\\n]", Code: "", Priority: 1},
					"abc":       {Regex: "abc", Code: "return TOKENA", Priority: 2},
					"(abc)|c":   {Regex: "(abc)|c", Code: "return TOKENB", Priority: 3},
				},
			}, // Define el valor esperado para un archivo válido
		},
		{
			name:     "Valid file for Python",
			filePath: "example/python/grammar.yal",
			wantErr:  false,
			want: LexFileData{
				Header: "import myToken\n",
				Footer: "",
				Rule: map[string]reg.DummyInfo{
					"[0-9]+": {Regex: "[0-9]+", Code: "return NUMBER", Priority: 1},
					"\\+":    {Regex: "\\+", Code: "return PLUS", Priority: 2},
					"-":      {Regex: "-", Code: "return MINUS", Priority: 3},
					"\\*":    {Regex: "\\*", Code: "return TIMES", Priority: 4},
					"/":      {Regex: "/", Code: "return DIV", Priority: 5},
					"\\(":    {Regex: "\\(", Code: "return LPAREN", Priority: 6},
					"\\)":    {Regex: "\\)", Code: "return RPAREN", Priority: 7},
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

			// Compara Header y Footer
			if strings.EqualFold(tt.want.Header, got.Header) {
				t.Errorf("LexParser() Header = %v, want %v", got.Header, tt.want.Header)
			}

			if got.Footer != tt.want.Footer {
				t.Errorf("LexParser() Footer = %v, want %v", got.Footer, tt.want.Footer)
			}

			// Compara las reglas ignorando el orden del mapa
			if len(got.Rule) != len(tt.want.Rule) {
				t.Errorf("LexParser() Rule length = %d, want %d", len(got.Rule), len(tt.want.Rule))
			} else {
				for key, wantValue := range tt.want.Rule {
					gotValue, exists := got.Rule[key]
					if !exists {
						t.Errorf("LexParser() missing key in Rule: %v", key)
					} else if !reflect.DeepEqual(gotValue, wantValue) {
						t.Errorf("LexParser() Rule[%v] = %v, want %v", key, gotValue, wantValue)
					}
				}
			}
		})
	}
}

func TestResolveRule(t *testing.T) {
	tests := []struct {
		name  string
		rule  string
		want  string
		rules map[string]string
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
