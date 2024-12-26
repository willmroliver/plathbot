package ds

type DoublyNode[T any] struct {
	Val  T
	Next *DoublyNode[T]
	Prev *DoublyNode[T]
}

func (d *DoublyNode[T]) Detach() {
	if d.Prev != nil {
		d.Prev.Next = d.Next
	}
	if d.Next != nil {
		d.Next.Prev = d.Prev
	}
}

type DoublyList[T any] struct {
	head, tail *DoublyNode[T]
}

func NewDoublyList[T any]() *DoublyList[T] {
	head, tail := &DoublyNode[T]{}, &DoublyNode[T]{}

	head.Next = tail
	tail.Prev = head

	return &DoublyList[T]{head, tail}
}

func (d *DoublyList[T]) Push(val T) *DoublyNode[T] {
	last := d.tail.Prev
	last.Next = &DoublyNode[T]{val, d.tail, last}
	d.tail.Prev = last.Next

	return d.tail.Prev
}

func (d *DoublyList[T]) Unshift(val T) *DoublyNode[T] {
	first := d.head.Next
	first.Prev = &DoublyNode[T]{val, first, d.head}
	d.head.Next = first.Prev

	return d.head.Next
}

func (d *DoublyList[T]) Pop() (val T, ok bool) {
	if d.tail.Prev == d.head {
		return
	}

	node := d.tail.Prev
	node.Prev.Next = d.tail
	d.tail.Prev = node.Prev

	val, ok = node.Val, true
	return
}

func (d *DoublyList[T]) Shift() (val T, ok bool) {
	if d.head.Next == d.tail {
		return
	}

	node := d.head.Next
	node.Next.Prev = d.head
	d.head.Next = node.Next

	val, ok = node.Val, true
	return
}

func (d *DoublyList[T]) First() (val T, ok bool) {
	if d.head.Next != d.tail {
		val, ok = d.head.Next.Val, true
	}

	return
}

func (d *DoublyList[T]) Last() (val T, ok bool) {
	if d.tail.Prev != d.head {
		val, ok = d.tail.Prev.Val, true
	}

	return
}

func (d *DoublyList[T]) Find(test func(T) bool) *DoublyNode[T] {
	at := d.head.Next

	for at != d.tail {
		if test(at.Val) {
			return at
		}

		at = at.Next
	}

	return nil
}

func (d *DoublyList[T]) ForEach(cb func(T) bool) {
	at := d.head.Next

	for at != d.tail {
		if !cb(at.Val) {
			break
		}

		at = at.Next
	}
}

func (d *DoublyList[T]) Len() (len int) {
	at := d.head.Next

	for at != d.tail {
		len++
		at = at.Next
	}

	return
}

func (d *DoublyList[T]) Empty() bool {
	return d.head.Next == d.tail
}
