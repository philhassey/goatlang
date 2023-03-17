package goatlang

import (
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"number", `42`, `42`},
		{"numbers", `4  2`, `4 2`},
		{"string", `"42" `, `"42"`},
		{"rawString", "`x`", "`x`"},
		{"char", ` 'c'`, `'c'`},
		{"name", ` test `, `test`},
		{"add", `2+3`, `2 + 3`},
		{"addAssign", `a+=3`, `a += 3`},
		{"shiftAssign", `a<<=3`, `a <<= 3`},
		{"comment", `hello//world`, `hello`},
		{"ellipsis", `a...`, `a ...`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tokens, err := tokenize(row.Name, row.In)
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}
			var ts []string
			for _, s := range tokens[:len(tokens)-1] {
				if strings.Contains(s.String(), " ") {
					t.Fatalf("token contains spaces %v", s)
				}
				ts = append(ts, s.String())
			}
			s := strings.Join(ts, " ")
			if s != row.Want {
				t.Fatalf("Tokenize got %v want %v", s, row.Want)
			}
		})
	}
}

func TestTokenize_error(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Err  string
	}{
		{"invalidString", `"`, `literal not terminated`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			_, err := tokenize(row.Name, row.In)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("Tokenize error got %v want %v", err, row.Err)
			}
		})
	}
}

func TestTokenize_symbols(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"int", `42`, "(int)"},
		{"float", `42.0`, "(float)"},
		{"string", `"42"`, "(string)"},
		{"rawString", "`42`", "(string)"},
		{"char", `'c'`, "(char)"},
		{"+", `+`, "+"},
		{"+=", `+=`, "+="},
		{"<<=", `<<=`, "<<="},
		{"name", `hello`, "(name)"},
		{"func", `func`, "func"},
		{"true", `true`, "true"},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tokens, err := tokenize(row.Name, row.In)
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}
			if tokens[0].Symbol != row.Want {
				t.Fatalf("Symbol got %v want %v", tokens[0].Symbol, row.Want)
			}
		})
	}
}

func TestToken_must(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		defer expectPanic(t, "invalid syntax")
		x := &token{Text: "panic"}
		_ = x.Int()
	})
	t.Run("Hex", func(t *testing.T) {
		defer expectPanic(t, "invalid syntax")
		x := &token{Text: "0xpanic"}
		_ = x.Int()
	})
	t.Run("Octal", func(t *testing.T) {
		defer expectPanic(t, "invalid syntax")
		x := &token{Text: "079"}
		_ = x.Int()
	})
	t.Run("Float64", func(t *testing.T) {
		defer expectPanic(t, "invalid syntax")
		x := &token{Text: "panic"}
		_ = x.Float64()
	})
	t.Run("Unquote", func(t *testing.T) {
		defer expectPanic(t, "invalid syntax")
		x := &token{Text: "panic"}
		_ = x.Unquote()
	})
}
