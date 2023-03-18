package goatlang

import (
	"testing"
)

func TestTreeSort(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"varConst", "x := y; const y = 42", `(const (, y) 42) (:= (, x) y)`},
		{"varType", "x := y; type T struct {}", `(type T struct) (:= (, x) y)`},
		{"varImport", `x := y; import "fmt"`, `(import fmt "fmt") (:= (, x) y)`},
		{"varMethod", "x := y; func (t *T) M() {}", `(method T M (func (arguments (t T)) returns block)) (:= (, x) y)`},
		{"typeMethod", "func (t *T) M() {}; type T struct {}", `(type T struct) (method T M (func (arguments (t T)) returns block))`},
		{"varFunc", "x := y; func f() {}", `(function f (func arguments returns block)) (:= (, x) y)`},
		{"typeFunc", "package a; func f() { x := &T{} } ; type T struct {}", `(package a) (type T struct) (function f (func arguments returns (block (:= (, x) (new T ;)))))`},
		{"varOrder", "x := y; y := z; z := q", `(:= (, x) y) (:= (, y) z) (:= (, z) q)`},
		{"unstableSortBug", `var m1 = 1;var m2 = 2;var m3 = 3;var m4 = 4;var m5 = 5;var m6 = 6;var m7 = 7;var m8 = 8;var m9 = 9;var m10 = 10;var m11 = 11;var m12 = 12;var m13 = 13;var m14 = 14;var m15 = 15;var m16 = 16;func main() {}`, `(function main (func arguments returns block)) (:= (, m1) 1) (:= (, m2) 2) (:= (, m3) 3) (:= (, m4) 4) (:= (, m5) 5) (:= (, m6) 6) (:= (, m7) 7) (:= (, m8) 8) (:= (, m9) 9) (:= (, m10) 10) (:= (, m11) 11) (:= (, m12) 12) (:= (, m13) 13) (:= (, m14) 14) (:= (, m15) 15) (:= (, m16) 16)`},
		// {"multiPackage", "package a; x := y; package b; y := z;", `(package a) (:= (, x) y) (package b) (:= (, y) z)`},
		// {"mergePackage", "package a; x := y; package a; const y = 42;", `(package a) (const (, y) 42) (:= (, x) y)`},
		// {"mixedPackage", "package a; x := y; package b; z := q; package a; const y = 42;", `(package a) (const (, y) 42) (:= (, x) y) (package b) (:= (, z) q)`},
		{"varInit", "func init() { x = 42 } ; var x = 43", `(:= (, x) 43) (init (func arguments returns (block (= (, x) 42))))`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tokens, err := tokenize(row.Name, row.In)
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}
			tree, err := parse(tokens)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			tree = treeSort(tree)
			s := tree.String()
			s = s[3 : len(s)-1]
			if s != row.Want {
				t.Fatalf("Parse\n got %v\nwant %v", s, row.Want)
			}
		})
	}
}
