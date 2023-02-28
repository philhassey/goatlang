package goatlang

import (
	"math/rand"
	"reflect"
	"testing"
)

// go test -bench=. -benchtime=100000000x -run Map
// go test -count 1 -run Map

var testIntMapData []int

const (
	testIntIndexMask = 0xffff
	testIntValueMask = 0xff
)

func init() {
	for i := 0; i < 65536; i++ {
		testIntMapData = append(testIntMapData, rand.Int()&testIntValueMask)
	}
}

func TestIntMap(t *testing.T) {
	im := newIntMap(256)
	for i := 0; i < 16; i++ {
		im.Set(1234, Int(42))
		value, ok := im.Get(1234)
		assert(t, "exists ok", ok, true)
		assert(t, "exists value", value.Int(), 42)
		im.Delete(1234)
		value, ok = im.Get(1234)
		assert(t, "deleted ok", ok, false)
		assert(t, "deleted value", value.Int(), 0)
	}
}

func TestIntMap_Assign(t *testing.T) {
	im := newIntMap(256)
	im.Set(1234, Int(32))
	im.Assign(1234, Int(42))
	im.Assign(4321, Int(52))
	value, ok := im.Get(1234)
	assert(t, "exists ok", ok, true)
	assert(t, "exists value", value.Int(), 42)
	_, ok = im.Get(4321)
	assert(t, "not exist ok", ok, false)

	collisionKey := 1234 + 512 // double map size
	im.Assign(collisionKey, Int(62))
	value, ok = im.Get(1234)
	assert(t, "exists ok", ok, true)
	assert(t, "exists value", value.Int(), 42)
	_, ok = im.Get(collisionKey)
	assert(t, "not exist ok", ok, false)
}

func TestIntMap_Copy(t *testing.T) {
	m := newIntMap(0)
	m.Set(1, Int(42))

	c := m.Copy()
	if !reflect.DeepEqual(c, m) {
		t.Fatalf("copy is not equal")
	}

	m.Set(1, Int(43))
	v, _ := c.Get(1)
	if v.Int() != 42 {
		t.Fatalf("copy is not distinct")
	}
}

func TestIntMap_Mix(t *testing.T) {
	im := newIntMap(0)
	for n := 0; n < 65536; n += 4 {
		im.Set(testIntMapData[n&testIntIndexMask], Int(testIntMapData[(n+1)&testIntIndexMask]))
		_, _ = im.Get(testIntMapData[(n+2)&testIntIndexMask])
		im.Delete(testIntMapData[(n+3)&testIntIndexMask])
	}
	gm := map[int]Value{}
	for n := 0; n < 65536; n += 4 {
		gm[testIntMapData[n&testIntIndexMask]] = Int(testIntMapData[(n+1)&testIntIndexMask])
		_ = gm[testIntMapData[(n+2)&testIntIndexMask]]
		delete(gm, testIntMapData[(n+3)&testIntIndexMask])
	}
	assert(t, "len", im.Len(), len(gm))
	for k := 0; k <= testIntValueMask; k++ {
		iv, iok := im.Get(k)
		gv, gok := gm[k]
		assert(t, "ok", iok, gok)
		assert(t, "value", iv.Int(), gv.Int())
	}
}

func BenchmarkIntMap_Set(b *testing.B) {
	m := newIntMap(0)
	for n := 0; n < b.N; n++ {
		m.Set(testIntMapData[n&testIntIndexMask], Int(testIntMapData[(n+1)&testIntIndexMask]))
	}
}

func BenchmarkIntMap_Get(b *testing.B) {
	m := newIntMap(0)
	for n := 0; n < b.N; n++ {
		m.Set(testIntMapData[n&testIntIndexMask], Int(testIntMapData[(n+1)&testIntIndexMask]))
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m.Get(testIntMapData[n&testIntIndexMask])
	}
}

func BenchmarkIntMap_Mix(b *testing.B) {
	m := newIntMap(0)
	for n := 0; n < b.N; n += 4 {
		m.Set(testIntMapData[n&testIntIndexMask], Int(testIntMapData[(n+1)&testIntIndexMask]))
		m.Get(testIntMapData[(n+2)&testIntIndexMask])
		m.Delete(testIntMapData[(n+3)&testIntIndexMask])
	}
}

func BenchmarkIntMap_Delete(b *testing.B) {
	m := newIntMap(0)
	for n := 0; n < b.N; n++ {
		m.Set(testIntMapData[n&testIntIndexMask], Int(testIntMapData[(n+1)&testIntIndexMask]))
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		m.Delete(testIntMapData[n&testIntIndexMask])
	}
}

func BenchmarkGoMap_Set(b *testing.B) {
	m := map[int]Value{}
	for n := 0; n < b.N; n++ {
		m[testIntMapData[n&testIntIndexMask]] = Int(testIntMapData[(n+1)&testIntIndexMask])
	}
}

func BenchmarkGoMap_Get(b *testing.B) {
	m := map[int]Value{}
	for n := 0; n < b.N; n++ {
		m[testIntMapData[n&testIntIndexMask]] = Int(testIntMapData[(n+1)&testIntIndexMask])
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = m[testIntMapData[n&testIntIndexMask]]
	}
}

func BenchmarkGoMap_Mix(b *testing.B) {
	m := map[int]Value{}
	for n := 0; n < b.N; n += 4 {
		m[testIntMapData[n&testIntIndexMask]] = Int(testIntMapData[(n+1)&testIntIndexMask])
		_ = m[testIntMapData[(n+2)&testIntIndexMask]]
		delete(m, testIntMapData[(n+3)&testIntIndexMask])
	}
}

func BenchmarkGoMap_Delete(b *testing.B) {
	m := map[int]Value{}
	for n := 0; n < b.N; n++ {
		m[testIntMapData[n&testIntIndexMask]] = Int(testIntMapData[(n+1)&testIntIndexMask])
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		delete(m, testIntMapData[n&testIntIndexMask])
	}
}

// func intMapHash(k int) int {
// https://en.wikipedia.org/wiki/Xorshift
// 32-bit
// k ^= k << 13
// k ^= k >> 17
// k ^= k << 5
// 64-bit
// k ^= k << 13
// k ^= k >> 7
// k ^= k << 17
// 16-bit http://www.retroprogramming.com/2017/07/xorshift-pseudorandom-numbers-in-z80.html
// k ^= k << 7
// k ^= k >> 9
// k ^= k << 8

// https://mostlymangling.blogspot.com/2019/12/stronger-better-morer-moremur-better.html
// k ^= k >> 27
// k *= 0x3C79AC492BA7B653
// k ^= k >> 33
// k *= 0x1C69B3F74AC4AE35
// k ^= k >> 27

// https://nullprogram.com/blog/2018/07/31/
// https://github.com/skeeto/hash-prospector
// k ^= k >> 16
// k *= 0x21f0aaad
// k ^= k >> 15
// k *= 0xd35a2d97
// k ^= k >> 16

// 	return k
// }
