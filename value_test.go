package goatlang

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func assert[T comparable](t *testing.T, name string, got, want T) {
	t.Helper()
	if got == want {
		return
	}
	t.Fatalf("%s got %v want %v", name, got, want)
}

func expectPanic(t *testing.T, want string) {
	t.Helper()
	r := recover()
	got := fmt.Sprint(r)
	if !strings.Contains(got, want) {
		t.Fatalf("panic got %v want %v", got, want)
	}
}

func Test_Value_convert_nil(t *testing.T) {
	v := Float64(42)
	r := v.convert(TypeMap)
	res := r.Type()
	assert(t, "res", res, TypeNil)
}

func Test_Value_Type(t *testing.T) {
	v := Int(42)
	res := v.Type()
	assert(t, "res", res, TypeInt32)
}

func Test_Struct(t *testing.T) {
	t.Run("GetAttr", func(t *testing.T) {
		vm := New()
		if _, err := vm.Eval(nil, "test", "package main; type T struct { X int }; t := &T{X:42}"); err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		val := vm.Get("main.t")
		res := val.GetAttr("X").Int()
		assert(t, "res", res, 42)
	})

	t.Run("SetAttr", func(t *testing.T) {
		vm := New()
		if _, err := vm.Eval(nil, "test", "package main; type T struct { X int }; t := &T{X:0}"); err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		val := vm.Get("main.t")
		val.SetAttr("X", Int(42))
		rets, err := vm.Eval(nil, "test", "t.X")
		if err != nil {
			t.Fatalf("Eval#2 error: %v", err)
		}
		res := rets[0].Int()
		assert(t, "res", res, 42)
	})

	t.Run("New", func(t *testing.T) {
		vm := New()
		if _, err := vm.Eval(nil, "test", "package main; type T struct { X int }; t := &T{X:0}"); err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		base := vm.Get("main.T")
		val := NewStruct(base, []Value{String("X"), Int(42)})
		res := val.GetAttr("X").Int()
		assert(t, "res", res, 42)
	})
}

type testStruct struct {
	Object
	X, Y int
}

func NewTestStruct() Value {
	return Wrap(&testStruct{})
}

func (t *testStruct) GetAttr(key string) Value {
	switch key {
	case "X":
		return Int(t.X)
	case "Y":
		return Int(t.Y)
	}
	return Nil()
}

func (t *testStruct) SetAttr(key string, value Value) {
	switch key {
	case "X":
		t.X = value.Int()
	case "Y":
		t.Y = value.Int()
	}
}

func Test_NativeStruct(t *testing.T) {
	t.Run("Attr", func(t *testing.T) {
		v := NewTestStruct()
		v.SetAttr("X", Int(42))
		res := v.GetAttr("X").Int()
		assert(t, "res", res, 42)
	})

	t.Run("Index", func(t *testing.T) {
		vm := New()
		v := NewTestStruct()
		vm.Set("main.t", v)
		rets, err := vm.Eval(nil, "test", "t.X = 42; t.X")
		if err != nil {
			t.Fatalf("Eval error: %v", err)
		}
		res := rets[0].Int()
		assert(t, "res", res, 42)
	})
}

func Test_Byte(t *testing.T) {
	x := Int(256 + 42)
	res := x.Byte()
	assert(t, "res", res, byte(42))
}

func Test_Func(t *testing.T) {
	t.Run("N->0", func(t *testing.T) {
		var calledFunc bool
		f := NewFunc(2, 0, func(v *VM, args []Value) {
			calledFunc = true
			res := args[0].Int() * args[1].Int()
			assert(t, "res", res, 42)
		})
		vm := New()
		vm.Set("main.F", f)
		_, err := vm.Call("main.F", 0, Int(6), Int(7))
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		assert(t, "calledFunc", calledFunc, true)
	})
	t.Run("N->M", func(t *testing.T) {
		var calledFunc bool
		f := NewFunc(2, 0, func(v *VM, args []Value) []Value {
			calledFunc = true
			res := args[0].Int() * args[1].Int()
			assert(t, "res", res, 42)
			return []Value{Int(4), Int(res), Int(2)}
		})
		vm := New()
		vm.Set("main.F", f)
		res, err := vm.Call("main.F", 2, Int(6), Int(7))
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		assert(t, "calledFunc", calledFunc, true)
		assert(t, "len(res)", len(res), 2)
		assert(t, "res[0]", res[0].String(), "4")
		assert(t, "res[1]", res[1].String(), "42")
	})
}

func Test_Slice(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(42)})
		resV, ok := x.Get(Int(0))
		if !ok {
			t.Fatalf("ok got false want true")
		}
		res := resV.Int()
		assert(t, "res", res, 42)
	})

	t.Run("Set", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(0)})
		x.Set(Int(0), Int(42))
		resV, ok := x.Get(Int(0))
		if !ok {
			t.Fatalf("ok not true")
		}
		res := resV.Int()
		assert(t, "res", res, 42)
	})

	t.Run("Len", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(0), Int(0)})
		res := x.Len()
		assert(t, "res", res, 2)
	})

	t.Run("Range", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(6), Int(7)})
		next := x.Range()
		res := fmt.Sprint(next()) + "; " + fmt.Sprint(next()) + "; " + fmt.Sprint(next())
		assert(t, "res", res, "0 6 true; 1 7 true; nil nil false")
	})

	t.Run("Append", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(0), Int(6)})
		x = x.Append(Int(7), Int(42))
		res := x.String()
		assert(t, "res", res, "[0 6 7 42]")
	})

	t.Run("Slice", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(0), Int(6), Int(7), Int(42)})
		res := x.Slice(1, 3).String()
		assert(t, "res", res, "[6 7]")
	})

	t.Run("Data", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(6), Int(7)})
		res := fmt.Sprint(x.data())
		assert(t, "res", res, "[6 7]")
	})

	t.Run("convert", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(42)})
		v := x.convert(TypeString)
		res := v.String()
		assert(t, "res", res, "*")
	})

	t.Run("convert_noop", func(t *testing.T) {
		x := NewSlice(TypeInt32, []Value{Int(42)})
		v := x.convert(TypeSlice)
		res := v.String()
		assert(t, "res", res, "[42]")
	})
}

func Test_Map(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		x := NewMap(TypeString, TypeInt32, []Value{String("x"), Int(42)})
		resV, ok := x.Get(String("x"))
		assert(t, "ok", ok, true)
		res := resV.Int()
		assert(t, "res", res, 42)
	})

	t.Run("Set", func(t *testing.T) {
		x := NewMap(TypeString, TypeInt32, []Value{})
		x.Set(String("x"), Int(42))
		resV, ok := x.Get(String("x"))
		assert(t, "ok", ok, true)
		res := resV.Int()
		assert(t, "res", res, 42)
	})

	t.Run("Len", func(t *testing.T) {
		x := NewMap(TypeString, TypeInt32, []Value{String("x"), Int(0)})
		res := x.Len()
		assert(t, "res", res, 1)
	})

	t.Run("Range", func(t *testing.T) {
		x := NewMap(TypeString, TypeInt32, []Value{String("x"), Int(42)})
		next := x.Range()
		res := fmt.Sprint(next()) + "; " + fmt.Sprint(next())
		assert(t, "res", res, "x 42 true; nil nil false")
	})

	t.Run("Delete", func(t *testing.T) {
		x := NewMap(TypeString, TypeInt32, []Value{String("x"), Int(42)})
		x.Delete(String("x"))
		resV, ok := x.Get(String("x"))
		assert(t, "ok", ok, false)
		res := resV.Int()
		assert(t, "res", res, 0)
	})
}

type testSlice struct {
	Object
	data []float64
}

func (t *testSlice) Len() int { return len(t.data) }

func (t *testSlice) Range() func() (Value, Value, bool) {
	r := t.data
	n := 0
	return func() (Value, Value, bool) {
		if n >= len(r) {
			return Nil(), Nil(), false
		}
		k, v := Int(n), Float64(r[n])
		n++
		return k, v, true
	}
}

func Test_NativeSlice(t *testing.T) {
	t.Run("data", func(t *testing.T) {
		r := Wrap(&testSlice{data: []float64{6, 7}})
		res := fmt.Sprint(r.data())
		assert(t, "res", res, "[6 7]")
	})
}

func Test_Nil(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		v := Nil()
		res := v.IsNil()
		assert(t, "res", res, true)
	})
	t.Run("func", func(t *testing.T) {
		v := Nil()
		v.t = TypeFunc
		res := v.IsNil()
		assert(t, "res", res, true)
	})
	t.Run("object", func(t *testing.T) {
		v := Wrap(nil)
		res := v.IsNil()
		assert(t, "res", res, true)
	})
}

func Test_Value_Unwrap(t *testing.T) {
	s := &testSlice{}
	v := Wrap(s)
	res := v.Unwrap()
	if res != s {
		t.Fatalf("res got %v want %v", res, s)
	}
}

func Test_Value_Unwrap_nil(t *testing.T) {
	v := Int(42)
	res := v.Unwrap()
	if res != nil {
		t.Fatalf("res got %v want %v", res, nil)
	}
}

func Test_String(t *testing.T) {
	t.Run("convert", func(t *testing.T) {
		x := String("*")
		v := x.convert(TypeSlice)
		res := v.String()
		assert(t, "res", res, "[42]")
	})

	t.Run("convert_noop", func(t *testing.T) {
		x := String("*")
		v := x.convert(TypeString)
		res := v.String()
		assert(t, "res", res, "*")
	})

	t.Run("Set_panic", func(t *testing.T) {
		defer expectPanic(t, "unsupported")
		v := String("*")
		v.Set(Int(0), Int(0))
	})

	t.Run("Append_panic", func(t *testing.T) {
		defer expectPanic(t, "unsupported")
		v := String("*")
		v.Append(Int(0))
	})

	t.Run("Delete_panic", func(t *testing.T) {
		defer expectPanic(t, "unsupported")
		v := String("*")
		v.Delete(Int(0))
	})

	t.Run("GetAttr_panic", func(t *testing.T) {
		defer expectPanic(t, "unsupported")
		v := String("*")
		_ = v.GetAttr("Attr")
	})

	t.Run("SetAttr_panic", func(t *testing.T) {
		defer expectPanic(t, "unsupported")
		v := String("*")
		v.SetAttr("Attr", String("value"))
	})
}

func Test_NilSlice(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		defer expectPanic(t, "runtime error")
		x := Nil().assign(TypeSlice)
		x.Get(Int(0))
	})

	t.Run("Set", func(t *testing.T) {
		defer expectPanic(t, "runtime error")
		x := Nil().assign(TypeSlice)
		x.Set(Int(0), Int(42))
	})

	t.Run("Len", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		res := x.Len()
		assert(t, "res", res, 0)
	})

	t.Run("Range", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		next := x.Range()
		res := fmt.Sprint(next())
		assert(t, "res", res, "nil nil false")
	})

	t.Run("Append", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		x = x.Append(Int(42))
		res := x.String()
		assert(t, "res", res, "[42]")
	})

	t.Run("Slice", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		res := x.Slice(0, 0).String()
		assert(t, "res", res, "[]")
	})

	t.Run("Slice_outOfRange", func(t *testing.T) {
		defer expectPanic(t, "runtime error")
		x := Nil().assign(TypeSlice)
		x.Slice(1, 1)
	})

	t.Run("Data", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		res := fmt.Sprint(x.data())
		assert(t, "res", res, "[]")
	})

	t.Run("convert", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		v := x.convert(TypeString)
		res := v.String()
		assert(t, "res", res, "")
	})

	t.Run("convert_noop", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		v := x.convert(TypeSlice)
		res := v.String()
		assert(t, "res", res, "[]")
	})

	t.Run("eqNil", func(t *testing.T) {
		x := Nil().assign(TypeSlice)
		res := x.Equals(Nil())
		assert(t, "res", res, true)
	})
}

func Test_NilMap(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		x := Nil().assign(TypeMap)
		resV, ok := x.Get(String("x"))
		assert(t, "ok", ok, false)
		res := resV.IsNil()
		assert(t, "res", res, true)
	})

	t.Run("Set", func(t *testing.T) {
		defer expectPanic(t, "runtime error")
		x := Nil().assign(TypeMap)
		x.Set(String("x"), Int(42))
	})

	t.Run("Len", func(t *testing.T) {
		x := Nil().assign(TypeMap)
		res := x.Len()
		assert(t, "res", res, 0)
	})

	t.Run("Range", func(t *testing.T) {
		x := Nil().assign(TypeMap)
		next := x.Range()
		res := fmt.Sprint(next())
		assert(t, "res", res, "nil nil false")
	})

	t.Run("Delete", func(t *testing.T) {
		x := Nil().assign(TypeMap)
		x.Delete(String("x"))
	})

	t.Run("eqNil", func(t *testing.T) {
		x := Nil().assign(TypeMap)
		res := x.Equals(Nil())
		assert(t, "res", res, true)
	})
}

func TestNumericOps(t *testing.T) {
	test := []struct {
		name string
		a, b Value
		op   string
		want Value
	}{
		{"Float64.Add", Float64(40), Float64(2), "Add", Float64(42)},
		{"Float64.Sub", Float64(44), Float64(2), "Sub", Float64(42)},
		{"Float64.Mul", Float64(21), Float64(2), "Mul", Float64(42)},
		{"Float64.Div", Float64(84), Float64(2), "Div", Float64(42)},
		{"Float64.Mod", Float64(85), Float64(43), "Mod", Float64(42)},
		{"Float64.BitLsh", Float64(21), Float64(1), "BitLsh", Float64(42)},
		{"Float64.BitRsh", Float64(84), Float64(1), "BitRsh", Float64(42)},
		{"Float64.BitAnd", Float64(106), Float64(63), "BitAnd", Float64(42)},
		{"Float64.BitOr", Float64(40), Float64(34), "BitOr", Float64(42)},
		{"Float64.BitXor", Float64(8), Float64(34), "BitXor", Float64(42)},
		{"Float64.cast", newUntypedInt(42), Float64(0), "cast", Float64(42)},

		{"Int.Add", Int(40), Int(2), "Add", Int(42)},
		{"Int.Sub", Int(44), Int(2), "Sub", Int(42)},
		{"Int.Mul", Int(21), Int(2), "Mul", Int(42)},
		{"Int.Div", Int(84), Int(2), "Div", Int(42)},
		{"Int.Mod", Int(85), Int(43), "Mod", Int(42)},
		{"Int.BitLsh", Int(21), Int(1), "BitLsh", Int(42)},
		{"Int.BitRsh", Int(84), Int(1), "BitRsh", Int(42)},
		{"Int.BitAnd", Int(106), Int(63), "BitAnd", Int(42)},
		{"Int.BitOr", Int(40), Int(34), "BitOr", Int(42)},
		{"Int.BitXor", Int(8), Int(34), "BitXor", Int(42)},
		{"Int.cast", newUntypedInt(42), Int(0), "cast", Int(42)},

		{"Int32.Add", Int32(40), Int32(2), "Add", Int32(42)},
		{"Int32.Sub", Int32(44), Int32(2), "Sub", Int32(42)},
		{"Int32.Mul", Int32(21), Int32(2), "Mul", Int32(42)},
		{"Int32.Div", Int32(84), Int32(2), "Div", Int32(42)},
		{"Int32.Mod", Int32(85), Int32(43), "Mod", Int32(42)},
		{"Int32.BitLsh", Int32(21), Int32(1), "BitLsh", Int32(42)},
		{"Int32.BitRsh", Int32(84), Int32(1), "BitRsh", Int32(42)},
		{"Int32.BitAnd", Int32(106), Int32(63), "BitAnd", Int32(42)},
		{"Int32.BitOr", Int32(40), Int32(34), "BitOr", Int32(42)},
		{"Int32.BitXor", Int32(8), Int32(34), "BitXor", Int32(42)},
		{"Int32.cast", newUntypedInt(42), Int32(0), "cast", Int32(42)},

		{"Uint32.Add", Uint32(40), Uint32(2), "Add", Uint32(42)},
		{"Uint32.Sub", Uint32(44), Uint32(2), "Sub", Uint32(42)},
		{"Uint32.Mul", Uint32(21), Uint32(2), "Mul", Uint32(42)},
		{"Uint32.Div", Uint32(84), Uint32(2), "Div", Uint32(42)},
		{"Uint32.Mod", Uint32(85), Uint32(43), "Mod", Uint32(42)},
		{"Uint32.BitLsh", Uint32(21), Uint32(1), "BitLsh", Uint32(42)},
		{"Uint32.BitRsh", Uint32(84), Uint32(1), "BitRsh", Uint32(42)},
		{"Uint32.BitAnd", Uint32(106), Uint32(63), "BitAnd", Uint32(42)},
		{"Uint32.BitOr", Uint32(40), Uint32(34), "BitOr", Uint32(42)},
		{"Uint32.BitXor", Uint32(8), Uint32(34), "BitXor", Uint32(42)},
		{"Uint32.cast", newUntypedInt(42), Uint32(0), "cast", Uint32(42)},

		{"Uint.Add", Uint(40), Uint(2), "Add", Uint(42)},
		{"Uint.Sub", Uint(44), Uint(2), "Sub", Uint(42)},
		{"Uint.Mul", Uint(21), Uint(2), "Mul", Uint(42)},
		{"Uint.Div", Uint(84), Uint(2), "Div", Uint(42)},
		{"Uint.Mod", Uint(85), Uint(43), "Mod", Uint(42)},
		{"Uint.BitLsh", Uint(21), Uint(1), "BitLsh", Uint(42)},
		{"Uint.BitRsh", Uint(84), Uint(1), "BitRsh", Uint(42)},
		{"Uint.BitAnd", Uint(106), Uint(63), "BitAnd", Uint(42)},
		{"Uint.BitOr", Uint(40), Uint(34), "BitOr", Uint(42)},
		{"Uint.BitXor", Uint(8), Uint(34), "BitXor", Uint(42)},
		{"Uint.cast", newUntypedInt(42), Uint(0), "cast", Uint(42)},

		{"Int8.Add", Int8(40), Int8(2), "Add", Int8(42)},
		{"Int8.Sub", Int8(44), Int8(2), "Sub", Int8(42)},
		{"Int8.Mul", Int8(21), Int8(2), "Mul", Int8(42)},
		{"Int8.Div", Int8(84), Int8(2), "Div", Int8(42)},
		{"Int8.Mod", Int8(85), Int8(43), "Mod", Int8(42)},
		{"Int8.BitLsh", Int8(21), Int8(1), "BitLsh", Int8(42)},
		{"Int8.BitRsh", Int8(84), Int8(1), "BitRsh", Int8(42)},
		{"Int8.BitAnd", Int8(106), Int8(63), "BitAnd", Int8(42)},
		{"Int8.BitOr", Int8(40), Int8(34), "BitOr", Int8(42)},
		{"Int8.BitXor", Int8(8), Int8(34), "BitXor", Int8(42)},
		{"Int8.cast", newUntypedInt(42), Int8(0), "cast", Int8(42)},

		{"Byte.Add", Byte(40), Byte(2), "Add", Byte(42)},
		{"Byte.Sub", Byte(44), Byte(2), "Sub", Byte(42)},
		{"Byte.Mul", Byte(21), Byte(2), "Mul", Byte(42)},
		{"Byte.Div", Byte(84), Byte(2), "Div", Byte(42)},
		{"Byte.Mod", Byte(85), Byte(43), "Mod", Byte(42)},
		{"Byte.BitLsh", Byte(21), Byte(1), "BitLsh", Byte(42)},
		{"Byte.BitRsh", Byte(84), Byte(1), "BitRsh", Byte(42)},
		{"Byte.BitAnd", Byte(106), Byte(63), "BitAnd", Byte(42)},
		{"Byte.BitOr", Byte(40), Byte(34), "BitOr", Byte(42)},
		{"Byte.BitXor", Byte(8), Byte(34), "BitXor", Byte(42)},
		{"Byte.cast", newUntypedInt(42), Byte(0), "cast", Byte(42)},

		{"Uint8.Add", Uint8(40), Uint8(2), "Add", Uint8(42)},
		{"Uint8.Sub", Uint8(44), Uint8(2), "Sub", Uint8(42)},
		{"Uint8.Mul", Uint8(21), Uint8(2), "Mul", Uint8(42)},
		{"Uint8.Div", Uint8(84), Uint8(2), "Div", Uint8(42)},
		{"Uint8.Mod", Uint8(85), Uint8(43), "Mod", Uint8(42)},
		{"Uint8.BitLsh", Uint8(21), Uint8(1), "BitLsh", Uint8(42)},
		{"Uint8.BitRsh", Uint8(84), Uint8(1), "BitRsh", Uint8(42)},
		{"Uint8.BitAnd", Uint8(106), Uint8(63), "BitAnd", Uint8(42)},
		{"Uint8.BitOr", Uint8(40), Uint8(34), "BitOr", Uint8(42)},
		{"Uint8.BitXor", Uint8(8), Uint8(34), "BitXor", Uint8(42)},
		{"Uint8.cast", newUntypedInt(42), Uint8(0), "cast", Uint8(42)},

		{"untypedInt.Add", newUntypedInt(40), newUntypedInt(2), "Add", newUntypedInt(42)},
		{"untypedInt.Sub", newUntypedInt(44), newUntypedInt(2), "Sub", newUntypedInt(42)},
		{"untypedInt.Mul", newUntypedInt(21), newUntypedInt(2), "Mul", newUntypedInt(42)},
		{"untypedInt.Div", newUntypedInt(84), newUntypedInt(2), "Div", newUntypedInt(42)},
		{"untypedInt.Mod", newUntypedInt(85), newUntypedInt(43), "Mod", newUntypedInt(42)},
		{"untypedInt.BitLsh", newUntypedInt(21), newUntypedInt(1), "BitLsh", newUntypedInt(42)},
		{"untypedInt.BitRsh", newUntypedInt(84), newUntypedInt(1), "BitRsh", newUntypedInt(42)},
		{"untypedInt.BitAnd", newUntypedInt(106), newUntypedInt(63), "BitAnd", newUntypedInt(42)},
		{"untypedInt.BitOr", newUntypedInt(40), newUntypedInt(34), "BitOr", newUntypedInt(42)},
		{"untypedInt.BitXor", newUntypedInt(8), newUntypedInt(34), "BitXor", newUntypedInt(42)},
		{"untypedInt.cast", newUntypedInt(42), Nil(), "cast", Int(42)},

		{"untypedInt.cast/float64", newUntypedInt(42), Float64(0), "cast", Float64(42)},
		{"untypedInt.cast/int", newUntypedInt(42), Int(0), "cast", Int(42)},
		{"untypedInt.cast/uint32", newUntypedInt(42), Uint32(0), "cast", Uint32(42)},
		{"untypedInt.cast/int8", newUntypedInt(42), Int8(0), "cast", Int8(42)},
		{"untypedInt.cast/byte", newUntypedInt(42), Byte(0), "cast", Byte(42)},

		{"untypedInt.convert/float64", newUntypedInt(42), Float64(0), "convert", Float64(42)},
		{"untypedInt.convert/int", newUntypedInt(42), Int(0), "convert", Int(42)},
		{"untypedInt.convert/uint32", newUntypedInt(42), Uint32(0), "convert", Uint32(42)},
		{"untypedInt.convert/int8", newUntypedInt(42), Int8(0), "convert", Int8(42)},
		{"untypedInt.convert/byte", newUntypedInt(42), Byte(0), "convert", Byte(42)},

		////////////////////////////////////////////////////////////////////////////

		// {"float64Min.convert/Uint8", Float64(float64(math.MaxFloat64)), Uint8(0), "convert", Uint8(0)},
		{"float64Max.convert/Uint8", Float64(float64(-math.MaxFloat64)), Uint8(0), "convert", Uint8(0)},
		{"float64Ex.convert/Uint8", Float64(float64(-1)), Uint8(0), "convert", Uint8(255)},
		{"int32Min.convert/Uint8", Int32(int32(math.MinInt32)), Uint8(0), "convert", Uint8(0)},
		{"int32Max.convert/Uint8", Int32(int32(math.MaxInt32)), Uint8(0), "convert", Uint8(255)},
		{"int32Ex.convert/Uint8", Int32(int32(-1)), Uint8(0), "convert", Uint8(255)},
		{"uint32Min.convert/Uint8", Uint32(uint32(0)), Uint8(0), "convert", Uint8(0)},
		{"uint32Max.convert/Uint8", Uint32(uint32(math.MaxUint32)), Uint8(0), "convert", Uint8(255)},
		{"uint32Ex.convert/Uint8", Uint32(uint32(math.MaxInt32 + 1)), Uint8(0), "convert", Uint8(0)},
		{"int8Min.convert/Uint8", Int8(int8(math.MinInt8)), Uint8(0), "convert", Uint8(128)},
		{"int8Max.convert/Uint8", Int8(int8(math.MaxInt8)), Uint8(0), "convert", Uint8(127)},
		{"int8Ex.convert/Uint8", Int8(int8(-1)), Uint8(0), "convert", Uint8(255)},
		{"uint8Min.convert/Uint8", Uint8(uint8(0)), Uint8(0), "convert", Uint8(0)},
		{"uint8Max.convert/Uint8", Uint8(uint8(math.MaxUint8)), Uint8(0), "convert", Uint8(255)},
		{"uint8Ex.convert/Uint8", Uint8(uint8(math.MaxInt8 + 1)), Uint8(0), "convert", Uint8(128)},

		// {"float64Min.convert/Int8", Float64(float64(math.MaxFloat64)), Int8(0), "convert", Int8(0)},
		{"float64Max.convert/Int8", Float64(float64(-math.MaxFloat64)), Int8(0), "convert", Int8(0)},
		{"float64Ex.convert/Int8", Float64(float64(-1)), Int8(0), "convert", Int8(-1)},
		{"int32Min.convert/Int8", Int32(int32(math.MinInt32)), Int8(0), "convert", Int8(0)},
		{"int32Max.convert/Int8", Int32(int32(math.MaxInt32)), Int8(0), "convert", Int8(-1)},
		{"int32Ex.convert/Int8", Int32(int32(-1)), Int8(0), "convert", Int8(-1)},
		{"uint32Min.convert/Int8", Uint32(uint32(0)), Int8(0), "convert", Int8(0)},
		{"uint32Max.convert/Int8", Uint32(uint32(math.MaxUint32)), Int8(0), "convert", Int8(-1)},
		{"uint32Ex.convert/Int8", Uint32(uint32(math.MaxInt32 + 1)), Int8(0), "convert", Int8(0)},
		{"int8Min.convert/Int8", Int8(int8(math.MinInt8)), Int8(0), "convert", Int8(-128)},
		{"int8Max.convert/Int8", Int8(int8(math.MaxInt8)), Int8(0), "convert", Int8(127)},
		{"int8Ex.convert/Int8", Int8(int8(-1)), Int8(0), "convert", Int8(-1)},
		{"uint8Min.convert/Int8", Uint8(uint8(0)), Int8(0), "convert", Int8(0)},
		{"uint8Max.convert/Int8", Uint8(uint8(math.MaxUint8)), Int8(0), "convert", Int8(-1)},
		{"uint8Ex.convert/Int8", Uint8(uint8(math.MaxInt8 + 1)), Int8(0), "convert", Int8(-128)},

		// {"float64Min.convert/Uint32", Float64(float64(math.MaxFloat64)), Uint32(0), "convert", Uint32(0)},
		{"float64Max.convert/Uint32", Float64(float64(-math.MaxFloat64)), Uint32(0), "convert", Uint32(0)},
		{"float64Ex.convert/Uint32", Float64(float64(-1)), Uint32(0), "convert", Uint32(4294967295)},
		{"int32Min.convert/Uint32", Int32(int32(math.MinInt32)), Uint32(0), "convert", Uint32(2147483648)},
		{"int32Max.convert/Uint32", Int32(int32(math.MaxInt32)), Uint32(0), "convert", Uint32(2147483647)},
		{"int32Ex.convert/Uint32", Int32(int32(-1)), Uint32(0), "convert", Uint32(4294967295)},
		{"uint32Min.convert/Uint32", Uint32(uint32(0)), Uint32(0), "convert", Uint32(0)},
		{"uint32Max.convert/Uint32", Uint32(uint32(math.MaxUint32)), Uint32(0), "convert", Uint32(4294967295)},
		{"uint32Ex.convert/Uint32", Uint32(uint32(math.MaxInt32 + 1)), Uint32(0), "convert", Uint32(2147483648)},
		{"int8Min.convert/Uint32", Int8(int8(math.MinInt8)), Uint32(0), "convert", Uint32(4294967168)},
		{"int8Max.convert/Uint32", Int8(int8(math.MaxInt8)), Uint32(0), "convert", Uint32(127)},
		{"int8Ex.convert/Uint32", Int8(int8(-1)), Uint32(0), "convert", Uint32(4294967295)},
		{"uint8Min.convert/Uint32", Uint8(uint8(0)), Uint32(0), "convert", Uint32(0)},
		{"uint8Max.convert/Uint32", Uint8(uint8(math.MaxUint8)), Uint32(0), "convert", Uint32(255)},
		{"uint8Ex.convert/Uint32", Uint8(uint8(math.MaxInt8 + 1)), Uint32(0), "convert", Uint32(128)},

		// {"float64Min.convert/Int32", Float64(float64(math.MaxFloat64)), Int32(0), "convert", Int32(-2147483648)},
		{"float64Max.convert/Int32", Float64(float64(-math.MaxFloat64)), Int32(0), "convert", Int32(-2147483648)},
		{"float64Ex.convert/Int32", Float64(float64(-1)), Int32(0), "convert", Int32(-1)},
		{"int32Min.convert/Int32", Int32(int32(math.MinInt32)), Int32(0), "convert", Int32(-2147483648)},
		{"int32Max.convert/Int32", Int32(int32(math.MaxInt32)), Int32(0), "convert", Int32(2147483647)},
		{"int32Ex.convert/Int32", Int32(int32(-1)), Int32(0), "convert", Int32(-1)},
		{"uint32Min.convert/Int32", Uint32(uint32(0)), Int32(0), "convert", Int32(0)},
		{"uint32Max.convert/Int32", Uint32(uint32(math.MaxUint32)), Int32(0), "convert", Int32(-1)},
		{"uint32Ex.convert/Int32", Uint32(uint32(math.MaxInt32 + 1)), Int32(0), "convert", Int32(-2147483648)},
		{"int8Min.convert/Int32", Int8(int8(math.MinInt8)), Int32(0), "convert", Int32(-128)},
		{"int8Max.convert/Int32", Int8(int8(math.MaxInt8)), Int32(0), "convert", Int32(127)},
		{"int8Ex.convert/Int32", Int8(int8(-1)), Int32(0), "convert", Int32(-1)},
		{"uint8Min.convert/Int32", Uint8(uint8(0)), Int32(0), "convert", Int32(0)},
		{"uint8Max.convert/Int32", Uint8(uint8(math.MaxUint8)), Int32(0), "convert", Int32(255)},
		{"uint8Ex.convert/Int32", Uint8(uint8(math.MaxInt8 + 1)), Int32(0), "convert", Int32(128)},

		{"float64Min.convert/Float64", Float64(float64(math.MaxFloat64)), Float64(0), "convert", Float64(math.MaxFloat64)},
		{"float64Max.convert/Float64", Float64(float64(-math.MaxFloat64)), Float64(0), "convert", Float64(-math.MaxFloat64)},
		{"float64Ex.convert/Float64", Float64(float64(-1)), Float64(0), "convert", Float64(-1)},
		{"int32Min.convert/Float64", Int32(int32(math.MinInt32)), Float64(0), "convert", Float64(-2.147483648e+09)},
		{"int32Max.convert/Float64", Int32(int32(math.MaxInt32)), Float64(0), "convert", Float64(2.147483647e+09)},
		{"int32Ex.convert/Float64", Int32(int32(-1)), Float64(0), "convert", Float64(-1)},
		{"uint32Min.convert/Float64", Uint32(uint32(0)), Float64(0), "convert", Float64(0)},
		{"uint32Max.convert/Float64", Uint32(uint32(math.MaxUint32)), Float64(0), "convert", Float64(4.294967295e+09)},
		{"uint32Ex.convert/Float64", Uint32(uint32(math.MaxInt32 + 1)), Float64(0), "convert", Float64(2.147483648e+09)},
		{"int8Min.convert/Float64", Int8(int8(math.MinInt8)), Float64(0), "convert", Float64(-128)},
		{"int8Max.convert/Float64", Int8(int8(math.MaxInt8)), Float64(0), "convert", Float64(127)},
		{"int8Ex.convert/Float64", Int8(int8(-1)), Float64(0), "convert", Float64(-1)},
		{"uint8Min.convert/Float64", Uint8(uint8(0)), Float64(0), "convert", Float64(0)},
		{"uint8Max.convert/Float64", Uint8(uint8(math.MaxUint8)), Float64(0), "convert", Float64(255)},
		{"uint8Ex.convert/Float64", Uint8(uint8(math.MaxInt8 + 1)), Float64(0), "convert", Float64(128)},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			var res Value
			switch tt.op {
			case "Add":
				res = tt.a.opAdd(tt.b)
			case "Sub":
				res = tt.a.opSub(tt.b)
			case "Mul":
				res = tt.a.opMul(tt.b)
			case "Div":
				res = tt.a.opDiv(tt.b)
			case "Mod":
				res = tt.a.opMod(tt.b)
			case "BitLsh":
				res = tt.a.opBitLsh(tt.b)
			case "BitRsh":
				res = tt.a.opBitRsh(tt.b)
			case "BitAnd":
				res = tt.a.opBitAnd(tt.b)
			case "BitOr":
				res = tt.a.opBitOr(tt.b)
			case "BitXor":
				res = tt.a.opBitXor(tt.b)
			case "cast":
				res = tt.a.assign(tt.b.t)
			case "convert":
				res = tt.a.convert(tt.b.t)
			default:
				t.Fatalf("unknown op: %v", tt.op)
			}
			assert(t, "t", res.t, tt.want.t)
			assert(t, "num", res.num, tt.want.num)
			switch tt.want.t {
			case TypeFloat64:
				got, want := res.Float64(), res.num
				assert(t, "eq", got, want)
			case TypeInt32:
				got, want := res.Int(), int(res.num)
				assert(t, "eq", got, want)
				gotN, wantN := res.Int32(), int32(res.num)
				assert(t, "eq32", gotN, wantN)
			case TypeUint32:
				got, want := res.Uint(), uint(res.num)
				assert(t, "eq", got, want)
				gotN, wantN := res.Uint32(), uint32(res.num)
				assert(t, "eq", gotN, wantN)
			case TypeInt8:
				got, want := res.Int8(), int8(res.num)
				assert(t, "eq", got, want)
			case TypeUint8:
				got, want := res.Byte(), byte(res.num)
				assert(t, "eq", got, want)
				got8, want8 := res.Uint8(), uint8(res.num)
				assert(t, "eq8", got8, want8)
			case untypedInt:
				got, want := res.Int(), int(res.num)
				assert(t, "eq", got, want)
			default:
				t.Fatalf("unknown t: %v", tt.want.t)
			}
		})
	}
}

func TestBool(t *testing.T) {
	t.Run("Eq", func(t *testing.T) {
		tests := []struct {
			name string
			a, b bool
			want bool
		}{
			{"trueEqTrue", true, true, true},
			{"trueNeqFalse", true, false, false},
			{"falseEqFalse", false, false, true},
			{"falseNeqTrue", false, true, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				a := Bool(tt.a)
				b := Bool(tt.b)
				res := a.opEq(b).Bool()
				assert(t, "res", res, tt.want)
			})
		}
	})
}
