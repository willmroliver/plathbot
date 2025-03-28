package ds

import "sync"

type KVPair[K comparable, V any] struct {
	key K
	val V
}

type LRUCache[K comparable, V any] struct {
	data map[K]*DoublyNode[*KVPair[K, V]]
	ll   *DoublyList[*KVPair[K, V]]
	lim  int
	mux  *sync.Mutex
}

func NewLRUCache[K comparable, V any](lim int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		map[K]*DoublyNode[*KVPair[K, V]]{},
		NewDoublyList[*KVPair[K, V]](),
		lim,
		&sync.Mutex{},
	}
}

func (c *LRUCache[K, V]) Save(k K, v V) {
	if n, ok := c.data[k]; ok {
		n.Detach()
		c.ll.Unshift(&KVPair[K, V]{k, v})
		return
	}

	c.data[k] = c.ll.Unshift(&KVPair[K, V]{k, v})

	if len(c.data) > c.lim {
		v, _ := c.ll.Pop()
		delete(c.data, v.key)
	}
}

func (c *LRUCache[K, V]) Load(k K) (val V, ok bool) {
	if node, ok := c.data[k]; ok {
		node.Detach()
		c.ll.Unshift(node.Val)
		return node.Val.val, ok
	}

	return
}

func (c *LRUCache[K, V]) Delete(k K) bool {
	if node, ok := c.data[k]; ok {
		node.Detach()
		delete(c.data, node.Val.key)
		return ok
	}

	return false
}

func (c *LRUCache[K, V]) ForEach(cb func(V) bool) {
	c.ll.ForEach(func(p *KVPair[K, V]) bool {
		return cb(p.val)
	})
}

func (c *LRUCache[K, V]) Lock() {
	c.mux.Lock()
}

func (c *LRUCache[K, V]) Unlock() {
	c.mux.Unlock()
}
