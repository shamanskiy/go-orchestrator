package queues

type Queue[T any] struct {
	data []T
}

func New[T any]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Enqueue(val T) {
	q.data = append(q.data, val)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.data) == 0 {
		var zero T
		return zero, false
	}
	val := q.data[0]
	q.data = q.data[1:]
	return val, true
}

func (q *Queue[T]) Len() int {
	return len(q.data)
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.data) == 0
}
