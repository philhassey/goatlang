package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/philhassey/goatlang"
	"github.com/radovskyb/watcher"
)

var profile = flag.String("profile", "", "write cpu profile to `file`, use `go tool pprof` to analyze")
var codeFlag = flag.Bool("code", false, "dump code")
var treeFlag = flag.Bool("tree", false, "dump tree")
var liveFlag = flag.Bool("live", false, "live coding features")

func main() {
	Main()
}

func Main(loaders ...func(*goatlang.VM)) {
	flag.Parse()
	args := flag.Args()

	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			log.Fatalln("could not create CPU profile:", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalln("could not start CPU profile:", err)
		}
		defer pprof.StopCPUProfile()
	}

	if len(args) > 0 {
		arg := args[0]
		if *liveFlag {
			live(arg, loaders)
			return
		}
		run(arg, loaders)
		return
	}
	repl(loaders)
}

func options(stdout io.Writer) []goatlang.RunOption {
	imports := map[string]string{}
	var opts []goatlang.RunOption
	if *treeFlag {
		opts = append(opts, goatlang.WithTreeDump(stdout))
	}
	if *codeFlag {
		opts = append(opts, goatlang.WithCodeDump(stdout))
	}
	opts = append(opts, goatlang.WithEvalImports(imports))
	return opts
}

func run(arg string, loaders []func(*goatlang.VM)) {
	root := "."
	sys := os.DirFS(root)
	opts := options(os.Stdout)
	vm := goatlang.NewVM(goatlang.WithLoaders(loaders...))
	if err := vm.Load(sys, arg, opts...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if _, err := vm.Call("main.main", 0); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func live(arg string, loaders []func(*goatlang.VM)) {
	root := "."
	sys := os.DirFS(root)
	liveCh := make(chan string, 256)
	const liveReloadCmd = "__liveReloadCmd"

	// init watcher
	w := watcher.New()
	defer w.Close()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename, watcher.Move)
	if err := w.AddRecursive(root); err != nil {
		log.Fatalln(err)
	}

	// init readline
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatalln(err)
	}
	defer rl.Close()
	log.SetOutput(rl.Stderr())
	defer log.SetOutput(os.Stderr)
	// HACK: because readline deadlocks on Close
	shutdown := func() {
		rl.Terminal.ExitRawMode()
		os.Exit(0)
	}
	defer shutdown()

	// init vm
	opts := options(rl.Stdout())
	vm := goatlang.NewVM(goatlang.WithStdout(rl.Stdout()), goatlang.WithLoaders(loaders...))
	vm.Set("builtin.__yield", goatlang.NewFunc(0, 0, func(v *goatlang.VM) {
		for {
			select {
			case line := <-liveCh:
				if line == liveReloadCmd {
					if err := vm.Load(sys, arg, opts...); err != nil {
						fmt.Fprintln(rl.Stderr(), err)
					}
				} else {
					eval(vm, sys, rl.Stdout(), rl.Stderr(), line, opts...)
				}
			default:
				return
			}
		}
	}))

	// load package
	if err := vm.Load(sys, arg, opts...); err != nil {
		fmt.Fprintln(rl.Stderr(), err)
		return
	}

	// start watcher
	go func() {
		for {
			select {
			case event := <-w.Event:
				_ = event
				liveCh <- liveReloadCmd
			case err := <-w.Error:
				log.Println(err)
			case <-w.Closed:
				return
			}
		}
	}()
	go func() {
		if err := w.Start(time.Second / 10); err != nil {
			log.Println(err)
		}
	}()
	// start repl
	go func() {
		for {
			line, err := input(rl)
			if err != nil {
				shutdown()
				return
			}
			liveCh <- line
		}
	}()

	// call main()
	if _, err := vm.Call("main.main", 0); err != nil {
		fmt.Fprintln(rl.Stderr(), err)
		return
	}
}

func input(rl *readline.Instance) (string, error) {
	var line string
	for {
		l, err := rl.Readline()
		if err != nil {
			return "", err
		}
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		line += l
		if !strings.HasSuffix(line, "\\") {
			break
		}
		line = line[:len(line)-1] + "\n"
	}
	return line, nil
}

func repl(loaders []func(*goatlang.VM)) {
	sys := os.DirFS(".")
	opts := options(os.Stdout)
	vm := goatlang.NewVM(goatlang.WithLoaders(loaders...))
	rl, err := readline.New("> ")
	if err != nil {
		log.Fatalln(err)
	}
	defer rl.Close()
	for {
		line, err := input(rl)
		if err != nil {
			return
		}
		eval(vm, sys, os.Stdout, os.Stderr, line, opts...)
	}
}

func eval(vm *goatlang.VM, sys fs.FS, stdout, stderr io.Writer, line string, opts ...goatlang.RunOption) {
	rets, err := vm.Eval(sys, "stdin", line, opts...)
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	stack := make([]goatlang.Value, len(rets))
	copy(stack, rets)
	vm.Set("$", goatlang.NewSlice(0, stack))
	var s []string
	for _, v := range stack {
		s = append(s, v.String())
	}
	if len(s) > 0 {
		fmt.Fprintln(stdout, "$ =", s)
	}
}
