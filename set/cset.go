package set

import (
	"encoding/json"
	"fmt"

	"github.com/jamestrandung/go-data-structure/ds"
	"github.com/jamestrandung/go-data-structure/emap"
)

var (
	defaultShardCount = 32
)

type ConcurrentSet[T comparable] emap.ConcurrentMap[T, struct{}]

// NewConcurrentSet returns a new instance of ConcurrentSet with the given elements.
func NewConcurrentSet[T comparable](elements ...T) ConcurrentSet[T] {
	return NewConcurrentSetWithConcurrencyLevel[T](defaultShardCount, elements...)
}

// NewConcurrentSetWithConcurrencyLevel returns a new instance of ConcurrentSet with the given
// amount of shards and contains the given elements.
func NewConcurrentSetWithConcurrencyLevel[T comparable](concurrencyLevel int, elements ...T) ConcurrentSet[T] {
	m := emap.NewConcurrentMapWithConcurrencyLevel[T, struct{}](concurrencyLevel)

	for _, element := range elements {
		m.Set(element, struct{}{})
	}

	return ConcurrentSet[T](m)
}

func (cs ConcurrentSet[T]) cast() emap.ConcurrentMap[T, struct{}] {
	return emap.ConcurrentMap[T, struct{}](cs)
}

// Add an element to this set.
func (cs ConcurrentSet[T]) Add(element T) {
	cs.cast().Set(element, struct{}{})
}

// AddAll adds all elements from the given slice to this set.
func (cs ConcurrentSet[T]) AddAll(elements []T) {
	for _, element := range elements {
		cs.cast().Set(element, struct{}{})
	}
}

// Count returns the size of this set.
func (cs ConcurrentSet[T]) Count() int {
	return cs.cast().Count()
}

// IsEmpty returns whether this set is empty.
func (cs ConcurrentSet[T]) IsEmpty() bool {
	return cs.Count() == 0
}

// Has returns whether given element is in this set.
func (cs ConcurrentSet[T]) Has(element T) bool {
	return cs.cast().Has(element)
}

// Contains returns whether all elements in `other` are in this set.
func (cs ConcurrentSet[T]) Contains(other ds.Collection[T]) bool {
	isSuperset := true
	other.ForEach(
		func(element T) bool {
			if !cs.Has(element) {
				isSuperset = false
				return true
			}

			return false
		},
	)

	return isSuperset
}

// Equals returns whether this and `other` collections have the same size and contain the same elements.
func (cs ConcurrentSet[T]) Equals(other ds.Collection[T]) bool {
	if cs.Count() != other.Count() {
		return false
	}

	// Avoid cases where `other` cannot perform `Has` efficiently
	if _, ok := other.(ds.Set[T]); ok {
		return other.Contains(cs)
	}

	return cs.Contains(other)
}

// Pop removes and returns an element from this set.
func (cs ConcurrentSet[T]) Pop() (v T, ok bool) {
	for _, shard := range cs.cast() {
		items, unlockFn := shard.GetItemsToWrite()

		for element := range items {
			delete(items, element)
			unlockFn()

			return element, true
		}

		unlockFn()
	}

	return
}

// Remove removes an element from this set, returns whether this set changed
// as a result of the call.
func (cs ConcurrentSet[T]) Remove(element T) bool {
	_, found := cs.cast().Remove(element)
	return found
}

// RemoveAll removes the given elements from this set, returns whether this set
// changed as a result of the call.
func (cs ConcurrentSet[T]) RemoveAll(elements []T) bool {
	return cs.cast().RemoveAll(elements)
}

// RemoveIf removes the given element from this set based on some condition, then returns
// true if the element was removed. If the given element doesn't exist in this set or the
// element was not removed because of the condition func, false will be returned.
//
// Note: Condition func is called while lock is held. Hence, it must NOT access this set
// as it may lead to deadlock since sync.RWLock in the underlying emap.ConcurrentMap is
// not reentrant.
func (cs ConcurrentSet[T]) RemoveIf(element T, conditionFn func() bool) bool {
	_, removed := cs.cast().RemoveIf(
		element,
		func(currentVal struct{}) bool {
			return conditionFn()
		},
	)

	return removed
}

// Clear removes all elements from this set.
func (cs ConcurrentSet[T]) Clear() {
	cs.cast().Clear()
}

// Iter returns a channel which could be used in a for range loop. The capacity of the returned
// channel is the same as the size of the set at the time Iter is called.
func (cs ConcurrentSet[T]) Iter() <-chan T {
	c := cs.cast().Iter()

	out := make(chan T, cap(c))

	go func() {
		defer close(out)

		for element := range c {
			out <- element.Key
		}
	}()

	return out
}

// Items returns all elements of this set as a slice.
func (cs ConcurrentSet[T]) Items() []T {
	var result []T

	cs.ForEach(
		func(element T) bool {
			result = append(result, element)

			return false
		},
	)

	return result
}

// ForEach executes the given doEachFn on every element in this set. If `doEachFn` returns true,
// stop execution immediately.
//
// Note: doEachFn is called while lock is held. Hence, it must NOT access this set as it may
// lead to deadlock since sync.RWLock in the underlying emap.ConcurrentMap is not reentrant.
func (cs ConcurrentSet[T]) ForEach(doEachFn func(element T) bool) {
	cs.cast().ForEach(
		func(key T, val struct{}) bool {
			return doEachFn(key)
		},
	)
}

// MarshalJSON returns the JSON bytes of this set.
func (cs ConcurrentSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(cs.Items())
}

// UnmarshalJSON consumes a slice of JSON bytes to populate this set.
func (cs ConcurrentSet[T]) UnmarshalJSON(b []byte) error {
	var tmp []T

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	cs.AddAll(tmp)

	return nil
}

// String returns a string representation of the current state of the set.
func (cs ConcurrentSet[T]) String() string {
	return fmt.Sprintf("%v", cs.Items())
}

// Difference returns all elements of this set that are not in `other`.
func (cs ConcurrentSet[T]) Difference(other ds.Set[T]) []T {
	var result []T

	cs.ForEach(
		func(element T) bool {
			if !other.Has(element) {
				result = append(result, element)
			}

			return false
		},
	)

	return result
}

// SymmetricDifference returns all elements that are in either this set or `other` but not in both.
func (cs ConcurrentSet[T]) SymmetricDifference(other ds.Set[T]) []T {
	var result []T

	cs.ForEach(
		func(element T) bool {
			if !other.Has(element) {
				result = append(result, element)
			}

			return false
		},
	)

	other.ForEach(
		func(element T) bool {
			if !cs.Has(element) {
				result = append(result, element)
			}

			return false
		},
	)

	return result
}

// Intersect returns all elements that exist in both sets.
func (cs ConcurrentSet[T]) Intersect(other ds.Set[T]) []T {
	var result []T

	// Optimization if `other` is a HashSet
	_, otherIsHashSet := other.(HashSet[T])

	if otherIsHashSet || cs.Count() < other.Count() {
		cs.ForEach(
			func(element T) bool {
				if other.Has(element) {
					result = append(result, element)
				}

				return false
			},
		)

		return result
	}

	other.ForEach(
		func(element T) bool {
			if cs.Has(element) {
				result = append(result, element)
			}

			return false
		},
	)

	return result
}

// Union returns all elements that are in both sets.
func (cs ConcurrentSet[T]) Union(other ds.Set[T]) []T {
	result := NewHashSetWithInitialSize[T](cs.Count() + other.Count())

	cs.ForEach(
		func(element T) bool {
			result.Add(element)

			return false
		},
	)

	other.ForEach(
		func(element T) bool {
			result.Add(element)

			return false
		},
	)

	return result.Items()
}

// IsProperSubset returns whether all elements in this set are in `other` but they are not equal.
func (cs ConcurrentSet[T]) IsProperSubset(other ds.Set[T]) bool {
	if cs.Count() >= other.Count() {
		return false
	}

	return other.Contains(cs)
}
