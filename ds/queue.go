package ds

// Queue is a queue of T.
type Queue[T any] interface {
    // Offer adds an element at the back of the queue, returns whether the element was
    // added to this queue.
    Offer(element T) bool
    // Pop removes and returns the element at the front of the queue.
    Pop() (T, bool)
    // Peek retrieves and returns the element at the front of the queue.
    Peek() (T, bool)
}