package goatlang

import (
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

/**
On performance ...

Best   v.value.(structT).GetIndex      cast to concrete type, access method
...    v.value.GetIndex                access interface method
Worst  v.value.(getSetIndex).GetIndex  cast to interface type, access interface method

Similar:
	if t, ok := v.value.(struct); ok { ... }  type assert, safer, more idiomatic
	if v.t == TypeStruct { ... }              type cast, unsafe, less idiomatic

On pointers:
	v.value.(*structT) is faster than v.value.(structT)
*/

type Type int

const (
	TypeNil          = Type(0b00000000)
	untypedInt       = Type(0b00000001)
	isNumericMask    = Type(0b00000011)
	typedNumberMask  = Type(0b00000010)
	signedNumberMask = Type(0b00010000)
	TypeUint8        = Type(0b00000011)
	TypeInt8         = Type(0b00010011)
	TypeUint32       = Type(0b00000111)
	TypeInt32        = Type(0b00010111)
	TypeFloat64      = Type(0b00011111)
	numericBitsMask  = Type(0b00011111)
	typeType         = Type(0b00000100) // hidden non-numeric
	typeNext         = Type(0b00001000) // hidden non-numeric
	TypeBool         = Type(0b00100000)
	TypeString       = Type(0b01000000)
	TypeObject       = Type(0b01100000)
	nillableMin      = Type(0b10000000)
	TypeSlice        = Type(0b10000000)
	TypeMap          = Type(0b10100000)
	TypeFunc         = Type(0b11000000)
	TypeStruct       = Type(0b11100000)

	typeMask  = Type(0xff)
	typeShift = 8
)

func newType(t Type) Value { return Value{t: typeType, num: float64(t)} }

var typeToString = map[Type]string{
	TypeNil:     "any",
	untypedInt:  "number",
	TypeBool:    "bool",
	TypeUint8:   "uint8",
	TypeInt32:   "int32",
	TypeFloat64: "float64",
	TypeUint32:  "uint32",
	TypeInt8:    "int8",
	TypeString:  "string",
	TypeSlice:   "slice",
	TypeMap:     "map",
	TypeFunc:    "func",
	TypeStruct:  "struct",
	TypeObject:  "object",
	typeNext:    "next",
	typeType:    "type",
}

// func (t Type) String() string {
// 	switch t.base() {
// 	case TypeSlice:
// 		v := t.value()
// 		return "[]" + v.String()
// 	case TypeMap:
// 		k, v := t.pair()
// 		return "map[" + k.String() + "]" + v.String()
// 	}
// 	return typeToString[t.base()]
// }

func (t Type) str(g *lookup) string {
	switch t.base() {
	case TypeSlice:
		v := t.value()
		return "[]" + v.str(g)
	case TypeMap:
		k, v := t.pair()
		return "map[" + k.str(g) + "]" + v.str(g)
	case TypeStruct:
		v := t.value()
		if v > 0 {
			return g.Key(int(v))
		}
		return "struct"
	}
	return typeToString[t.base()]
}

func (t Type) base() Type {
	return t & typeMask
}

func (t Type) value() Type {
	return t >> typeShift
}

func (t Type) pair() (Type, Type) {
	return (t >> typeShift) & typeMask, (t >> (typeShift * 2))
}

func (t Type) isSafeStr() bool {
	switch t.base() {
	case TypeSlice, TypeMap, TypeStruct:
		return false
	}
	return true
}

func sliceType(value Type) Type {
	return value<<typeShift | TypeSlice
}

func mapType(key, value Type) Type {
	return value<<(typeShift*2) | key<<typeShift | TypeMap
}

func structType(value Type) Type {
	return value<<typeShift | TypeStruct
}

var _ Object = &Value{}

type Value struct {
	t     Type
	num   float64
	value Object
}

type Object interface {
	// String() string

	Get(k Value) (Value, bool)
	Set(k, v Value)
	Len() int
	Range() func() (Value, Value, bool)

	Append(items ...Value) Value
	Delete(k Value)
	Slice(i, j int) Value

	GetAttr(k string) Value
	SetAttr(k string, v Value)
}

func (v Value) Get(key Value) (Value, bool) {
	if v.value != nil || v.t.base() == TypeSlice {
		return v.value.Get(key)
	}
	_, vt := v.t.pair()
	return newZero(vt), false
}
func (v Value) Set(key, value Value) { v.value.Set(key, value) }
func (v Value) Len() int {
	if v.value != nil {
		return v.value.Len()
	} else {
		return 0
	}
}
func (v Value) Range() func() (Value, Value, bool) {
	if v.value != nil {
		return v.value.Range()
	} else {
		return func() (Value, Value, bool) { return Nil(), Nil(), false }
	}
}
func (v Value) Append(items ...Value) Value {
	if v.value != nil {
		return v.value.Append(items...)
	} else {
		return newSlice(v.t.value(), items)
	}
}
func (v Value) Delete(key Value) {
	if v.value != nil {
		v.value.Delete(key)
	}
}
func (v Value) Slice(i, j int) Value {
	if v.value != nil || i != 0 || j != 0 {
		return v.value.Slice(i, j)
	}
	return newSlice(v.t.value(), nil)
}
func (v Value) GetAttr(key string) Value        { return v.value.GetAttr(key) }
func (v Value) SetAttr(key string, value Value) { v.value.SetAttr(key, value) }

func Wrap(o Object) Value { return Value{t: TypeObject, value: o} }

func (v Value) Unwrap() Object {
	if v.t == TypeObject {
		return v.value
	}
	return nil
}

type funcT struct {
	Object
	Args, Rets   int
	Variadic     bool
	VariadicType Type
	Value        func(v *VM)
}

func (v Value) getFunc() *funcT {
	return v.value.(*funcT)
}

func Nil() Value { return Value{} }

func Float64(v float64) Value { return Value{t: TypeFloat64, num: v} }
func Int(v int) Value         { return Value{t: TypeInt32, num: float64(int32(v))} }
func Int32(v int32) Value     { return Value{t: TypeInt32, num: float64(v)} }
func Uint(v uint) Value       { return Value{t: TypeUint32, num: float64(uint32(v))} }
func Uint32(v uint32) Value   { return Value{t: TypeUint32, num: float64(v)} }
func Int8(v int8) Value       { return Value{t: TypeInt8, num: float64(v)} }
func Byte(v byte) Value       { return Value{t: TypeUint8, num: float64(v)} }
func Uint8(v uint8) Value     { return Value{t: TypeUint8, num: float64(v)} }

func (v Value) Float64() float64 { return v.num }
func (v Value) Int() int         { return int(v.num) }
func (v Value) Int32() int32     { return int32(v.num) }
func (v Value) Uint() uint       { return uint(v.num) }
func (v Value) Uint32() uint32   { return uint32(v.num) }
func (v Value) Int8() int8       { return int8(v.num) }
func (v Value) Byte() byte       { return byte(v.num) }
func (v Value) Uint8() uint8     { return uint8(v.num) }

type safeStr interface {
	SafeStr() string
}

func (v Value) safeStr() string {
	if v, ok := v.value.(safeStr); ok {
		return v.SafeStr()
	}
	return v.String()
}
func newZero(t Type) Value {
	switch t {
	case TypeString:
		return String("")
	default:
		return Value{t: t}
	}
}
func newUntypedInt(v int) Value {
	return Value{t: untypedInt, num: float64(v)}
}

func newFunc(argc, rets int, f func(vm *VM)) Value {
	variadic := argc < 0
	if variadic {
		argc = -argc
	}
	return Value{t: TypeFunc, value: &funcT{Args: argc, Rets: rets, Variadic: variadic, Value: f}}
}

func NewFunc[F func(vm *VM) | func(vm *VM) Value | func(vm *VM, args []Value) | func(v *VM, args []Value) Value | func(v *VM, args []Value) []Value | func(v *VM, args []Value, vargs ...Value) []Value](argc, rets int, fnc F) (res Value) {
	switch f := any(fnc).(type) {
	case func(vm *VM): // 0->0
		res = newFunc(argc, rets, f)
	case func(vm *VM) Value: // 0->1
		res = newFunc(argc, rets, func(vm *VM) {
			vm.stack = append(vm.stack, f(vm))
		})
	case func(vm *VM, args []Value): // N->0
		res = newFunc(argc, rets, func(vm *VM) {
			i := len(vm.stack) - argc
			a := vm.stack[i:]
			vm.stack = vm.stack[:i]
			f(vm, a)
		})
	case func(v *VM, args []Value) Value: // N->1
		res = newFunc(argc, rets, func(vm *VM) {
			i := len(vm.stack) - argc
			a := vm.stack[i:]
			vm.stack = vm.stack[:i]
			vm.stack = append(vm.stack, f(vm, a))
		})
	case func(v *VM, args []Value) []Value: // N->M
		res = newFunc(argc, rets, func(vm *VM) {
			i := len(vm.stack) - argc
			a := vm.stack[i:]
			vm.stack = vm.stack[:i]
			vm.stack = append(vm.stack, f(vm, a)...)
		})
	case func(v *VM, args []Value, vargs ...Value) []Value: // N,...->M
		res = newFunc(-argc, rets, func(vm *VM) {
			i := len(vm.stack) - argc
			a := vm.stack[i:]
			vm.stack = vm.stack[:i]
			vm.stack = append(vm.stack, f(vm, a[:argc-1], a[argc-1].data()...)...)
		})
	}
	return res
}

func Bool(v bool) Value {
	if v {
		return Value{t: TypeBool, num: 1}
	}
	return Value{t: TypeBool, num: 0}
}

func (v Value) Type() Type { return v.t.base() }

func (v Value) String() string {
	switch v.t.base() {
	case TypeNil:
		return "nil"
	case TypeBool:
		return fmt.Sprint(v.Bool())
	case TypeInt32, TypeUint32, TypeInt8, TypeUint8, untypedInt:
		return fmt.Sprint(int(v.num))
	case TypeFloat64:
		return fmt.Sprint(v.num)
	case TypeString:
		return string(v.value.(stringT))
	case TypeStruct, TypeFunc:
		if v.value == nil {
			return "nil"
		}
	case TypeSlice:
		if v.value == nil {
			return "[]"
		}
	case TypeMap:
		if v.value == nil {
			return "map[]"
		}
	}
	return fmt.Sprint(v.value)
}

func (v Value) Bool() bool {
	return v.num != 0
}

// func (v Value) copy() Value { return v }

// func (v Value) copy() Value {
// 	res := v
// 	if v.t != TypeStruct || v.value == nil {
// 		return res
// 	}
// 	s := v.value.(*structT)
// 	res.value = &structT{
// 		Lookup:  s.Lookup,
// 		Fields:  s.Fields.Copy(),
// 		Methods: s.Methods,
// 	}
// 	return res
// }

func (v Value) assign(t Type) Value {
	switch {
	case v.t == t:
		return v
	case v.t == untypedInt:
		switch t {
		case TypeFloat64:
			return Value{t: t, num: v.num}
		case TypeInt32:
			return Value{t: t, num: float64(int32(v.num))}
		case TypeUint32:
			return Value{t: t, num: float64(uint32(v.num))}
		case TypeInt8:
			return Value{t: t, num: float64(int8(v.num))}
		case TypeUint8:
			return Value{t: t, num: float64(uint8(v.num))}
		default:
			return Value{t: TypeInt32, num: float64(int32(v.num))}
		}
	case v.t != TypeNil:
		return v
	case t >= nillableMin:
		return Value{t: t}
	default:
		return Value{}
	}
}

func mixType(a, b Type) Type {
	return a | b
}

func (v Value) opAdd(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: v.num + b.num}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) + int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) + uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) + int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) + byte(b.num))}
	case TypeString:
		return String(string(v.value.(stringT) + b.value.(stringT)))
	default:
		return Value{t: untypedInt, num: v.num + b.num}
	}
}
func (v Value) opSub(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: v.num - b.num}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) - int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) - uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) - int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) - byte(b.num))}
	default:
		return Value{t: untypedInt, num: v.num - b.num}
	}
}
func (v Value) opMul(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: v.num * b.num}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) * int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) * uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) * int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) * byte(b.num))}
	default:
		return Value{t: untypedInt, num: v.num * b.num}
	}
}
func (v Value) opDiv(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: v.num / b.num}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) / int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) / uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) / int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) / byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) / int(b.num))}
	}
}
func (v Value) opMod(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: float64(int(v.num) % int(b.num))}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) % int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) % uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) % int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) % byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) % int(b.num))}
	}
}
func (v Value) opBitLsh(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: float64(int(v.num) << int(b.num))}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) << int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) << uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) << int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) << byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) << int(b.num))}
	}
}
func (v Value) opBitRsh(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: float64(int(v.num) >> int(b.num))}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) >> int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) >> uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) >> int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) >> byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) >> int(b.num))}
	}
}
func (v Value) opBitAnd(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: float64(int(v.num) & int(b.num))}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) & int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) & uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) & int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) & byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) & int(b.num))}
	}
}

func (v Value) opBitOr(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: float64(int(v.num) | int(b.num))}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) | int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) | uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) | int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) | byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) | int(b.num))}
	}
}
func (v Value) opBitXor(b Value) Value {
	t := mixType(v.t, b.t)
	switch t {
	case TypeFloat64:
		return Value{t: t, num: float64(int(v.num) ^ int(b.num))}
	case TypeInt32:
		return Value{t: t, num: float64(int32(v.num) ^ int32(b.num))}
	case TypeUint32:
		return Value{t: t, num: float64(uint32(v.num) ^ uint32(b.num))}
	case TypeInt8:
		return Value{t: t, num: float64(int8(v.num) ^ int8(b.num))}
	case TypeUint8:
		return Value{t: t, num: float64(byte(v.num) ^ byte(b.num))}
	default:
		return Value{t: untypedInt, num: float64(int(v.num) ^ int(b.num))}
	}
}

func (v Value) opLt(b Value) Value {
	if v.t != TypeString {
		return Bool(v.num < b.num)
	}
	return Bool(v.value.(stringT) < b.value.(stringT))
}

func (v Value) opLte(b Value) Value {
	if v.t != TypeString {
		return Bool(v.num <= b.num)
	}
	return Bool(v.value.(stringT) <= b.value.(stringT))
}

func (v Value) opNeq(b Value) Value { return Bool(!v.Equals(b)) }

func (v Value) Equals(b Value) bool {
	switch {
	case v.t == TypeBool:
		return v.num == b.num
	case (v.t & TypeFloat64) > 0:
		return v.num == b.num
	case v.t == TypeString:
		return v.value.(stringT) == b.value.(stringT)
	case v.t.base() == TypeStruct, v.t == TypeFunc:
		return (b.t == TypeNil && v.value == nil) || v.value == b.value
	case v.t == TypeNil && b.t == TypeNil:
		return true
	case v.t.base() == TypeSlice && b.t == TypeNil:
		return v.value == nil
	case v.t.base() == TypeMap && b.t == TypeNil:
		return v.value == nil
	default:
		return false
	}
}
func (v Value) opEq(b Value) Value { return Bool(v.Equals(b)) }

func (v Value) convert(t Type) (res Value) {
	switch t {
	case TypeUint8:
		return Uint8(uint8(int64(v.num)))
	case TypeInt8:
		return Int8(int8(int64(v.num)))
	case TypeInt32:
		if v.t == TypeFloat64 {
			return Int32(int32(v.num))
		}
		return Int32(int32(int64(v.num)))
	case TypeUint32:
		return Uint32(uint32(v.num))
	case TypeFloat64:
		return Float64(v.num)
	case TypeString:
		if v.t == TypeString {
			return v
		} else if v.t&isNumericMask != 0 {
			return String(string(rune(v.num)))
		}
		data := v.data()
		b := make([]byte, len(data))
		for k, v := range data {
			b[k] = byte(v.num)
		}
		return String(string(b))
	case TypeSlice:
		if v.t.base() == TypeSlice {
			return v
		}
		data := []byte(v.String())
		s := make([]Value, len(data))
		for k, v := range data {
			s[k] = Byte(v)
		}
		return newSlice(TypeUint8, s)
	default:
		return Value{}
	}
}

func (v Value) IsNil() bool { return v.t == TypeNil }

func (v Value) data() []Value {
	if t, ok := v.value.(*sliceT); ok {
		return t.data
	}
	res := make([]Value, v.Len())
	next := v.Range()
	for {
		kk, vv, ok := next()
		if !ok {
			break
		}
		res[kk.Int()] = vv
	}
	return res
}

func (v Value) getIndex(vm *VM, idx int) Value {
	if t, ok := v.value.(*structT); ok {
		return t.GetIndex(idx)
	}
	return v.value.GetAttr(vm.globals.Key(idx))
}

func (v Value) setIndex(vm *VM, idx int, val Value) {
	if t, ok := v.value.(*structT); ok {
		t.SetIndex(idx, val)
		return
	}
	v.value.SetAttr(vm.globals.Key(idx), val)
}

type stringT string

func String(v string) Value {
	return Value{t: TypeString, value: stringT(v)}
}

func (s stringT) Get(a Value) (Value, bool) { return Int(int(s[a.Int()])), true }
func (s stringT) Set(k, v Value)            { panic("unsupported") }
func (s stringT) Len() int                  { return len(s) }
func (s stringT) Range() func() (Value, Value, bool) {
	var r []rune
	for _, v := range s {
		r = append(r, v)
	}
	n := 0
	return func() (Value, Value, bool) {
		if n >= len(r) {
			return Nil(), Nil(), false
		}
		k, v := Int(n), r[n]
		n++
		return k, Int32(v), true
	}
}
func (s stringT) Append(items ...Value) Value { panic("unsupported") }
func (s stringT) Delete(k Value)              { panic("unsupported") }
func (s stringT) Slice(i, j int) Value        { return String(string(s[i:j])) }
func (s stringT) GetAttr(k string) Value      { panic("unsupported") }
func (s stringT) SetAttr(k string, v Value)   { panic("unsupported") }

type nextT struct {
	Object
	next func() (Value, Value, bool)
}

func (v Value) next() (Value, Value, bool) {
	return v.value.(*nextT).next()
}

type sliceT struct {
	Object
	valueType Type
	data      []Value
}

func (s *sliceT) Len() int { return len(s.data) }

func (s *sliceT) Get(k Value) (Value, bool) {
	return s.data[k.Int()], true
}

func (s *sliceT) Slice(i, j int) Value {
	return newSlice(s.valueType, s.data[i:j])
}

func (s *sliceT) Set(k, v Value) {
	s.data[k.Int()] = v.assign(s.valueType)
}

func (s *sliceT) Range() func() (Value, Value, bool) {
	r := s.data
	n := 0
	return func() (Value, Value, bool) {
		if n >= len(r) {
			return Nil(), Nil(), false
		}
		k, v := Int(n), r[n]
		n++
		return k, v, true
	}
}

func (s *sliceT) Append(items ...Value) Value {
	return NewSlice(s.valueType, append(s.data, items...))
}

func (s *sliceT) String() string {
	var p []string
	for _, v := range s.data {
		p = append(p, v.safeStr())
	}
	return "[" + strings.Join(p, " ") + "]"
}

func (s *sliceT) SafeStr() string {
	var p []string
	for _, v := range s.data {
		if !v.t.isSafeStr() {
			return "[...]"
		}
		p = append(p, v.safeStr())
	}
	return "[" + strings.Join(p, " ") + "]"
}

func NewSlice(valueType Type, data []Value) Value {
	for i, v := range data {
		data[i] = v.assign(valueType)
	}
	return newSlice(valueType, data)
}

func newSlice(valueType Type, data []Value) Value {
	return Value{t: sliceType(valueType), value: &sliceT{valueType: valueType, data: data}}
}

func NewMap(keyType, valueType Type, in []Value) Value {
	if keyType == TypeString {
		return newStringMap(keyType, valueType, in)
	}
	return newNumericMap(keyType, valueType, in)
}

func newStringMap(keyType, valueType Type, in []Value) Value {
	m := &stringMap{valueType: valueType, data: map[string]Value{}}
	m.keys = make([]string, len(in)/2)
	for i := 0; i < len(in); i += 2 {
		k, v := string(in[i].value.(stringT)), in[i+1]
		m.keys[i/2] = k
		m.data[k] = v.assign(valueType)
	}
	return Value{t: mapType(keyType, valueType), value: m}
}

type stringMap struct {
	Object
	valueType Type
	data      map[string]Value
	keys      []string
}

func (m *stringMap) Len() int { return len(m.data) }

func (m *stringMap) Get(k Value) (Value, bool) {
	v, ok := m.data[string(k.value.(stringT))]
	if !ok {
		return newZero(m.valueType), false
	}
	return v, true
}

func (m *stringMap) Set(k, v Value) {
	key := string(k.value.(stringT))
	if _, ok := m.data[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.data[key] = v.assign(m.valueType)
}

func (m *stringMap) Delete(k Value) {
	delete(m.data, string(k.value.(stringT)))
	if len(m.data) >= (len(m.keys) >> 1) {
		return
	}
	m.keys = maps.Keys(m.data)
}

func (m *stringMap) Range() func() (Value, Value, bool) {
	r := m.keys
	n := 0
	return func() (Value, Value, bool) {
		for n < len(r) {
			k := r[n]
			v, ok := m.data[k]
			n++
			if ok {
				return String(k), v, true
			}
		}
		return Nil(), Nil(), false
	}
}

func (m *stringMap) String() string {
	var p []string
	for k, v := range m.data {
		p = append(p, k+":"+v.safeStr())
	}
	return "map[" + strings.Join(p, " ") + "]"
}

func (m *stringMap) SafeStr() string {
	var p []string
	for k, v := range m.data {
		if !v.t.isSafeStr() {
			return "map[...]"
		}
		p = append(p, k+":"+v.safeStr())
	}
	return "map[" + strings.Join(p, " ") + "]"
}

type numericMap struct {
	Object
	keyType   Type
	valueType Type
	data      map[float64]Value
	keys      []float64
}

func newNumericMap(keyType, valueType Type, in []Value) Value {
	m := &numericMap{keyType: keyType, valueType: valueType, data: map[float64]Value{}}
	m.keys = make([]float64, len(in)/2)
	for i := 0; i < len(in); i += 2 {
		k, v := in[i].num, in[i+1]
		m.keys[i/2] = k
		m.data[k] = v.assign(valueType)
	}
	return Value{t: mapType(keyType, valueType), value: m}
}

func (m *numericMap) Len() int { return len(m.data) }

func (m *numericMap) Get(k Value) (Value, bool) {
	v, ok := m.data[k.num]
	if !ok {
		return newZero(m.valueType), false
	}
	return v, true
}

func (m *numericMap) Set(k, v Value) {
	key := k.num
	if _, ok := m.data[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.data[key] = v.assign(m.valueType)
}

func (m *numericMap) Delete(k Value) {
	delete(m.data, k.num)
	if len(m.data) >= (len(m.keys) >> 1) {
		return
	}
	m.keys = maps.Keys(m.data)
}

func (m *numericMap) Range() func() (Value, Value, bool) {
	r := m.keys
	n := 0
	return func() (Value, Value, bool) {
		for n < len(r) {
			k := r[n]
			v, ok := m.data[k]
			n++
			if ok {
				return Value{t: m.keyType, num: k}, v, true
			}
		}
		return Nil(), Nil(), false
	}
}

func (m *numericMap) String() string {
	var p []string
	for k, v := range m.data {
		p = append(p, Value{t: m.keyType, num: k}.String()+":"+v.safeStr())
	}
	return "map[" + strings.Join(p, " ") + "]"
}

func (m *numericMap) SafeStr() string {
	var p []string
	for k, v := range m.data {
		if !v.t.isSafeStr() {
			return "map[...]"
		}
		p = append(p, Value{t: m.keyType, num: k}.String()+":"+v.safeStr())
	}
	return "map[" + strings.Join(p, " ") + "]"
}

type structT struct {
	Object
	TypeN   int
	Lookup  map[string]int
	Fields  intMap
	Methods *intMap
}

func NewStruct(base Value, data []Value) Value {
	b := base.value.(*structT)
	lookup, fields, methods := b.Lookup, b.Fields.Copy(), b.Methods
	s := newStruct(b.TypeN, lookup, fields, methods)
	st := s.value.(*structT)
	for n := 0; n < len(data); n += 2 {
		st.SetAttr(data[n].String(), data[n+1])
	}
	return s
}

func newStructByIndex(base Value, data []Value) Value {
	b := base.value.(*structT)
	lookup, fields, methods := b.Lookup, b.Fields.Copy(), b.Methods
	s := newStruct(b.TypeN, lookup, fields, methods)
	st := s.value.(*structT)
	for n := 0; n < len(data); n += 2 {
		st.SetIndex(data[n].Int(), data[n+1])
	}
	return s
}

func newStruct(typeN int, lookup map[string]int, data intMap, methods *intMap) Value {
	return Value{t: TypeStruct | Type(typeN<<8), value: &structT{Lookup: lookup, Fields: data, Methods: methods}}
}

func (s *structT) GetAttr(k string) Value {
	return s.GetIndex(s.Lookup[k])
}
func (s *structT) SetAttr(k string, v Value) {
	s.SetIndex(s.Lookup[k], v)
}

func (s *structT) GetIndex(k int) Value {
	v, ok := s.Fields.Get(k)
	if ok {
		return v
	}
	raw, _ := s.Methods.Get(k)
	return newMethod(Value{t: TypeStruct, value: s}, raw.getFunc())
}

func (s *structT) String() string {
	items := []string{}
	for k, v := range s.Lookup {
		vv, ok := s.Fields.Get(v)
		if !ok {
			continue
		}
		items = append(items, k+":"+vv.safeStr())
	}
	slices.Sort(items)
	return "&{" + strings.Join(items, " ") + "}"
}

func (s *structT) SafeStr() string {
	items := []string{}
	for k, v := range s.Lookup {
		vv, ok := s.Fields.Get(v)
		if !ok {
			continue
		}
		if !vv.t.isSafeStr() {
			return "&{...}"
		}
		items = append(items, k+":"+vv.safeStr())
	}
	slices.Sort(items)
	return "&{" + strings.Join(items, " ") + "}"
}

func newMethod(obj Value, f *funcT) Value {
	xArgs := f.Args - 1
	vArgs := xArgs
	if f.Variadic {
		vArgs = -vArgs
	}
	return newFunc(vArgs, f.Rets, func(v *VM) {
		args := make([]Value, xArgs)
		copy(args, v.stack[len(v.stack)-xArgs:])
		v.stack = v.stack[:len(v.stack)-xArgs]
		v.stack = append(v.stack, obj)
		v.stack = append(v.stack, args...)
		f.Value(v)
	})
}

func (s *structT) SetIndex(k int, v Value) {
	s.Fields.Assign(k, v)
}

func nilRange() func() (Value, Value, bool) {
	return func() (Value, Value, bool) {
		return Nil(), Nil(), false
	}
}

func newNext(next func() (Value, Value, bool)) Value {
	return Value{t: typeNext, value: &nextT{next: next}}
}

func (v Value) addField(key string, idx int, val Value) {
	v.value.(*structT).Lookup[key] = idx
	v.value.(*structT).Fields.Set(idx, val)
}

func (v Value) syncFields(b Value) {
	cur := b.value.(*structT)
	for key, idx := range cur.Lookup {
		value, _ := cur.Fields.Get(idx)
		v.addField(key, idx, value)
	}
}

func (v Value) addMethod(key string, idx int, val Value) {
	s := v.value.(*structT)
	s.Lookup[key] = idx
	if fnc, ok := s.Methods.Get(idx); ok {
		f, ok := fnc.value.(*funcT)
		if ok {
			*f = *val.value.(*funcT)
			return
		}
	}
	s.Methods.Set(idx, val)
}
