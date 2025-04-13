package common

type Queue[T any] []T

func (q *Queue[T]) Push(el T) {
	*q = append(*q, el)
}

func (q *Queue[T]) Deque() T {
	self := *q
	var el T
	el, *q = self[0], self[1:]
	return el
}
