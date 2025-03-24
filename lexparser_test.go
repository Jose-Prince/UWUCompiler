package main

import (
	"testing"
)

func TestLexParser(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
	}{
		{
			name:     "Valid file with rules",
			filePath: "testfile1.yalex",
			wantErr:  false,
		},
		{
			name:     "Invalid file path",
			filePath: "invalid.yalex",
			wantErr:  true,
		},
		{
			name:     "Empty file",
			filePath: "emptyfile.yalex",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simula el comportamiento de LexParser
			err := LexParser(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LexParser() error = %v, wantErr %v", err, tt.wantErr)
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


