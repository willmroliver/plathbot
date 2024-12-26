package ds_test

import (
	"testing"

	"github.com/willmroliver/plathbot/src/ds"
)

func TestPriorityQueue(t *testing.T) {
	pq := ds.NewPriorityQueue[int](func(a, b int) int {
		if a > b {
			return 1
		}

		return -1
	})

	pq.Push(1)
	pq.Push(2)
	pq.Push(3)
	pq.Push(4)
	pq.Push(5)
	pq.Push(6)
	pq.Push(7)

	if pq.Pop() != 7 {
		t.Error("Max Heap did not return max value")
	}
	if pq.Top() != 6 {
		t.Error("Max Heap did not re-adjust after pop")
	}

	for _, n := range []int{6, 5, 4, 3, 2, 1} {
		m := pq.Pop()
		if m != n {
			t.Errorf("Max Heap did not return max value. Expected %d, Got %d", n, m)
		}
	}

	pq.Push(-12)
	pq.Push(43)
	pq.Push(2)
	pq.Push(-12)
	pq.Push(100)
	pq.Push(55)
	pq.Push(-1)
	pq.Push(0)

	for _, n := range []int{100, 55, 43, 2, 0, -1, -12, -12} {
		m := pq.Pop()
		if m != n {
			t.Errorf("Max Heap did not return max value. Expected %d, Got %d", n, m)
		}
	}
}
