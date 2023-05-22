package utils

type (
	Map[K, V any] interface {
		Get(k K) (V, bool)
		Put(k K, v V)
	}
	Cmp[H comparable, Self any] interface {
		Hash() H
		Eq(Self) bool
	}
	customMap[K, V any, H comparable] struct {
		hash  func(K) H
		cmp   func(K, K) bool
		inner map[H][]customMapEntry[K, V]
	}
	customMapEntry[K, V any] struct {
		key K
		val V
	}
)

func (m *customMap[K, V, H]) Get(k K) (v V, found bool) {
	l := m.inner[m.hash(k)]
	for i := len(l) - 1; i >= 0; i-- {
		e := l[i]
		if m.cmp(k, e.key) {
			v, found = e.val, true
			return
		}
	}
	return
}

func (m *customMap[K, V, H]) Put(k K, v V) {
	h := m.hash(k)
	l := m.inner[h]
	/*
		WARN: we disable uniqueness check here for MASSIVE speed gain.
		the alternative would have to make many deep-eq checks using reflection.
		since we know the input is unique, this does not waste space.
		Just to ensure correctness in either case; we iterate the inner array
		in reverse in Get().
		for _, e := range l {
			if m.cmp(k, e.key) {
				e.val = v
				return
			}
		}
	*/
	m.inner[h] = append(l, customMapEntry[K, V]{key: k, val: v})
}

func NewMap[K, V any, H comparable](
	hashFn func(K) H, cmpFn func(K, K) bool,
) Map[K, V] {
	return &customMap[K, V, H]{
		hash: hashFn, cmp: cmpFn,
		inner: make(map[H][]customMapEntry[K, V]),
	}
}
