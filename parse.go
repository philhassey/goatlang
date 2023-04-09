package goatlang

import (
	"fmt"
	"text/scanner"

	"golang.org/x/exp/slices"
)

// parse a slice of Tokens into a tree of Tokens
// https://crockford.com/javascript/tdop/tdop.html
func parse(tokens []*token) (res *token, err error) {
	res = &token{Text: "_"}
	p := &parser{Tokens: tokens}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v: %v", p.Token.Pos, r)
		}
	}()
	p.Next()
	for p.Token.Symbol != "(eof)" {
		if tok := p.Statement(); tok != nil {
			res.Append(tok)
		}
	}
	return res, nil
}

type parser struct {
	Token  *token
	Tokens []*token
	N      int
	mask   []string
	Depth  int
}

func symAtPos(pos scanner.Position, symbol string) *token {
	return &token{Pos: pos, Symbol: symbol, Text: symbol}
}

func panicf(msg string, args ...interface{}) {
	panic(fmt.Sprintf(msg, args...))
}

func (p *parser) Advance(sym string) *token {
	t := p.Token
	if t.Symbol != sym {
		panicf("advance got %v want %v", t.Symbol, sym)
	}
	p.Next()
	return t
}

func (p *parser) Statement() *token {
	tok := p.Expression(0)
	if tok == nil {
		return nil
	}
	if tok.Symbol == "call" {
		tok.Tokens[2].Text = "0"
	}
	return tok
}

func (p *parser) Block(sym, pre, post string) *token {
	p.Advance(pre)
	res := symAtPos(p.Token.Pos, sym)
	for p.Token.Symbol != post {
		if tok := p.Statement(); tok != nil {
			res.Append(tok)
		}
	}
	p.Advance(post)
	return res
}

func (p *parser) Next() *token {
	p.Token = p.Tokens[p.N]
	p.N++
	return p.Token
}

func (p *parser) Expression(rbp int, mask ...string) *token {
	p.Depth++
	tmp := p.mask
	p.mask = mask
	tok := p.doExpression(rbp)
	p.mask = tmp
	p.Depth--
	return tok
}

func (p *parser) doExpression(rbp int) *token {
	t := p.Token
	p.Next()
	left := getSymbol(t).Nud(p, t)
	for rbp < getSymbol(p.Token).Lbp && !slices.Contains(p.mask, p.Token.Symbol) {
		t = p.Token
		p.Next()
		left = getSymbol(t).Led(p, t, left)
	}
	return left
}
