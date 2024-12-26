package ds

type LRUCache[K comparable, V any] struct {
	data map[K]V
	lim  int
}

func NewLRUCache[K comparable, T any](lim int) *LRUCache[K, T] {
	return &LRUCache[K, T]{
		map[K]T{},
		lim,
	}
}
