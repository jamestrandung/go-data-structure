package ds

// Stack is a stack of T.
type Stack[T any] interface {
    // Push adds an element on top of the stack.
    Push(element T)
    // Pop removes and returns the element at the top of the stack.
    Pop() (T, bool)
    // Peek retrieves and returns the element at the top of the stack.
    Peek() (T, bool)
}
