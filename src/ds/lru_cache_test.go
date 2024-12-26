package ds_test

import (
	"testing"

	"github.com/willmroliver/plathbot/src/ds"
)

func TestLRUCache(t *testing.T) {
	cache := ds.NewLRUCache[int, string](3)

	cache.Save(3, "two")
	cache.Save(3, "three")
	cache.Save(1, "one")
	cache.Save(2, "two")

	if val, ok := cache.Load(2); !ok || val != "two" {
		t.Errorf("Load() - Expected ok, %s; Got %v, %s", "two", ok, val)
	}

	cache.Save(4, "four")

	if val, ok := cache.Load(3); ok {
		t.Errorf("Load() - Expected !ok; Got %v, %s", ok, val)
	}

	cache.Save(5, "five")

	if val, ok := cache.Load(1); ok {
		t.Errorf("Load() - Expected !ok; Got %v, %s", ok, val)
	}

	if val, ok := cache.Load(4); !ok || val != "four" {
		t.Errorf("Load() - Expected ok, %s; Got %v, %s", "four", ok, val)
	}

	for _, k := range []int{1, 3} {
		if cache.Delete(k) {
			t.Errorf("Delete() - Expected !ok for key: %d", k)
		}
	}

	for _, k := range []int{2, 4, 5} {
		if !cache.Delete(k) {
			t.Errorf("Delete() - Expected ok for key: %d", k)
		}
	}

	for _, k := range []int{1, 2, 3, 4, 5} {
		if val, ok := cache.Load(k); ok {
			t.Errorf("Load() - Expected !ok; Got %v, %s", ok, val)
		}
	}
}
