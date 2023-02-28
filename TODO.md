# To do

# Later

# Way later (if ever)
- cache parse / compile data so live reload is ultra fast
- add custom byte slice type so string <-> []byte isn't a mess
- types that refer to specific structs, global or local (local structs must be globally defined path/to/pkg.FuncName.TypeName)
- auto-init features for type aliases `type T struct{X int}; type A []T; x := A{{X:42}}`
- also local type aliases
- add proper field ordering according to struct def (a tiny bit useful)
- properly handle ignoring _test files, and maybe other os/arch specific files
- type switch, type assertions (trying to avoid using these anyways)

- safe.Run package (escape valve for lack of defer, recover)
- init structure without field names (not that useful except for unit tests)
- anonymous structures `x := []struct{name string}{...}` (not that useful except for unit tests)

- lambda & closure functions (exact go behavior is very tricky)
- support named return values (not useful except for defer/recover)
- defer, recover (depends on closures, named returns) - maybe useful w/o closures, f.Close(), etc
- concurrency primitives (go, chan, select, wg, mutex) (depends on closures)

# Never
- methods on non-structs - costs memory without compile-time type tracking
- proper int64, uint64, int16, uint16 - these just don't seem that useful and they take up an extra bit
- real non-pointer structs ??? (maybe copy on get semantics?) - the catch is that
    the memory allocation pattern will be the opposite of expected.
- pointers to non-structs
- complex numbers, who uses these
- goto, fallthrough is kinda toxic anyways
- custom types with methods - really neat, but complicated and costly, use a struct instead
- re-add Stringer support - easy, but makes the .String() vs fmt.Sprint have different results
- embedded structs (depends on non-pointer structs maybe?)
- generics - ??? might require changes to how we do map typing
- arrays
- numeric slice types / string slice types - better as a custom type then a builtin

# Out of scope
- runtime type checking
- register based VM - maybe not, for balls.go, this would only reduce 20% instructions from 199 -> 160 (elim localget/localset)
- compile time type checking (see: d11d554f3fa501e7b7b1a0a52a29d9ca4780f1bf)

# Done
- named complex types `type Matrix []float64`, `type X []*T`, `type X = T` (auto-init may be tricky)
- int16, int64 as alias to int32 (lossy, but helpful if depending on a lib using them)
- cleanup func definition
- local struct definition
- make VM tests, etc override Stdout and validate results "val val;stdout"
- support varargs (a bit tricky, might have performance cost, slice... is pass-through)
- support multiple returns from native functions
- make safe recursive stringer
- fix stdout in live mode
- x := ^1
- const (a = 40; b = a+2 )
- callstack on panic isn't always very accurate (due to optimization)
- add some slices/maps package functions
- int -> int32
- byte -> uint8
- uint -> uint32
- rune -> int32
- add uint32, uint, int8 types (signed bit, rm next, typetype is unsigend float, no nillableMask)
- mask v.t&b.t to ensure a numeric type results
- invert for/range loops for instructions per cycle
- add FASTSETINT, FASTGETINT for p[0], etc
- lock it down to int32, ensure add,mul,sub,div are accurate
- make live code work better by just replacing funcs & methods
- allow non-pointer struct for initializing pointer structs (kinda sketchy)
- figure out pattern for live reload code
- 100% coverage again
- better support for deep casting [][]byte, map[int][]byte
- make append nil to [][]byte work correctly
- make set nil to map[int][]byte work correctly
- typed nils - already exists for slices
- make copy/append/make/print/println/delete/panic/len be able to be methods
- ensure casting, etc, only applies to nil / constInt, try and simplify
- make shadowing work maybe
- fix lookup reuse bug 1.0/60
- support import aliases
- copy, panic
- ensure int vs byte conversions are correct
- for range x { } 
- iota - a bit complicated
- call casts conversions
- non-struct types (sans methods)
- consider changing "NewFloat64" (etc) to "Float64"
- make string type be just a string
- make all other types be *typeT
- consider having Wrap(Object) and Value.UnWrap() -> Object
- consider adding IsNil() bool
- remove Data method, or make it a private helper method
- remove type info from NewSlice / NewMap if possible - seems somewhat useful?
- hack in varargs for fmt.Print and fmt.Println
- remove "problems"
- try having slice be a *sliceT
- try having func be a *funcT
- remove ability to create anonymous structs from NewStruct(), no need for *VM
- tags to control real modules vs goatlang imports
- remove Stringer support
- have .String() not be required so fmt.Sprint() can be the fallback
- embed a lookup in the structs, can't set something unknown of course
- add Get/Set/etc methods to Value 
- fix all "problems"
- more consistent print / String() support
- remove Yield() function, replace with time.Sleep()
- make all compile/parse be private
- review if all Lookup should be on VM
- consolidate file/module loading into core (vm.Eval() -> Stateful?, vm.Load(), vm.Main())
- live code reload
- track Pos in a single uint64
- import other packages from CLI
- pass functions in as parameters
- stack traces
- .String() support in fmt
- compile time error if a foreign global does not exist
- error support
- init() support
- support subset of math, fmt, strings, rand, sort
- shadow package test
- repl should remember imports
- repl should $1 -> $[1]
- run package
- run file (with imports)
- repl should only print stack if len > 0
- key only ranges
- evaluate consts (types, consts, methods, funcs) before vars, etc. 
- support multiple files in a module
- support importing other .go modules
- add readline style execution
- add switch support
- ensure variables/fields can be funcs
- make sure eq properly compares things like structs
- make sure map init can use the untyped struct form
- make sure "make" uses zero values for its type (float, string, etc)
- untyped struct init within slice init / map init fmt.Println(map[int]S{2: {X:1, Y:2}})
- recursive string print for maps, slices, structs
- print for structs
- add make builtin for slice and map
- support nil comparison
- support nil assignment
- elipsis vararg support for builtins
- ensure all *=, etc type ops are supported, and all bit ops
- support packages (but not aliases)
- support structs (but not embedding) (always pointers)
- support interfaces (but not embedding)
- support float64, int(32), byte, string, bool
- support slices of anything
- support map[native]anything
- support functions (not lambdas)
- support operators
- support multiple returns
- support native functions

- support non-go running of code without being in a function
- support hot-reloading
- single pass generation of code
- stack based VM