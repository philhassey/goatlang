package goatlang

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestBuiltins(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Want string
	}{
		{"math.Abs", `import "math"; v = math.Abs(-42); v`, `42`},
		{"math.Atan", `import "math"; v = math.Atan(0); v`, `0`},
		{"math.Atan2", `import "math"; v = math.Atan2(0,1); v`, `0`},
		{"math.Ceil", `import "math"; v = math.Ceil(41.5); v`, `42`},
		{"math.Cos", `import "math"; v = math.Cos(0); v`, `1`},
		{"math.Floor", `import "math"; v = math.Floor(42.5); v`, `42`},
		{"math.Hypot", `import "math"; v = math.Hypot(3,4); v`, `5`},
		{"math.Log", `import "math"; v = math.Log(1); v`, `0`},
		{"math.Max", `import "math"; v = math.Max(41,42); v`, `42`},
		{"math.Min", `import "math"; v = math.Min(42,43); v`, `42`},
		{"math.Mod", `import "math"; v = math.Mod(85,43); v`, `42`},
		{"math.Pow", `import "math"; v = math.Pow(4,2); v`, `16`},
		{"math.Round", `import "math"; v = math.Round(41.5); v`, `42`},
		{"math.Signbit", `import "math"; v = math.Signbit(-42); v`, `true`},
		{"math.Sin", `import "math"; v = math.Sin(0); v`, `0`},
		{"math.Sqrt", `import "math"; v = math.Sqrt(1764); v`, `42`},
		{"math.Tan", `import "math"; v = math.Tan(0); v`, `0`},
		{"math.Pi", `import "math"; math.Pi`, `3.141592653589793`},

		{"rand.Float64", `import "math/rand"; v = rand.Float64(); v < 1`, `true`},
		{"rand.Uint32", `import "math/rand"; v = rand.Uint32(); v != 0`, `true`},
		{"rand.Intn", `import "math/rand"; v = rand.Intn(42); v < 42`, `true`},
		{"rand.Int", `import "math/rand"; v = rand.Int()%42; v < 42`, `true`},
		{"rand.Int31n", `import "math/rand"; v = rand.Int31n(42); v < 42`, `true`},
		{"rand.Int31", `import "math/rand"; v = rand.Int31()%42; v < 42`, `true`},
		{"rand.Seed", `import "math/rand"; rand.Seed(0); v = rand.Intn(42); v`, `12`},

		{"fmt.Sprint", `import "fmt"; v = fmt.Sprint(42); v`, `42`},
		{"fmt.Print", `import "fmt"; fmt.Print(42)`, `;42`},
		{"fmt.Println", `import "fmt"; fmt.Println(42)`, ";42\n"},
		{"print", `print(42)`, `;42`},
		{"println", `println(42)`, ";42\n"},
		{"fmt.Print/vargs", `import "fmt"; fmt.Print(40,2)`, `;40 2`},
		{"fmt.Print/ellipsis", `import "fmt"; fmt.Print([]int{40,2}...)`, `;40 2`},

		{"strings.Split", `import "strings"; v := strings.Split("a,b,c", ","); v`, `[a b c]`},
		{"strings.Join", `import "strings"; v := strings.Join([]string{"a","b","c"},","); v`, `a,b,c`},
		{"strings.TrimRight", `import "strings"; v := strings.TrimRight("42333","3"); v`, `42`},
		{"strings.TrimSuffix", `import "strings"; v := strings.TrimSuffix("4233","33"); v`, `42`},
		{"strings.ReplaceAll", `import "strings"; v := strings.ReplaceAll("41","1","2"); v`, `42`},
		{"strings.Contains", `import "strings"; v := strings.Contains("x42y","42"); v`, `true`},

		{"__type", `v = __type(42); v`, `number`},

		{"time.Sleep", `import "time"; time.Sleep(0)`, ``},
		{"time.Time.UnixMilli", `import "time"; v := time.Now().UnixMilli(); v!=0`, `true`},

		{"maps.Clone", `import "golang.org/x/exp/maps"; a := map[string]int{"k":40}; b := maps.Clone(a); c := maps.Clone(a); c["k"] = 42; a; b; c`, `map[k:40] map[k:40] map[k:42]`},
		{"maps.Keys", `import "golang.org/x/exp/maps"; m := map[string]int{"k":40,"v":2}; n := maps.Keys(m); n`, `[k v]`},

		{"slices.Contains/true", `import "golang.org/x/exp/slices"; s := []int{1,42,3}; v := slices.Contains(s, 42); v`, `true`},
		{"slices.Contains/false", `import "golang.org/x/exp/slices"; s := []int{1,2,3}; v := slices.Contains(s, 42); v`, `false`},
		{"slices.Delete", `import "golang.org/x/exp/slices"; s := []int{1,2,3,4}; s = slices.Delete(s,1,3); s`, `[1 4]`},
		{"slices.Sort", `import "golang.org/x/exp/slices"; s := []int{4,2,1,3}; slices.Sort(s); s`, `[1 2 3 4]`},
		{"slices.SortFunc", `import "golang.org/x/exp/slices"; func f(a, b int) bool { return b < a }; s := []int{4,2,1,3}; slices.SortFunc(s, f); s`, `[4 3 2 1]`},
		{"slices.SortStableFunc", `import "golang.org/x/exp/slices"; func f(a, b int) bool { false }; s := []int{4,2,1,3}; slices.SortStableFunc(s, f); s`, `[4 2 1 3]`},
		{"slices.Equal/true", `import "golang.org/x/exp/slices"; a := []int{4,2,1,3}; b := []int{4,2,1,3}; v := slices.Equal(a,b); v`, `true`},
		{"slices.Equal/false", `import "golang.org/x/exp/slices"; a := []int{1,2,3,4}; b := []int{4,2,1,3}; v := slices.Equal(a,b); v`, `false`},

		{"strconv.ParseFloat", `import "strconv"; v, err := strconv.ParseFloat("42",64); v, err`, `42 nil`},
		{"strconv.ParseFloat/error", `import "strconv"; v, err := strconv.ParseFloat("asdf",64); v, err!=nil`, `0 true`},
		{"strconv.Itoa", `import "strconv"; v := strconv.Itoa(42); v`, `42`},
		{"strconv.FormatFloat", `import "strconv"; v := strconv.FormatFloat(42,'E',-1,64); v`, `4.2E+01`},

		{"os.Args", `import "os"; v := len(os.Args); v > 0`, `true`},
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
			stdout := &bytes.Buffer{}
			vm := NewVM(WithStdout(stdout))
			codes, slots, err := compile(vm.globals, tree, false)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}
			rets, err := vm.run(codes, slots)
			if err != nil {
				t.Fatalf("Exec error: %v", err)
			}
			var ts []string
			for _, s := range rets {
				ts = append(ts, s.String())
			}
			s := strings.Join(ts, " ")
			if stdout.Len() > 0 {
				s += ";" + stdout.String()
			}
			assert(t, "Exec", s, row.Want)
		})
	}
}

func testEval(t *testing.T, in string) string {
	tokens, err := tokenize("test", in)
	if err != nil {
		t.Fatalf("Tokenize error: %v", err)
	}
	tree, err := parse(tokens)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	stdout := &bytes.Buffer{}
	vm := NewVM(WithStdout(stdout))
	codes, slots, err := compile(vm.globals, tree, false)
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}
	rets, err := vm.run(codes, slots)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}
	var ts []string
	for _, s := range rets {
		ts = append(ts, s.String())
	}
	s := strings.Join(ts, " ")
	if stdout.Len() > 0 {
		s += ";" + stdout.String()
	}
	return s
}

func TestBuiltins_Mocks(t *testing.T) {
	t.Run("os.ReadFile", func(t *testing.T) {
		osReadFile = func(name string) ([]byte, error) {
			assert(t, "name", name, "test.txt")
			return []byte{42}, nil
		}
		res := testEval(t, `import "os"; b, err := os.ReadFile("test.txt"); b, err`)
		assert(t, "res", res, `[42] nil`)
	})

	t.Run("os.ReadFile/error", func(t *testing.T) {
		osReadFile = func(name string) ([]byte, error) {
			return nil, fmt.Errorf("errReadFile")
		}
		res := testEval(t, `import "os"; b, err := os.ReadFile("test.txt"); b, err`)
		assert(t, "res", res, `nil errReadFile`)
	})

	t.Run("os.WriteFile", func(t *testing.T) {
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			assert(t, "name", name, "test.txt")
			assert(t, "data", string(data), "*")
			assert(t, "perm", int(perm), 0666)
			return nil
		}
		res := testEval(t, `import "os"; err := os.WriteFile("test.txt",[]byte{42},0666); err`)
		assert(t, "res", res, `nil`)
	})

	t.Run("os.WriteFile/error", func(t *testing.T) {
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("errWriteFile")
		}
		res := testEval(t, `import "os"; err := os.WriteFile("test.txt",[]byte{42},0666); err`)
		assert(t, "res", res, `errWriteFile`)
	})

}

func TestBuiltins_error(t *testing.T) {
	tests := []struct {
		Name string
		In   string
		Err  string
	}{
		{"slices.SortFunc/panic", `import "golang.org/x/exp/slices"; func f(a, b int) bool { panic("panic") }; s := []int{4,2,1,3}; slices.SortFunc(s, f)`, `panic`},
		{"slices.SortStableFunc/panic", `import "golang.org/x/exp/slices"; func f(a, b int) bool { panic("panic") }; s := []int{4,2,1,3}; slices.SortStableFunc(s, f)`, `panic`},
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
