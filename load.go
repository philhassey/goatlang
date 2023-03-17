package goatlang

import (
	"errors"
	"fmt"
	"go/build/constraint"
	"io/fs"
	"os"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type pkgList []*token

func (p pkgList) String() string {
	var res []string
	for _, t := range p {
		s := t.String()
		res = append(res, s[3:len(s)-1])
	}
	return strings.Join(res, " ")
}

func loadImports(sys fs.FS, topPkg string, top *token) (pkgList, error) {
	packages := map[string]*token{}
	deps := map[string]map[string]bool{}
	todo := []string{topPkg}
	for len(todo) > 0 {
		pkg := todo[len(todo)-1]
		todo = todo[:len(todo)-1]
		if _, ok := packages[pkg]; ok {
			continue
		}
		var p *token
		if pkg == topPkg {
			p = top
		} else {
			var err error
			p, err = rawLoadPackage(sys, pkg)
			if errors.Is(err, os.ErrNotExist) {
				packages[pkg] = &token{}
				continue
			} else if err != nil {
				return nil, fmt.Errorf("error in loadPackage: %w", err)
			}
			p = treeSort(p)
		}
		packages[pkg] = p
		deps[pkg] = map[string]bool{}
		for _, t := range p.Tokens {
			if t.Symbol != "import" {
				continue
			}
			for i := 1; i < len(t.Tokens); i += 2 {
				pk := t.Tokens[i]
				todo = append(todo, pk.Unquote())
				deps[pkg][pk.Unquote()] = true
			}
		}
	}
	keys := maps.Keys(packages)
	slices.Sort(keys)
	var res []*token
	for len(packages) > 0 {
		var pkg string
		for _, k := range keys {
			if len(deps[k]) > 0 {
				continue
			}
			pkg = k
			break
		}
		delete(deps, pkg)
		for _, d := range deps {
			delete(d, pkg)
		}
		tok := packages[pkg]
		delete(packages, pkg)
		idx := slices.Index(keys, pkg)
		keys = slices.Delete(keys, idx, idx+1)
		if len(tok.Tokens) == 0 {
			tok = &token{
				Symbol: "_",
				Text:   "_",
			}
			tok.Append(&token{
				Symbol: "package",
				Text:   "package",
			})
			tok.Tokens[0].Append(&token{Text: "_"})
			tok.Tokens[0].Append(&token{Text: pkg})
		}
		res = append(res, tok)
	}
	// for _, pkg := range res {
	// 	if len(pkg.Tokens) > 0 && pkg.Symbol == "_" {
	// 		fmt.Println("load", pkg.Symbol, pkg.Tokens[0].Text, pkg.Tokens[0].Tokens)
	// 	}
	// }
	return res, nil
}

func loadPackage(sys fs.FS, topPkg string) (pkgList, error) {
	p, err := rawLoadPackage(sys, topPkg)
	if err != nil {
		return nil, fmt.Errorf("error in loadPackage: %w", err)
	}
	p = treeSort(p)
	return loadImports(sys, topPkg, p)
}

func loadFile(sys fs.FS, fname string) (pkgList, error) {
	p, err := rawLoadFile(sys, fname, false)
	if err != nil {
		return nil, fmt.Errorf("error in loadFile: %w", err)
	}
	p = treeSort(p)
	return loadImports(sys, "_", p) // not sure why _
}

func checkConstraint(s string) (bool, error) {
	line := strings.Split(strings.TrimSpace(s), "\n")[0]
	if !constraint.IsGoBuild(line) {
		return true, nil
	}
	expr, err := constraint.Parse(line)
	if err != nil {
		return false, err
	}
	ok := func(t string) bool { return t == "goat" }
	return expr.Eval(ok), nil

}
func rawLoadFile(sys fs.FS, fname string, checkBC bool) (*token, error) {
	b, err := fs.ReadFile(sys, fname)
	if err != nil {
		return nil, fmt.Errorf("error in ReadFile: %w", err)
	}
	if checkBC {
		ok, err := checkConstraint(string(b))
		if err != nil {
			return nil, fmt.Errorf("error in constraint: %w", err)
		}
		if !ok {
			return &token{}, nil
		}
	}
	tokens, err := tokenize(fname, string(b))
	if err != nil {
		return nil, fmt.Errorf("error in tokenize: %w", err)
	}
	tree, err := parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("error in parse: %w", err)
	}
	return tree, nil
}

func rawLoadPackage(sys fs.FS, pkg string) (*token, error) {
	var matches []string
	parts := append([]string{"vendor"}, strings.Split(pkg, "/")...)
	for len(parts) > 0 {
		shortPkg := strings.Join(parts, "/")
		var err error
		matches, err = fs.Glob(sys, shortPkg+"/*.go")
		if err != nil {
			return nil, fmt.Errorf("error in Glob: %w", err)
		}
		if len(matches) > 0 {
			break
		}
		parts = parts[1:]
	}
	if len(matches) == 0 {
		return nil, os.ErrNotExist
	}
	var files []*token
	pkgs := map[string]bool{}
	for _, fname := range matches {
		tree, err := rawLoadFile(sys, fname, true)
		if err != nil {
			return nil, fmt.Errorf("error in loadFile: %w", err)
		}
		if len(tree.Tokens) == 0 {
			continue
		}
		first := tree.Tokens[0]
		if first.Symbol != "package" {
			return nil, fmt.Errorf("expected package in: %v", fname)
		}
		pkgs[first.Tokens[0].Text] = true
		files = append(files, tree)
	}
	if len(pkgs) == 0 {
		return nil, os.ErrNotExist
	}
	if len(pkgs) > 1 {
		return nil, fmt.Errorf("multiple packages found in: %v", pkg)
	}
	tree := joinFiles(files)
	for _, tok := range tree.Tokens {
		if tok.Symbol == "package" && tok.Tokens[0].Text != "main" && tok.Tokens[0].Text != pkg {
			exp := symAtPos(tok.Pos, "(string)")
			exp.Text = pkg
			tok.Append(exp)
		}
	}

	return tree, nil
}

func joinFiles(files []*token) *token {
	tok := symAtPos(files[0].Pos, "_")
	for i, t := range files {
		if i == 0 {
			tok.Tokens = append(tok.Tokens, t.Tokens...)
			continue
		}
		tok.Tokens = append(tok.Tokens, t.Tokens[1:]...)
	}
	return tok
}
