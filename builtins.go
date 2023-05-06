package goatlang

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

func get2Pop1f(v *VM) (float64, float64) {
	a, b := v.stack[len(v.stack)-2].Float64(), v.stack[len(v.stack)-1].Float64()
	v.stack = v.stack[:len(v.stack)-1]
	return a, b
}
func get1f(v *VM) float64     { return v.stack[len(v.stack)-1].Float64() }
func set1f(v *VM, a float64)  { v.stack[len(v.stack)-1] = Float64(a) }
func pop1f(v *VM) float64     { a := get1f(v); v.stack = v.stack[:len(v.stack)-1]; return a }
func push1f(v *VM, a float64) { v.stack = append(v.stack, Float64(a)) }

func get2Pop1v(v *VM) (Value, Value) {
	a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
	v.stack = v.stack[:len(v.stack)-1]
	return a, b
}
func get1v(v *VM) Value    { return v.stack[len(v.stack)-1] }
func set1v(v *VM, a Value) { v.stack[len(v.stack)-1] = a }

// func pop1v(v *VM) Value    { a := get1v(v); v.stack = v.stack[:len(v.stack)-1]; return a }

func loadMath(g *lookup) {
	g.Set("math.Abs", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Abs(a)) }))
	g.Set("math.Atan", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Atan(a)) }))
	g.Set("math.Atan2", NewFunc(2, 1, func(v *VM) { a, b := get2Pop1f(v); set1f(v, math.Atan2(a, b)) }))
	g.Set("math.Ceil", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Ceil(a)) }))
	g.Set("math.Cos", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Cos(a)) }))
	g.Set("math.Floor", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Floor(a)) }))
	g.Set("math.Hypot", NewFunc(2, 1, func(v *VM) { a, b := get2Pop1f(v); set1f(v, math.Hypot(a, b)) }))
	g.Set("math.Log", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Log(a)) }))
	g.Set("math.Max", NewFunc(2, 1, func(v *VM) { a, b := get2Pop1f(v); set1f(v, math.Max(a, b)) }))
	g.Set("math.Min", NewFunc(2, 1, func(v *VM) { a, b := get2Pop1f(v); set1f(v, math.Min(a, b)) }))
	g.Set("math.Mod", NewFunc(2, 1, func(v *VM) { a, b := get2Pop1f(v); set1f(v, math.Mod(a, b)) }))
	g.Set("math.Pow", NewFunc(2, 1, func(v *VM) { a, b := get2Pop1f(v); set1f(v, math.Pow(a, b)) }))
	g.Set("math.Round", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Round(a)) }))
	g.Set("math.Signbit", NewFunc(1, 1, func(v *VM) { a := get1f(v); v.stack[len(v.stack)-1] = Bool(math.Signbit(a)) }))
	g.Set("math.Sin", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Sin(a)) }))
	g.Set("math.Sqrt", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Sqrt(a)) }))
	g.Set("math.Tan", NewFunc(1, 1, func(v *VM) { a := get1f(v); set1f(v, math.Tan(a)) }))

	g.Set("math.Pi", Float64(math.Pi))
}

func loadMathRand(g *lookup) {
	g.Set("math/rand.Float64", NewFunc(0, 1, func(v *VM) { push1f(v, rand.Float64()) }))
	g.Set("math/rand.Int", NewFunc(0, 1, func(v *VM) { v.stack = append(v.stack, Int(rand.Int())) }))
	g.Set("math/rand.Intn", NewFunc(1, 1, func(v *VM) { a := int(get1f(v)); v.stack[len(v.stack)-1] = Int(rand.Intn(a)) }))
	g.Set("math/rand.Int31", NewFunc(0, 1, func(v *VM) { v.stack = append(v.stack, Int32(rand.Int31())) }))
	g.Set("math/rand.Int31n", NewFunc(1, 1, func(v *VM) { a := int(get1f(v)); v.stack[len(v.stack)-1] = Int32(rand.Int31n(int32(a))) }))

	g.Set("math/rand.Uint32", NewFunc(0, 1, func(v *VM) { v.stack = append(v.stack, Uint32(rand.Uint32())) }))
	g.Set("math/rand.Seed", NewFunc(1, 0, func(v *VM) { a := int64(pop1f(v)); rand.Seed(a) }))
}

func sprint(v *VM, a Value) string { return a.String() }
func vaSprint(v *VM, va []Value) string {
	res := make([]string, len(va))
	for i, a := range va {
		res[i] = sprint(v, a)
	}
	return strings.Join(res, " ")
}

func loadFmt(g *lookup) {
	g.Set("fmt.Sprint", NewFunc(1, 1, func(v *VM, args []Value, vargs ...Value) []Value {
		return []Value{String(vaSprint(v, vargs))}
	}))
	g.Set("fmt.Print", NewFunc(1, 0, func(v *VM, args []Value, vargs ...Value) []Value {
		fmt.Fprint(v.stdout, vaSprint(v, vargs))
		return nil
	}))
	g.Set("fmt.Println", NewFunc(1, 0, func(v *VM, args []Value, vargs ...Value) []Value {
		fmt.Fprintln(v.stdout, vaSprint(v, vargs))
		return nil
	}))
	g.Set("fmt.Sprintf", NewFunc(2, 1, func(v *VM, args []Value, vargs ...Value) []Value {
		var va []any
		for _, v := range vargs {
			va = append(va, v)
		}
		return []Value{String(fmt.Sprintf(args[0].String(), va...))}
	}))
}

func loadErrors(g *lookup) {
	g.Set("errors.New", NewFunc(1, 1, func(v *VM, args []Value) Value {
		return Wrap(&errorT{err: errors.New(args[0].String())})
	}))
}

type errorT struct {
	Object
	err error
}

func (e *errorT) GetAttr(k string) (v Value) {
	switch k {
	case "Error", "String":
		v = NewFunc(0, 1, func(v *VM) Value {
			return String(e.Error())
		})
	}
	return
}

func (e *errorT) Error() string  { return e.String() }
func (e *errorT) String() string { return e.err.Error() }

func Error(err error) Value {
	return Wrap(&errorT{err: err})
}

func loadStrings(g *lookup) {
	g.Set("strings.Split", NewFunc(2, 1, func(v *VM) {
		s, sep := get2Pop1v(v)
		res := strings.Split(s.String(), sep.String())
		data := make([]Value, len(res))
		for i, val := range res {
			data[i] = String(val)
		}
		set1v(v, NewSlice(TypeString, data))
	}))
	g.Set("strings.Join", NewFunc(2, 1, func(v *VM) {
		elem, sep := get2Pop1v(v)
		data := make([]string, elem.Len())
		iter := elem.Range()
		for {
			key, val, ok := iter()
			if !ok {
				break
			}
			data[key.Int()] = val.String()
		}
		set1v(v, String(strings.Join(data, sep.String())))
	}))
	g.Set("strings.ReplaceAll", NewFunc(3, 1, func(v *VM, args []Value) Value {
		a, b, c := args[0].String(), args[1].String(), args[2].String()
		return String(strings.ReplaceAll(a, b, c))
	}))
	g.Set("strings.TrimRight", NewFunc(2, 1, func(v *VM, args []Value) Value {
		return String(strings.TrimRight(args[0].String(), args[1].String()))
	}))
	g.Set("strings.TrimSuffix", NewFunc(2, 1, func(v *VM, args []Value) Value {
		return String(strings.TrimSuffix(args[0].String(), args[1].String()))
	}))
	g.Set("strings.Contains", NewFunc(2, 1, func(v *VM, args []Value) Value {
		return Bool(strings.Contains(args[0].String(), args[1].String()))
	}))
	g.Set("strings.Repeat", NewFunc(2, 1, func(v *VM, args []Value) Value {
		return String(strings.Repeat(args[0].String(), args[1].Int()))
	}))
}

const builtinYield = "builtin.__yield"

func loadBuiltin(g *lookup) {
	g.Set("builtin.__type", NewFunc(1, 1, func(v *VM) { a := get1v(v); set1v(v, String(a.t.str(g))) }))
	g.Set("builtin.print", NewFunc(1, 0, func(v *VM, args []Value, vargs ...Value) []Value {
		fmt.Fprint(v.stdout, vaSprint(v, vargs))
		return nil
	}))
	g.Set("builtin.println", NewFunc(1, 0, func(v *VM, args []Value, vargs ...Value) []Value {
		fmt.Fprintln(v.stdout, vaSprint(v, vargs))
		return nil
	}))
	g.Set(builtinYield, NewFunc(0, 0, func(v *VM) {}))
}

func loadTime(g *lookup) {
	// NOTE: int32 isn't adequate, using float64 where we can
	g.Set("time.Sleep", NewFunc(1, 0, func(v *VM) { v.Yield(); a := pop1f(v); time.Sleep(time.Duration(a)) }))
	g.Set("time.Now", NewFunc(0, 1, func(v *VM) { v.stack = append(v.stack, Wrap(newTime(time.Now()))) }))
	g.Set("time.Second", Float64(float64(time.Second)))
}

type timeTime struct {
	Object
	v time.Time
}

func newTime(t time.Time) *timeTime {
	return &timeTime{v: t}
}

func (t *timeTime) GetAttr(k string) (res Value) {
	switch k {
	case "UnixMilli":
		res = NewFunc(0, 1, func(vm *VM, args []Value) Value {
			// wraps around every few days, due to truncation
			return Int32(int32(t.v.UnixMilli()))
		})
	}
	return res
}

func loadMaps(g *lookup) {
	g.Set("golang.org/x/exp/maps.Clone", NewFunc(1, 1, func(vm *VM, args []Value) Value {
		src := args[0]
		var in []Value
		next := src.Range()
		for {
			key, value, ok := next()
			if !ok {
				break
			}
			in = append(in, key, value)
		}
		kt, vt := src.t.pair()
		return NewMap(kt, vt, in)
	}))
	g.Set("golang.org/x/exp/maps.Keys", NewFunc(1, 1, func(vm *VM, args []Value) Value {
		src := args[0]
		var in []Value
		next := src.Range()
		for {
			key, _, ok := next()
			if !ok {
				break
			}
			in = append(in, key)
		}
		kt, _ := src.t.pair()
		return NewSlice(kt, in)
	}))
}

func loadStrconv(g *VM) {
	g.Set("strconv.ParseFloat", NewFunc(2, 2, func(vm *VM, args []Value) []Value {
		res, err := strconv.ParseFloat(args[0].String(), args[0].Int())
		if err != nil {
			return []Value{Float64(0), Error(err)}
		}
		return []Value{Float64(res), Nil()}
	}))
	g.Set("strconv.Itoa", NewFunc(1, 1, func(vm *VM, args []Value) Value {
		return String(strconv.Itoa(args[0].Int()))
	}))
	g.Set("strconv.FormatFloat", NewFunc(4, 1, func(vm *VM, args []Value) Value {
		return String(strconv.FormatFloat(args[0].Float64(), args[1].Uint8(), args[2].Int(), args[3].Int()))
	}))
}

func loadOs(g *VM) {
	var args []Value
	for _, v := range os.Args {
		args = append(args, String(v))
	}
	g.Set("os.Args", NewSlice(TypeString, args))
	g.Set("os.ReadFile", NewFunc(1, 2, func(vm *VM, args []Value) []Value {
		b, err := osReadFile(args[0].String())
		if err != nil {
			return []Value{Nil(), Error(err)}
		}
		var res []Value
		for _, v := range b {
			res = append(res, Byte(v))
		}
		return []Value{NewSlice(TypeUint8, res), Nil()}
	}))
	g.Set("os.WriteFile", NewFunc(3, 1, func(vm *VM, args []Value) Value {
		name := args[0].String()
		var b []byte
		for _, v := range args[1].data() {
			b = append(b, v.Uint8())
		}
		err := osWriteFile(name, b, os.FileMode(args[2].Uint32()))
		if err != nil {
			return Error(err)
		}
		return Nil()
	}))
}

var osWriteFile = os.WriteFile
var osReadFile = os.ReadFile

func loadSlices(g *VM) {
	g.Set("golang.org/x/exp/slices.Contains", NewFunc(2, 1, func(vm *VM, args []Value) Value {
		a := args[1]
		for _, b := range args[0].data() {
			if a.opEq(b).Bool() {
				return Bool(true)
			}
		}
		return Bool(false)
	}))
	g.Set("golang.org/x/exp/slices.Delete", NewFunc(3, 1, func(vm *VM, args []Value) Value {
		s := args[0].data()
		vt := args[0].t.value()
		return newSlice(vt, slices.Delete(s, args[1].Int(), args[2].Int()))
	}))
	g.Set("golang.org/x/exp/slices.SortStableFunc", NewFunc(2, 0, func(vm *VM, args []Value) {
		s := args[0].data()
		slices.SortStableFunc(s, func(a, b Value) bool {
			rets, err := g.Func(args[1], 1, a, b)
			if err != nil {
				panic(err)
			}
			return rets[0].Bool()
		})
	}))
	g.Set("golang.org/x/exp/slices.SortFunc", NewFunc(2, 0, func(vm *VM, args []Value) {
		s := args[0].data()
		slices.SortFunc(s, func(a, b Value) bool {
			rets, err := g.Func(args[1], 1, a, b)
			if err != nil {
				panic(err)
			}
			return rets[0].Bool()
		})
	}))
	g.Set("golang.org/x/exp/slices.Sort", NewFunc(1, 0, func(vm *VM, args []Value) {
		s := args[0].data()
		slices.SortFunc(s, func(a, b Value) bool {
			return a.opLt(b).Bool()
		})
	}))
	g.Set("golang.org/x/exp/slices.Equal", NewFunc(2, 1, func(vm *VM, args []Value) Value {
		a := args[0].data()
		b := args[1].data()
		res := slices.EqualFunc(a, b, func(a, b Value) bool {
			return a.opEq(b).Bool()
		})
		return Bool(res)
	}))
}

/*
type ioWriter struct {
	Object
	v io.Writer
}

func (w *ioWriter) GetAttr(k string) (v Value) {
	switch k {
	case "Write":
		return NewFunc(1, 0, func(vm *VM, args []Value) Value {
			data := args[0].data()
			b := make([]byte, len(data))
			for i, val := range data {
				b[i] = val.Byte()
			}
			n, err := w.v.Write(b)
			_, _ = n, err
			// TODO: return both values
		})
	}
	return
}
*/
