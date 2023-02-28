package goatlang

type lookup struct {
	keyToIndex map[string]int
	indexToKey []string
	data       []Value
	cap        int
}

func newLookup() *lookup {
	return &lookup{
		keyToIndex: map[string]int{},
	}
}

func (l *lookup) Len() int {
	return len(l.data)
}

func (l *lookup) Cap() int {
	return l.cap
}

func (l *lookup) Read(index int) Value {
	return l.data[index]
}

func (l *lookup) Write(index int, v Value) {
	l.data[index] = v
}

func (l *lookup) Assign(index int, v Value) {
	l.data[index] = v.assign(l.data[index].t)
}

func (l *lookup) Set(key string, v Value) {
	n := l.Index(key)
	l.data[n] = v
}

func (l *lookup) Get(key string) Value {
	n := l.Index(key)
	return l.data[n]
}

// Index returns the index of the key, creating it if required
func (l *lookup) Index(key string) int {
	n, ok := l.keyToIndex[key]
	if ok {
		return n
	}
	n = int(len(l.data))
	l.data = append(l.data, Value{})
	l.indexToKey = append(l.indexToKey, key)
	if len(l.data) > l.cap {
		l.cap = len(l.data)
	}
	l.keyToIndex[key] = n
	return n
}

func (l *lookup) shadow(key string) {
	if n, ok := l.keyToIndex[key]; ok {
		l.shadow("~" + key)
		l.keyToIndex["~"+key] = n
		delete(l.keyToIndex, key)
	}
}

func (l *lookup) unshadow(key string) {
	if n, ok := l.keyToIndex["~"+key]; ok {
		l.keyToIndex[key] = n
		l.unshadow("~" + key)
	}
}

func (l *lookup) Shadow(key string) int {
	l.shadow(key)
	return l.Index(key)
}

// func (l *lookup) Pop() {
// 	n := len(l.data) - 1
// 	key := l.indexToKey[n]
// 	l.data = l.data[:n]
// 	l.indexToKey = l.indexToKey[:n]
// 	delete(l.keyToIndex, key)
// }

func (l *lookup) Drop(t int) {
	for i := 1; i <= t; i++ {
		n := len(l.data) - i
		key := l.indexToKey[n]
		if key == "" {
			continue
		}
		delete(l.keyToIndex, key)
		l.indexToKey[n] = ""
		l.unshadow(key)
	}
}

func (l *lookup) Exists(key string) bool {
	_, ok := l.keyToIndex[key]
	return ok
}

func (l *lookup) Key(index int) string {
	return l.indexToKey[index]
}
