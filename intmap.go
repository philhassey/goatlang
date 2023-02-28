package goatlang

// https://programming.guide/robin-hood-hashing.html

type intMapPair struct {
	distance int
	key      int
	value    Value
}

type intMap struct {
	pairs      []intMapPair
	total      int
	size, mask int
	min, max   int
}

const (
	intMapMin = 16
)

// newIntMap creats an IntMap large enough to insert that many without resizing.
func newIntMap(alloc int) intMap {
	size := intMapMin
	for size < (alloc << 1) {
		size <<= 1
	}
	m := intMap{}
	m.init(size, 0)
	return m
}

func (m *intMap) Copy() intMap {
	pairs := make([]intMapPair, len(m.pairs))
	copy(pairs, m.pairs)
	// for i := 0; i < len(pairs); i++ {
	// 	p := m.pairs[i]
	// 	if p.distance == 0 {
	// 		continue
	// 	}
	// 	pairs[i] = intMapPair{
	// 		distance: p.distance,
	// 		key:      p.key,
	// 		value:    p.value.copy(),
	// 	}
	// }
	return intMap{
		pairs: pairs,
		total: m.total,
		size:  m.size,
		mask:  m.mask,
		min:   m.min,
		max:   m.max,
	}
}

func (m *intMap) Len() int {
	return m.total
}

func (m *intMap) Set(key int, value Value) {
	hash := intMapHash(key)
	i := hash
	for {
		i &= m.mask
		if m.pairs[i].distance == 0 {
			m.insert(hash, key, value)
			m.total++
			if m.total > m.max {
				m.resize(m.size << 1)
			}
			return
		}
		if m.pairs[i].key == key {
			m.pairs[i].value = value
			return
		}
		i++
	}
}

func (m *intMap) Assign(key int, value Value) {
	hash := intMapHash(key)
	i := hash
	for {
		i &= m.mask
		if m.pairs[i].distance == 0 {
			return
		}
		if m.pairs[i].key == key {
			m.pairs[i].value = value.assign(m.pairs[i].value.t)
			return
		}
		i++
	}
}

func (m *intMap) insert(i, key int, value Value) {
	pair := intMapPair{distance: 1, key: key, value: value}
	for {
		i &= m.mask
		if m.pairs[i].distance < pair.distance {
			pair, m.pairs[i] = m.pairs[i], pair
			if pair.distance == 0 {
				return
			}
		}
		pair.distance++
		i++
	}
}

func (m *intMap) Get(key int) (Value, bool) {
	i := intMapHash(key)
	for {
		i &= m.mask
		if m.pairs[i].distance == 0 {
			return Value{}, false
		}
		if m.pairs[i].key == key {
			return m.pairs[i].value, true
		}
		i++
	}
}

func (m *intMap) Delete(key int) {
	i := intMapHash(key)
	for {
		i &= m.mask
		pair := m.pairs[i]
		if pair.distance == 0 {
			return
		}
		if pair.key == key {
			prev, i := i, i+1
			for {
				i &= m.mask
				if m.pairs[i].distance <= 1 {
					m.pairs[prev].distance = 0
					m.total--
					if m.total < m.min {
						m.resize(m.size >> 1)
					}
					return
				}
				m.pairs[prev] = m.pairs[i]
				m.pairs[prev].distance--
				prev, i = i, i+1
			}
		}
		i++
	}
}

func (m *intMap) init(size, total int) {
	m.pairs = make([]intMapPair, size)
	m.total = total
	m.size, m.mask = size, size-1
	m.max = size * 3 / 4
	m.min = size / 4
}

func (m *intMap) resize(size int) {
	if size < intMapMin {
		size = intMapMin
	}
	if size == m.size {
		return
	}
	pairs, total := m.pairs, m.total
	m.init(size, total)
	for _, pair := range pairs {
		if pair.distance == 0 {
			continue
		}
		hash := intMapHash(pair.key)
		m.insert(hash, pair.key, pair.value)
	}
}

// intMapHash alternatives in the test file
func intMapHash(k int) int {
	return k
}
