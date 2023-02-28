package goatlang

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

func TestVM(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"number", "42", "42"},
		{"add", "2 + 3", "5"},
		{"addMul", "2 + 3 * 4", "14"},
		{"mulAdd", "2 * 3 + 4", "10"},
		{"parens", "(2+3) * 4", "20"},
		{"twoNumbers", "4 2", "4 2"},
		{"assign", "a := 2; b := 3; b - a", "1"},
		{"multiAssign", "a,b,c := 2,3,5 ; a + b * c", `17`},
		{"string", `"test"`, `test`},
		{"addStrings", `"hello" + "world"`, `helloworld`},
		{"func", "func square(a int) int { return a*a } ; x := square(16); x", `256`},
		{"argTypes", "func test(a int) int { c := a ; return c }", ``},
		{"fancyFunc", "func fancy(a int, b int) int { c := (a + 1) * (b + 1) ; return c * c } ; x := fancy(2,3) ; x", `144`},
		{"ifElse", "if true { 42 } else { 43 }", `42`},
		{"ifFalse", "if false { 42 }", ``},
		{"multiReturn", "func test() (int, int, int) { return 2,3,5 } ; a,b,c := test() ; a + b * c", `17`},
		{"for", "x := 0 ; for i:=0; i<5; i++ { x += i } ; x", `10`},
		{"inc", "x := 42; x++; x", `43`},
		{"intDiv", "1 / 2", `0`},
		{"convert", "float64(1) / 2", `0.5`},
		{"var", "var x int; x", `0`},
		{"sliceRange", "x := []int{2,3,5} ; func f() int { res := 0 ; for k,v := range x { res += v } return res }; i := f(); i", `10`},
		{"mapRange", `x := map[string]int{"a":2,"b":3} ; func f() (string, int) { rk, rv := "", 0; for k,v := range x { rk += k; rv += v } return rk, rv } ; a,b := f(); a; b`, `ab 5`},
		{"index", `x := map[int]int{40:2}; y := x[40]; y`, `2`},
		{"indexOk", `x := map[int]int{40:2}; y, ok := x[40]; y; ok`, `2 true`},
		{"indexNotOk", `x := map[int]int{}; y, ok := x[40]; y; ok`, `0 false`},
		{"indexNot", `x := map[int]int{}; y := x[40]; y`, `0`},
		{"append", `x := []int{1}; x = append(x,2,3); x[0]; x[1]; x[2]`, `1 2 3`},
		{"set", `x := map[string]int{}; x["b"] = 42; x["b"]`, `42`},
		{"logicalOrLeft", `if false || true { 42 }`, `42`},
		{"logicalOrRight", `if true && true { 42 }`, `42`},
		{"gt", `if 4 > 2 { 42 }`, `42`},
		{"eq", `if 1 == 1 { 42 }`, `42`},
		{"negate", `x := -42; y := -x; y`, `42`},
		{"mapPlusEquals", `o := map[string]int{"x":40, "dx":2} ; o["x"] += o["dx"]; o["x"]`, `42`},
		{"localSetGet", `func f() int { o := map[string]int{"a":40}; o["b"] = 2; return o["a"]+o["b"] } ; x := f(); x`, `42`},
		{"numericMap", `o := map[int]int{1:40}; o[2] = 2; o[1]+o[2]`, `42`},
		{"mathSqrt", `import "math"; v := math.Sqrt(4); v`, `2`},
		{"mathHypot", `import "math"; v := math.Hypot(3,4); v`, `5`},
		{"randFloat64", `import "math/rand"; v := rand.Float64(); v < 1.0`, `true`},
		{"andJump", `false && true`, `false`},
		{"orJump", `true || false`, `true`},
		{"forBeak", `v := 0; for i:=1; i<5; i++ { v += i; break } ; v`, `1`},
		{"forNothing", `for i:=0; i<5; i++ { }`, ``},
		{"rangeBreak", `func f(o map[int]int) int { r := 0 ; for k,v := range o { r += k+v ; break } ; return r } ; v := f(map[int]int{40:2,2:40}); v `, `42`},
		{"zeroByte", `var x byte; x`, `0`},
		{"zeroString", `var x string; x`, ``},
		{"zeroSlice", `var x []int; x`, `[]`},
		{"mod", `9 % 7`, `2`},
		{"bitLsh", `16<<1`, `32`},
		{"bitRsh", `16>>1`, `8`},
		{"bitAnd", `0x1f&0x27`, `7`},
		{"bitOr", `1|2`, `3`},
		{"bitXor", `1^3`, `2`},
		{"stringLt", `"a"<"b"`, `true`},
		{"stringGt", `"a">"b"`, `false`},
		{"stringLte", `"a"<="b"`, `true`},
		{"stringGte", `"a">="b"`, `false`},
		{"numericLte", `4 <= 2`, `false`},
		{"numericGte", `4 >= 2`, `true`},
		{"not", `!true`, `false`},
		{"stringEq", `"a"=="b"`, `false`},
		{"sliceEq", `[]int{40} ==[]int{2}`, `false`},
		{"convertInt", `int(42.5)`, `42`},
		{"convertByte", `byte(0x101)`, `1`},
		{"convertByte2", `x := 256+42; byte(x)`, `42`},
		{"convertString", `string([]byte{52,50})`, `42`},
		{"convertSlice", `[]byte("*")[0]`, `42`},
		{"stringIndex", `"*42"[0]`, `42`},
		{"stringSlice", `"*42"[1:3]`, `42`},
		{"neq", `4!=2`, `true`},
		{"stringRange", `func f() int { res := 0 ; for k,v := range "42" { res += k + int(v) } ; return res } ; x := f(); x`, `103`},
		{"sliceSlice", `x := []string{"x","4","2","y"}; y := x[1:3]; y[0]+ y[1]`, `42`},
		{"autoNegativeSlice", `"x42"[1:]`, `42`},
		{"autoZeroSlice", `"42x"[:2]`, `42`},
		{"autoWholeSlice", `"42"[:]`, `42`},
		{"sliceSet", `x := []int{1,2,3}; x[1] = 42; x[1]`, `42`},
		{"sliceString", `import "fmt"; x := fmt.Sprint([]int{40,2}); x`, `[40 2]`},
		{"stringMapLen", `len(map[string]int{"a":40,"b":2})`, `2`},
		{"stringMapGet", `m := map[string]int{"a":42}; m["a"]`, `42`},
		{"stringMapGetZero", `m := map[string]int{"a":42}; m["b"]`, `0`},
		{"stringMapDel", `m := map[string]int{"a":42}; delete(m,"a"); len(m)`, `0`},
		{"stringMapDel2", `m := map[string]int{"a":42,"b":43,"c":44,"d":45}; delete(m,"a"); delete(m,"b");delete(m,"c");len(m)`, `1`},
		{"stringMapString", `import "fmt"; m := map[string]int{"a":4}; m`, `map[a:4]`},
		{"stringMapSprint", `import "fmt"; m := map[string]int{"a":4,"b":2}; v := fmt.Sprint(m); v == "map[a:4 b:2]" || v == "map[b:2 a:4]"`, `true`},
		{"stringMapRange", `func f() int { res:=0; m := map[string]int{"1":40,"2":2} ; for _,v := range m { res += v }; return res } ; v := f(); v`, `42`},

		{"numericMapLen", `len(map[int]int{1:40,2:2})`, `2`},
		{"numericMapGet", `m := map[int]int{1:42}; m[1]`, `42`},
		{"numericMapGetZero", `m := map[int]int{1:42}; m[2]`, `0`},
		{"numericMapDel", `m := map[int]int{1:42}; delete(m,1); len(m)`, `0`},
		{"numericMapDel2", `m := map[int]int{1:42,2:43,3:44,4:45}; delete(m,1); delete(m,2);delete(m,3);len(m)`, `1`},
		{"numericMapString", `import "fmt"; m := map[int]int{1:4}; m`, `map[1:4]`},
		{"numericMapSprint", `import "fmt"; m := map[int]int{1:4,2:2}; v := fmt.Sprint(m); v == "map[1:4 2:2]" || v == "map[2:2 1:4]"`, `true`},
		{"numericMapRange", `func f() int { res:=0; m := map[int]int{1:38,2:1} ; for k,v := range m { res += k+v }; return res } ; v := f(); v`, `42`},

		{"structZero", `type T struct { X int }; t := &T{}; t.X`, `0`},
		{"structInit", `type T struct { X int }; t := &T{X:42}; t.X`, `42`},
		{"structSet", `type T struct { X int }; t := &T{}; t.X = 42; t.X`, `42`},
		{"structGetSet", `type T struct { X int }; t := &T{X:6}; t.X *= t.X + 1; t.X`, `42`},
		{"localStructGetSet", `type T struct { X int }; func f() int  { t := &T{X:6}; t.X *= t.X + 1; return t.X }; v := f(); v`, `42`},
		{"callMethod", `type T struct { X int }; func (t *T) Test(y int) int { return t.X*y }; t := &T{X:6}; v := t.Test(7); v`, `42`},
		{"localCallMethod", `type T struct { X int }; func (t *T) Test(y int) int { return t.X*y }; func f() int { t := &T{X:6}; return t.Test(7) } v:= f(); v`, `42`},
		{"blank", `_ = 42`, ``},
		{"numericTypeFloat64", `func f() string { m := map[float64]bool{42:true}; for k,v := range m { return __type(k) } } ; v := f(); v`, `float64`},
		{"numericTypeInt", `func f() string { m := map[int]bool{42:true}; for k,v := range m { return __type(k) } } ; v := f(); v`, `int32`},
		{"numericTypeByte", `func f() string { m := map[byte]bool{42:true}; for k,v := range m { return __type(k) } } ; v := f(); v`, `uint8`},
		{"numericTypeBool", `func f() string { m := map[bool]bool{true:true}; for k,v := range m { return __type(k) } } ; v := f(); v`, `bool`},
		{"any", `func f() any { return "42" } ; v := __type(f()); v`, `string`},
		{"globalRange", `m := map[int]int{10:30,7:-5}; var res int; for k,v := range m { res += k+v; } ; res`, `42`},
		{"appendEllipsis", `a := []int{1,2}; b := []int{3,4}; c := append(a, b...); c`, `[1 2 3 4]`},
		{"zeroAppend", `var a []int; a = append(a, 42); a`, `[42]`},
		{"nilAppend", `a := []int{0}; a = nil; a = append(a, 42); a`, `[42]`},
		{"printUndefined", `asdf`, `nil`},
		{"makeSlice", "x := make([]int, 42); v := len(x); v", `42`},
		{"structPrint", `import "fmt"; type T struct { X,Y,Z int }; t := &T{X:42}; v:=fmt.Sprint(t); v`, `&{X:42 Y:0 Z:0}`},
		{"structString", `import "fmt"; type T struct { X,Y,Z int }; t := &T{X:42}; t`, `&{X:42 Y:0 Z:0}`},
		{"sliceStructPrint", `import "fmt"; type T struct { X int}; s := []*T{&T{X:42}}; v := fmt.Sprint(s); v`, `[&{X:42}]`},
		{"mapStructPrint", `import "fmt"; type T struct { X int}; s := map[int]*T{1:&T{X:42}}; v := fmt.Sprint(s); v`, `map[1:&{X:42}]`},
		{"append2x", `var res[]int; res = append(res,0); res = append(res,1); res`, `[0 1]`},
		{"forAppend", `import "fmt"; type T struct { X int }; var res []*T; for i:=0; i<3; i++ { res = append(res,&T{X:i}) } ; v := fmt.Sprint(res); v`, `[&{X:0} &{X:1} &{X:2}]`},
		{"makeZeroInt", `v := make([]int,1); v[0]`, `0`},
		{"makeZeroString", `v := make([]string,1); v[0]`, ``},
		{"makeZeroBool", `v := make([]bool,1); v[0]`, `false`},
		{"emptyStruct", `type T struct { }; x := &T{}; x`, `&{}`},
		{"emptySlice", `x:=[]int{}; len(x)`, `0`},
		{"emptyMap", `x:=map[int]int{}; len(x)`, `0`},
		{"compStructEq", `type T struct { }; x := &T{}; x == x`, `true`},
		{"compStructNeq", `type T struct { }; x,y := &T{}, &T{}; x != y`, `true`},
		{"compSliceNil", `var x []int; x == nil`, `true`},
		{"compMapNil", `var x map[int]int; x == nil`, `true`},
		{"compStructNil", `type T struct {}; var x *T; x == nil`, `true`},
		{"varFunc", `func f() int { return 42 } ; var x func() int; x = f; v = x(); v`, `42`},
		{"fieldFunc", `type T struct { F func() int }; func f() int { return 42 }; t := &T{F:f}; v = t.F(); v`, `42`},
		{"switchValue", `v := 2; switch v { case 1: 41; case 2: 42; default: 43 }`, `42`},
		{"switchTrue", `switch { case false: 42; case true: 42; default: 43; }`, `42`},
		{"structMultiField", `type T struct { X, Y, Z int } ; t := &T{X:1,Y:2,Z:3} ; t.X; t.Y ; t.Z`, `1 2 3`},
		{"shadowParam", `x := 42; func f(x int) { x = 43 } ; f(44); x`, `42`},
		{"shadowLocal", `x := 42; func f() { x := 43 } ; f(); x`, `42`},
		{"charToString", `string(42)`, `*`},
		{"shadowPackage", `import "fmt"; type T struct { Sprint int }; func f() int { fmt := &T{Sprint:42}; return fmt.Sprint }; v := f(); v`, `42`},
		{"mapString", `m := map[int]string{2:"hi"}; m`, `map[2:hi]`},
		{"rangeNilSlice", `var x []int; for k,v := range x { k; v }`, ``},
		{"rangeNilMember", `type T struct { x []int }; t := &T{}; for k,v := range t.x { k ; v }`, ``},
		{"localMultiRet", `func f() (int,int) { return 6,7 } ; func g() int { a,b := f(); return a*b }; v := g(); v`, `42`},
		{"retMultiRet", `func f() (int,int) { return 6,7 } ; func g() (int,int) { return f() }; x,y := g(); x*y`, `42`},
		{"forContinue", `for i:=0; i<1; i++ { continue ; 42 }`, ``},
		{"initFunc", "var x = 43 ; func init() { x = 42 } ; x", `42`},
		{"err", `import "errors"; err = errors.New("42"); v := err.String(); v`, `42`},
		{"errReturn", `package main; import "errors"; func f() error { return errors.New("42") } ; v = f().String(); v`, `42`},
		{"varErr", `import "errors"; var err error; err = errors.New("42"); v := err.String(); v`, `42`},
		// {"stringer", `import "fmt"; type T struct { }; func (t *T) String() string { return "42" }; v := fmt.Sprint(&T{}); v`, `42`},
		{"localDiv", `func f() int { a := 84; b := 2; return a/b }; v = f(); v`, `42`},
		{"localSub", `func f() int { a := 44; b := 2; return a-b }; v := f(); v`, `42`},
		{"structMethodsString", `type T struct { X int }; func (t *T) M() { }; t := &T{X:42}; t`, `&{X:42}`},
		{"lenNil", `var v []int; len(v)`, `0`},
		{"callSkipReturns", `func f() int { return 42 }; f()`, ``},
		{"byteMul", `x := byte(42)*byte(42); x`, `228`},
		{"intConstAddFloat", `x := float64(2.5); y := 40 + x; y`, `42.5`},
		{"byteConstAddType", `x := byte(6); y := 7 * x; t := __type(y); t`, `uint8`},
		{"globalConstAssign", `var x byte; x = 42; t := __type(x); t`, `uint8`},
		{"localConstAssign", `func f() { var x byte; x = 42; t := __type(x); return t }; t := f(); t`, `uint8`},
		{"intMapConstAssign", `m := map[int]byte{}; m[0] = 42; t := __type(m[0]); t`, `uint8`},
		{"stringMapConstAssign", `m := map[string]byte{}; m["k"] = 42; t := __type(m["k"]); t`, `uint8`},
		{"sliceConstAssign", `s := []byte{0}; s[0] = 42; t := __type(s[0]); t`, `uint8`},
		{"structConstAssign", `type T struct { V byte }; s := &T{}; s.V = 42; t := __type(s.V); t`, `uint8`},
		{"intMapConstInit", `m := map[int]byte{0:42}; t := __type(m[0]); t`, `uint8`},
		{"stringMapConstInit", `m := map[string]byte{"k":42}; t := __type(m["k"]); t`, `uint8`},
		{"sliceConstInit", `s := []byte{42}; t := __type(s[0]); t`, `uint8`},
		{"structConstInit", `type T struct { V byte }; s := &T{V:42}; t := __type(s.V); t`, `uint8`},
		{"sliceConstAppend", `s := []byte{}; s = append(s, 42); t := __type(s[0]); t`, `uint8`},
		{"nilSliceConstAppend", `var s []byte; s = append(s, 42); t := __type(s[0]); t`, `uint8`},
		{"varConstIntAssign", `x := 42; t := __type(x); t`, `int32`},
		{"varConstFloatAssign", `x := 42.0; t := __type(x); t`, `float64`},
		{"funcConstArg", `func f(x byte) string { return __type(x) }; t := f(42); t`, `uint8`},
		{"funcConstReturn", `func f() byte { return 42 }; t := __type(f()); t`, `uint8`},
		{"funcConstReturnAssign", `func f() byte { return 42 }; x := f(); t := __type(x); t`, `uint8`},
		{"funcNilSliceArg", `func f(x []byte) string { return __type(x) }; t := f(nil); t`, `[]uint8`},
		{"funcNilSliceReturn", `func f() []byte { return nil }; t := __type(f()); t`, `[]uint8`},
		{"funcNilSliceArgAppend", `func f(x []byte) string { x = append(x,42); return __type(x[0]) }; t := f(nil); t`, `uint8`},
		{"funcNilSliceReturnAppend", `func f() []byte { return nil }; t := __type(append(f(),42)[0]); t`, `uint8`},
		{"byteOrFix", `b := byte(42) | 256; b`, `42`},
		{"copy", `a := []int{0,0}; b := []int{4,2,3}; copy(a,b); a`, `[4 2]`},
		{"copyString", `a := []int{0}; copy(a,"*") a`, `[42]`},
		{"setVar", `var x int = 42; x`, `42`},
		{"setByte", `var x byte = 42; t := __type(x); t`, `uint8`},
		{"assignBug", `func f() int { for { n := 0; break } ; t := 1.0 / 60; return int(t*1000) } ; x := f(); x`, `16`},
		{"nameShadow", "func f() int { x := 1; for i:=0; i<1; i++ { x := 2 } ; return x } ; y := f(); y", `1`},
		{"copyMethod", `type T struct { }; func (t *T) copy() int { return 42 }; t := &T{}; v := t.copy(); v`, `42`},
		{"compoundSliceSliceType", `var x [][]string; v := __type(x); v`, `[][]string`},
		{"compoundSliceMapType", `var x []map[int]string; v := __type(x); v`, `[]map[int32]string`},
		{"compoundMapSliceType", `var x map[int][]string; v := __type(x); v`, `map[int32][]string`},
		{"compoundMapMapType", `var x map[int]map[byte]string; v := __type(x); v`, `map[int32]map[uint8]string`},
		{"compoundAppendType", `var x []map[int]string; x = append(x,nil); v := __type(x[0]); v`, `map[int32]string`},
		{"compoundMapGetDefaultType", `var x map[int][]string; v := __type(x[0]); v`, `[]string`},
		{"compoundMapSetDefaultType", `x := map[int][]string{}; x[0] = nil; v := __type(x[0]); v`, `[]string`},
		{"compoundStructFieldsType", `type T struct { X map[int]string }; t := &T{}; v := __type(t.X); v`, `map[int32]string`},
		{"anyType", `var x any; v := __type(x); v`, `any`},
		{"incOpt", `func f() int { x := 40; x += 2; return x } ; v := f(); v`, `42`},
		{"decOpt", `func f() int { x := 44; x -= 2; return x } ; v := f(); v`, `42`},
		{"varSliceToNilType", `var x []byte; x = nil; v := __type(x); v`, `[]uint8`},
		{"varMapToNilType", `var x map[int]string; x = nil; v := __type(x); v`, `map[int32]string`},
		{"funcToNil", `var x func(); x = nil; x`, `nil`},
		{"structToNil", `type T struct{} var x *T; x = nil; x`, `nil`},
		{"anyToIntToString", `var x any; x = 42; x = "test"; x`, `test`},
		{"anyToStringToInt", `var x any; x = "test"; x = 42; x`, `42`},
		{"nilMapString", `var x map[string]int; x`, `map[]`},
		{"anyToStringToNil", `var x any; x = "test"; x = nil; x`, `nil`},
		{"anyToIntToNil", `var x any; x = 0; x = nil; x`, `nil`},
		{"anyToFuncToNil", `func f() {} var x any; x = f; x = nil; x`, `nil`},
		{"anyToStructToNil", `type T struct {} ; var x any; x = &T{}; x = nil; x`, `nil`},
		{"nilStructEqNil", `type T struct {} ; var x *T; x == nil`, `true`},
		{"nilFuncEqNil", `var x func (); x == nil`, `true`},
		{"emptySliceType", `x := []byte{}; t := __type(x); t`, `[]uint8`},
		{"emptyStringMapType", `x := map[string]int{}; t := __type(x); t`, `map[string]int32`},
		{"emptyNumericMapType", `x := map[int]string{}; t := __type(x); t`, `map[int32]string`},
		{"emojiToByteSlice", `x := []byte("🐐"); x`, `[240 159 144 144]`},
		{"invalidGlobalLookup", `package ext; type T struct { } ; type S struct { T * T } ; func f() *T { t := &T{} ; return t } ; x := f(); x`, `&{}`},
		{"makeSliceStruct", `type T struct{} ; x := []*T{nil,nil,nil} ; v := make([]*T,len(x)); v`, `[nil nil nil]`},
		{"nilEqNil", `var a, b any; a == b`, `true`},
		{"structPointers", `type T struct { V int } ; x := &T{}; y = x; y.V = 42; x.V`, `42`},
		{"longSliceFloat64", `x := []float64{1,2,3}; v := __type(x[1]); v`, `float64`},
		{"int32", `var x int32; v := __type(x); v`, `int32`},
		{"uint8", `var x uint8; v := __type(x); v`, `uint8`},
		{"fastGetInt", `func f() int { x := []int{42}; return x[0] }; v := f(); v`, `42`},
		{"fastSetInt", `func f() int { x := []int{0}; x[0] = 42; return x[0] }; v := f(); v`, `42`},
		{"runeType", `x := rune(42); v := __type(x); v`, `int32`},
		{"stringIterType", `func f() string { for _,v := range "*" { return __type(v) } }; v := f(); v`, `int32`},
		{"uintType", `var x uint; t := __type(x); t`, `uint32`},
		{"charStar", `x := '*'; x`, `42`},
		{"charBackSlash", `x := '\\'; x`, `92`},
		{"charNewLine", `x := '\n'; x`, `10`},
		{"incField", `type T struct { N int }; func (t *T) F() { t.N ++ }; t := &T{}; t.F(); t.N`, `1`},
		{"constRefConst", `const (a = 40; b = a+2 ); b`, `42`},
		{"constSingle", `const a = 42; a`, `42`},
		{"constMulti", `const a,b = 40,2; a;b`, `40 2`},
		{"constBlockMulti", `const (a,b = 40,2); a;b`, `40 2`},
		{"constWeirdBug", `const a = -42; const b=-a; b`, `42`},
		{"complementInt", `a := ^42; a`, `-43`},
		{"complementIntNeg", `a := ^-43; a`, `42`},
		{"complementByte", `a := ^byte(42); a`, `213`},
		{"complementByteNeg", `a := ^int8(-43); a`, `42`},
		{"localConst", `func f() int { const a = 42; return a }; v = f(); v`, `42`},
		{"interfaceReloadBug", `type T interface { F() }; type T interface { F() }; v := __type(T); v`, `struct`},
		{"structStructUnsafeString", `type T struct { X *T }; t:=&T{X:&T{}}; t`, `&{X:&{...}}`},
		{"structStructSafeString", `type P struct { X, Y, Z int }; type T struct { P *P }; t:=&T{P:&P{}}; t`, `&{P:&{X:0 Y:0 Z:0}}`},
		{"sliceSliceSafeString", `x := [][]int{{1,2,3},{4,5,6}}; x`, `[[1 2 3] [4 5 6]]`},
		{"sliceSliceUnsafeString", `type T struct {} ; x := [][]*T{{&T{}}}; x`, `[[...]]`},
		{"stringMapMapSafeString", `x := map[string]map[string]int{"a":{"b":42}}; x`, `map[a:map[b:42]]`},
		{"stringMapMapUnsafeString", `type T struct {} ; x := map[string]map[string]*T{"a":{"b":&T{}}}; x`, `map[a:map[...]]`},
		{"numericMapMapSafeString", `x := map[int]map[int]int{4:{2:42}}; x`, `map[4:map[2:42]]`},
		{"numericMapMapUnsafeString", `type T struct {} ; x := map[int]map[int]*T{4:{2:&T{}}}; x`, `map[4:map[...]]`},
		{"structStructMethodsString", `type T struct { X int }; func (t *T) M() { }; t := []*T{&T{X:42}}; t`, `[&{X:42}]`},
		{"variadicFunc", `func f(a ...int) int { return a[0]+a[1] }; v := f(40,2); v`, `42`},
		{"variadicSlice", `func f(a ...int) int { return a[0]+a[1] }; a:=[]int{40,2}; v := f(a...); v`, `42`},
		{"variadicInlineSlice", `func f(a ...int) int { return a[0]+a[1] }; v := f([]int{40,2}...); v`, `42`},
		{"variadicMethod", `type T struct {}; func (t *T) M(a ...int) { return a[0]+a[1] }; t := &T{}; v := t.M(40,2); v`, `42`},
		{"variadicType", `func f(a ...int) string { return __type(a) }; v := f(); v`, `[]int32`},
		{"variadicTypeArg", `func f(a ...int) string { return __type(a) }; v := f(42); v`, `[]int32`},
		{"localStruct", `func f() int { type T struct { X int }; t := &T{X:42}; return t.X }; v := f(); v`, `42`},
		{"localType", `func f() any { type T byte; var v T = 42; return v }; v := f(); x := __type(v); x`, `uint8`},
		{"caseBreak", `v := 0; for i:=0; i<1; i++ { switch i { case 0: break } ; v = 42 ; break }; v`, `42`},
		{"caseBreakDefault", `v := 42; switch true { case true: break; default: v = 0 } ; v`, `42`},
		{"passCoverage", `v := 0; switch true { case true: v = 42; default: } v`, `42`},
		{"caseTrueBug", `v := 0; switch true { case true: v = 42 } v`, `42`},
		{"trueEqTrue", `v := 0; if true == true { v = 42 }; v`, `42`},
		{"gtTrue", `1>0`, `true`},
		{"gtFalse1", `0>0`, `false`},
		{"gtFalse2", `0>1`, `false`},
		{"ltTrue", `0<1`, `true`},
		{"ltFalse1", `0<0`, `false`},
		{"ltFalse2", `1<0`, `false`},
		{"gteTrue1", `1>=0`, `true`},
		{"gteTrue2", `0>=0`, `true`},
		{"gteFalse", `0>1`, `false`},
		{"lteTrue1", `0<=1`, `true`},
		{"lteTrue2", `0<=0`, `true`},
		{"lteFalse", `1<=0`, `false`},

		// approximate according to Go
		{"~funcType", `func t() {}; v := __type(t); v`, `func`},
		{"~funcToNilType", `func f() {}; x := f; x = nil; v := __type(x); v`, `func`},
		{"~nilFuncType", `var t func(); v := __type(t); v`, `func`},
		{"~nilFuncToNilType", `var x func(); x = nil; v := __type(x); v`, `func`},
		{"~structType", `type T struct{}; t := &T{}; v := __type(t); v`, `struct`},
		{"~structToNilType", `type T struct {} x := &T{}; x = nil; v := __type(x); v`, `struct`},
		{"~nilStructType", `type T struct {}; var t *T; v := __type(t); v`, `struct`},
		{"~nilStructToNilType", `type T struct {} var x *T; x = nil; v := __type(x); v`, `struct`},

		// undefined according to Go
		{"?intToString", `x := 42; x = "test"; x`, `test`},
		{"?stringToInt", `x := "test"; x = 42; x`, `42`},
		{"?stringToNil", `var x string; x = nil; x`, `nil`},
		{"?stringToNilType", `var x string; x = nil; v := __type(x); v`, `any`},
		{"?varStringToNil", `var x string = nil; x`, `nil`},
		{"?varStringToNilType", `var x string = nil; v = __type(x); v`, `any`},
		{"?intToNil", `var x int; x = nil; x`, `nil`},
		{"?intToNilType", `var x int; x = nil; v := __type(x); v`, `any`},
		{"?lessReturns", `func f() (int,int) { return 4,2 }; x := f(); x`, `4`},
		{"?upsertFunc", `func f() int { return 0 }; x := f; func f() int { return 42 }; v := x(); v`, `42`},
		{"?upsertMethod", `type T struct {}; func (t *T) f() int { return 0; }; t := &T{}; x := t.f; func (t*T) f() int { return 42; }; v := x(); v`, `42`},
		{"?upsertTypeMethodBefore", `type T struct {}; func (t *T) f() int { return 42; }; t := &T{}; type T struct {}; v := t.f(); v`, `42`},
		{"?upsertTypeMethodAfter", `type T struct {}; func (t *T) f() int { return 42; }; type T struct {}; t := &T{}; v := t.f(); v`, `42`},
		{"?upsertTypeFieldBefore", `type T struct { X int }; func (t *T) f() any { return t.X }; t := &T{}; type T struct { X string }; v := __type(t.f()); v`, `int32`},
		{"?upsertTypeFieldAfter", `type T struct { X int }; func (t *T) f() any { return t.X }; type T struct { X string }; t := &T{}; v := __type(t.f()); v`, `string`},

		// incorrect according to Go
		{"!anyToSliceToNil", `var x any; x = []byte{}; x = nil; x`, `[]`},
		{"!anyToMapToNil", `var x any; x = map[string]int{}; x = nil; x`, `map[]`},
		{"!anyToSliceToNilType", `var x any; x = []byte{}; x = nil; t := __type(x); t`, `[]uint8`},
		{"!anyToMapToNilType", `var x any; x = map[string]int{}; x = nil; t := __type(x); t`, `map[string]int32`},
		{"!anyToFuncToNilType", `func f() {} var x any; x = f; x = nil; t := __type(x); t`, `func`},
		{"!anyToStructToNilType", `type T struct {} ; var x any; x = &T{}; x = nil; t := __type(x); t`, `struct`},
		{"!structNonPointer", `type T struct { V int } ; x := T{}; y = x; y.V = 42; x.V`, `42`},
	}
	opts := []struct {
		suffix   string
		optimize bool
	}{
		{"", false},
		{"/optimized", true},
	}
	for _, row := range tests {
		for _, opt := range opts {
			t.Run(row.Name+opt.suffix, func(t *testing.T) {
				tokens, err := tokenize(row.Name, row.In)
				if err != nil {
					t.Fatalf("Tokenize error: %v", err)
				}
				tree, err := parse(tokens)
				if err != nil {
					t.Fatalf("Parse error: %v", err)
				}
				vm := NewVM()
				codes, slots, err := compile(vm.globals, tree, opt.optimize)
				if err != nil {
					t.Fatalf("Compile error: %v", err)
				}
				// {
				// 	fmt.Println(row.Name + opt.suffix)
				// 	var ts []string
				// 	for _, s := range codes {
				// 		ts = append(ts, s.String(vm.globals))
				// 	}
				// 	s := strings.Join(ts, "; ")
				// 	fmt.Println(s)
				// }
				rets, err := vm.run(codes, slots)
				if err != nil {
					t.Fatalf("Exec error: %v", err)
				}
				var ts []string
				for _, s := range rets {
					ts = append(ts, s.String())
				}
				s := strings.Join(ts, " ")
				if s != row.Want {
					t.Fatalf("Exec got %v want %v", s, row.Want)
				}
			})
		}
	}
}

func TestVM_error(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Err  string
	}{
		{"tooManyArgs", `func f() { } ; f(42)`, `incorrect args`},
		{"notEnoughArgs", `func f(a int) { } ; f()`, `incorrect args`},
		{"notEnoughReturns", `func f() {} ; x := f()`, `incorrect returns`},
		{"backtrace", `package main; func f() { g() } func g() { die() } f()`, `main.g(...)`},
		{"backtraceBottom", `package main; func f() { g() } func g() { die() } f()`, `main.f(...)`},
		{"panic", `panic("hello")`, `hello`},
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
			vm := NewVM()
			codes, slots, err := compile(vm.globals, tree, false)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}
			// for _, c := range codes {
			// 	fmt.Println(c.String(g))
			// }
			_, err = vm.run(codes, slots)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("Exec error got %v want %v", err, row.Err)
			}
		})
	}
}

func TestVM_WithStdout(t *testing.T) {
	stdout := &bytes.Buffer{}
	vm := NewVM(WithStdout(stdout))
	_, err := vm.Eval(nil, "stdin", `print("42")`)
	if err != nil {
		t.Fatalf("Eval err got %v want nil", err)
	}
	assert(t, "stdout", string(stdout.Bytes()), "42")
}

func TestVM_WithLoaders(t *testing.T) {
	var loaded bool
	loader := func(*VM) { loaded = true }
	_ = NewVM(WithLoaders(loader))
	assert(t, "loaded", loaded, true)
}

func TestVM_unknown(t *testing.T) {
	want := "unknown code"
	codes := []instruction{{Code: -42}}
	vm := NewVM()
	_, err := vm.run(codes, 0)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("Exec error got %v want %v", err, want)
	}
}

func TestVM_Call(t *testing.T) {
	vm := NewVM()
	res, err := vm.Call("math.Sqrt", 1, Float64(256))
	if err != nil {
		t.Fatalf("Call error got %v want nil", err)
	}
	if len(res) != 1 {
		t.Fatalf("results len got %v want 1", len(res))
	}
	if res[0].Float64() != 16 {
		t.Fatalf("result got %v want 16", res[0].String())
	}
}

func TestVM_incorrect_stack(t *testing.T) {
	vm := NewVM()
	vm.Set("f", NewFunc(0, 1, func(vm *VM) {}))
	_, err := vm.Call("f", 1)
	want := "incorrect returns" // was "incorrect stack"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Call error got %v want %v", err, want)
	}
}

func TestVM_Call_error(t *testing.T) {
	vm := NewVM()
	_, err := vm.Call("math.Sqrt", 2, Float64(256))
	want := "incorrect returns"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Call error got %v want %v", err, want)
	}
}

func TestVM_Func(t *testing.T) {
	vm := NewVM()
	res, err := vm.Func(vm.globals.Get("math.Sqrt"), 1, Float64(256))
	if err != nil {
		t.Fatalf("Func error got %v want nil", err)
	}
	if len(res) != 1 {
		t.Fatalf("results len got %v want 1", len(res))
	}
	if res[0].Float64() != 16 {
		t.Fatalf("result got %v want 16", res[0].String())
	}
}

func TestVM_Func_error(t *testing.T) {
	vm := NewVM()
	_, err := vm.Func(vm.globals.Get("math.Sqrt"), 2, Float64(256))
	want := "incorrect returns"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Func error got %v want %v", err, want)
	}
}

func TestVM_Eval(t *testing.T) {
	t.Run("Happy", func(t *testing.T) {
		vm := NewVM()
		rets, err := vm.Eval(nil, "test", "42", testWithNoop())
		if err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		res := rets[0].Int()
		assert(t, "res", res, 42)
	})
	t.Run("WithImports", func(t *testing.T) {
		vm := NewVM()
		imports := map[string]string{}
		_, err := vm.Eval(mapFS{}, "test", `import "fmt"`, WithEvalImports(imports))
		if err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		res := fmt.Sprint(imports)
		assert(t, "res", res, "map[fmt:fmt]")
	})
	t.Run("WithTreeDump", func(t *testing.T) {
		vm := NewVM()
		w := &bytes.Buffer{}
		_, err := vm.Eval(mapFS{}, "test", `7 + 2 * 3`, WithTreeDump(w))
		if err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		res := w.String()
		assert(t, "res", res, "(+ 7 (* 2 3))\n")
	})

	t.Run("WithCodeDump", func(t *testing.T) {
		vm := NewVM()
		w := &bytes.Buffer{}
		_, err := vm.Eval(mapFS{}, "test", `42`, WithCodeDump(w))
		if err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		res := w.String()
		assert(t, "res", res, "test:1:1: PUSH 42\n")
	})
}

func TestVM_Eval_error(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Err  string
	}{
		{"runErr", `f()`, `error in run:`},
		{"compileErr", `import "ext"; ext.F()`, `error in compile:`},
		{"loadImportsErr", `import "\\"`, `error in loadImports:`},
		{"parseErr", `++`, `error in parse:`},
		{"tokenizeErr", `"`, `error in tokenize:`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			vm := NewVM()
			_, err := vm.Eval(mapFS{}, "eval", row.In)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("Eval error got %v want %v", err, row.Err)
			}
		})
	}
}

func TestVM_Eval_importError(t *testing.T) {
	tests := []struct {
		Name string
		FS   fs.FS
		In   string
		Err  string
	}{
		{"loadImports",
			mapFS{
				"ext/ext.go": `++`,
			},
			`import "ext"`, `error in loadImports:`},
		{"compile",
			mapFS{
				"ext/ext.go": `package ext; import "fail"; fail.F()`,
			},
			`import "ext"`, `error in compile (imports):`},
		{"run",
			mapFS{
				"ext/ext.go": `package ext; var x = fail()`,
			},
			`import "ext"`, `error in run (imports):`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			vm := NewVM()
			_, err := vm.Eval(row.FS, "eval", row.In)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("Eval error got %v want %v", err, row.Err)
			}
		})
	}
}

func testWithNoop() RunOption { return func(o *runConfig) {} }

func TestVM_Load(t *testing.T) {
	t.Run("package", func(t *testing.T) {
		vm := NewVM()
		err := vm.Load(mapFS{"main/main.go": "package main"}, "main", testWithNoop())
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
	})
	t.Run("file", func(t *testing.T) {
		vm := NewVM()
		err := vm.Load(mapFS{"main/main.go": "package main"}, "main/main.go", testWithNoop())
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
	})

	t.Run("buildConstraintsBug", func(t *testing.T) {
		vm := NewVM()
		err := vm.Load(mapFS{
			"point/point.go": `package point; type Point struct {}`,
			"point/goat.go": `//go:build !goat
package point
`,
			"matrix/matrix.go": `package matrix; import "point"; func F(p *point.Point) {}`,
			"main/main.go":     `package main; import "matrix"; func main() {}`,
		}, "main", testWithNoop())
		if err != nil {
			t.Fatalf("Load error: %v", err)
		}
	})
}

func TestVM_Load_error(t *testing.T) {
	tests := []struct {
		Name string
		Arg  string
		FS   fs.FS
		Err  string
	}{
		{"runErr", "main", mapFS{"main/main.go": `package main; f()`}, `error in run:`},
		{"compileErr", "main", mapFS{"main/main.go": `package main; import "ext"; ext.F()`}, `error in compile:`},
		{"loadImportsErr", "main", mapFS{"main/main.go": `package main; import "\\"`}, `error in load:`},
		{"parseErr", "main", mapFS{"main/main.go": `++`}, `error in parse:`},
		{"tokenizeErr", "main", mapFS{"main/main.go": `package main; "`}, `error in tokenize:`},
		{"unexpectedReturns", "main", mapFS{"main/main.go": `package main; 42`}, `unexpected returns:`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			vm := NewVM()
			err := vm.Load(row.FS, row.Arg)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("Load error got %v want %v", err, row.Err)
			}
		})
	}
}
