package goatlang

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
)

type code int
type reg int
type pos uint64
type instruction struct {
	Code    code
	A, B, C reg
	Pos     pos
}

func joinParams(a, b reg) reg      { return (((a + 32768) & 0xffff) << 16) | ((b + 32768) & 0xffff) }
func splitParams(v reg) (reg, reg) { return ((v >> 16) & 0xffff) - 32768, (v & 0xffff) - 32768 }

// func joinParams(a, b Reg) Reg      { return ((a & 0xff) << 8) | (b & 0xff) }
// func splitParams(v Reg) (Reg, Reg) { return (v >> 8) & 0xff, v & 0xff }

func (i *instruction) String(g *lookup) string {
	var p []string
	p = append(p, i.Code.String())
	switch i.Code {
	case codePush, codeReturn, codeJumpFalse, codeJumpTrue, codeJump, codeIncDec, codeAnd, codeOr, codeStruct:
		p = append(p, fmt.Sprint(i.A))
	case codeGlobalGet, codeGlobalSet, codeConst, codeGlobalRef, codeGetAttr, codeSetAttr, codeGlobalFunc, codeGlobalStruct:
		p = append(p, g.Key(int(i.A)))
	case codeLocalGet, codeLocalSet:
		p = append(p, "$"+fmt.Sprint(i.A))
	case codeLocalIncDec, codeFastGetInt, codeFastSetInt, codeRange:
		p = append(p, "$"+fmt.Sprint(i.A), fmt.Sprint(i.B))
	case codeFastGet, codeFastSet, codeFastGetAttr, codeFastSetAttr:
		p = append(p, "$"+fmt.Sprint(i.A), g.Key(int(i.B)))
	case codeLocalAdd, codeLocalMul, codeLocalSub, codeLocalDiv:
		p = append(p, "$"+fmt.Sprint(i.A), "$"+fmt.Sprint(i.B))
	case codeFastCallAttr:
		c1, c2 := splitParams(i.C)
		p = append(p, "$"+fmt.Sprint(i.A), g.Key(int(i.B)), fmt.Sprintf("%d:%d", c1, c2))
	case codeNewStruct:
		p = append(p, g.Key(int(i.A)), fmt.Sprint(i.B))
	case codeNewLocalStruct:
		p = append(p, "$"+fmt.Sprint(i.A), fmt.Sprint(i.B))
	case codeSetMethod:
		p = append(p, g.Key(int(i.A)))
	case codeFastCall:
		p = append(p, g.Key(int(i.A)), fmt.Sprint(i.B), fmt.Sprint(i.C))
	case codeNewSlice:
		p = append(p, Type(i.A).String(), fmt.Sprint(i.B))
	case codeNewMap:
		p = append(p, Type(i.A).String(), Type(i.B).String(), fmt.Sprint(i.C))
	case codeZero, codeType, codeMake:
		p = append(p, Type(i.A).String())
	case codeConvert, codeCast:
		p = append(p, Type(i.A).String())
	case codeCall, codeAppend, codeCallVariadic:
		p = append(p, fmt.Sprint(i.A), fmt.Sprint(i.B))
	case codeIter:
		b1, b2 := splitParams(i.B)
		p = append(p, "$"+fmt.Sprint(i.A), fmt.Sprintf("$%d:$%d", b1, b2), fmt.Sprint(i.C))
	case codeTODO:
		p = append(p, fmt.Sprint(i.A), fmt.Sprint(i.B), fmt.Sprint(i.C))
	case codeFunc:
		a1, a2 := splitParams(i.A)
		p = append(p, fmt.Sprintf("%d:%d", a1, a2), fmt.Sprint(i.B), fmt.Sprint(i.C))
	}
	return strings.Join(p, " ")
}

func newPos(l *lookup, fileName, funcName string, line, column int) pos {
	// return struct{}{}
	fileNameIdx := l.Index("#" + fileName)
	funcNameIdx := l.Index("#" + funcName)
	return pos((fileNameIdx << 48) | (funcNameIdx << 32) | (line << 16) | column)
}

func (p pos) IsZero() bool {
	// return false
	return p == 0
}

func (p pos) info(l *lookup) (fileName, funcName string, line, column int) {
	// return "", "", 0, 0
	fileNameIdx := int((p >> 48) & 0xffff)
	funcNameIdx := int((p >> 32) & 0xffff)
	line = int((p >> 16) & 0xffff)
	column = int(p & 0xffff)
	fileName = l.Key(fileNameIdx)[1:]
	funcName = l.Key(funcNameIdx)[1:]
	return fileName, funcName, line, column
}

func (p pos) String(l *lookup) string {
	fileName, funcName, line, column := p.info(l)
	if funcName != "" {
		return fmt.Sprintf("%v(...) %v:%v:%v", funcName, fileName, line, column)
	}
	return fmt.Sprintf("%v:%v:%v", fileName, line, column)
}

func newGlobals() *lookup {
	g := newLookup()
	g.Set("nil", Nil())
	g.Set("true", Bool(true))
	g.Set("false", Bool(false))
	return g
}

func loadBuiltins(g *VM) {
	loadMath(g.globals)
	loadMathRand(g.globals)
	loadFmt(g.globals)
	loadStrings(g.globals)
	loadErrors(g.globals)
	loadBuiltin(g.globals)
	loadTime(g.globals)
	loadMaps(g.globals)
	loadSlices(g)
	loadOs(g)
	loadStrconv(g)
}

type compiler struct {
	Globals     *lookup
	PackageName string
	ExportName  string
	Locals      *lookup
	scope       []int
	cur         *token
	Imports     map[string]string // alias -> package
	Optimize    bool
	Returns     []int
	FuncName    string
}

func compilePkgs(g *lookup, pkgs []*token, optimize bool) (ins []instruction, slots int, err error) {
	locals := newLookup()
	for _, tok := range pkgs {
		cmp := &compiler{
			Globals:  g,
			Locals:   locals,
			Imports:  map[string]string{},
			Optimize: optimize,
		}
		var res []instruction
		res, slots, err = cmp.run(tok)
		if err != nil {
			return nil, 0, err
		}
		ins = append(ins, res...)
	}
	return ins, slots, nil

}

func compile(g *lookup, tok *token, optimize bool) (ins []instruction, slots int, err error) {
	cmp := &compiler{
		Globals:  g,
		Locals:   newLookup(),
		Imports:  map[string]string{},
		Optimize: optimize,
	}
	return cmp.run(tok)
}

func (c *compiler) run(tok *token) (ins []instruction, slots int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v: %v", c.cur.Pos, r)
		}
	}()
	res := c.optimize(c.compileAll(tok.Tokens))
	return res, c.Locals.Cap(), nil
}

func (c *compiler) isLocal() bool {
	return len(c.scope) > 0
}

func (c *compiler) Begin() {
	c.scope = append(c.scope, c.Locals.Len())
}

func (c *compiler) Shadow(key string) int {
	n, ok := c.Locals.keyToIndex[key]
	if ok && n < c.scope[len(c.scope)-1] {
		return c.Locals.Shadow(key)
	}
	return c.Locals.Index(key)
}

func (c *compiler) End() {
	b := c.Locals.Len()
	a := c.scope[len(c.scope)-1]
	c.scope = c.scope[:len(c.scope)-1]
	c.Locals.Drop(b - a)
	// for i := 0; i < b-a; i++ {
	// 	c.Locals.Pop()
	// }
}

func (c *compiler) expPrefix(key string) string {
	if c.ExportName == "" {
		return key
	}
	return c.ExportName + "." + key
}

func (c *compiler) pkgPrefix(key string) string {
	if c.PackageName == "" {
		return key
	}
	return c.PackageName + "." + key
}

var infixMap = map[string]code{
	"||": codeOr,
	"&&": codeAnd,

	"<":  codeLt,
	">":  codeGt,
	"<=": codeLte,
	">=": codeGte,
	"==": codeEq,
	"!=": codeNeq,

	"|":  codeBitOr,
	"^":  codeBitXor,
	"&":  codeBitAnd,
	"<<": codeBitLsh,
	">>": codeBitRsh,

	"+": codeAdd,
	"-": codeSub,
	"*": codeMul,
	"/": codeDiv,
	"%": codeMod,
}

var prefixMap = map[string]code{
	"negate":     codeNegate,
	"!":          codeNot,
	"complement": codeBitComplement,
}

var convMap = map[string]Type{
	"bool":    TypeBool,
	"byte":    TypeUint8,
	"uint8":   TypeUint8,
	"uint32":  TypeUint32,
	"uint":    TypeUint32,
	"uint64":  TypeUint32,
	"uint16":  TypeUint32,
	"rune":    TypeInt32,
	"int":     TypeInt32,
	"int8":    TypeInt8,
	"int32":   TypeInt32,
	"int64":   TypeInt32,
	"int16":   TypeInt32,
	"float64": TypeFloat64,
	"string":  TypeString,
	"[]":      TypeSlice,
	"map":     TypeMap,
	"func":    TypeFunc,
	"struct":  TypeStruct,
	"error":   TypeStruct,
}

var builtinMap = map[string]code{
	"len":    codeLen,
	"delete": codeDelete,
	"append": codeAppend,
	"panic":  codePanic,
	"copy":   codeCopy,
}

func (c *compiler) compile(tok *token) []instruction {
	c.cur = tok
	var res []instruction
	switch tok.Symbol {
	case "(int)":
		res = append(res, instruction{Code: codePush, A: reg(tok.Int())})
	case "(char)":
		res = append(res, instruction{Code: codePush, A: reg(tok.Char())})
	case "(float)":
		c.Globals.Set(tok.Text, Float64(tok.Float64()))
		res = append(res, instruction{Code: codeConst, A: reg(c.Globals.Index(tok.Text))})
	case "(string)":
		c.Globals.Set(tok.Text, String(tok.Unquote()))
		res = append(res, instruction{Code: codeConst, A: reg(c.Globals.Index(tok.Text))})
	case "<", ">", "<=", ">=", "==", "!=", "|", "^", "&", "<<", ">>", "+", "-", "*", "/", "%":
		res = append(res, c.compileAll(tok.Tokens)...)
		res = append(res, instruction{Code: infixMap[tok.Symbol]})
	case "&&", "||":
		res = append(res, c.compile(tok.Tokens[0])...)
		right := c.optimize(c.compile(tok.Tokens[1]))
		res = append(res, instruction{Code: infixMap[tok.Symbol], A: reg(len(right))})
		res = append(res, right...)
	case "|=", "^=", "&=", "<<=", ">>=", "+=", "-=", "*=", "/=", "%=", "++", "--":
		var todo []instruction
		switch tok.Symbol {
		case "++":
			todo = append(todo, instruction{Code: codeIncDec, A: 1})
		case "--":
			todo = append(todo, instruction{Code: codeIncDec, A: -1})
		default:
			todo = append(todo, c.compile(tok.Tokens[1])...)
			todo = append(todo, instruction{Code: infixMap[tok.Symbol[:len(tok.Symbol)-1]]})
		}
		arg := tok.Tokens[0]
		if arg.Symbol == "index" {
			const indexItem, indexKey = 0, 1
			res = append(res, c.compile(arg.Tokens[indexItem])...)
			res = append(res, c.compile(arg.Tokens[indexKey])...)
			res = append(res, instruction{Code: codeGet})
			res = append(res, todo...)
			res = append(res, c.compile(arg.Tokens[indexItem])...)
			res = append(res, c.compile(arg.Tokens[indexKey])...)
			res = append(res, instruction{Code: codeSet})
		} else if arg.Symbol == "." {
			const indexItem, indexKey = 0, 1
			res = append(res, c.compile(arg.Tokens[indexItem])...)
			res = append(res, instruction{Code: codeGetAttr, A: reg(c.Globals.Index(arg.Tokens[indexKey].Text))})
			res = append(res, todo...)
			res = append(res, c.compile(arg.Tokens[indexItem])...)
			res = append(res, instruction{Code: codeSetAttr, A: reg(c.Globals.Index(arg.Tokens[indexKey].Text))})
		} else {
			getter := codeGlobalGet
			setter := codeGlobalSet
			lookup := c.Globals
			key := arg.Text
			if c.Locals.Exists(key) {
				getter = codeLocalGet
				setter = codeLocalSet
				lookup = c.Locals
			} else {
				key = c.expPrefix(key)
			}
			res = append(res, instruction{Code: getter, A: reg(lookup.Index(key))})
			res = append(res, todo...)
			res = append(res, instruction{Code: setter, A: reg(lookup.Index(key))})
		}

	case "const":
		var values []*token
		if tok.Tokens[1].Symbol != "," {
			values = []*token{tok.Tokens[1]}
		} else {
			values = tok.Tokens[1].Tokens
		}
		for i := 0; i < len(tok.Tokens[0].Tokens); i++ {
			target := tok.Tokens[0].Tokens[i]
			key := target.Text
			value := values[i]
			res = append(res, c.compile(value)...)

			code := codeGlobalSet
			var idx int
			if c.isLocal() {
				code = codeLocalSet
				idx = c.Shadow(key)
			} else {
				lookup := c.Globals
				key = c.expPrefix(key)
				idx = lookup.Index(key)
			}
			res = append(res, instruction{Code: code, A: reg(idx)})
		}
	case ":=", "var":
		values := c.compile(tok.Tokens[1])
		res = append(res, values...)
		if len(values) == 0 {
			for _, arg := range tok.Tokens[0].Tokens {
				typ := arg.Tokens[0]
				ct := typeFromToken(c, typ)
				res = append(res, instruction{Code: codeZero, A: reg(ct)})
			}
		}
		for i := 1; i <= len(tok.Tokens[0].Tokens); i++ {
			target := tok.Tokens[0].Tokens[len(tok.Tokens[0].Tokens)-i]
			key := target.Text
			code := codeGlobalSet
			lookup := c.Globals
			var idx int
			if key == "_" {
				res = append(res, instruction{Code: codePop})
				continue
			} else if c.isLocal() {
				code = codeLocalSet
				idx = c.Shadow(key)
			} else {
				key = c.expPrefix(key)
				idx = lookup.Index(key)
			}
			if len(values) > 0 && len(target.Tokens) > 0 {
				typ := typeFromToken(c, target.Tokens[0])
				if slices.Contains([]Type{TypeUint8, TypeInt32, TypeFloat64}, typ) {
					res = append(res, instruction{Code: codeCast, A: reg(typ)})
				}
			}
			res = append(res, instruction{Code: code, A: reg(idx)})
		}

	case "function":
		target := tok.Tokens[0]
		c.FuncName = c.pkgPrefix(target.Text)
		res = append(res, c.compile(tok.Tokens[1])...)
		idx := c.Globals.Index(c.expPrefix(target.Text))
		res = append(res, instruction{Code: codeGlobalFunc, A: reg(idx)})
		c.FuncName = ""

	case "=":
		res = append(res, c.compile(tok.Tokens[1])...)
		for i := 1; i <= len(tok.Tokens[0].Tokens); i++ {
			arg := tok.Tokens[0].Tokens[len(tok.Tokens[0].Tokens)-i]
			if arg.Text == "_" {
				res = append(res, instruction{Code: codePop})
				continue
			} else if arg.Symbol == "index" {
				const indexItem, indexKey = 0, 1
				res = append(res, c.compile(arg.Tokens[indexItem])...)
				res = append(res, c.compile(arg.Tokens[indexKey])...)
				res = append(res, instruction{Code: codeSet})
			} else if arg.Symbol == "." {
				const indexItem, indexKey = 0, 1
				res = append(res, c.compile(arg.Tokens[indexItem])...)
				res = append(res, instruction{Code: codeSetAttr, A: reg(c.Globals.Index(arg.Tokens[indexKey].Text))})
			} else {
				code := codeGlobalSet
				lookup := c.Globals
				key := arg.Text
				if c.Locals.Exists(key) {
					code = codeLocalSet
					lookup = c.Locals
				} else {
					key = c.expPrefix(key)
				}
				res = append(res, instruction{Code: code, A: reg(lookup.Index(key))})
			}
		}
	case "true", "false", "nil":
		res = append(res, instruction{Code: codeConst, A: reg(c.Globals.Index(tok.Text))})
	case "(name)":
		key := c.expPrefix(tok.Text)
		if tok.Text == "$" {
			res = append(res, instruction{Code: codeGlobalGet, A: reg(c.Globals.Index("$"))})
		} else if c.Locals.Exists(tok.Text) {
			res = append(res, instruction{Code: codeLocalGet, A: reg(c.Locals.Index(tok.Text))})
		} else if c.Globals.Exists(key) {
			res = append(res, instruction{Code: codeGlobalGet, A: reg(c.Globals.Index(key))})
		} else if c.Globals.Exists("builtin." + tok.Text) {
			res = append(res, instruction{Code: codeGlobalGet, A: reg(c.Globals.Index("builtin." + tok.Text))})
		} else {
			res = append(res, instruction{Code: codeGlobalGet, A: reg(c.Globals.Index(key))})
		}
	case ".":
		const dotLeft, dotRight = 0, 1
		left, right := tok.Tokens[dotLeft], tok.Tokens[dotRight]
		if left.Symbol == "(name)" && !c.Locals.Exists(left.Text) {
			if pkg, ok := c.Imports[left.Text]; ok {
				key := pkg + "." + right.Text
				if !c.Globals.Exists(key) {
					panicf("undefined: %v", key)
				}
				res = append(res, instruction{Code: codeGlobalGet, A: reg(c.Globals.Index(key))})
				break
			}
		}
		res = append(res, c.compile(left)...)
		res = append(res, instruction{Code: codeGetAttr, A: reg(c.Globals.Index(right.Text))})
	case "slice":
		const sliceObj, sliceBegin, sliceEnd = 0, 1, 2
		res = append(res, c.compile(tok.Tokens[sliceObj])...)
		res = append(res, c.compile(tok.Tokens[sliceBegin])...)
		res = append(res, c.compile(tok.Tokens[sliceEnd])...)
		res = append(res, instruction{Code: codeSlice})
	case "func":
		const funcArguments, funcReturns, funcBlock = 0, 1, 2
		tmp := c.Locals
		c.Locals = newLookup()
		c.Begin()
		arguments := len(tok.Tokens[funcArguments].Tokens)
		for _, arg := range tok.Tokens[funcArguments].Tokens {
			c.Locals.Index(arg.Text)
		}
		if arguments > 0 && tok.Tokens[funcArguments].Tokens[arguments-1].Tokens[0].Text == "..." {
			arguments = -arguments
		}
		returns := len(tok.Tokens[funcReturns].Tokens)
		c.Returns = append(c.Returns, returns)
		block := c.optimize(c.compile(tok.Tokens[funcBlock]))
		res = append(res, instruction{Code: codeFunc,
			A: joinParams(reg(arguments), reg(returns)),
			B: reg(c.Locals.Cap()),
			C: reg(len(block)),
		})
		for _, arg := range tok.Tokens[funcArguments].Tokens {
			t := c.toType(arg.Tokens[0])
			res = append(res, t)
		}
		for _, ret := range tok.Tokens[funcReturns].Tokens {
			t := c.toType(ret)
			res = append(res, t)
		}
		res = append(res, block...)
		c.Returns = c.Returns[:len(c.Returns)-1]
		c.End()
		c.Locals = tmp
	case "block", ",":
		res = append(res, c.compileAll(tok.Tokens)...)
	case "return":
		if len(tok.Tokens) == 1 && tok.Tokens[0].Symbol == "call" {
			returns := c.compileAll(tok.Tokens)
			returns[len(returns)-1].B = reg(c.Returns[len(c.Returns)-1])
			res = append(res, returns...)
			res = append(res, instruction{Code: codeReturn, A: reg(c.Returns[len(c.Returns)-1])})
			break
		}
		returns := c.compileAll(tok.Tokens)
		res = append(res, returns...)
		res = append(res, instruction{Code: codeReturn, A: reg(len(tok.Tokens))})
	case "call":
		const callName, callArguments, callReturns = 0, 1, 2
		res = append(res, c.compileAll(tok.Tokens[callArguments].Tokens)...)
		if slices.Contains([]string{"byte", "uint8", "int8", "int", "int32", "rune", "uint32", "uint", "int64", "uint64", "int16", "uint16", "float64", "string", "[]"}, tok.Tokens[callName].Symbol) {
			res = append(res, instruction{Code: codeConvert, A: reg(convMap[tok.Tokens[callName].Symbol])})
		} else if code := builtinMap[tok.Tokens[callName].Text]; code != 0 {
			ellipsis := 0
			args := tok.Tokens[callArguments].Tokens
			if len(args) > 0 && args[len(args)-1].Symbol == "..." {
				ellipsis = 1
			}
			res = append(res, instruction{Code: code, A: reg(len(args)), B: reg(ellipsis)})
		} else {
			fnc := c.compile(tok.Tokens[callName])
			if tok.Tokens[callName].Symbol == "(name)" {
				typ := c.Globals.Read(int(fnc[0].A))
				if typ.t == typeType {
					res = append(res, instruction{Code: codeConvert, A: reg(typ.Int())})
					break
				}
			}
			res = append(res, fnc...)

			args := tok.Tokens[callArguments].Tokens
			if len(args) > 0 && args[len(args)-1].Symbol == "..." {
				res = append(res, instruction{Code: codeCallVariadic,
					A: reg(len(tok.Tokens[callArguments].Tokens)),
					B: reg(tok.Tokens[callReturns].Int()),
				})
				break
			}
			res = append(res, instruction{Code: codeCall,
				A: reg(len(tok.Tokens[callArguments].Tokens)),
				B: reg(tok.Tokens[callReturns].Int()),
			})
		}
	case "init":
		const initFunc = 0
		c.FuncName = c.pkgPrefix("init")
		res = append(res, c.compile(tok.Tokens[initFunc])...)
		res = append(res, instruction{Code: codeCall})
		c.FuncName = ""
	case "if":
		const ifStmt, ifCond, ifThen, ifElse = 0, 1, 2, 3
		c.Begin()
		res = append(res, c.compile(tok.Tokens[ifStmt])...)
		res = append(res, c.compile(tok.Tokens[ifCond])...)
		c.Begin()
		thenI := c.optimize(c.compile(tok.Tokens[ifThen]))
		c.End()
		var elseI []instruction
		if len(tok.Tokens) > ifElse {
			c.Begin()
			elseI = c.optimize(c.compile(tok.Tokens[ifElse]))
			c.End()
		}
		if len(elseI) == 0 {
			res = append(res, instruction{Code: codeJumpFalse, A: reg(len(thenI))})
			res = append(res, thenI...)
		} else {
			res = append(res, instruction{Code: codeJumpFalse, A: reg(len(thenI) + 1)})
			res = append(res, thenI...)
			res = append(res, instruction{Code: codeJump, A: reg(len(elseI))})
			res = append(res, elseI...)
		}
		c.End()
	case "switch":
		const switchStmt, switchCases, switchDefault = 0, 1, 2
		c.Begin()
		stmt := c.compile(tok.Tokens[switchStmt])
		res = append(res, stmt...)
		isValue := len(stmt) > 0
		var v int
		if isValue {
			v = c.Locals.Index(tok.Pos.String())
			res = append(res, instruction{Code: codeLocalSet, A: reg(v)})
		}

		var out []instruction
		c.Begin()
		defBlock := c.optimize(c.compileAll(tok.Tokens[switchDefault].Tokens))
		c.End()
		for i := len(tok.Tokens[switchCases].Tokens) - 1; i >= 0; i-- {
			cs := tok.Tokens[switchCases].Tokens[i]
			const caseStmt, caseBlock = 0, 1
			csStmt := c.optimize(c.compile(cs.Tokens[caseStmt]))
			c.Begin()
			csBlock := c.optimize(c.compileAll(cs.Tokens[caseBlock].Tokens))
			for n, ins := range csBlock {
				switch ins.Code {
				case codeBreak:
					csBlock[n].Code, csBlock[n].A = codeJump, reg((len(csBlock)-n)+len(out)+len(defBlock))
				}
			}
			c.End()
			var chunk []instruction
			chunk = append(chunk, csStmt...)
			if isValue {
				chunk = append(chunk, instruction{Code: codeLocalGet, A: reg(v)})
				chunk = append(chunk, instruction{Code: codeEq})
			}
			chunk = append(chunk, instruction{Code: codeJumpFalse, A: reg(len(csBlock) + 1)})
			chunk = append(chunk, csBlock...)
			chunk = append(chunk, instruction{Code: codeJump, A: reg(len(out) + len(defBlock))})
			out = append(chunk, out...)
		}
		c.End()
		res = append(res, out...)
		res = append(res, defBlock...)

	case "for":
		const forInit, forCond, forPost, forBlock = 0, 1, 2, 3
		c.Begin()
		res = append(res, c.compile(tok.Tokens[forInit])...)
		cond := c.optimize(c.compile(tok.Tokens[forCond]))
		block := c.optimize(c.compile(tok.Tokens[forBlock]))
		post := c.optimize(c.compile(tok.Tokens[forPost]))
		if len(cond) > 0 {
			res = append(res, instruction{Code: codeJump, A: reg((len(block) + len(post)))})
		}
		for n, ins := range block {
			switch ins.Code {
			case codeBreak:
				block[n].Code, block[n].A = codeJump, reg((len(block)-n)+len(post)+len(cond))
			case codeContinue:
				block[n].Code, block[n].A = codeJump, reg(len(block)-n-1)
			}
		}
		res = append(res, block...)
		res = append(res, post...)
		if len(cond) > 0 {
			res = append(res, cond...)
			res = append(res, instruction{Code: codeJumpTrue, A: reg(-(len(block) + len(post) + len(cond) + 1))})
		} else {
			res = append(res, instruction{Code: codeJump, A: reg(-(len(block) + len(post) + 1))})
		}

		c.End()
	case "...":
		res = append(res, c.compile(tok.Tokens[0])...)
	case "make":
		const makeType, makeLen = 0, 1
		var typ Type
		if tok.Tokens[makeType].Symbol == "(name)" {
			ref := c.compile(tok.Tokens[makeType])
			val := c.Globals.Read(int(ref[0].A))
			typ = Type(val.Int())
		} else {
			typ = typeFromToken(c, tok.Tokens[makeType])

		}
		switch typ.base() {
		case TypeSlice:
			res = append(res, c.compile(tok.Tokens[makeLen])...)
			res = append(res, instruction{Code: codeMake, A: reg(typ.value())})
		case TypeMap:
			kt, vt := typ.pair()
			res = append(res, instruction{Code: codeNewMap, A: reg(kt), B: reg(vt), C: 0})
		}
	case "package":
		c.PackageName = tok.Tokens[0].Text
		c.ExportName = tok.Tokens[0].Text
		if len(tok.Tokens) > 1 {
			c.ExportName = tok.Tokens[1].Text
		}
	case "import":
		for i := 0; i < len(tok.Tokens); i += 2 {
			key := tok.Tokens[i].Text
			t := tok.Tokens[i+1]
			c.Imports[key] = t.Unquote()
		}
	case "[]":
		const newType, newData = 0, 1
		res = append(res, c.compile(tok.Tokens[newData])...)
		res = append(res, instruction{Code: codeNewSlice, A: reg(convMap[tok.Tokens[newType].Symbol]), B: reg(len(tok.Tokens[newData].Tokens))})

	case "map":
		const newKeyType, newValueType, newData = 0, 1, 2
		res = append(res, c.compile(tok.Tokens[newData])...)
		res = append(res, instruction{Code: codeNewMap, A: reg(typeFromToken(c, tok.Tokens[newKeyType])), B: reg(typeFromToken(c, tok.Tokens[newValueType])), C: reg(len(tok.Tokens[newData].Tokens))})
	case "range":
		c.Begin()
		const rangeKey, rangeValue, rangeItem, rangeBlock = 0, 1, 2, 3
		res = append(res, c.compile(tok.Tokens[rangeItem])...)
		r := c.Locals.Index(tok.Pos.String())
		k := c.Locals.Index(tok.Tokens[rangeKey].Text)
		v := c.Locals.Index(tok.Tokens[rangeValue].Text)
		block := c.optimize(c.compile(tok.Tokens[rangeBlock]))
		for n, ins := range block {
			switch ins.Code {
			case codeBreak:
				block[n].Code, block[n].A = codeJump, reg(len(block)-n)
			case codeContinue:
				block[n].Code, block[n].A = codeJump, reg(len(block)-n-1)
			}
		}
		res = append(res, instruction{Code: codeRange, A: reg(r), B: reg(len(block))})
		res = append(res, block...)
		res = append(res, instruction{Code: codeIter, A: reg(r), B: joinParams(reg(k), reg(v)), C: reg(-(len(block) + 1))})
		c.End()
	case "break":
		res = append(res, instruction{Code: codeBreak})
	case "continue":
		res = append(res, instruction{Code: codeContinue})

	case "index", "indexOk":
		const indexItem, indexKey = 0, 1
		res = append(res, c.compile(tok.Tokens[indexItem])...)
		res = append(res, c.compile(tok.Tokens[indexKey])...)
		code := codeGet
		if tok.Symbol == "indexOk" {
			code = codeGetOk
		}
		res = append(res, instruction{Code: code})
	case "negate", "!":
		res = append(res, c.compile(tok.Tokens[0])...)
		res = append(res, instruction{Code: prefixMap[tok.Symbol]})
	case "complement":
		res = append(res, c.compile(tok.Tokens[0])...)
		res = append(res, instruction{Code: codeBitComplement})
	case "type":
		const typeName, typeStruct = 0, 1
		ts := tok.Tokens[typeStruct].Symbol

		key := tok.Tokens[typeName].Text
		var idx int
		setStruct := codeGlobalStruct
		getStruct := codeGlobalGet

		if c.isLocal() {
			setStruct = codeLocalSet
			getStruct = codeLocalGet
			idx = c.Shadow(key)
		} else {
			key = c.expPrefix(key)
			idx = c.Globals.Index(key)
		}

		if getStruct == codeLocalGet {
			setStruct = codeLocalSet
		}
		if ts == "interface" {
			res = append(res, instruction{Code: codeStruct, A: 0})
			res = append(res, instruction{Code: setStruct, A: reg(idx)})
			for i := 0; i < len(tok.Tokens[typeStruct].Tokens); i += 3 {
				res = append(res, instruction{Code: codeZero, A: reg(TypeFunc)})
				res = append(res, instruction{Code: getStruct, A: reg(idx)})
				res = append(res, instruction{Code: codeSetMethod, A: reg(c.Globals.Index(tok.Tokens[typeStruct].Tokens[i].Text))})
			}
			break
		}
		if ts == "bool" || ts == "byte" || ts == "uint8" || ts == "int8" || ts == "int" || ts == "int32" || ts == "rune" || ts == "uint32" || ts == "uint" || ts == "float64" || ts == "string" || ts == "int16" || ts == "int64" || ts == "uint16" || ts == "uint64" {
			c.Globals.Write(int(reg(idx)), newType(convMap[ts]))
			break
		}
		if ts == "struct" {
			for i := 0; i < len(tok.Tokens[typeStruct].Tokens); i += 2 {
				t := tok.Tokens[typeStruct].Tokens[i]
				res = append(res, instruction{Code: codeGlobalRef, A: reg(c.Globals.Index(t.Text))})
				t = tok.Tokens[typeStruct].Tokens[i+1]
				res = append(res, instruction{Code: codeZero, A: reg(typeFromToken(c, t))})
			}
			res = append(res, instruction{Code: codeStruct, A: reg(len(tok.Tokens[typeStruct].Tokens))})
			res = append(res, instruction{Code: setStruct, A: reg(idx)})
			break
		}
		typ := typeFromToken(c, tok.Tokens[typeStruct])
		if typ != TypeStruct {
			c.Globals.Write(int(reg(idx)), newType(typ))
		} else {
			ref := c.compile(tok.Tokens[typeStruct])
			c.Globals.Write(int(reg(idx)), c.Globals.Read(int(ref[0].A)))
		}

	case "new":
		const newType, newData = 0, 1
		ref := c.compile(tok.Tokens[newType])
		if ref[0].Code == codeGlobalGet {
			val := c.Globals.Read(int(ref[0].A))
			if val.Type() == typeType {
				typ := Type(val.Int())
				if typ.base() == TypeSlice {
					res = append(res, c.compile(tok.Tokens[newData])...)
					res = append(res, instruction{Code: codeNewSlice, A: reg(typ.value()), B: reg(len(tok.Tokens[newData].Tokens))})
					break

				}
				if typ.base() == TypeMap {
					res = append(res, c.compile(tok.Tokens[newData])...)
					kt, vt := typ.pair()
					res = append(res, instruction{Code: codeNewMap, A: reg(kt), B: reg(vt), C: reg(len(tok.Tokens[newData].Tokens))})
					break
				}
			}
		}

		for i := 0; i < len(tok.Tokens[newData].Tokens); i += 2 {
			t := tok.Tokens[newData].Tokens[i]
			res = append(res, instruction{Code: codeGlobalRef, A: reg(c.Globals.Index(t.Text))})
			res = append(res, c.compile(tok.Tokens[newData].Tokens[i+1])...)
		}
		newStruct := codeNewStruct
		if ref[0].Code == codeLocalGet {
			newStruct = codeNewLocalStruct
		}
		res = append(res, instruction{Code: newStruct, A: ref[0].A, B: reg(len(tok.Tokens[newData].Tokens))})
	case "method":
		const methodType, methodName, methodFunc = 0, 1, 2
		c.FuncName = c.pkgPrefix(tok.Tokens[methodType].Text) + "." + tok.Tokens[methodName].Text
		res = append(res, c.compile(tok.Tokens[methodFunc])...)
		res = append(res, instruction{Code: codeGlobalGet, A: reg(c.Globals.Index(c.expPrefix(tok.Tokens[methodType].Text)))})
		res = append(res, instruction{Code: codeSetMethod,
			A: reg(c.Globals.Index(tok.Tokens[methodName].Text)),
		})
		c.FuncName = ""
	case "~":
	default:
		panicf("unknown symbol: %v", tok.Symbol)
	}
	for n, i := range res {
		if !i.Pos.IsZero() {
			continue
		}
		res[n].Pos = newPos(c.Globals, tok.Pos.Filename, c.FuncName, tok.Pos.Line, tok.Pos.Column)
	}
	return res
}

func (c *compiler) toType(tok *token) instruction {
	return instruction{Code: codeType, A: reg(typeFromToken(c, tok))}
}

func (c *compiler) compileAll(tokens []*token) []instruction {
	var res []instruction
	for _, t := range tokens {
		res = append(res, c.compile(t)...)
	}
	return res
}

func (c *compiler) optimize(in []instruction) []instruction {
	if !c.Optimize {
		return in
	}
	return c.doOptimize(c.doOptimize(in))
}
func (c *compiler) doOptimize(in []instruction) []instruction {
	var out []instruction
	for n := 0; n < len(in); n++ {
		switch {
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeIncDec && in[n+2].Code == codeLocalSet && in[n].A == in[n+2].A:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeLocalIncDec, A: in[n].A, B: in[n+1].A})
			n += 2

		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeLocalGet && in[n+2].Code == codeAdd:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeLocalAdd, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeLocalGet && in[n+2].Code == codeMul:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeLocalMul, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeLocalGet && in[n+2].Code == codeDiv:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeLocalDiv, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeLocalGet && in[n+2].Code == codeSub:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeLocalSub, A: in[n].A, B: in[n+1].A})
			n += 2

		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeConst && in[n+2].Code == codeGet:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastGet, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeConst && in[n+2].Code == codeSet:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastSet, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codePush && in[n+2].Code == codeGet:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastGetInt, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codePush && in[n+2].Code == codeSet:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastSetInt, A: in[n].A, B: in[n+1].A})
			n += 2
		case n < len(in)-2 && in[n].Code == codeLocalGet && in[n+1].Code == codeGetAttr && in[n+2].Code == codeCall:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastCallAttr, A: in[n].A, B: in[n+1].A, C: joinParams(in[n+2].A, in[n+2].B)})
			n += 2
		case n < len(in)-1 && in[n].Code == codeGlobalGet && in[n+1].Code == codeCall:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastCall, A: in[n].A, B: in[n+1].A, C: in[n+1].B})
			n += 1

		case n < len(in)-1 && in[n].Code == codeLocalGet && in[n+1].Code == codeGetAttr:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastGetAttr, A: in[n].A, B: in[n+1].A})
			n += 1
		case n < len(in)-1 && in[n].Code == codeLocalGet && in[n+1].Code == codeSetAttr:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeFastSetAttr, A: in[n].A, B: in[n+1].A})
			n += 1

		case n < len(in)-1 && in[n].Code == codePush && in[n+1].Code == codeAdd:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeIncDec, A: in[n].A})
			n += 1
		case n < len(in)-1 && in[n].Code == codePush && in[n+1].Code == codeSub:
			out = append(out, instruction{Pos: in[n].Pos, Code: codeIncDec, A: -in[n].A})
			n += 1

		case n < len(in) && in[n].Code == codeJump && in[n].A == 0:
			out = append(out, instruction{Pos: in[n].Pos, Code: codePass})

		default:
			out = append(out, in[n])
		}
	}
	return out
}

func typeFromToken(c *compiler, tok *token) Type {
	switch tok.Symbol {
	case "[]":
		return sliceType(typeFromToken(c, tok.Tokens[0]))
	case "map":
		return mapType(typeFromToken(c, tok.Tokens[0]), typeFromToken(c, tok.Tokens[1]))
	case "(name)", ".":
		ref := c.compile(tok)
		typ := c.Globals.Read(int(ref[0].A))
		if typ.t == typeType {
			return Type(typ.Int())
		}
		return TypeStruct
	case "...":
		return sliceType(typeFromToken(c, tok.Tokens[0]))
	default:
		return convMap[tok.Symbol]
	}
}
