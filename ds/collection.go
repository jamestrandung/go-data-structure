package ds

// Collection is a collection of T.
type Collection[T comparable] interface {
    // Add an element to this collection.
    Add(element T)
    // AddAll adds all elements from the given slice to this collection.
    AddAll(elements []T)
    // Count returns the size of this collection.
    Count() int
    // IsEmpty returns whether this collection is empty.
    IsEmpty() bool
    // Has returns whether given element is in this collection.
    Has(element T) bool
    // Contains returns whether all elements in `other` are in this collection.
    Contains(other Collection[T]) bool
    // Equals returns whether this and `other` collections have the same size and contain the same elements.
    Equals(other Collection[T]) bool
    // Pop removes and returns an element from this collection.
    Pop() (T, bool)
    // Remove removes an element from this collection, returns whether this collection
    // changed as a result of the call.
    Remove(element T) bool
    // RemoveAll removes the given elements from this collection, returns whether this collection
    // changed as a result of the call.
    RemoveAll(elements []T) bool
    // RemoveIf removes the given element from this collection based on some condition, then returns
    // true if the element was removed. If the given element doesn't exist in this collection or the
    // element was not removed because of the condition func, false will be returned.
    RemoveIf(element T, conditionFn func() bool) bool
    // Clear removes all elements from this collection.
    Clear()
    // Iter returns a channel which could be used in a for range loop. The capacity of the returned
    // channel is the same as the size of the collection at the time Iter is called.
    Iter() <-chan T
    // Items returns all elements of this collection as a slice.
    Items() []T
    // ForEach executes the given doEachFn on every element in this collection. If `doEachFn` returns
    // true, stop execution immediately.
    ForEach(doEachFn func(element T) bool)
    // String returns a string representation of the current state of the collection.
    String() string
}
