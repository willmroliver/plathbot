package ds

type SinglyNode[T any] struct {
	val  T
	next *SinglyNode[T]
}

type DoublyNode[T any] struct {
	val  T
	next *DoublyNode[T]
	prev *DoublyNode[T]
}
