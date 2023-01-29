package ds

// Tuple represents a k-v pair.
type Tuple[K any, V any] struct {
	Key K
	Val V
}

// Map is a map for K -> V.
type Map[K comparable, V any] interface {
	// Set sets a new k-v pair in this map, then returns the previous value associated with
	// key, and whether such value exists.
	Set(key K, value V) (V, bool)
	// SetAll sets all k-v pairs from the given map in this map.
	SetAll(data map[K]V)
	// Get gets a value based on the given key.
	Get(key K) (V, bool)
	// GetAndSetIf gets a value based on the given key and sets a new value based on some condition,
	// returning all inset and outset params in the condition func.
	GetAndSetIf(key K, conditionFn func(currentVal V, found bool) (newVal V, shouldSet bool)) (currentVal V, found bool, newVal V, shouldSet bool)
	// GetElseCreate return the value associated with the given key and true. If the key doesn't
	// exist in this map, create a new value, set it in the map and then return the new value
	// together with false instead.
	GetElseCreate(key K, newValueFn func() V) (V, bool)
	// Count returns size of this map.
	Count() int
	// IsEmpty returns whether this map is empty.
	IsEmpty() bool
	// Has returns whether given key is in this map.
	Has(key K) bool
	// Remove pops a K-V pair from this map, then returns it.
	Remove(key K) (V, bool)
	// RemoveAll removes all given keys from this map, then returns whether this map changed as a
	// result of the call.
	RemoveAll(keys []K) bool
	// RemoveIf removes the given key from this map based on some condition, then returns the value
	// associated with the removed key and true. If the given key doesn't exist in this map or the
	// key was not removed because of the condition func, a zero-value and false will be returned.
	RemoveIf(key K, conditionFn func(currentVal V) bool) (V, bool)
	// Clear removes all k-v pairs from this map.
	Clear()
	// Iter returns a channel which could be used in a for range loop. The capacity of the returned
	// channel is the same as the size of the map at the time Iter() is called.
	Iter() <-chan Tuple[K, V]
	// Items returns all k-v pairs as a slice of core.Tuple.
	Items() []Tuple[K, V]
	// ForEach executes the given doEachFn on every element in this map. If `doEachFn` returns true,
	// stop execution immediately.
	ForEach(doEachFn func(key K, val V) bool)
	// String returns a string representation of the current state of this map.
	String() string
}
