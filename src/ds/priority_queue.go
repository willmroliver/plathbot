package ds

type PriorityQueue[T any] struct {
	data    []T
	compare func(T, T) int
}

func NewPriorityQueue[T any](compare func(T, T) int) *PriorityQueue[T] {
	pq := make([]T, 0, 8)

	return &PriorityQueue[T]{
		pq,
		compare,
	}
}

func (q *PriorityQueue[T]) Push(val T) {
	q.data = append(q.data, val)
	i := len(q.data) - 1
	j := (i - 1) / 2
	a, b := q.data[i], q.data[j]

	for i > 0 && q.compare(a, b) > 0 {
		q.data[i], q.data[j] = b, a
		i = j
		j = (i - 1) / 2
		a, b = q.data[i], q.data[j]
	}
}

func (q *PriorityQueue[T]) Pop() (val T) {
	val = q.data[0]
	n, at, i, j := len(q.data), 0, 1, 2
	q.data[0] = q.data[n-1]

	for i < n {
		if q.compare(q.data[i], q.data[at]) < 0 && j < n &&
			q.compare(q.data[j], q.data[at]) < 0 {
			break
		}

		if j < n && q.compare(q.data[j], q.data[i]) > 0 {
			q.data[at], q.data[j] = q.data[j], q.data[at]
			at = j
		} else {
			q.data[at], q.data[i] = q.data[i], q.data[at]
			at = i
		}

		i, j = 2*at+1, 2*at+2
	}

	q.data = q.data[:n-1]
	return
}

func (q *PriorityQueue[T]) Top() (val T) {
	return q.data[0]
}

func (q *PriorityQueue[T]) Len() int {
	return len(q.data)
}
