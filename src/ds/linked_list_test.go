package ds_test

import (
	"testing"

	"github.com/willmroliver/plathbot/src/ds"
)

func TestDoublyLinkedList(t *testing.T) {
	list := ds.NewDoublyList[int]()

	if len := list.Len(); len != 0 {
		t.Errorf("Len() - Expected 0; Got %d", len)
	}

	n5, n1 := list.Push(5), list.Unshift(1)

	if n, ok := list.First(); !ok || n != 1 {
		t.Errorf("First() - Expected ok, %d; Got %v, %d", 1, ok, n)
	}

	if n, ok := list.Last(); !ok || n != 5 {
		t.Errorf("Last() - Expected ok, %d; Got %v, %d", 5, ok, n)
	}

	n9 := list.Push(9)

	if len := list.Len(); len != 3 {
		t.Errorf("Len() - Expected %d; Got %d", 3, len)
	}

	list.ForEach(func(n int) bool { t.Logf("%d ", n); return true })

	n5.Detach()

	if n, ok := list.First(); !ok || n != 1 {
		t.Errorf("First() - Expected ok, %d; Got %v, %d", 1, ok, n)
	}

	if n, ok := list.Last(); !ok || n != 9 {
		t.Errorf("Last() - Expected ok, %d; Got %v, %d", 9, ok, n)
	}

	if f := list.Find(func(v int) bool { return v == 9 }); f != n9 {
		t.Errorf("Find() - Expected %+v; Got %+v", n9, f)
	}

	if n, ok := list.Shift(); !ok || n != n1.Val {
		t.Errorf("Shift() - Expected ok, %d; Got %v, %d", n1.Val, ok, n)
	}

	n9.Detach()

	if !list.Empty() {
		t.Error("Empty() - Expected empty list")
	}
}
