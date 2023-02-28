package goatlang

import (
	"sort"
)

func treeSort(top *token) *token {
	tt := top.Tokens
	priority := map[string]int{
		"package":  100, // BUG: package never seen ???
		"import":   90,
		"type":     80,
		"const":    70,
		"method":   60,
		"function": 50,
		"init":     -10,
	}
	sort.SliceStable(tt, func(ai, bi int) bool {
		a, b := tt[ai], tt[bi]
		am, bm := priority[a.Symbol], priority[b.Symbol]
		return am > bm
	})
	return top
}
