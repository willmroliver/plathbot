package cache

import (
	"sync"
	"time"
)

type ExpireWrapper[T any] struct {
	id        any
	value     T
	expiresAt time.Time
}

type ExpireCache[T any] struct {
	keyFunc func(T) any
	values  sync.Map
}

func (c *ExpireCache[T]) save(value T, lifespan time.Duration) {
	id := c.keyFunc(value)

	w := &ExpireWrapper[T]{
		ID:        id,
		Value:     value,
		ExpiresAt: time.Now().Add(lifespan),
	}

}
