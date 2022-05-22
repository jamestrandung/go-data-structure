package ds

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
