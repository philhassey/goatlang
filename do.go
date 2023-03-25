package goatlang

import (
	"fmt"
)

func (v *VM) exec() {
	codes := v.frame.Codes
	baseN := v.frame.BaseN
	l := len(codes)
	for v.frame.N = 0; v.frame.N < l; v.frame.N++ {
		switch codes[v.frame.N].Code {
		case codePush, codeGlobalRef:
			v.stack = append(v.stack, newUntypedInt(int(codes[v.frame.N].A)))

		case codePop:
			v.stack = v.stack[:len(v.stack)-1]

		case codeAdd:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opAdd(b)
		case codeSub:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opSub(b)
		case codeMul:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opMul(b)
		case codeDiv:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opDiv(b)

		case codeMod:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opMod(b)

		case codeLte:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opLte(b)

		case codeGte:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = b.opLte(a)

		case codeNeq:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opNeq(b)

		case codeBitAnd:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opBitAnd(b)

		case codeBitOr:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opBitOr(b)

		case codeBitLsh:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opBitLsh(b)

		case codeBitRsh:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opBitRsh(b)

		case codeBitXor:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opBitXor(b)

		case codeIncDec:
			i := &codes[v.frame.N]
			a := v.stack[len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opAdd(Int(int(i.A)))

		case codeLocalIncDec:
			i := &codes[v.frame.N]
			v.stack[baseN+int(i.A)] = v.stack[baseN+int(i.A)].opAdd(Int(int(i.B)))

		case codeConvert:
			i := &codes[v.frame.N]
			a := v.stack[len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.convert(Type(i.A))
		case codeCast:
			i := &codes[v.frame.N]
			a := v.stack[len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.assign(Type(i.A))

		case codeNegate:
			v.stack[len(v.stack)-1] = v.stack[len(v.stack)-1].opMul(newUntypedInt(-1))
		case codeBitComplement:
			a := v.stack[len(v.stack)-1].assign(TypeNil)
			b := Uint32(0xffffffff).convert(a.t)
			v.stack[len(v.stack)-1] = a.opBitXor(b)

		case codeNot:
			v.stack[len(v.stack)-1] = Bool(!v.stack[len(v.stack)-1].Bool())

		case codeZero:
			i := &codes[v.frame.N]
			v.stack = append(v.stack, newZero(Type(i.A)))

		case codeAnd:
			i := &codes[v.frame.N]
			a := v.stack[len(v.stack)-1]
			if !a.Bool() {
				v.frame.N += int(i.A)
			} else {
				v.stack = v.stack[:len(v.stack)-1]
			}

		case codeOr:
			i := &codes[v.frame.N]
			a := v.stack[len(v.stack)-1]
			if a.Bool() {
				v.frame.N += int(i.A)
			} else {
				v.stack = v.stack[:len(v.stack)-1]
			}

		case codeEq:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opEq(b)

		case codeGlobalSet:
			a := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.globals.Assign(int(codes[v.frame.N].A), a)

		case codeGlobalZero:
			i := &codes[v.frame.N]
			a := v.globals.Read(int(i.A))
			if a.IsNil() {
				v.globals.Assign(int(i.A), newZero(Type(i.B)))
			}

		case codeGlobalFunc:
			val := v.stack[len(v.stack)-1]
			idx := int(codes[v.frame.N].A)
			v.stack = v.stack[:len(v.stack)-1]
			if fnc := v.globals.Read(idx); !fnc.IsNil() {
				*fnc.value.(*funcT) = *val.value.(*funcT)
				break
			}
			v.globals.Write(idx, val)

		case codeGlobalGet, codeConst:
			v.stack = append(v.stack, v.globals.Read(int(codes[v.frame.N].A)))

		case codeFunc:
			i := &codes[v.frame.N]
			args, rets := splitParams(i.A)
			nargs := args
			if nargs < 0 {
				nargs = -nargs
			}
			slots, jump := i.B, i.C

			tokens := codes[v.frame.N+1 : v.frame.N+1+int(nargs+rets+jump)]
			v.frame.N += int(nargs + rets + jump)
			f := newFunc(int(args), int(rets), mkFunc(int(nargs), int(rets), int(slots), tokens))
			if args < 0 {
				f.getFunc().VariadicType = Type(tokens[nargs-1].A)
			}
			v.stack = append(v.stack, f)

		case codeCall:
			i := &codes[v.frame.N]
			f := v.stack[len(v.stack)-1].getFunc()
			v.stack = v.stack[:len(v.stack)-1]
			call(v, f, int(i.A), int(i.B))

		case codeCallVariadic:
			i := &codes[v.frame.N]
			f := v.stack[len(v.stack)-1].getFunc()
			v.stack = v.stack[:len(v.stack)-1]
			callReady(v, f, int(i.A), int(i.B))

		case codeLocalGet:
			i := &codes[v.frame.N]
			v.stack = append(v.stack, v.stack[baseN+int(i.A)])

		case codeLocalSet:
			i := &codes[v.frame.N]
			v.stack, v.stack[baseN+int(i.A)] = v.stack[:len(v.stack)-1], v.stack[len(v.stack)-1].assign(v.stack[baseN+int(i.A)].t)

		case codeLocalZero:
			i := &codes[v.frame.N]
			v.stack[baseN+int(i.A)] = newZero(Type(i.B))

		case codeReturn:
			return

		case codeJumpFalse:
			a := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			if a.Bool() {
				break
			}
			i := &codes[v.frame.N]
			v.frame.N += int(i.A)

		case codeJumpTrue:
			a := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			if !a.Bool() {
				break
			}
			i := &codes[v.frame.N]
			v.frame.N += int(i.A)
		case codeJump:
			i := &codes[v.frame.N]
			v.frame.N += int(i.A)

		case codeLt:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = a.opLt(b)

		case codeGt:
			a, b := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1] = b.opLt(a)

		case codeGet:
			r, k := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			v.stack[len(v.stack)-1], _ = r.Get(k)

		case codeLen:
			r := v.stack[len(v.stack)-1]
			if r.value != nil {
				v.stack[len(v.stack)-1] = Int(r.Len())
			} else {
				v.stack[len(v.stack)-1] = Int(0)
			}
		case codeDelete:
			r, k := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-2]
			r.Delete(k)

		case codeSlice:
			r, a, b := v.stack[len(v.stack)-3], v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			i, j := a.Int(), b.Int()
			if j < 0 {
				j += 1 + r.Len()
			}
			v.stack = v.stack[:len(v.stack)-2]
			v.stack[len(v.stack)-1] = r.Slice(i, j)

		case codeGetOk:
			r, k := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			var ok bool
			v.stack[len(v.stack)-2], ok = r.Get(k)
			v.stack[len(v.stack)-1] = Bool(ok)

		case codeSet:
			value, obj, key := v.stack[len(v.stack)-3], v.stack[len(v.stack)-2], v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-3]
			obj.Set(key, value)

		case codeFastGet:
			i := &codes[v.frame.N]
			r, k := v.stack[baseN+int(i.A)], v.globals.Read(int(i.B))
			val, _ := r.Get(k)
			v.stack = append(v.stack, val)

		case codeFastSet:
			i := &codes[v.frame.N]
			val := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			r, k := v.stack[baseN+int(i.A)], v.globals.Read(int(i.B))
			r.Set(k, val)

		case codeFastGetInt:
			i := &codes[v.frame.N]
			r := v.stack[baseN+int(i.A)]
			val, _ := r.Get(Int(int(i.B)))
			v.stack = append(v.stack, val)

		case codeFastSetInt:
			i := &codes[v.frame.N]
			val := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			r := v.stack[baseN+int(i.A)]
			r.Set(Int(int(i.B)), val)

		case codeFastCall:
			i := &codes[v.frame.N]
			f := v.globals.Read(int(i.A)).getFunc()
			call(v, f, int(i.B), int(i.C))

		case codeAppend:
			i := &codes[v.frame.N]
			s := v.stack[len(v.stack)-int(i.A)]
			vs := v.stack[len(v.stack)-int(i.A)+1:]
			v.stack = v.stack[:len(v.stack)-int(i.A)+1]
			if i.B == 1 {
				tmp := vs[len(vs)-1]
				vs = append(vs[:len(vs)-1], tmp.data()...)
			}
			if s.value != nil {
				v.stack[len(v.stack)-1] = s.Append(vs...)
			} else {
				vsCopy := make([]Value, len(vs))
				copy(vsCopy, vs)
				v.stack[len(v.stack)-1] = NewSlice(s.t.value(), vsCopy)
			}

		case codeNewSlice:
			i := &codes[v.frame.N]
			s := make([]Value, i.B)
			copy(s, v.stack[len(v.stack)-int(i.B):])
			v.stack = v.stack[:len(v.stack)-int(i.B)]
			value := NewSlice(Type(i.A), s)
			v.stack = append(v.stack, value)

		case codeMake:
			i := &codes[v.frame.N]
			l := v.stack[len(v.stack)-1].Int()
			s := make([]Value, l)
			for j := 0; j < l; j++ {
				s[j] = newZero(Type(i.A))
			}
			value := NewSlice(Type(i.A), s)
			v.stack[len(v.stack)-1] = value

		case codeNewMap:
			i := &codes[v.frame.N]
			value := NewMap(Type(i.A), Type(i.B), v.stack[len(v.stack)-int(i.C):])
			v.stack = v.stack[:len(v.stack)-int(i.C)]
			v.stack = append(v.stack, value)

		case codeRange:
			i := &codes[v.frame.N]
			a := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			if a.value != nil {
				v.stack[baseN+int(i.A)] = newNext(a.Range())
			} else {
				v.stack[baseN+int(i.A)] = newNext(nilRange())
			}
			v.frame.N += int(i.B)

		case codeIter:
			i := &codes[v.frame.N]
			key, value, ok := v.stack[baseN+int(i.A)].next()
			if !ok {
				break
			}
			v.frame.N += int(i.C)
			b1, b2 := splitParams(i.B)
			v.stack[baseN+int(b1)] = key
			v.stack[baseN+int(b2)] = value

		case codeStruct:
			i := &codes[v.frame.N]
			lookup := map[string]int{}
			data := newIntMap(int(i.A) / 2)
			methods := newIntMap(0)
			s := newStruct(0, lookup, nil, data, &methods)
			for n := 0; n < int(i.A); n += 2 {
				k := v.stack[len(v.stack)-int(i.A)+n].Int()
				s.addField(v.globals.Key(k), k, v.stack[len(v.stack)-int(i.A)+n+1])
			}
			v.stack = v.stack[:len(v.stack)-int(i.A)]
			v.stack = append(v.stack, s)

		case codeGlobalStruct:
			i := &codes[v.frame.N]
			prev := v.globals.Read(int(i.A))
			cur := v.stack[len(v.stack)-1]
			cur.value.(*structT).TypeN = int(i.A)
			v.stack = v.stack[:len(v.stack)-1]
			if prev.IsNil() {
				v.globals.Write(int(i.A), cur)
			} else {
				prev.syncFields(cur)
			}

		case codeNewStruct:
			i := &codes[v.frame.N]
			parent := v.globals.Read(int(i.A))
			s := newStructByIndex(parent, v.stack[len(v.stack)-int(i.B):])
			v.stack = v.stack[:len(v.stack)-int(i.B)]
			v.stack = append(v.stack, s)

		// case codeNewLocalStruct:
		// 	i := &codes[v.frame.N]
		// 	parent := v.stack[baseN+int(i.A)]
		// 	s := newStructByIndex(parent, v.stack[len(v.stack)-int(i.B):])
		// 	v.stack = v.stack[:len(v.stack)-int(i.B)]
		// 	v.stack = append(v.stack, s)

		case codeSetMethod:
			i := &codes[v.frame.N]
			k := int(i.A)
			m := v.stack[len(v.stack)-2]
			s := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-2]
			s.addMethod(v.globals.Key(k), k, m)

		case codeGetAttr:
			i := &codes[v.frame.N]
			r, k := v.stack[len(v.stack)-1], i.A
			v.stack[len(v.stack)-1] = r.getIndex(v, int(k))

		case codeSetAttr:
			i := &codes[v.frame.N]
			value, obj, key := v.stack[len(v.stack)-2], v.stack[len(v.stack)-1], i.A
			v.stack = v.stack[:len(v.stack)-2]
			obj.setIndex(v, int(key), value)

		case codeFastGetAttr:
			i := &codes[v.frame.N]
			r := v.stack[baseN+int(i.A)]
			v.stack = append(v.stack, r.getIndex(v, int(i.B)))

		case codeFastSetAttr:
			i := &codes[v.frame.N]
			obj, value, key := v.stack[baseN+int(i.A)], v.stack[len(v.stack)-1], i.B
			v.stack = v.stack[:len(v.stack)-1]
			obj.setIndex(v, int(key), value)

		case codeFastCallAttr:
			i := &codes[v.frame.N]
			obj := v.stack[baseN+int(i.A)]
			k := i.B
			f := obj.getIndex(v, int(k)).getFunc()
			c1, c2 := splitParams(i.C)
			call(v, f, int(c1), int(c2))

		case codeLocalMul:
			i := &codes[v.frame.N]
			a, b := v.stack[baseN+int(i.A)], v.stack[baseN+int(i.B)]
			v.stack = append(v.stack, a.opMul(b))

		case codeLocalAdd:
			i := &codes[v.frame.N]
			a, b := v.stack[baseN+int(i.A)], v.stack[baseN+int(i.B)]
			v.stack = append(v.stack, a.opAdd(b))

		case codeLocalDiv:
			i := &codes[v.frame.N]
			a, b := v.stack[baseN+int(i.A)], v.stack[baseN+int(i.B)]
			v.stack = append(v.stack, a.opDiv(b))

		case codeLocalSub:
			i := &codes[v.frame.N]
			a, b := v.stack[baseN+int(i.A)], v.stack[baseN+int(i.B)]
			v.stack = append(v.stack, a.opSub(b))

		case codePanic:
			a := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-1]
			panic(a.String())

		case codeCopy:
			a := v.stack[len(v.stack)-2]
			b := v.stack[len(v.stack)-1]
			v.stack = v.stack[:len(v.stack)-2]
			if b.t.base() == TypeSlice {
				copy(a.data(), b.data())
			} else {
				copy(a.data(), b.convert(TypeSlice).data())
			}

		case codePass:

		default: // TODO: comment out
			panic(fmt.Sprintf("unknown code: %v", v.frame.Codes[v.frame.N].Code))
		}
	}
}
