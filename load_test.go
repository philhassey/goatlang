package goatlang

import (
	"bytes"
	"io"
	"io/fs"
	"strings"
	"testing"

	"golang.org/x/exp/slices"
)

func TestLoadPackage(t *testing.T) {
	tests := []struct {
		Name    string
		Package string
		FS      mapFS
		Want    string
	}{
		{"singleFile", "main", mapFS{
			"main/main.go": `package main; func main(){}`,
		}, `(package main) (function main (func arguments returns block))`},
		{"multiFile", "main", mapFS{
			"main/main.go":   `package main; func main(){}`,
			"main/consts.go": `package main; const C = 42`,
		}, `(package main) (const (, C) 42) (function main (func arguments returns block))`},
		{"import", "main", mapFS{
			"main/main.go": `package main; import "util"`,
			"util/util.go": `package util; const U = 42`,
		}, `(package util) (const (, U) 42) (package main) (import util "util")`},
		{"rhombus", "main", mapFS{
			"main/main.go": `package main; import ("a" "b")`,
			"a/pkg.go":     `package a; import "c"`,
			"b/pkg.go":     `package b;`,
			"c/pkg.go":     `package c;`,
		}, `(package b) (package c) (package a) (import c "c") (package main) (import a "a" b "b")`},
		{"diamond", "main", mapFS{
			"main/main.go": `package main; import ("a" "b")`,
			"a/pkg.go":     `package a; import "c"`,
			"b/pkg.go":     `package b; import "c"`,
			"c/pkg.go":     `package c;`,
		}, `(package c) (package a) (import c "c") (package b) (import c "c") (package main) (import a "a" b "b")`},
		{"virtual", "main", mapFS{
			"main/main.go": `package main; import "fmt"`,
		}, `(package _ fmt) (package main) (import fmt "fmt")`},
		{"shorten", "main", mapFS{
			"main/main.go":    `package main; import "example.com/test/ext"`,
			"test/ext/pkg.go": `package ext;`,
		}, `(package ext example.com/test/ext) (package main) (import ext "example.com/test/ext")`},
		{"contraint", "main", mapFS{
			"main/main_skip.go":    "//go:build !goat\npackage main; const Skip = true",
			"main/main_include.go": "//go build goat\npackage main; const Include = true",
			"main/main.go":         `package main; const None = true`,
		}, `(package main) (const (, None) true) (const (, Include) true)`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tree, err := loadPackage(row.FS, row.Package)
			if err != nil {
				t.Fatalf("LoadPackage error: %v", err)
			}
			// tree = treeSort(tree)
			s := tree.String()
			if s != row.Want {
				t.Fatalf("Parse\n got %v\nwant %v", s, row.Want)
			}
		})
	}
}

func TestLoadPackage_Exec(t *testing.T) {
	tests := []struct {
		Name    string
		Package string
		FS      mapFS
		Want    string
	}{
		{"shorten", "main", mapFS{
			"main/main.go":    `package main; import "example.com/test/ext"; func test() any { return ext.F() }`,
			"test/ext/pkg.go": `package ext; func F() int { return 42 }`,
		}, `42`},
		{"lostImportsBug", "main", mapFS{
			"main/main.go":    `package main; import "example.com/test/ext"; func test() any { return ext.F() }`,
			"main/a.go":       `package main; import "fmt"`,
			"main/b.go":       `package main; import "fmt"`,
			"main/c.go":       `package main; import "fmt"`,
			"main/x.go":       `package main; import "fmt"`,
			"main/y.go":       `package main; import "fmt"`,
			"main/z.go":       `package main; import "fmt"`,
			"test/ext/pkg.go": `package ext; import ( "fmt" "math" ); func F() int { return 42 }`,
		}, `42`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tree, err := loadPackage(row.FS, row.Package)
			if err != nil {
				t.Fatalf("LoadPackage error: %v", err)
			}
			// tree = treeSort(tree)
			vm := NewVM()
			codes, slots, err := compilePkgs(vm.globals, tree, false)
			if err != nil {
				t.Fatalf("Compile error: %v", err)
			}
			_, err = vm.run(codes, slots)
			if err != nil {
				t.Fatalf("Exec error: %v", err)
			}
			rets, err := vm.Call("main.test", 1)
			if err != nil {
				t.Fatalf("Call error: %v", err)
			}
			var ts []string
			for _, s := range rets {
				ts = append(ts, s.String())
			}
			s := strings.Join(ts, " ")
			if s != row.Want {
				t.Fatalf("Call got %v want %v", s, row.Want)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		Name     string
		FileName string
		FS       mapFS
		Want     string
	}{
		{"fileWithImport", "main/main.go", mapFS{
			"main/main.go":  `package main; import "util"`,
			"main/extra.go": `package main; const junk = "garbage"`,
			"util/util.go":  `package util; const U = 42`,
		}, `(package util) (const (, U) 42) (package main) (import util "util")`},
		{"fileWithIgnore", "main/main.go", mapFS{
			"main/main.go": "//go:build ignore\npackage main; const OK = true",
		}, `(package main) (const (, OK) true)`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			tree, err := loadFile(row.FS, row.FileName)
			if err != nil {
				t.Fatalf("LoadFile error: %v", err)
			}
			// tree = treeSort(tree)
			s := tree.String()
			if s != row.Want {
				t.Fatalf("Parse\n got %v\nwant %v", s, row.Want)
			}
		})
	}
}

func TestLoadPackage_error(t *testing.T) {
	tests := []struct {
		Name    string
		Package string
		FS      mapFS
		Err     string
	}{
		{"parseError", "main", mapFS{
			"main/main.go": `++`,
		}, `error in parse:`},
		{"tokenizeError", "main", mapFS{
			"main/main.go": `"`,
		}, `error in tokenize:`},
		{"globErr", "\\", mapFS{}, `error in Glob:`},
		{"readFileError", "main", mapFS{
			"main/main.go": ``,
		}, `error in ReadFile:`},
		{"nothing", "main", mapFS{}, `error in loadPackage:`},
		{"badDep", "main", mapFS{
			"main/main.go":   `package main; import "bad"`,
			"bad/invalid.go": `"`,
		}, `error in loadPackage:`},
		{"badConstraints", "main", mapFS{
			"main/main.go": `//go:build ???`,
		}, `error in constraint:`},

		{"noPackages", "main", mapFS{
			"main/main.go": `//`,
		}, `file does not exist`},
		{"multiplePackages", "main", mapFS{
			"main/a.go": `package a`,
			"main/b.go": `package b`,
		}, `multiple packages found in: main`},
		{"expectedPackage", "main", mapFS{
			"main/main.go": `func f() {}`,
		}, `expected package in: main`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			_, err := loadPackage(row.FS, row.Package)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("LoadPackage error got %v want %v", err, row.Err)
			}
		})
	}
}

func TestLoadFile_error(t *testing.T) {
	tests := []struct {
		Name     string
		FileName string
		FS       mapFS
		Err      string
	}{
		{"notExist", "main/main.go", mapFS{}, `file does not exist`},
	}
	for _, row := range tests {
		t.Run(row.Name, func(t *testing.T) {
			_, err := loadFile(row.FS, row.FileName)
			if err == nil || !strings.Contains(err.Error(), row.Err) {
				t.Fatalf("LoadFile error got %v want %v", err, row.Err)
			}
		})
	}
}

// https://pkg.go.dev/testing/fstest official alternative, doesn't simulate errors

type mapFS map[string]string

func (m mapFS) Open(name string) (fs.File, error) {
	b, err := m.ReadFile(name)
	return fsFile{r: bytes.NewReader(b)}, err
}

func (m mapFS) ReadFile(name string) ([]byte, error) {
	if v, ok := m[name]; ok {
		if len(v) == 0 {
			return nil, fs.ErrPermission // for error tests
		}
		return []byte(v), nil
	}
	return nil, fs.ErrNotExist
}

func (m mapFS) ReadDir(name string) ([]fs.DirEntry, error) {
	prefix := name + "/"
	var res []fs.DirEntry
	for name := range m {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		res = append(res, dirEntry{strings.TrimPrefix(name, prefix)})
	}
	if len(res) == 0 {
		return nil, fs.ErrNotExist
	}
	slices.SortFunc(res, func(a, b fs.DirEntry) bool {
		return a.Name() < b.Name()
	})
	return res, nil
}

type fsFile struct{ r io.Reader }

func (f fsFile) Stat() (fs.FileInfo, error) { return nil, fs.ErrPermission }
func (f fsFile) Read(b []byte) (int, error) { return f.r.Read(b) }
func (f fsFile) Close() error               { return nil }

type dirEntry struct{ name string }

func (d dirEntry) Name() string               { return d.name }
func (d dirEntry) IsDir() bool                { return false }
func (d dirEntry) Type() fs.FileMode          { return 0 }
func (d dirEntry) Info() (fs.FileInfo, error) { return nil, fs.ErrPermission }
