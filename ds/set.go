package ds

// Set is an unordered set of T.
type Set[T comparable] interface {
	// AddAll adds all elements from the given slice to this set.
	AddAll(data []T)
	// Add an element to the set, returns whether the element is new to the set.
	Add(element T) bool
	// Count returns the size of this set.
	Count() int
	// IsEmpty returns whether this set is empty.
	IsEmpty() bool
	// Has returns whether given element is in this set.
	Has(element T) bool
	// Remove removes an element from this set, returns whether such element exists.
	Remove(element T) bool
	// Pop removes and returns an arbitrary element from this set.
	Pop() (T, bool)
	// Clear removes all elements from this set.
	Clear()
	// Difference returns all elements of this set that are not in `other`.
	Difference(other Set[T]) []T
	// SymmetricDifference returns all elements that are in either this set or `other` but not in both.
	SymmetricDifference(other Set[T]) []T
	// Intersect returns all elements that exist in both sets.
	Intersect(other Set[T]) []T
	// Union returns all elements that are in both sets.
	Union(other Set[T]) []T
	// Equals returns whether this and `other` sets have the same size and contain the same elements.
	Equals(other Set[T]) bool
	// IsProperSubset returns whether all elements in this set are in `other` but they are not equal.
	IsProperSubset(other Set[T]) bool
	// Contains returns whether all elements in `other` are in this set.
	Contains(other Set[T]) bool
	// Iter returns a channel which could be used in a for range loop. The capacity of the returned
	// channel is the same as the size of the set at the time Iter is called.
	Iter() <-chan T
	// Items returns all elements of this set as a slice.
	Items() []T
	// ForEach executes the given doEachFn on every element in this set. If `doEachFn` returns true,
	// stop execution immediately.
	ForEach(doEachFn func(element T) bool)
	// MarshalJSON returns the JSON bytes of this set.
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON consumes a slice of JSON bytes to populate this set.
	UnmarshalJSON(b []byte) error
	// String returns a string representation of the current state of the set.
	String() string
}
