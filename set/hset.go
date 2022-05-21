package set

import (
	"encoding/json"
	"fmt"
)

type HashSet[T comparable] map[T]struct{}

// NewHashSet returns a new instance of HashSet with the given elements.
func NewHashSet[T comparable](elements ...T) HashSet[T] {
	if len(elements) == 0 {
		return make(map[T]struct{})
	}

	var result HashSet[T] = make(map[T]struct{}, len(elements))
	result.AddAll(elements)

	return result
}

// NewHashSetWithInitialSize returns a new instance of HashSet with the given initial size.
func NewHashSetWithInitialSize[T comparable](initialSize int) HashSet[T] {
	return make(map[T]struct{}, initialSize)
}

// AddAll adds all elements from the given slice to this set.
func (hs HashSet[T]) AddAll(data []T) {
	for _, value := range data {
		hs[value] = struct{}{}
	}
}

// Add an element to the set, returns whether the element is new to the set.
func (hs HashSet[T]) Add(element T) bool {
	curSize := hs.Count()
	hs[element] = struct{}{}
	return hs.Count() > curSize
}

// Count returns the size of this set.
func (hs HashSet[T]) Count() int {
	return len(hs)
}

// IsEmpty returns whether this set is empty.
func (hs HashSet[T]) IsEmpty() bool {
	return hs.Count() == 0
}

// Has returns whether given element is in this set.
func (hs HashSet[T]) Has(element T) bool {
	_, ok := hs[element]
	return ok
}

// Remove removes an element from this set, returns whether such element exists.
func (hs HashSet[T]) Remove(element T) bool {
	if _, ok := hs[element]; ok {
		delete(hs, element)
		return true
	}

	return false
}

// Pop removes and returns an arbitrary element from this set.
func (hs HashSet[T]) Pop() (v T, ok bool) {
	for element := range hs {
		delete(hs, element)
		return element, true
	}

	return
}

// Clear removes all elements from this set.
func (hs HashSet[T]) Clear() {
	for element := range hs {
		delete(hs, element)
	}
}

// Difference returns all elements of this set that are not in `other`.
func (hs HashSet[T]) Difference(other Set[T]) []T {
	var result []T

	for element := range hs {
		if !other.Has(element) {
			result = append(result, element)
		}
	}

	return result
}

// SymmetricDifference returns all elements that are in either this set or `other` but not in both.
func (hs HashSet[T]) SymmetricDifference(other Set[T]) []T {
	var result []T

	for element := range hs {
		if !other.Has(element) {
			result = append(result, element)
		}
	}

	other.ForEach(
		func(element T) bool {
			if !hs.Has(element) {
				result = append(result, element)
			}

			return false
		},
	)

	return result
}

// Intersect returns all elements that exist in both sets.
func (hs HashSet[T]) Intersect(other Set[T]) []T {
	var result []T

	// Optimization if `other` is a ConcurrentSet
	_, otherIsConcurrentSet := other.(ConcurrentSet[T])

	if !otherIsConcurrentSet && hs.Count() < other.Count() {
		for element := range hs {
			if other.Has(element) {
				result = append(result, element)
			}
		}

		return result
	}

	other.ForEach(
		func(element T) bool {
			if hs.Has(element) {
				result = append(result, element)
			}

			return false
		},
	)

	return result
}

// Union returns all elements that are in both sets.
func (hs HashSet[T]) Union(other Set[T]) []T {
	result := NewHashSetWithInitialSize[T](hs.Count() + other.Count())

	for element := range hs {
		result.Add(element)
	}

	other.ForEach(
		func(element T) bool {
			result.Add(element)

			return false
		},
	)

	return result.Items()
}

// Equals returns whether this and `other` sets have the same size and contain the same elements.
func (hs HashSet[T]) Equals(other Set[T]) bool {
	if hs.Count() != other.Count() {
		return false
	}

	return hs.Contains(other)
}

// IsProperSubset returns whether all elements in this set are in `other` but they are not equal.
func (hs HashSet[T]) IsProperSubset(other Set[T]) bool {
	if hs.Count() >= other.Count() {
		return false
	}

	return other.Contains(hs)
}

// Contains returns whether all elements in `other` are in this set.
func (hs HashSet[T]) Contains(other Set[T]) bool {
	isSuperset := true
	other.ForEach(
		func(element T) bool {
			if !hs.Has(element) {
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
func (hs HashSet[T]) Iter() <-chan T {
	out := make(chan T, hs.Count())

	go func() {
		defer close(out)

		for element := range hs {
			out <- element
		}
	}()

	return out
}

// Items returns all elements of this set as a slice.
func (hs HashSet[T]) Items() []T {
	result := make([]T, 0, hs.Count())

	for element := range hs {
		result = append(result, element)
	}

	return result
}

// ForEach executes the given doEachFn on every element in this set. If `doEachFn` returns true,
// stop execution immediately.
func (hs HashSet[T]) ForEach(doEachFn func(element T) bool) {
	for element := range hs {
		if doEachFn(element) {
			break
		}
	}
}

// MarshalJSON returns the JSON bytes of this set.
func (hs HashSet[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(hs.Items())
}

// UnmarshalJSON consumes a slice of JSON bytes to populate this set.
func (hs HashSet[T]) UnmarshalJSON(b []byte) error {
	var tmp []T

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	hs.AddAll(tmp)

	return nil
}

// String returns a string representation of the current state of the set.
func (hs HashSet[T]) String() string {
	return fmt.Sprintf("%v", hs.Items())
}
