# Set

## Usage

```go
import (
"github.com/jamestrandung/go-data-structure/set"
)

```

```bash
go get "github.com/jamestrandung/go-data-structure/set"
```

## Implemented interface

```go
// Set is an unordered set of T.
type Set[T comparable] interface {
    Collection[T]
    // Difference returns all elements of this set that are not in `other`.
    Difference(other Set[T]) []T
    // SymmetricDifference returns all elements that are in either this set or `other` but not in both.
    SymmetricDifference(other Set[T]) []T
    // Intersect returns all elements that exist in both sets.
    Intersect(other Set[T]) []T
    // Union returns all elements that are in both sets.
    Union(other Set[T]) []T
    // IsProperSubset returns whether all elements in this set are in `other` but they are not equal.
    IsProperSubset(other Set[T]) bool
}

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
    // Equals returns whether this and `other` sets have the same size and contain the same elements.
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
    // MarshalJSON returns the JSON bytes of this collection.
    MarshalJSON() ([]byte, error)
    // UnmarshalJSON consumes a slice of JSON bytes to populate this collection.
    UnmarshalJSON(b []byte) error
    // String returns a string representation of the current state of the collection.
    String() string
}
```

## Hash Set

Go does not have a built-in data type for Set. This package aims to fill in this gap by providing a `HashSet`
implementation that is backed by the basic `map` type.

```go
// Create a new set
hs := set.NewHashSet[string]()

// Add element to set
hs.Add("foo")

// Check if the set contains "bar"
found := hs.Has("bar")

// Remove "foo" element
hs.Remove("foo")
```

For more examples, have a look at `hset_test.go`.

## Concurrent Set

This is a thread-safe alternative for `HashSet` that is backed by `emap.ConcurrentMap`.

```go
// Create a new set
cs := set.NewConcurrentSet[string]()

// Add element to set
cs.Add("foo")

// Check if the set contains "bar"
found := cs.Has("bar")

// Remove "foo" element
cs.Remove("foo")
```

For more examples, have a look at `cset_test.go`.