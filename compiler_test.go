package goatlang

import (
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"number", "42", "PUSH 42"},
		{"add", "2 + 3", "PUSH 2; PUSH 3; ADD"},
		{"addMul", "2 + 3 * 4", "PUSH 2; PUSH 3; PUSH 4; MUL; ADD"},
		{"mulAdd", "2 * 3 + 4", "PUSH 2; PUSH 3; MUL; PUSH 4; ADD"},
		{"parens", "(2+3) * 4", "PUSH 2; PUSH 3; ADD; PUSH 4; MUL"},
		{"twoNumbers", "4 2", "PUSH 4; PUSH 2"},
		{"assign", "a := 42 a", "PUSH 42; GLOBALSET a; GLOBALGET a"},
		{"string", `"test"`, `CONST "test"`},
		{"func", "func square(a int) int { return a*a }",
			`FUNC 1:1 1 4; TYPE int32; TYPE int32; LOCALGET $0; LOCALGET $0; MUL; RETURN 1; GLOBALFUNC square`},
		{"multiReturn", "func test() (int, int) { return 1,2 }",
			`FUNC 0:2 0 3; TYPE int32; TYPE int32; PUSH 1; PUSH 2; RETURN 2; GLOBALFUNC test`,
		},
		{"call", "test(2,4)", "PUSH 2; PUSH 4; GLOBALGET test; CALL 2 0"},
		{"callAssign", "x := test(42)", `PUSH 42; GLOBALGET test; CALL 1 1; GLOBALSET x`},
		{"localSet", "func test(a int) int { c := a }",
			`FUNC 1:1 2 2; TYPE int32; TYPE int32; LOCALGET $0; LOCALSET $1; GLOBALFUNC test`},
		{"localType", "func test(a int) { c := a + 1 }",
			`FUNC 1:0 2 4; TYPE int32; LOCALGET $0; PUSH 1; ADD; LOCALSET $1; GLOBALFUNC test`,
		},
		{"retTypes", "func test() int { return 1 } ; c := test() + 1",
			`FUNC 0:1 0 2; TYPE int32; PUSH 1; RETURN 1; GLOBALFUNC test; GLOBALGET test; CALL 0 1; PUSH 1; ADD; GLOBALSET c`,
		},
		{"if", "if true { 42 }", `CONST true; JUMPFALSE 1; PUSH 42`},
		{"ifElse", "if true { 42 } else { 43 }", `CONST true; JUMPFALSE 2; PUSH 42; JUMP 1; PUSH 43`},
		{"for", "for i:=0; i<10; i++ { println(i) }",
			`PUSH 0; LOCALSET $0; JUMP 6; LOCALGET $0; GLOBALGET builtin.println; CALL 1 0; LOCALGET $0; INCDEC 1; LOCALSET $0; LOCALGET $0; PUSH 10; LT; JUMPTRUE -10`},
		{"inc", "x := 0; x++", "PUSH 0; GLOBALSET x; GLOBALGET x; INCDEC 1; GLOBALSET x"},
		{"localDec", "func f(x int) { x-- }", `FUNC 1:0 1 3; TYPE int32; LOCALGET $0; INCDEC -1; LOCALSET $0; GLOBALFUNC f`},
		{"convert", "float64(42)", "PUSH 42; CONVERT float64"},
		{"var", "var x int", "GLOBALZERO x int32"},

		{"sliceInit", "x := []int{1,2,3}", "PUSH 1; PUSH 2; PUSH 3; NEWSLICE int32 3; GLOBALSET x"},
		{"map", `x := map[string]int{"a":1,"b":2}`, `CONST "a"; PUSH 1; CONST "b"; PUSH 2; NEWMAP string int32 4; GLOBALSET x`},
		{"range", "func f() { for k,v := range r { println(k,v) } }", `FUNC 0:0 3 7; GLOBALGET r; RANGE $0 4; LOCALGET $1; LOCALGET $2; GLOBALGET builtin.println; CALL 2 0; ITER $0 $1:$2 -5; GLOBALFUNC f`},
		{"get", "x := m[1]", `GLOBALGET m; PUSH 1; GET; GLOBALSET x`},
		{"getOk", "x, ok := m[1]", `GLOBALGET m; PUSH 1; GETOK; GLOBALSET ok; GLOBALSET x`},
		{"set", "m[1] = x", `GLOBALGET x; GLOBALGET m; PUSH 1; SET`},
		{"forEver", "for { x := 42 }", `PUSH 42; LOCALSET $0; JUMP -3`},
		{"forWhile", "for true { x := 42 }", `JUMP 2; PUSH 42; LOCALSET $0; CONST true; JUMPTRUE -4`},
		{"noSlotReuse", "func f() { x := 1 ; for { y := 2 } ; z := 3 }",
			`FUNC 0:0 3 7; PUSH 1; LOCALSET $0; PUSH 2; LOCALSET $1; JUMP -3; PUSH 3; LOCALSET $2; GLOBALFUNC f`},
		{"nameReuse", "func f() { for { x := 1 } ; x := 2 }",
			`FUNC 0:0 2 5; PUSH 1; LOCALSET $0; JUMP -3; PUSH 2; LOCALSET $1; GLOBALFUNC f`},
		{"nameShadow", "func f() { x := 1; for { x := 2 } ; x = 3 }",
			`FUNC 0:0 2 7; PUSH 1; LOCALSET $0; PUSH 2; LOCALSET $1; JUMP -3; PUSH 3; LOCALSET $0; GLOBALFUNC f`},
		{"noShadow", "func f() { x := 1; x, y := 2, 3; }",
			`FUNC 0:0 2 6; PUSH 1; LOCALSET $0; PUSH 2; PUSH 3; LOCALSET $1; LOCALSET $0; GLOBALFUNC f`},
		{"continue", `for { 42 continue 43 }`, `PUSH 42; JUMP 1; PUSH 43; JUMP -4`},
		{"break", `for { 42 break 43 }`, `PUSH 42; JUMP 2; PUSH 43; JUMP -4`},
		{"rangeContinue", `func f() { for x := range y { 42 continue 43 } }`,
			`FUNC 0:0 3 6; GLOBALGET y; RANGE $0 3; PUSH 42; JUMP 1; PUSH 43; ITER $0 $1:$2 -4; GLOBALFUNC f`},
		{"rangeBreak", `func f() { for x := range y { 42 break 43 } }`,
			`FUNC 0:0 3 6; GLOBALGET y; RANGE $0 3; PUSH 42; JUMP 2; PUSH 43; ITER $0 $1:$2 -4; GLOBALFUNC f`},
		{"globalOrder", `func f() { x = 3 } var x = 4`, `FUNC 0:0 0 2; PUSH 3; GLOBALSET x; GLOBALFUNC f; PUSH 4; GLOBALSET x`},
		{"packageVars", `package main; var x = 5`, `PUSH 5; GLOBALSET main.x`},
		{"packageFncs", `package main; func f() {}`, `FUNC 0:0 0 0; GLOBALFUNC main.f`},
		{"importVars", `package main; import "math"; var yum = math.Pi`, `GLOBALGET math.Pi; GLOBALSET main.yum`},
		{"importFncs", `package main; import ("strings" "math"); var big = math.Max(1.0,2.0);`,
			`CONST 1.0; CONST 2.0; GLOBALGET math.Max; CALL 2 1; GLOBALSET main.big`},
		{"importFncsFlat", `package main; import ("math/rand"); var v = rand.Float64();`,
			`GLOBALGET math/rand.Float64; CALL 0 1; GLOBALSET main.v`},
		{"varSliceMap", "var a []map[string]int", `GLOBALZERO a []map[string]int32`},
		{"negate", "x := -42; y := -x;", `PUSH -42; GLOBALSET x; GLOBALGET x; NEGATE; GLOBALSET y`},
		{"mapPlusEquals", `o["x"] += o["dx"]`, `GLOBALGET o; CONST "x"; GET; GLOBALGET o; CONST "dx"; GET; ADD; GLOBALGET o; CONST "x"; SET`},
		{"and", `p && q`, `GLOBALGET p; AND 1; GLOBALGET q`},
		{"or", `p || q`, `GLOBALGET p; OR 1; GLOBALGET q`},
		{"localSetArg", `func f(x int) { x = 42 }`, `FUNC 1:0 1 2; TYPE int32; PUSH 42; LOCALSET $0; GLOBALFUNC f`},
		{"not", `!true`, `CONST true; NOT`},
		{"slice", `a[2:4]`, `GLOBALGET a; PUSH 2; PUSH 4; SLICE`},
		{"getGet", `a[4][2] = 3`, `PUSH 3; GLOBALGET a; PUSH 4; GET; PUSH 2; SET`},
		{"typeStruct", `type T struct { X,Y,Z int; Name string }`,
			`GLOBALREF X; ZERO int32; GLOBALREF Y; ZERO int32; GLOBALREF Z; ZERO int32; GLOBALREF Name; ZERO string; STRUCT 8; GLOBALSTRUCT T`},
		{"typePackage", `package main; type T struct {}`, `STRUCT 0; GLOBALSTRUCT main.T`},
		{"newData", `package main; v := &T{ X:1, Y:2, Z:3, Name:"42"}`,
			`GLOBALREF X; PUSH 1; GLOBALREF Y; PUSH 2; GLOBALREF Z; PUSH 3; GLOBALREF Name; CONST "42"; NEWSTRUCT main.T 8; GLOBALSET main.v`},
		{"importNewData", `package ext; type T struct { X int } ; package main; import "ext"; v := &ext.T{ X:1 }`,
			`GLOBALREF X; ZERO int32; STRUCT 2; GLOBALSTRUCT ext.T; GLOBALREF X; PUSH 1; NEWSTRUCT ext.T 2; GLOBALSET main.v`},
		{"method", `package main; func (t *T) test(a, b int) int { return 40 + 2 }`,
			`FUNC 3:1 3 4; TYPE main.T; TYPE int32; TYPE int32; TYPE int32; PUSH 40; PUSH 2; ADD; RETURN 1; GLOBALGET main.T; SETMETHOD test`},
		{"getAttr", `x := obj.attr`, `GLOBALGET obj; GETATTR attr; GLOBALSET x`},
		{"setAttr", `obj.attr = 42`, `PUSH 42; GLOBALGET obj; SETATTR attr`},
		{"callAttr", `x := obj.attr()`, `GLOBALGET obj; GETATTR attr; CALL 0 1; GLOBALSET x`},
		{"localCallAttr", `func f(obj T) { obj.attr() }`, `FUNC 1:0 1 3; TYPE T; LOCALGET $0; GETATTR attr; CALL 0 0; GLOBALFUNC f`},
		{"structGetSet", `t.X *= t.X + 1`, `GLOBALGET t; GETATTR X; GLOBALGET t; GETATTR X; PUSH 1; ADD; MUL; GLOBALGET t; SETATTR X`},
		{"blankNew", `x, _, y := 1, 2, 3`, `PUSH 1; PUSH 2; PUSH 3; GLOBALSET y; POP; GLOBALSET x`},
		{"blankSet", `x, _, y = 1, 2, 3`, `PUSH 1; PUSH 2; PUSH 3; GLOBALSET y; POP; GLOBALSET x`},
		{"typeInterface", `type T interface { X(k int) int ; Z(k int) ; Y(k string)int }`, `STRUCT 0; GLOBALSTRUCT T; ZERO func; GLOBALGET T; SETMETHOD X; ZERO func; GLOBALGET T; SETMETHOD Z; ZERO func; GLOBALGET T; SETMETHOD Y`},
		{"typeInterfacePkg", `package main; type T interface { X(k int) int ; Z(k int) ; Y(k string)int }`, `STRUCT 0; GLOBALSTRUCT main.T; ZERO func; GLOBALGET main.T; SETMETHOD X; ZERO func; GLOBALGET main.T; SETMETHOD Z; ZERO func; GLOBALGET main.T; SETMETHOD Y`},
		{"globalRange", `for k,v := range m { res += k+v; }`,
			`GLOBALGET m; RANGE $0 6; GLOBALGET res; LOCALGET $1; LOCALGET $2; ADD; ADD; GLOBALSET res; ITER $0 $1:$2 -7`},
		{"appendEllipsis", `c := append(a, b...)`, `GLOBALGET a; GLOBALGET b; APPEND 2 1; GLOBALSET c`},
		{"nilEq", "a == nil", `GLOBALGET a; CONST nil; EQ`},
		{"nilAssign", "a = nil", `CONST nil; GLOBALSET a`},
		{"makeSlice", "make([]int, 42)", `PUSH 42; MAKE int32`},
		{"makeMap", "make(map[int]string)", `NEWMAP int32 string 0`},
		{"printlnJustNL", "println()", `GLOBALGET builtin.println; CALL 0 0`},
		{"forAppend", `for i:=0; i<3; i++ { res = append(res,&T{X:i}) }`,
			`PUSH 0; LOCALSET $0; JUMP 9; GLOBALGET res; GLOBALREF X; LOCALGET $0; NEWSTRUCT T 2; APPEND 2 0; GLOBALSET res; LOCALGET $0; INCDEC 1; LOCALSET $0; LOCALGET $0; PUSH 3; LT; JUMPTRUE -13`},
		{"varMap", `var x map[int]int`, `GLOBALZERO x map[int32]int32`},
		{"varObject", `type T struct {} ; var t *T`, `STRUCT 0; GLOBALSTRUCT T; GLOBALZERO t T`},
		{"varFunc", `var x func() int`, `GLOBALZERO x func`},
		{"switchValue", `switch v { case 1: 41; case 2: 42; default: 43 }`,
			`GLOBALGET v; LOCALSET $0; PUSH 1; LOCALGET $0; EQ; JUMPFALSE 2; PUSH 41; JUMP 7; PUSH 2; LOCALGET $0; EQ; JUMPFALSE 2; PUSH 42; JUMP 1; PUSH 43`},
		{"switchTrue", `switch { case false: 42; case true: 42; default: 43; }`,
			`CONST false; JUMPFALSE 2; PUSH 42; JUMP 5; CONST true; JUMPFALSE 2; PUSH 42; JUMP 1; PUSH 43`},
		{"shadowPackage", `import "fmt"; func f() int { fmt := &T{Sprint:42}; return fmt.Sprint }`,
			`FUNC 0:1 1 7; TYPE int32; GLOBALREF Sprint; PUSH 42; NEWSTRUCT T 2; LOCALSET $0; LOCALGET $0; GETATTR Sprint; RETURN 1; GLOBALFUNC f`},
		{"retMultiRet", `func g() (int,int) { return f() }`, `FUNC 0:2 0 3; TYPE int32; TYPE int32; GLOBALGET f; CALL 0 2; RETURN 2; GLOBALFUNC g`},
		{"forContinue", `for i:=0; i<1; i++ { continue ; 42 }`, `PUSH 0; LOCALSET $0; JUMP 5; JUMP 1; PUSH 42; LOCALGET $0; INCDEC 1; LOCALSET $0; LOCALGET $0; PUSH 1; LT; JUMPTRUE -9`},
		{"initFunc", "func init() { x = 42 }", `FUNC 0:0 0 2; PUSH 42; GLOBALSET x; CALL 0 0`},
		{"varErr", `var err error`, `GLOBALZERO err struct`},
		{"stack", `$[0]`, `GLOBALGET $; PUSH 0; GET`},
		{"intTypes", `type T byte; var t T; t = T(42)`, `GLOBALZERO t uint8; PUSH 42; CONVERT uint8; GLOBALSET t`},
		{"panic", `panic("hello")`, `CONST "hello"; PANIC`},
		{"copy", `copy(a,b)`, `GLOBALGET a; GLOBALGET b; COPY`},
		{"sliceArg", `func f(v []byte) {}`, `FUNC 1:0 1 0; TYPE []uint8; GLOBALFUNC f`},
		{"sliceRet", `func f() []byte {}`, `FUNC 0:1 0 0; TYPE []uint8; GLOBALFUNC f`},
		{"mapArg", `func f(v map[string]byte) {}`, `FUNC 1:0 1 0; TYPE map[string]uint8; GLOBALFUNC f`},
		{"mapRet", `func f() map[string]byte {}`, `FUNC 0:1 0 0; TYPE map[string]uint8; GLOBALFUNC f`},
		{"structArg", `func f(v *T) {}`, `FUNC 1:0 1 0; TYPE T; GLOBALFUNC f`},
		{"structRet", `func f() *T {}`, `FUNC 0:1 0 0; TYPE T; GLOBALFUNC f`},
		{"extStructArg", `package ext; type T struct { } ; package main; import "ext"; func f(v *ext.T) {}`, `STRUCT 0; GLOBALSTRUCT ext.T; FUNC 1:0 1 0; TYPE ext.T; GLOBALFUNC main.f`},
		{"extStructRet", `package ext; type T struct { } ; package main; import "ext"; func f() *ext.T {}`, `STRUCT 0; GLOBALSTRUCT ext.T; FUNC 0:1 0 0; TYPE ext.T; GLOBALFUNC main.f`},
		{"anyArg", `func f(v any) {}`, `FUNC 1:0 1 0; TYPE any; GLOBALFUNC f`},
		{"anyRet", `func f() any {}`, `FUNC 0:1 0 0; TYPE any; GLOBALFUNC f`},
		{"varByte", `var x byte = 42`, `PUSH 42; CAST uint8; GLOBALSET x`},
		{"intTypesArg", `type T byte; func f(t T) { }`, `FUNC 1:0 1 0; TYPE uint8; GLOBALFUNC f`},
		{"internalType", `package main; __type(42)`, `PUSH 42; GLOBALGET builtin.__type; CALL 1 0`},
		{"emptyMap", `x := map[string]int{}`, `NEWMAP string int32 0; GLOBALSET x`},
		{"callCallNegativeBug", `f(-1).m()`, `PUSH -1; GLOBALGET f; CALL 1 1; GETATTR m; CALL 0 0`},
		{"iotaCast", `const (val = code(-(iota + 1)))`, `PUSH 0; PUSH 1; ADD; NEGATE; GLOBALGET code; CALL 1 1; GLOBALSET val`},
		{"memberInc", `type T struct { N int }; func (t *T) F() { t.N ++ }`, `GLOBALREF N; ZERO int32; STRUCT 2; GLOBALSTRUCT T; FUNC 1:0 1 5; TYPE T; LOCALGET $0; GETATTR N; INCDEC 1; LOCALGET $0; SETATTR N; GLOBALGET T; SETMETHOD F`},
		{"constRefConst", `const (a = 40; b = a+2 )`, `PUSH 40; GLOBALSET a; GLOBALGET a; PUSH 2; ADD; GLOBALSET b`},
		{"constWeirdBug", `const a = -42; const b=-a`, `PUSH -42; GLOBALSET a; GLOBALGET a; NEGATE; GLOBALSET b`},
		{"funcEllipsis", `func f(a int, b ...int) { }`, `FUNC -2:0 2 0; TYPE int32; TYPE []int32; GLOBALFUNC f`},
		{"funcEllipsisCode", `func f(a ...int) int { return a[0]+a[1] }`, `FUNC -1:1 1 8; TYPE []int32; TYPE int32; LOCALGET $0; PUSH 0; GET; LOCALGET $0; PUSH 1; GET; ADD; RETURN 1; GLOBALFUNC f`},
		{"callVariadic", `f(a...)`, `GLOBALGET a; GLOBALGET f; CALLVARIADIC 1 0`},
		{"localStruct", `func f() any { type T struct {}; return &T{} }`, `FUNC 0:1 0 4; TYPE any; STRUCT 0; GLOBALSTRUCT f.T; NEWSTRUCT f.T 0; RETURN 1; GLOBALFUNC f`},
		{"localStructNotGlobal", `type T struct {}; func f() any { type T struct {}; return &T{} }`, `STRUCT 0; GLOBALSTRUCT T; FUNC 0:1 0 4; TYPE any; STRUCT 0; GLOBALSTRUCT f.T; NEWSTRUCT f.T 0; RETURN 1; GLOBALFUNC f`},
		{"localAliasNotGlobal", `type T string; func f() any { type T int; var x T = 42; return x }; var x T`,
			`FUNC 0:1 1 5; TYPE any; PUSH 42; CAST int32; LOCALSET $0; LOCALGET $0; RETURN 1; GLOBALFUNC f; GLOBALZERO x string`},
		{"caseBreak", `for { switch true { case true: break } ; v = 42 ; break }`, `CONST true; LOCALSET $0; CONST true; LOCALGET $0; EQ; JUMPFALSE 2; JUMP 1; JUMP 0; PUSH 42; GLOBALSET v; JUMP 1; JUMP -12`},
		{"caseBreakDefault", `switch true { case true: break; default: v = 0 }`, `CONST true; LOCALSET $0; CONST true; LOCALGET $0; EQ; JUMPFALSE 2; JUMP 3; JUMP 2; PUSH 0; GLOBALSET v`},
		{"typeAliasSlice", `type Matrix []float64 ; x := Matrix{1,2,3}`, `PUSH 1; PUSH 2; PUSH 3; NEWSLICE float64 3; GLOBALSET x`},
		{"typeAliasEmptySlice", `type Matrix []float64 ; x := Matrix{}`, `NEWSLICE float64 0; GLOBALSET x`},
		{"typeAliasMap", `type Matrix map[int]string ; x := Matrix{1:"test"}`, `PUSH 1; CONST "test"; NEWMAP int32 string 2; GLOBALSET x`},
		{"typeAliasEmptyMap", `type Matrix map[int]string ; x := Matrix{}`, `NEWMAP int32 string 0; GLOBALSET x`},
		{"typeAliasStruct", `type T struct { K int }; type Matrix T ; x := Matrix{K:42}`, `GLOBALREF K; ZERO int32; STRUCT 2; GLOBALSTRUCT T; GLOBALREF K; PUSH 42; NEWSTRUCT T 2; GLOBALSET x`},
		{"typeAliasEmptyStruct", `type T struct { K int }; type Matrix T ; x := Matrix{}`, `GLOBALREF K; ZERO int32; STRUCT 2; GLOBALSTRUCT T; NEWSTRUCT T 0; GLOBALSET x`},
		{"typeAliasMakeSlice", `type Matrix []float64 ; x = make(Matrix, 16)`, `PUSH 16; MAKE float64; GLOBALSET x`},
		{"typeAliasMakeAlias", `type T struct{}; type B T; type M []B; x = make(M,16)`, `STRUCT 0; GLOBALSTRUCT T; PUSH 16; MAKE T; GLOBALSET x`},
		{"castToAlias", `type T int; const C = T(42)`, `PUSH 42; CONVERT int32; GLOBALSET C`},
		{"argNamedType", `type typ struct {}; func f(typ *typ) { }`, `STRUCT 0; GLOBALSTRUCT typ; FUNC 1:0 1 0; TYPE typ; GLOBALFUNC f`},
		{"fieldNamedType", `type typ struct {}; func f() { var typ typ }`, `STRUCT 0; GLOBALSTRUCT typ; FUNC 0:0 1 1; LOCALZERO $0 typ; GLOBALFUNC f`},
		{"sliceMapStringStructInit", `type T struct { X int }; type B T; v := []map[string]B{{"x":{X:1}},{"y":{X:2}}}`,
			`GLOBALREF X; ZERO int32; STRUCT 2; GLOBALSTRUCT T; CONST "x"; GLOBALREF X; PUSH 1; NEWSTRUCT T 2; NEWMAP string T 2; CONST "y"; GLOBALREF X; PUSH 2; NEWSTRUCT T 2; NEWMAP string T 2; NEWSLICE map[string]T 2; GLOBALSET v`},
		{"sliceStructInit", `type T struct { X int}; s := []*T{&T{X:42}}`, `GLOBALREF X; ZERO int32; STRUCT 2; GLOBALSTRUCT T; GLOBALREF X; PUSH 42; NEWSTRUCT T 2; NEWSLICE T 1; GLOBALSET s`},
		{"sliceBug", `f := []*F{p2f([]*P{a}),}`, `GLOBALGET a; NEWSLICE P 1; GLOBALGET p2f; CALL 1 1; NEWSLICE F 1; GLOBALSET f`},
		{"varBlank", `var _ int`, ``},
		{"lambdaFunc", `func f() func() int { return func() int { return 42 }}`, `FUNC 0:1 0 5; TYPE func; FUNC 0:1 0 2; TYPE int32; PUSH 42; RETURN 1; RETURN 1; GLOBALFUNC f`},
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
			vm := New()
			g := vm.globals
			codes, _, err := compile(g, tree, false)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}
			var ts []string
			for _, s := range codes {
				ts = append(ts, s.String(g))
			}
			s := strings.Join(ts, "; ")
			assert(t, "Compile", s, row.Want)
		})
	}
}

func TestCompile_error(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Err  string
	}{
		{"undefined", `import "math"; math.Garbage()`, `undefined`},
		{"invalidType", `func f() { var T int; var x T }`, `invalid type: T`},
		{"untypedData", `v := []any{{}}`, `untyped data`},
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
			g := newLookup()
			_, _, err = compile(g, tree, false)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("error got %v want %v", err, row.Err)
			}
		})
	}
}

func TestCompile_unknown(t *testing.T) {
	want := "unknown symbol"
	tree := &token{Tokens: []*token{{Symbol: "unknown"}}}
	g := newLookup()
	_, _, err := compile(g, tree, false)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error got %v want %v", err, want)
	}
}

func TestInstruction_String(t *testing.T) {
	lookup := newLookup()
	idx := lookup.Index("idx")
	tests := []struct {
		In   instruction
		Want string
	}{
		{instruction{Code: codeFastGet, A: 42, B: reg(idx)}, "FASTGET $42 idx"},
		{instruction{Code: codeFastCall, A: reg(idx), B: 1, C: 2}, "FASTCALL idx 1 2"},
		{instruction{Code: codeFastCallAttr, A: 42, B: reg(idx), C: joinParams(4, 2)}, "FASTCALLATTR $42 idx 4:2"},
		{instruction{Code: codeLocalAdd, A: 42, B: 43}, "LOCALADD $42 $43"},
		{instruction{Code: codeLocalIncDec, A: 42, B: 1}, "LOCALINCDEC $42 1"},
		{instruction{Code: codeTODO, A: 1, B: 2, C: 3}, "TODO 1 2 3"},
	}

	for _, row := range tests {
		t.Run(row.In.Code.String(), func(t *testing.T) {
			s := row.In.String(lookup)
			if s != row.Want {
				t.Fatalf("Instruction.String\n got %v\nwant %v", s, row.Want)
			}
		})
	}
}
