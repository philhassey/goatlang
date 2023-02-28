package goatlang

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type frame struct {
	BaseN int
	Codes []instruction
	N     int
}

type VM struct {
	stack   []Value
	globals *lookup
	stdout  io.Writer

	backtrace []pos
	frame     frame
}

func (v *VM) Set(key string, value Value) { v.globals.Set(key, value) }
func (v *VM) Get(key string) Value        { return v.globals.Get(key) }

type VMOption func(*vmConfig)
type vmConfig struct {
	stdout io.Writer
}

func WithStdout(v io.Writer) VMOption { return func(c *vmConfig) { c.stdout = v } }

func NewVM(options ...VMOption) *VM {
	vm := &VM{
		globals: newGlobals(),
	}
	loadBuiltins(vm)
	config := vmConfig{
		stdout: os.Stdout,
	}
	for _, o := range options {
		o(&config)
	}
	vm.stdout = config.stdout
	return vm
}

func (v *VM) btErr(r any) error {
	bt := v.backtrace
	var lines []string
	i := v.frame.Codes[v.frame.N]
	lines = append(lines, fmt.Sprintf("%v: %v: %v", i.Pos.String(v.globals), i.Code, r))
	for n := len(bt) - 1; n >= 0; n-- {
		pos := bt[n]
		if pos == 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("\t%v", pos.String(v.globals)))
	}
	return errors.New(strings.Join(lines, "\n"))
}

func (v *VM) run(codes []instruction, slots int) (rets []Value, err error) {
	vm := VM{
		globals: v.globals,
		stdout:  v.stdout,
		stack:   make([]Value, slots),
		frame:   frame{Codes: codes},
	}
	defer func() {
		if r := recover(); r != nil {
			err = vm.btErr(r)
		}
	}()
	vm.exec()
	rets = vm.stack[slots:]
	return rets, nil
}

func (v *VM) Func(fnc Value, xRets int, params ...Value) (rets []Value, err error) {
	vm := VM{
		globals: v.globals,
		stdout:  v.stdout,
		stack:   append(params, fnc),
		frame: frame{Codes: []instruction{{
			Code: codeCall,
			A:    reg(len(params)),
			B:    reg(xRets),
		}}},
	}
	defer func() {
		if r := recover(); r != nil {
			err = vm.btErr(r)
		}
	}()
	vm.exec()
	return vm.stack[len(vm.stack)-xRets:], nil
}

func (v *VM) Call(name string, xRets int, params ...Value) (rets []Value, err error) {
	return v.Func(v.globals.Get(name), xRets, params...)
}

func (v *VM) Yield() { v.Call(builtinYield, 0) }

type RunOption func(*runConfig)

type runConfig struct {
	codeDump    io.Writer
	treeDump    io.Writer
	evalImports map[string]string
}

func WithEvalImports(v map[string]string) RunOption { return func(c *runConfig) { c.evalImports = v } }
func WithCodeDump(v io.Writer) RunOption            { return func(c *runConfig) { c.codeDump = v } }
func WithTreeDump(v io.Writer) RunOption            { return func(c *runConfig) { c.treeDump = v } }

func (v *VM) Load(sys fs.FS, arg string, options ...RunOption) error {
	arg = strings.Replace(filepath.Clean(arg), string(os.PathSeparator), "/", -1)
	var config runConfig
	for _, o := range options {
		o(&config)
	}
	f := loadPackage
	if strings.HasSuffix(arg, ".go") {
		f = loadFile
	}
	pkgs, err := f(sys, arg)
	if err != nil {
		return fmt.Errorf("error in load: %w", err)
	}

	v.treeDump(config.treeDump, pkgs)
	codes, slots, err := compilePkgs(v.globals, pkgs, true)
	if err != nil {
		return fmt.Errorf("error in compile: %w", err)
	}
	v.codeDump(config.codeDump, codes)
	rets, err := v.run(codes, slots)
	if err != nil {
		return fmt.Errorf("error in run: %w", err)
	}
	if len(rets) > 0 {
		return fmt.Errorf("unexpected returns: %v", rets)
	}
	return nil
}

func (v *VM) treeDump(w io.Writer, tree []*token) {
	if w == nil {
		return
	}
	for _, t := range tree {
		s := t.String()
		s = s[3 : len(s)-1]
		w.Write([]byte(s + "\n"))
	}
}

func (v *VM) codeDump(w io.Writer, codes []instruction) {
	if w == nil {
		return
	}
	for _, s := range codes {
		fmt.Fprintf(w, "%s: %s\n", s.Pos.String(v.globals), s.String(v.globals))
	}
}

func (v *VM) Eval(sys fs.FS, fname, input string, options ...RunOption) (rets []Value, err error) {
	var opts runConfig
	for _, o := range options {
		o(&opts)
	}
	const pkgName = "main"
	tokens, err := tokenize(fname, input)
	if err != nil {
		return nil, fmt.Errorf("error in tokenize: %w", err)
	}
	tree, err := parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("error in parse: %w", err)
	}
	pkgs, err := loadImports(sys, "", tree)
	if err != nil {
		return nil, fmt.Errorf("error in loadImports: %w", err)
	}
	codes, slots, err := compilePkgs(v.globals, pkgs[:len(pkgs)-1], true)
	if err != nil {
		return nil, fmt.Errorf("error in compile (imports): %w", err)
	}
	_, err = v.run(codes, slots)
	if err != nil {
		return nil, fmt.Errorf("error in run (imports): %w", err)
	}

	v.treeDump(opts.treeDump, pkgs[len(pkgs)-1:])
	if opts.evalImports == nil {
		opts.evalImports = map[string]string{}
	}
	cmp := &compiler{
		Globals:     v.globals,
		Locals:      newLookup(),
		Imports:     opts.evalImports,
		Optimize:    true,
		PackageName: pkgName,
		ExportName:  pkgName,
	}
	codes, slots, err = cmp.run(pkgs[len(pkgs)-1])
	if err != nil {
		return nil, fmt.Errorf("error in compile: %w", err)
	}
	v.codeDump(opts.codeDump, codes)
	rets, err = v.run(codes, slots)
	if err != nil {
		return nil, fmt.Errorf("error in run: %w", err)
	}
	return rets, nil
}

func mkFunc(args, rets, slots int, tokens []instruction) func(v *VM) {
	empty := make([]Value, slots-args)
	codes := tokens[args+rets:]
	return func(v *VM) {
		v.backtrace = append(v.backtrace, v.frame.Codes[v.frame.N].Pos)
		prev := v.frame
		v.frame = frame{
			Codes: codes,
			BaseN: len(v.stack) - args,
		}
		for i := 0; i < args; i++ {
			v.stack[len(v.stack)-args+i] = v.stack[len(v.stack)-args+i].assign(Type(tokens[i].A))
		}
		v.stack = append(v.stack, empty...)
		topN := len(v.stack)
		v.exec()
		v.stack = append(v.stack[:v.frame.BaseN], v.stack[topN:]...)
		for i := 0; i < rets; i++ {
			v.stack[len(v.stack)-rets+i] = v.stack[len(v.stack)-rets+i].assign(Type(tokens[args+i].A))
		}
		v.frame = prev
		v.backtrace = v.backtrace[:len(v.backtrace)-1]
	}
}

func call(v *VM, ft *funcT, xArgs, xRets int) {
	if !ft.Variadic {
		callReady(v, ft, xArgs, xRets)
		return
	}
	nVarArgs := xArgs - ft.Args + 1
	varArgs := make([]Value, nVarArgs)
	end := len(v.stack) - len(varArgs)
	copy(varArgs, v.stack[end:])
	v.stack = v.stack[:end]
	v.stack = append(v.stack, NewSlice(ft.VariadicType.value(), varArgs))
	xArgs = xArgs - len(varArgs) + 1
	callReady(v, ft, xArgs, xRets)
}

func callReady(v *VM, ft *funcT, xArgs, xRets int) {
	if xArgs != ft.Args {
		panic("incorrect args")
	}
	// if xRets > ft.Rets {
	// 	panic("incorrect returns")
	// }
	top := len(v.stack) - xArgs
	ft.Value(v)
	fRets := len(v.stack) - top
	if fRets < xRets {
		panic("incorrect returns")
	} else if fRets > xRets {
		v.stack = v.stack[:top+xRets]
	}
}
