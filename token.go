package goatlang

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type token struct {
	Pos    scanner.Position
	Symbol string
	Text   string
	Tokens []*token
}

func (t *token) Copy() *token {
	toks := make([]*token, len(t.Tokens))
	for i, tt := range t.Tokens {
		toks[i] = tt.Copy()
	}
	return &token{
		Pos:    t.Pos,
		Symbol: t.Symbol,
		Text:   t.Text,
		Tokens: toks,
	}
}

func (t *token) Replace(sym, newSym, newText string) {
	if t.Symbol == sym {
		t.Symbol = newSym
		t.Text = newText
		t.Tokens = nil
		return
	}
	for _, tt := range t.Tokens {
		tt.Replace(sym, newSym, newText)
	}
}

func (t *token) String() string {
	if len(t.Tokens) > 0 {
		var tt []string
		for _, v := range t.Tokens {
			tt = append(tt, v.String())
		}
		return fmt.Sprintf("(%v %v)", t.Text, strings.Join(tt, " "))
	}
	return t.Text
}

func (t *token) Char() rune {
	value, _, _, _ := strconv.UnquoteChar(t.Text[1:len(t.Text)-1], '\\')
	return value
}

func (t *token) rename(v string) {
	t.Symbol, t.Text = v, v
}

func (t *token) Int() int {
	if len(t.Text) > 2 && t.Text[:2] == "0x" {
		v, err := strconv.ParseInt(t.Text[2:], 16, 0)
		if err != nil {
			panicf("error parsing hex: %v", err)
		}
		return int(v)
	}
	if len(t.Text) > 1 && t.Text[0] == '0' {
		v, err := strconv.ParseInt(t.Text[1:], 8, 0)
		if err != nil {
			panicf("error parsing octal: %v", err)
		}
		return int(v)
	}
	v, err := strconv.Atoi(t.Text)
	if err != nil {
		panicf("error parsing int: %v", err)
	}
	return int(v)
}

func (t *token) Unquote() string {
	v, err := strconv.Unquote(t.Text)
	if err != nil {
		panicf("error parsing string: %v", err)
	}
	return v
}

func (t *token) Float64() float64 {
	v, err := strconv.ParseFloat(t.Text, 64)
	if err != nil {
		panicf("error parsing float: %v", err)
	}
	return v
}

func (t *token) Append(b *token) {
	t.Tokens = append(t.Tokens, b)
}

var symMap map[int]string = map[int]string{
	scanner.Int:       "(int)",
	scanner.Float:     "(float)",
	scanner.Char:      "(char)",
	scanner.String:    "(string)",
	scanner.RawString: "(string)",
}

// tokenize a string into a list of Tokens
func tokenize(filename string, in string) ([]*token, error) {
	var res []*token
	var err error
	var s scanner.Scanner
	s.Init(strings.NewReader(in))
	s.Mode = scanner.ScanIdents | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanFloats | scanner.ScanComments | scanner.SkipComments
	s.Error = func(s *scanner.Scanner, msg string) {
		err = fmt.Errorf("%v: %v", s.Position, msg)
	}
	const symChars = "`~!.#$%^&*()-=+[{]}\\|;:,<.>/?"
	s.Filename = filename
	for ch := s.Scan(); err == nil && ch != scanner.EOF; ch = s.Scan() {
		if strings.ContainsRune(symChars, ch) {
			sym := string(ch)
			if strings.ContainsRune(symChars, ch) {
				pk := s.Peek()
				sym2 := sym + string(pk)
				if _, ok := symbols[sym2]; ok || sym2 == ".." {
					sym = sym2
					s.Scan()
					sym3 := sym + string(s.Peek())
					if _, ok := symbols[sym3]; ok {
						sym = sym3
						s.Scan()
					}
				}
			}
			res = append(res, &token{Pos: s.Position, Symbol: sym, Text: sym})
			continue
		}
		sym := "(name)"
		text := s.TokenText()
		if v, ok := symMap[int(ch)]; ok {
			sym = v
		} else if symbols[text] != nil {
			sym = text
		}
		res = append(res, &token{Pos: s.Position, Symbol: sym, Text: text})
	}
	res = append(res, &token{Pos: s.Pos(), Symbol: "(eof)", Text: "(eof)"})
	return res, err
}
