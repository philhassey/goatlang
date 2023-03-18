package goatlang

import (
	"fmt"
	"strings"
	"text/scanner"
)

type symbol struct {
	Lbp int                                  // Left Binding Power
	Nud func(*parser, *token) *token         // Null Denotation: values & prefix ops
	Led func(*parser, *token, *token) *token // Left Denotation: infix & suffix ops
}

func nudSelf(p *parser, t *token) *token { return t }
func ledInfix(p *parser, t *token, left *token) *token {
	t.Append(left)
	t.Append(p.doExpression(getSymbol(t).Lbp))
	return t
}
func ledPostfix(p *parser, t *token, left *token) *token {
	t.Append(left)
	return t
}
func nudNil(p *parser, t *token) *token  { return nil }
func nullNud(p *parser, t *token) *token { panicf("null nud: %v", t.Symbol); return nil }
func nullLed(p *parser, t *token, left *token) *token {
	panicf("null led: %v", t.Symbol)
	return nil
}

func plural(t *token) *token {
	if t.Symbol != "," {
		tmp := t
		t = symAtPos(t.Pos, ",")
		t.Tokens = []*token{tmp}
	}
	return t
}

func blankAtPos(pos scanner.Position) *token {
	tok := symAtPos(pos, "(name)")
	tok.Text = "_"
	return tok
}

func assignResize(left, right *token) {
	if right.Symbol == "call" {
		right.Tokens[2].Text = fmt.Sprint(len(left.Tokens))
	}
	if right.Symbol == "index" && len(left.Tokens) > 1 {
		right.rename("indexOk")
	}
}

func assignLed(p *parser, t *token, left *token) *token {
	t.Append(left)
	t.Append(p.Expression(getSymbol(t).Lbp))
	t.Tokens[0] = plural(t.Tokens[0])
	assignResize(t.Tokens[0], t.Tokens[1])
	return t
}

func getArgs(p *parser) (*token, *token) {
	p.Advance("(")
	args := symAtPos(p.Token.Pos, "arguments")
	for p.Token.Symbol != ")" {
		var name *token
		if p.Token.Symbol == "(name)" {
			name = p.Advance("(name)")
		} else {
			name = blankAtPos(p.Token.Pos)
		}
		args.Append(name)
		if p.Token.Symbol != "," && p.Token.Symbol != ")" {
			name.Append(getType(p))
		}
		if p.Token.Symbol != "," {
			break
		}
		p.Advance(",")
	}
	var typ *token
	for i := len(args.Tokens) - 1; i >= 0; i-- {
		arg := args.Tokens[i]
		if len(arg.Tokens) > 0 {
			typ = arg.Tokens[0]
		} else {
			arg.Append(typ)
		}
	}
	last := p.Token
	p.Advance(")")
	return args, last
}

func getReturns(last *token, p *parser) *token {
	returns := symAtPos(p.Token.Pos, "returns")
	if p.Token.Symbol == "(" {
		p.Advance("(")
		for p.Token.Symbol != ")" {
			returns.Append(getType(p))
			if p.Token.Symbol != "," {
				break
			}
			p.Advance(",")
		}
		p.Advance(")")
	} else if p.Token.Symbol != ")" && p.Token.Symbol != "," && p.Token.Symbol != "{" && p.Token.Symbol != "}" && p.Token.Symbol != ";" && p.Token.Pos.Line == last.Pos.Line {
		returns.Append(getType(p))
	}
	return returns
}

func funcNud(p *parser, t *token) *token {
	var klass *token
	wrap := symAtPos(t.Pos, "function")
	if p.Token.Symbol == "(" {
		tmp, _ := getArgs(p)
		klass = tmp.Tokens[0]
		wrap.rename("method")
		name := symAtPos(klass.Pos, "(name)")
		name.Text = klass.Tokens[0].Text
		wrap.Append(name)
		wrap.Append(p.Advance("(name)"))
	} else {
		wrap.Append(p.Advance("(name)"))
		if wrap.Tokens[0].Text == "init" {
			wrap = symAtPos(t.Pos, "init")
		}
	}
	args, last := getArgs(p)
	if klass != nil {
		args.Tokens = append([]*token{klass}, args.Tokens...)
	}
	t.Append(args)
	returns := getReturns(last, p)
	t.Append(returns)
	t.Append(p.Block("block", "{", "}"))
	wrap.Append(t)
	return wrap
}

func returnNud(p *parser, t *token) *token {
	for p.Token.Symbol != "}" && p.Token.Symbol != ";" && p.Token.Symbol != "case" && p.Token.Symbol != "default" {
		t.Append(p.Expression(commaBP))
		if p.Token.Symbol != "," {
			break
		}
		p.Advance(",")
	}
	if len(t.Tokens) == 1 {
		if t.Tokens[0].Symbol == "call" {
			t.Tokens[0].Tokens[2].Text = "-1"
		}
	}
	return t
}

func callLed(p *parser, t *token, left *token) *token {
	call := symAtPos(p.Token.Pos, "call")
	call.Append(left)
	arguments := symAtPos(p.Token.Pos, "arguments")
	call.Append(arguments)
	for p.Token.Symbol != ")" {
		arguments.Append(p.Expression(commaBP))
		if p.Token.Symbol != "," {
			break
		}
		p.Advance(",")
	}
	call.Append(&token{Pos: p.Token.Pos, Symbol: "returns", Text: "1"})
	p.Advance(")")
	return call
}

func ifNud(p *parser, t *token) *token {
	top := t
	for {
		first := p.Expression(0, "{")
		if p.Token.Symbol == ";" {
			t.Append(first)
			p.Advance(";")
			t.Append(p.Expression(0, "{"))
		} else {
			t.Append(symAtPos(first.Pos, "~"))
			t.Append(first)
		}
		t.Append(p.Block("block", "{", "}"))
		if p.Token.Symbol != "else" {
			break
		}
		p.Advance("else")
		if p.Token.Symbol != "if" {
			t.Append(p.Block("block", "{", "}"))
			break
		}
		t.Append(p.Token)
		t = p.Token
		p.Advance("if")
	}
	return top
}

func forNud(p *parser, t *token) *token {
	if p.Token.Symbol == "{" {
		t.Append(symAtPos(t.Pos, "~"))
		t.Append(symAtPos(t.Pos, "~"))
		t.Append(symAtPos(t.Pos, "~"))
		t.Append(p.Block("block", "{", "}"))
		return t
	}

	first := p.Expression(0, "{")
	if first.Symbol == "range" {
		tok := first
		tok.Append(blankAtPos(t.Pos))
		tok.Append(blankAtPos(t.Pos))
		tok.Append(p.Expression(0, "{"))
		tok.Append(p.Block("block", "{", "}"))
		return tok
	}

	if first.Symbol == ":=" && first.Tokens[1].Symbol == "range" {
		tok := symAtPos(t.Pos, "range")
		left := plural(first.Tokens[0])
		if len(left.Tokens) < 2 {
			left.Append(blankAtPos(t.Pos))
		}
		tok.Tokens = left.Tokens
		tok.Append(p.Expression(0, "{"))
		tok.Append(p.Block("block", "{", "}"))
		return tok
	}

	if p.Token.Symbol == "{" {
		t.Append(symAtPos(t.Pos, "~"))
		t.Append(first)
		t.Append(symAtPos(t.Pos, "~"))
		t.Append(p.Block("block", "{", "}"))
		return t
	}

	t.Append(first)
	p.Advance(";")
	t.Append(p.Expression(0, "{"))
	p.Advance(";")
	t.Append(p.Expression(0, "{"))
	t.Append(p.Block("block", "{", "}"))
	return t
}

// func variadicNud(p *parser, t *token) *token {
// 	p.Advance("(")
// 	for p.Token.Symbol != ")" {
// 		t.Append(p.Expression(commaBP))
// 		if p.Token.Symbol != ")" {
// 			p.Advance(",")
// 		}
// 	}
// 	p.Advance(")")
// 	return t
// }

func makeNud(p *parser, t *token) *token {
	p.Advance("(")
	typ := getType(p)
	t.Append(typ)
	if p.Token.Symbol == "," {
		p.Advance(",")
		t.Append(p.Expression(commaBP))
	}
	p.Advance(")")
	return t
}

func packageNud(p *parser, t *token) *token {
	t.Append(p.Advance("(name)"))
	return t
}

func appendAlias(p *parser, t *token) {
	tok := p.Advance("(string)")
	alias := symAtPos(tok.Pos, "(name)")
	parts := strings.Split(tok.Unquote(), "/")
	alias.Text = parts[len(parts)-1]
	t.Append(alias)
	t.Append(tok)
}

func importNud(p *parser, t *token) *token {
	if p.Token.Symbol == "(" {
		p.Advance("(")
		for p.Token.Symbol != ")" {
			if p.Token.Symbol == "(name)" {
				t.Append(p.Advance("(name)"))
				t.Append(p.Advance("(string)"))
			} else {
				appendAlias(p, t)
			}
		}
		p.Advance(")")
		return t
	}
	appendAlias(p, t)
	return t
}

func commaLed(p *parser, t *token, left *token) *token {
	t.Append(left)
	for {
		t.Append(p.Expression(commaBP))
		if p.Token.Symbol != "," {
			break
		}
		p.Advance(",")
	}
	return t
}

func getType(p *parser) *token {
	t := p.Token
	p.Next()
	switch t.Symbol {
	case "[]":
		t.Append(getType(p))
	case "map":
		p.Advance("[")
		t.Append(getType(p))
		p.Advance("]")
		t.Append(getType(p))
	case "any", "float64", "int", "int32", "uint32", "uint", "rune", "byte", "uint8", "int8", "uint16", "int16", "uint64", "int64", "bool", "string", "error":
	case "(name)":
		if p.Token.Symbol == "." {
			p.Advance(".")
			tmp := t
			t = symAtPos(t.Pos, ".")
			t.Append(tmp)
			t.Append(p.Advance("(name)"))
		}
	case "*":
		t = getType(p)
	case "interface":
		p.Advance("{")
		for p.Token.Symbol != "}" {
			if p.Token.Symbol == ";" { // HACK: for tests
				p.Advance(";")
				continue
			}
			name := p.Advance("(name)")
			t.Append(name)
			args, last := getArgs(p)
			t.Append(args)
			rets := getReturns(last, p)
			t.Append(rets)
		}
		p.Advance("}")
	case "func":
		args, last := getArgs(p)
		t.Append(args)
		rets := getReturns(last, p)
		t.Append(rets)
	case "struct":
		p.Advance("{")
		for p.Token.Symbol != "}" {
			var names []*token
			for {
				if p.Token.Symbol == ";" { // HACK: for tests
					p.Advance(";")
					continue
				}
				names = append(names, p.Advance("(name)"))
				if p.Token.Symbol != "," {
					break
				}
				p.Advance(",")
			}
			typ := getType(p)
			for _, n := range names {
				t.Append(n)
				t.Append(typ)
			}
		}
		p.Advance("}")
	case "...":
		t.Append(getType(p))
	default:
		panicf("type: unexpected symbol: %v", t.Symbol)
	}
	return t
}

func getDecl(p *parser, kind string) *token {
	if p.Token.Symbol == ";" { // HACK: for tests
		p.Advance(";")
		return nil
	}
	decl := symAtPos(p.Token.Pos, kind)
	left := symAtPos(p.Token.Pos, ",")
	for {
		left.Append(p.Advance("(name)"))
		if p.Token.Symbol != "," {
			break
		}
		p.Advance(",")
	}
	decl.Append(left)
	if p.Token.Symbol == ";" || p.Token.Symbol == ")" || p.Token.Pos.Line != left.Pos.Line {
		return decl
	}
	if p.Token.Symbol == "=" {
		if kind == "var" {
			decl.rename(":=")
		}
		p.Advance("=")
		right := p.Expression(0)
		decl.Append(right)
		assignResize(left, right)
		return decl
	}
	typ := getType(p)
	lt := left.Tokens
	for i := 0; i < len(lt); i++ {
		left.Tokens[i].Append(typ)
	}
	if p.Token.Symbol == "=" {
		p.Advance("=")
		right := p.Expression(0)
		decl.Append(right)
		assignResize(left, right)
		return decl
	}
	right := symAtPos(typ.Pos, ",")
	decl.Append(right)
	return decl
}

func declareNud(p *parser, t *token) *token {
	kind := t.Symbol
	if p.Token.Symbol == "(" {
		p.Advance("(")
		block := symAtPos(t.Pos, "block")
		for p.Token.Symbol != ")" {
			if decl := getDecl(p, kind); decl != nil {
				block.Append(decl)
			}
		}
		p.Advance(")")
		return block
	}
	decl := getDecl(p, kind)
	return decl
}

func constNud(p *parser, t *token) *token {
	kind := t.Symbol
	if p.Token.Symbol == "(" {
		t.Append(plural(symAtPos(t.Pos, ",")))
		t.Append(plural(symAtPos(t.Pos, ",")))
		p.Advance("(")
		var prev *token
		for p.Token.Symbol != ")" {
			if decl := getDecl(p, kind); decl != nil {
				for _, tt := range plural(decl.Tokens[0]).Tokens {
					t.Tokens[0].Append(tt)
				}
				if len(decl.Tokens) > 1 {
					for _, tt := range plural(decl.Tokens[1]).Tokens {
						prev = tt.Copy()
						tt.Replace("iota", "(int)", fmt.Sprint(len(t.Tokens[1].Tokens)))
						t.Tokens[1].Append(tt)
					}
				} else {
					for range plural(decl.Tokens[0]).Tokens {
						tt := prev.Copy()
						tt.Replace("iota", "(int)", fmt.Sprint(len(t.Tokens[1].Tokens)))
						t.Tokens[1].Append(tt)
					}
				}
			}
		}
		p.Advance(")")
		return t
	}
	decl := getDecl(p, kind)
	return decl
}

func sliceNud(p *parser, t *token) *token {
	typ := getType(p)
	t.Append(typ)
	data := symAtPos(p.Token.Pos, ",")
	t.Append(data)
	if p.Token.Symbol == "(" { // convert
		return t
	}
	data.Tokens = getData(p).Tokens
	return t
}
func mapNud(p *parser, t *token) *token {
	p.Advance("[")
	t.Append(getType(p))
	p.Advance("]")
	typ := getType(p)
	t.Append(typ)
	t.Append(getData(p))
	return t
}

func dataNud(p *parser, t *token) *token {
	p.N -= 2
	p.Next()
	return getData(p)
}

func getData(p *parser) *token {
	t := symAtPos(p.Token.Pos, ";")
	p.Advance("{")
	for p.Token.Symbol != "}" {
		t.Append(p.Expression(commaBP))
		if p.Token.Symbol == ":" {
			t.Symbol, t.Text = ":", ":"
			p.Advance(":")
			t.Append(p.Expression(commaBP))
		}
		if p.Token.Symbol != "," {
			break
		}
		p.Advance(",")
	}
	p.Advance("}")
	return t
}

func ellipsisLed(p *parser, t *token, left *token) *token {
	t.Append(left)
	return t
}

func indexLed(p *parser, t *token, left *token) *token {
	t.rename("index")
	t.Append(left)
	if p.Token.Symbol != ":" {
		t.Append(p.Expression(0))
	} else {
		t.Append(&token{Pos: p.Token.Pos, Symbol: "(int)", Text: "0"})
	}
	if p.Token.Symbol == ":" {
		t.rename("slice")
		p.Advance(":")
		if p.Token.Symbol != "]" {
			t.Append(p.Expression(0))
		} else {
			t.Append(&token{Pos: p.Token.Pos, Symbol: "(int)", Text: "-1"})
		}
	}
	p.Advance("]")
	return t
}

func negateNud(p *parser, t *token) *token {
	expr := p.doExpression(130) // higher BP for negation
	if expr.Symbol == "(int)" || expr.Symbol == "(float64)" {
		expr.Text = "-" + expr.Text
		return expr
	}
	t.rename("negate")
	t.Append(expr)
	return t
}

func complementNud(p *parser, t *token) *token {
	expr := p.doExpression(130) // higher BP for negation
	t.rename("complement")
	t.Append(expr)
	return t
}

func parenNud(p *parser, t *token) *token {
	expr := p.Expression(commaBP)
	p.Advance(")")
	return expr
}

func notNud(p *parser, t *token) *token {
	expr := p.doExpression(getSymbol(t).Lbp)
	t.Append(expr)
	return t
}

func typeNud(p *parser, t *token) *token {
	t.Append(p.Advance("(name)"))
	if p.Token.Symbol == "=" {
		p.Advance("=")
	}
	t.Append(getType(p))
	return t
}

func switchNud(p *parser, t *token) *token {
	if p.Token.Symbol != "{" {
		t.Append(p.Expression(0, "{"))
	} else {
		t.Append(symAtPos(t.Pos, "~"))
	}
	p.Advance("{")
	cases := symAtPos(p.Token.Pos, ",")
	t.Append(cases)
	for {
		if p.Token.Symbol == "case" {
			c := p.Advance("case")
			c.Append(p.Statement())
			p.Advance(":")
			c.Append(getCase(p))
			cases.Append(c)
		} else if p.Token.Symbol == "default" {
			c := p.Advance("default")
			p.Advance(":")
			c.Append(getCase(p))
			t.Append(c)
		} else {
			break
		}
	}
	if len(t.Tokens) < 3 {
		t.Append(symAtPos(p.Token.Pos, "~"))
	}
	p.Advance("}")
	return t
}

func getCase(p *parser) *token {
	res := symAtPos(p.Token.Pos, "block")
	for p.Token.Symbol != "}" && p.Token.Symbol != "case" && p.Token.Symbol != "default" {
		if tok := p.Statement(); tok != nil {
			res.Append(tok)
		}
	}
	return res
}

func newLed(p *parser, t *token, left *token) *token {
	tok := symAtPos(t.Pos, "new")
	tok.Append(left)
	p.N -= 2
	p.Next()
	tok.Append(getData(p))
	return tok
}

func stackNud(p *parser, t *token) *token {
	t.Symbol = "(name)"
	if p.Token.Symbol != "(int)" {
		return t
	}
	tok := symAtPos(t.Pos, "index")
	tok.Append(t)
	tok.Append(p.Advance("(int)"))
	return tok
}

func skipNud(p *parser, t *token) *token {
	tok := p.Token
	p.Next()
	return tok
}

var symbols map[string]*symbol

const commaBP = 20

func init() {
	symbols = map[string]*symbol{
		":=":  {Lbp: 10, Led: assignLed},
		"=":   {Lbp: 10, Led: assignLed},
		"+=":  {Lbp: 10, Led: ledInfix},
		"-=":  {Lbp: 10, Led: ledInfix},
		"*=":  {Lbp: 10, Led: ledInfix},
		"/=":  {Lbp: 10, Led: ledInfix},
		"%=":  {Lbp: 10, Led: ledInfix},
		"|=":  {Lbp: 10, Led: ledInfix},
		"^=":  {Lbp: 10, Led: ledInfix},
		"&=":  {Lbp: 10, Led: ledInfix},
		"<<=": {Lbp: 10, Led: ledInfix},
		">>=": {Lbp: 10, Led: ledInfix},

		"||": {Lbp: 30, Led: ledInfix},
		"&&": {Lbp: 40, Led: ledInfix},
		"!":  {Lbp: 50, Nud: notNud},
		"<":  {Lbp: 60, Led: ledInfix},
		">":  {Lbp: 60, Led: ledInfix},
		"<=": {Lbp: 60, Led: ledInfix},
		">=": {Lbp: 60, Led: ledInfix},
		"==": {Lbp: 60, Led: ledInfix},
		"!=": {Lbp: 60, Led: ledInfix},

		"|":  {Lbp: 70, Led: ledInfix},
		"^":  {Lbp: 80, Led: ledInfix, Nud: complementNud},
		"&":  {Lbp: 90, Nud: skipNud, Led: ledInfix},
		"<<": {Lbp: 100, Led: ledInfix},
		">>": {Lbp: 100, Led: ledInfix},

		"+": {Lbp: 110, Led: ledInfix},
		"-": {Lbp: 110, Led: ledInfix, Nud: negateNud},
		"*": {Lbp: 120, Nud: skipNud, Led: ledInfix},
		"/": {Lbp: 120, Led: ledInfix},
		"%": {Lbp: 120, Led: ledInfix},

		"++":  {Lbp: 140, Led: ledPostfix},
		"--":  {Lbp: 140, Led: ledPostfix},
		".":   {Lbp: 150, Led: ledInfix},
		"...": {Lbp: 150, Led: ellipsisLed},
		"(":   {Lbp: 150, Nud: parenNud, Led: callLed},
		"[":   {Lbp: 150, Led: indexLed},
		"{":   {Lbp: 150, Led: newLed, Nud: dataNud},

		"[]":      {Nud: sliceNud},
		"map":     {Nud: mapNud},
		",":       {Lbp: commaBP, Led: commaLed},
		"func":    {Nud: funcNud},
		"return":  {Nud: returnNud},
		"if":      {Nud: ifNud},
		"for":     {Nud: forNud},
		"package": {Nud: packageNud},
		"import":  {Nud: importNud},
		"const":   {Nud: constNud},
		"var":     {Nud: declareNud},
		"type":    {Nud: typeNud},
		"switch":  {Nud: switchNud},
		"$":       {Nud: stackNud},

		"make": {Nud: makeNud},
		// "println": {Nud: nudSelf},
		// "print":   {Nud: nudSelf},
		// "append":  {Nud: nudSelf},
		// "copy":    {Nud: nudSelf},
		// "panic":   {Nud: nudSelf},
		// "len":     {Nud: nudSelf},
		// "delete":  {Nud: nudSelf},

		"(float)":   {Nud: nudSelf},
		"(int)":     {Nud: nudSelf},
		"(name)":    {Nud: nudSelf},
		"(string)":  {Nud: nudSelf},
		"(char)":    {Nud: nudSelf},
		"true":      {Nud: nudSelf},
		"false":     {Nud: nudSelf},
		"nil":       {Nud: nudSelf},
		"error":     {Nud: nudSelf},
		"range":     {Nud: nudSelf},
		"float64":   {Nud: nudSelf},
		"any":       {Nud: nudSelf},
		"int":       {Nud: nudSelf},
		"int32":     {Nud: nudSelf},
		"byte":      {Nud: nudSelf},
		"uint8":     {Nud: nudSelf},
		"rune":      {Nud: nudSelf},
		"uint32":    {Nud: nudSelf},
		"uint":      {Nud: nudSelf},
		"int8":      {Nud: nudSelf},
		"int16":     {Nud: nudSelf},
		"int64":     {Nud: nudSelf},
		"uint16":    {Nud: nudSelf},
		"uint64":    {Nud: nudSelf},
		"bool":      {Nud: nudSelf},
		"string":    {Nud: nudSelf},
		"continue":  {Nud: nudSelf},
		"break":     {Nud: nudSelf},
		"struct":    {Nud: nudSelf},
		"interface": {Nud: nudSelf},
		"case":      {Nud: nudSelf},
		"default":   {Nud: nudSelf},
		"iota":      {Nud: nudSelf},

		";": {Nud: nudNil},

		"(eof)": {},
		":":     {},
		"}":     {},
		")":     {},
		"]":     {},
		"else":  {},

		// unsupported
		"chan": {},
		"go":   {},
		"<-":   {},
		"->":   {},
	}
	for _, s := range symbols {
		if s.Nud == nil {
			s.Nud = nullNud
		}
		if s.Led == nil {
			s.Led = nullLed
		}
	}
}

func getSymbol(t *token) *symbol {
	s, ok := symbols[t.Symbol]
	if !ok {
		panicf("unknown symbol %v", t.Symbol)
	}
	return s
}
