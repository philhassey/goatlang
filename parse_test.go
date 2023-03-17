package goatlang

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"number", "42", "42"},
		{"add", "2 + 3", "(+ 2 3)"},
		{"addMul", "2 + 3 * 4", "(+ 2 (* 3 4))"},
		{"mulAdd", "2 * 3 + 4", "(+ (* 2 3) 4)"},
		{"parens", "(2+3) * 4", "(* (+ 2 3) 4)"},
		{"twoNumbers", "4 2", "4 2"},
		{"assign", "a := 42", "(:= (, a) 42)"},
		{"stmt", "2 ; 3", "2 3"},
		{"string", `"test"`, `"test"`},
		{"func", "func square(a int) int { return a*a }", `(function square (func (arguments (a int)) (returns int) (block (return (* a a)))))`},
		{"multiReturn", "func test() (int, int) { return 1,2 }",
			`(function test (func arguments (returns int int) (block (return 1 2))))`,
		},
		{"multiAssign", "a, b := 40, 2", `(:= (, a b) (, 40 2))`},
		{"multiAssignCall", "a, b := test()", `(:= (, a b) (call test arguments 2))`},
		{"call", "test(4,2)", `(call test (arguments 4 2) 0)`},
		{"callAssign", "x := test(42)", `(:= (, x) (call test (arguments 42) 1))`},
		{"retUse", "c := test() + 1", `(:= (, c) (+ (call test arguments 1) 1))`},
		{"if", "if true { 42 }", `(if ~ true (block 42))`},
		{"ifElse", "if true { 42 } else { 43 }", `(if ~ true (block 42) (block 43))`},
		{"ifElseIf", "if true { 42 } else if false { 43 }", `(if ~ true (block 42) (if ~ false (block 43)))`},
		{"ifElseIfElse", "if true { 42 } else if false { 43 } else { 44 }", `(if ~ true (block 42) (if ~ false (block 43) (block 44)))`},
		{"for", "for i:=0; i<10; i+=1 { println(i) }", `(for (:= (, i) 0) (< i 10) (+= i 1) (block (call println (arguments i) 0)))`},
		{"package", "package main", `(package main)`},
		{"inc", "x++", `(++ x)`},
		{"localDec", "func f(x int) { x-- }", `(function f (func (arguments (x int)) returns (block (-- x))))`},
		{"funcArgs", "func test(a, b int) {}", `(function test (func (arguments (a int) (b int)) returns block))`},

		{"const", "const x = 42", `(const (, x) 42)`},
		{"multiConst", "const a, b = 40, 2", `(const (, a b) (, 40 2))`},
		{"multiConstList", "const ( a = 42 ; b, c = 6, 7 )", `(const (, a b c) (, 42 6 7))`},
		{"constList", "const ( a = 42 )", `(const (, a) (, 42))`},

		{"var", "var x = 42", `(:= (, x) 42)`},
		{"multiVar", "var a, b = 40, 2", `(:= (, a b) (, 40 2))`},
		{"varList", "var ( a = 42 ; b, c = 6, 7 )", `(block (:= (, a) 42) (:= (, b c) (, 6 7)))`},
		{"varType", "var a int", `(var (, (a int)) ,)`},
		{"varTypes", "var a, b int", `(var (, (a int) (b int)) ,)`},
		{"varSlice", "var a []byte", `(var (, (a ([] byte))) ,)`},
		{"varMap", "var a map[string]int", `(var (, (a (map string int))) ,)`},
		{"varSliceMap", "var a []map[string]int", `(var (, (a ([] (map string int)))) ,)`},
		{"varMapSlice", "var a map[string][]int", `(var (, (a (map string ([] int)))) ,)`},
		{"varNamedType", "var a T", `(var (, (a T)) ,)`},
		{"varAssign", "var a = T", `(:= (, a) T)`},
		{"varNamedPtrType", "var a *P", `(var (, (a P)) ,)`},

		{"slice", "[]int{1,2,3}", `([] int (, 1 2 3))`},
		{"map", `map[string]int{"a":1,"b":2}`, `(map string int (, "a" 1 "b" 2))`},
		{"index", "x[5]", `(index x 5)`},
		{"indexOk", "x, ok := y[5]", `(:= (, x ok) (indexOk y 5))`},
		{"set", "m[1] = x", `(= (, (index m 1)) x)`},
		{"indexSlice", "x[2:3]", `(slice x 2 3)`},
		{"leftSlice", "x[2:]", `(slice x 2 -1)`},
		{"rightSlice", "x[:3]", `(slice x 0 3)`},
		{"append", "append(x, 5, 6)", `(call append (arguments x 5 6) 0)`},

		{"range", "for k,v := range r { x }", `(range k v r (block x))`},
		{"keyRange", "for k := range r { x }", `(range k _ r (block x))`},
		{"noRange", "for range r { x }", `(range _ _ r (block x))`},

		{"forEver", "for { x := 42 }", `(for ~ ~ ~ (block (:= (, x) 42)))`},
		{"forWhile", "for true { x := 42 }", `(for ~ true ~ (block (:= (, x) 42)))`},
		{"ifStmt", "if x:=5; x>4 { true }", `(if (:= (, x) 5) (> x 4) (block true))`},

		{"importOne", `import "math"`, `(import math "math")`},
		{"importMany", `import ("math" "strings")`, `(import math "math" strings "strings")`},
		{"importAlias", `import (alias "math")`, `(import alias "math")`},
		{"importLong", `import "math/rand"`, `(import rand "math/rand")`},

		{"dot", `a.b.c`, `(. (. a b) c)`},
		{"negateNumber", `-42`, `-42`},
		{"negateVar", `-a`, `(negate a)`},
		{"negate", "x := -42; y := -x;", `(:= (, x) -42) (:= (, y) (negate x))`},
		{"mapPlusEquals", `o["x"] += o["dx"]`, `(+= (index o "x") (index o "dx"))`},
		{"and", `p && q`, `(&& p q)`},
		{"or", `p || q`, `(|| p q)`},
		{"andOr", `p || q && r`, `(|| p (&& q r))`},
		{"not", `!x`, `(! x)`},
		{"byteConvert", `[]byte("*")`, `(call ([] byte ,) (arguments "*") 0)`},
		{"typeStruct", `type T struct { X,Y,Z int; Name string }`, `(type T (struct X int Y int Z int Name string))`},
		{"newData1", `v := &T{ X:1, Y:2, Z:3, Name:"42"}`, `(:= (, v) (new T (, X 1 Y 2 Z 3 Name "42")))`},
		{"newData2", `import "ext"; v := &ext.T{ X:1, Y:2, Z:3, Name:"42"}`, `(import ext "ext") (:= (, v) (new (. ext T) (, X 1 Y 2 Z 3 Name "42")))`},
		{"method1", `func (t *T) test(a, b int) {}`, `(method T test (func (arguments (t T) (a int) (b int)) returns block))`},
		{"method2", `func (t T) test(a, b int) {}`, `(method T test (func (arguments (t T) (a int) (b int)) returns block))`},
		{"methodCall", `obj.test(1,2,3)`, `(call (. obj test) (arguments 1 2 3) 0)`},
		{"extraCommaSlice", `[]int{1,2,3,}`, `([] int (, 1 2 3))`},
		{"extraCommaMap", `map[int]int{1:1,2:2,3:3,}`, `(map int int (, 1 1 2 2 3 3))`},
		{"extraCommaArgs", `f(1,2,3,)`, `(call f (arguments 1 2 3) 0)`},
		{"extraCommaStruct", `&T{a:1,b:2,c:3,}`, `(new T (, a 1 b 2 c 3))`},
		{"typeInterface", `type T interface { X(k int) int ; Z(k int) ; Y(k string)int }`, `(type T (interface X (arguments (k int)) (returns int) Z (arguments (k int)) returns Y (arguments (k string)) (returns int)))`},
		{"typeInterfaceBlanks", `type T interface { X(int) }`, `(type T (interface X (arguments (_ int)) returns))`},
		{"typeInterfaceNamed", `type T interface { X(t int) ; Y() }`, `(type T (interface X (arguments (t int)) returns Y arguments returns))`},
		{"typeInterfaceNL", `type State interface {
			loop(t float64)
			draw()
		}`, `(type State (interface loop (arguments (t float64)) returns draw arguments returns))`},
		{"appendEllipsis", `append(x, y...)`, `(call append (arguments x (... y)) 0)`},
		{"nilEq", "a == nil", `(== a nil)`},
		{"nilAssign", "a = nil", `(= (, a) nil)`},
		{"makeSlice", "make([]int, 42)", `(make ([] int) 42)`},
		{"makeMap", "make(map[int]string)", `(make (map int string))`},
		{"structSliceInit", "x := []*T{&T{X:1}}", `(:= (, x) ([] T (, (new T (, X 1)))))`},
		{"structSliceAutoInit", "x := []*T{{X:1}}", `(:= (, x) ([] T (, (new T (, X 1)))))`},
		{"manyStructSliceAutoInit", "x := []*T{{X:1},{X:2}}", `(:= (, x) ([] T (, (new T (, X 1)) (new T (, X 2)))))`},
		{"forAppend", `for i:=0; i<3; i++ { res = append(res,&T{X:i}) }`, `(for (:= (, i) 0) (< i 3) (++ i) (block (= (, res) (call append (arguments res (new T (, X i))) 1))))`},
		{"structMapAutoInit", "x := map[int]*T{42:{X:1}}", `(:= (, x) (map int T (, 42 (new T (, X 1)))))`},
		{"varFunc", `var x func() int`, `(var (, (x (func arguments (returns int)))) ,)`},
		{"switchValue", `switch v { case 1: 41; case 2: 42; default: 43 }`, `(switch v (, (case 1 (block 41)) (case 2 (block 42))) (default (block 43)))`},
		{"switchTrue", `switch { case false: 42; case true: 42; default: 43; }`, `(switch ~ (, (case false (block 42)) (case true (block 42))) (default (block 43)))`},
		{"switchTrueNoDefault", `switch { case 2: 42;}`, `(switch ~ (, (case 2 (block 42))) ~)`},
		{"stackIndex", "$0", `(index $ 0)`},
		{"stackRegular", "$[0]", `(index $ 0)`},
		{"pkgTypeVar", `import "ext"; var x *ext.T`, `(import ext "ext") (var (, (x (. ext T))) ,)`},
		{"pkgTypeArg", `import "ext"; func f(x *ext.T) { }`, `(import ext "ext") (function f (func (arguments (x (. ext T))) returns block))`},
		{"2dSliceAutoInit", `x := [][]int{{1,2,3},{4,5,6}}`, `(:= (, x) ([] ([] int) (, ([] int (, 1 2 3)) ([] int (, 4 5 6)))))`},
		{"3dSliceAutoInit", `x := [][][]int{{{1,2,3},{2,3,4}},{{3,4,5}}}`, `(:= (, x) ([] ([] ([] int)) (, ([] ([] int) (, ([] int (, 1 2 3)) ([] int (, 2 3 4)))) ([] ([] int) (, ([] int (, 3 4 5)))))))`},
		{"mapSliceAutoInit", `x := map[int][]int{1:{1,2,3},2:{2,3,4}}`, `(:= (, x) (map int ([] int) (, 1 ([] int (, 1 2 3)) 2 ([] int (, 2 3 4)))))`},
		{"sliceMapAutoInit", `x := []map[int]int{{1:2},{3:4}}`, `(:= (, x) ([] (map int int) (, (map int int (, 1 2)) (map int int (, 3 4)))))`},
		{"sliceStructAutoInit", `x := []*T{{X:42},{X:43}}`, `(:= (, x) ([] T (, (new T (, X 42)) (new T (, X 43)))))`},
		{"sliceTypeAutoInit", `x := []T{42,43,44}`, `(:= (, x) ([] T (, 42 43 44)))`},
		{"retMultiRet", `func g() (int,int) { return f() }`, `(function g (func arguments (returns int int) (block (return (call f arguments -1)))))`},
		{"initFunc", "func init() { x = 42 }", `(init (func arguments returns (block (= (, x) 42))))`},
		{"errReturn", `func f() error { return errors.New("42") }`, `(function f (func arguments (returns error) (block (return (call (. errors New) (arguments "42") -1)))))`},
		{"funcParam", `func f(g func(), h func(string) int) { }`, `(function f (func (arguments (g (func arguments returns)) (h (func (arguments (_ string)) (returns int)))) returns block))`},
		{"fancyFuncParam", `func Run(tick func(), event func(*Event)) { }`, `(function Run (func (arguments (tick (func arguments returns)) (event (func (arguments (_ Event)) returns))) returns block))`},
		{"fmtPrintlnHack", `fmt.Println(4,2)`, `(call (. fmt Println) (arguments 4 2) 0)`},
		{"fmtPrintHack", `fmt.Print(4,2)`, `(call (. fmt Print) (arguments 4 2) 0)`},
		{"intTypes", `type T int; var t T; t = T(42)`, `(type T int) (var (, (t T)) ,) (= (, t) (call T (arguments 42) 1))`},
		{"constAuto", `const (A = 42;B;C)`, `(const (, A B C) (, 42 42 42))`},
		{"constIota", `const (A = iota + 42; B; C)`, `(const (, A B C) (, (+ 0 42) (+ 1 42) (+ 2 42)))`},
		{"constIotaNL", "const (\nA = iota\nB\nC\n)", `(const (, A B C) (, 0 1 2))`},
		{"panic", `panic("hello")`, `(call panic (arguments "hello") 0)`},
		{"copy", `copy(a,b)`, `(call copy (arguments a b) 0)`},
		{"convertByte", `var x = byte(42)`, `(:= (, x) (call byte (arguments 42) 1))`},
		{"varByte", `var x byte`, `(var (, (x byte)) ,)`},
		{"varByteAssign", `var x byte = 42`, `(var (, (x byte)) 42)`},
		{"makeSliceStruct", `make([]*T,n)`, `(make ([] T) n)`},
		{"callCallNegativeBug", `f(-1).m()`, `(call (. (call f (arguments -1) 1) m) arguments 0)`},
		{"structPointer", `x := &T{K:42}`, `(:= (, x) (new T (, K 42)))`},
		{"extStructPointer", `x := &ext.T{K:42}`, `(:= (, x) (new (. ext T) (, K 42)))`},
		{"!structNonPointer", `x := T{K:42}`, `(:= (, x) (new T (, K 42)))`},
		{"!extStructNonPointer", `x := ext.T{K:42}`, `(:= (, x) (new (. ext T) (, K 42)))`},
		{"sliceStructPointer", `[]*T{{K:42}}`, `([] T (, (new T (, K 42))))`},
		{"!sliceStructNonPointer", `[]T{{K:42}}`, `([] T (, (new T (, K 42))))`},
		{"sliceExtStructPointer", `[]*ext.T{{K:42}}`, `([] (. ext T) (, (new (. ext T) (, K 42))))`},
		{"!sliceExtStructNonPointer", `[]ext.T{{K:42}}`, `([] (. ext T) (, (new (. ext T) (, K 42))))`},
		{"mapStructPointer", `map[int]*T{0:{K:42}}`, `(map int T (, 0 (new T (, K 42))))`},
		{"!mapStructNonPointer", `map[int]T{0:{K:42}}`, `(map int T (, 0 (new T (, K 42))))`},
		{"mapExtStructPointer", `map[int]*ext.T{0:{K:42}}`, `(map int (. ext T) (, 0 (new (. ext T) (, K 42))))`},
		{"!mapExtStructNonPointer", `map[int]ext.T{0:{K:42}}`, `(map int (. ext T) (, 0 (new (. ext T) (, K 42))))`},
		{"!ignoreRefs", `&x`, `x`},
		{"!ignoreDerefs", `*x`, `x`},
		{"negateStructConfusion", `if X < -128 { 42 }`, `(if ~ (< X -128) (block 42))`},
		{"notStructConfusion", `if !visible { 42 }`, `(if ~ (! visible) (block 42))`},
		{"nlFuncSignature", `func test(
			x int) bool { 42 }`, `(function test (func (arguments (x int)) (returns bool) (block 42)))`},
		{"caseReturn", `switch x { case A: return x ; case B: }`, `(switch x (, (case A (block (return x))) (case B block)) ~)`},
		{"iotaCast", `const (codeBreak = code(-(iota + 1)))`, `(const (, codeBreak) (, (call code (arguments (negate (+ 0 1))) 1)))`},
		{"memberInc", `type T struct { N int }; func (t *T) F() { t.N ++ }`, `(type T (struct N int)) (method T F (func (arguments (t T)) returns (block (++ (. t N)))))`},
		{"constRefConst", `const (a = 40; b = a+2 )`, `(const (, a b) (, 40 (+ a 2)))`},
		{"constWeirdBug", `const a = -42; const b=-a`, `(const (, a) -42) (const (, b) (negate a))`},
		{"funcEllipsis", `func f(a int, b ...int) { }`, `(function f (func (arguments (a int) (b (... int))) returns block))`},
		{"caseBreak", `for { switch true { case true: break } ; v = 42 ; break }`, `(for ~ ~ ~ (block (switch true (, (case true (block break))) ~) (= (, v) 42) break))`},
		{"fancyTypeAlias", `type Matrix []float64 ; x := Matrix{1,2,3}`, `(type Matrix ([] float64)) (:= (, x) (new Matrix (, 1 2 3)))`},
		{"typeAliasExact", `type A = T`, `(type A T)`},
		{"typeAliasMakeSlice", `x = make(Matrix, 16)`, `(= (, x) (make Matrix 16))`},
		// {"typeAliasInit", `x := A{{X:42}}`, ``},
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
			s := tree.String()
			s = s[3 : len(s)-1]
			if s != row.Want {
				t.Fatalf("Parse\n got %v\nwant %v", s, row.Want)
			}
		})
	}
}

func TestParse_error(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Err  string
	}{
		{"nilNud", "++", "null nud"},
		{"advance", "f(z}", "advance got"},
		{"nullLed", "1 ! 2", "null led"},
		{"type", "[]else", "type: unexpected symbol"},
		{"autoItem", "[]int{{1}}", "unexpected item type: int"},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tokens, err := tokenize(row.Name, row.In)
			if err != nil {
				t.Fatalf("Tokenize error: %v", err)
			}
			_, err = parse(tokens)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("Parse error got %v want %v", err, row.Err)
			}
		})
	}
}

func TestParse_unknown(t *testing.T) {
	want := "unknown symbol"
	tokens := []*token{{Symbol: "unknown"}, {Symbol: "(eof)"}}
	_, err := parse(tokens)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("Parse error got %v want %v", err, want)
	}
}
