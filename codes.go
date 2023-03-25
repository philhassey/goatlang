package goatlang

const (
	codeTODO = code(-(iota + 1))
	codeBreak
	codeContinue
)
const (
	codePass = code(iota)
	codePush
	codeAdd
	codeSub
	codeMul
	codeGlobalSet
	codeGlobalGet
	codeGlobalFunc
	codeGlobalZero
	codeLocalSet
	codeLocalGet
	codeLocalZero
	codeConst
	codeFunc
	codeReturn
	codeCall
	codeConvert
	codeCast
	codeJumpFalse
	codeJumpTrue
	codeJump
	codeLt
	codeDiv
	codeGt
	codeIncDec
	codeZero
	codeType
	codeNewSlice
	codeNewMap
	codeRange
	codeIter
	codeGet
	codeGetOk
	codeSet
	codeNegate
	codeAnd
	codeOr
	codeEq
	codeAppend
	codeFastGet
	codeFastSet
	codeFastGetInt
	codeFastSetInt
	codeFastCall
	codeMod
	codeLte
	codeGte
	codeNeq
	codeBitAnd
	codeBitOr
	codeBitLsh
	codeBitRsh
	codeBitXor
	codeBitComplement
	codeNot
	codeSlice
	codeLen
	codeDelete
	codeGlobalRef
	codeStruct
	codeGlobalStruct
	codeNewStruct
	codeNewLocalStruct
	codeSetMethod
	codeGetAttr
	codeSetAttr

	codeFastGetAttr
	codeFastSetAttr
	codeFastCallAttr
	codeCallVariadic

	codePop

	codeMake
	codeCopy
	codePanic

	codeLocalMul
	codeLocalDiv
	codeLocalAdd
	codeLocalSub
	codeLocalIncDec
)

var codeToString = map[code]string{
	codeBreak:    "BREAK",
	codeContinue: "CONTINUE",
	codeTODO:     "TODO",

	codePass: "PASS",
	codePush: "PUSH",
	codeAdd:  "ADD",
	codeSub:  "SUB",
	codeMul:  "MUL",
	codeMod:  "MOD",

	codeGlobalSet:  "GLOBALSET",
	codeGlobalGet:  "GLOBALGET",
	codeGlobalFunc: "GLOBALFUNC",
	codeGlobalZero: "GLOBALZERO",
	codeLocalSet:   "LOCALSET",
	codeLocalGet:   "LOCALGET",
	codeLocalZero:  "LOCALZERO",
	codeConst:      "CONST",
	codeFunc:       "FUNC",
	codeReturn:     "RETURN",
	codeCall:       "CALL",
	codeJumpFalse:  "JUMPFALSE",
	codeJumpTrue:   "JUMPTRUE",
	codeJump:       "JUMP",
	codeDiv:        "DIV",
	codeIncDec:     "INCDEC",
	codeConvert:    "CONVERT",
	codeCast:       "CAST",
	codeZero:       "ZERO",
	codeType:       "TYPE",
	codeNewSlice:   "NEWSLICE",
	codeNewMap:     "NEWMAP",
	codeRange:      "RANGE",
	codeIter:       "ITER",
	codeGet:        "GET",
	codeGetOk:      "GETOK",
	codeSet:        "SET",
	codeNegate:     "NEGATE",
	codeAnd:        "AND",
	codeOr:         "OR",

	codeLt:           "LT",
	codeGt:           "GT",
	codeLte:          "LTE",
	codeGte:          "GTE",
	codeNeq:          "NEQ",
	codeEq:           "EQ",
	codeAppend:       "APPEND",
	codeFastGet:      "FASTGET",
	codeFastSet:      "FASTSET",
	codeFastGetInt:   "FASTGETINT",
	codeFastSetInt:   "FASTSETINT",
	codeFastCall:     "FASTCALL",
	codeCallVariadic: "CALLVARIADIC",

	codeBitOr:         "BITOR",
	codeBitXor:        "BITXOR",
	codeBitAnd:        "BITAND",
	codeBitLsh:        "BITLSH",
	codeBitRsh:        "BITRSH",
	codeBitComplement: "BITCOMPLEMENT",

	codeNot:    "NOT",
	codeSlice:  "SLICE",
	codeLen:    "LEN",
	codeDelete: "DELETE",

	codeGlobalRef:      "GLOBALREF", // alias for Push, but prints out ref for dumps
	codeStruct:         "STRUCT",
	codeGlobalStruct:   "GLOBALSTRUCT",
	codeNewStruct:      "NEWSTRUCT",
	codeNewLocalStruct: "NEWLOCALSTRUCT",
	codeSetMethod:      "SETMETHOD",
	codeGetAttr:        "GETATTR",
	codeSetAttr:        "SETATTR",
	codeFastGetAttr:    "FASTGETATTR",
	codeFastSetAttr:    "FASTSETATTR",
	codeFastCallAttr:   "FASTCALLATTR",

	codePop: "POP",

	codeMake:  "MAKE",
	codeCopy:  "COPY",
	codePanic: "PANIC",

	codeLocalMul:    "LOCALMUL",
	codeLocalDiv:    "LOCALDIV",
	codeLocalAdd:    "LOCALADD",
	codeLocalSub:    "LOCALSUB",
	codeLocalIncDec: "LOCALINCDEC",
}

func (c code) String() string {
	return codeToString[c]
}
