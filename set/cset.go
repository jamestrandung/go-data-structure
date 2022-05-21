package set

import (
	"encoding/json"
	"fmt"

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

// AddAll adds all elements from the given slice to this set.
func (cs ConcurrentSet[T]) AddAll(data []T) {
	for _, element := range data {
		cs.cast().Set(element, struct{}{})
	}
}

// Add an element to the set, returns whether the element is new to the set.
func (cs ConcurrentSet[T]) Add(element T) bool {
	_, found := cs.cast().Set(element, struct{}{})
	return !found
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

// Remove removes an element from this set, returns whether such element exists.
func (cs ConcurrentSet[T]) Remove(element T) bool {
	_, found := cs.cast().Remove(element)
	return found
}

// Pop removes and returns an arbitrary element from this set.
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

// Clear removes all elements from this set.
func (cs ConcurrentSet[T]) Clear() {
	cs.cast().Clear()
}

// Difference returns all elements of this set that are not in `other`.
func (cs ConcurrentSet[T]) Difference(other Set[T]) []T {
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
func (cs ConcurrentSet[T]) SymmetricDifference(other Set[T]) []T {
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
func (cs ConcurrentSet[T]) Intersect(other Set[T]) []T {
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
func (cs ConcurrentSet[T]) Union(other Set[T]) []T {
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

// Equals returns whether this and `other` sets have the same size and contain the same elements.
func (cs ConcurrentSet[T]) Equals(other Set[T]) bool {
	if cs.Count() != other.Count() {
		return false
	}

	return other.Contains(cs)
}

// IsProperSubset returns whether all elements in this set are in `other` but they are not equal.
func (cs ConcurrentSet[T]) IsProperSubset(other Set[T]) bool {
	if cs.Count() >= other.Count() {
		return false
	}

	return other.Contains(cs)
}

// Contains returns whether all elements in `other` are in this set.
func (cs ConcurrentSet[T]) Contains(other Set[T]) bool {
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
