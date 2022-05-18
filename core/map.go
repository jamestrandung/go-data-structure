package core

// Tuple represents a k-v pair.
type Tuple[K comparable, V any] struct {
	Key K
	Val V
}

// Map is a map for K -> V.
type Map[K comparable, V any] interface {
	// SetAll sets all k-v pairs from the given map in this map.
	SetAll(data map[K]V)
	// Set sets a new k-v pair in this map, then returns the previous value associated with
	// key, and whether such value exists.
	Set(key K, value V) (V, bool)
	// Get gets a value based on the given key.
	Get(key K) (V, bool)
	// GetAndSetIf gets a value based on the given key and sets a new value based on some condition,
	// returning all inset and outset params in the condition func.
	GetAndSetIf(key K, conditionFn func(currentVal V, found bool) (newVal V, shouldSet bool)) (currentVal V, found bool, newVal V, shouldSet bool)
	// GetElseCreate return the value associated with the given key and true. If the key doesn't
	// exist in this map, create a new value, set it in the map and then return the new value
	// together with false instead.
	GetElseCreate(key K, newValueFn func() V) (V, bool)
	// Count returns the count of all items in this map.
	Count() int
	// Has returns true if the given key is in this map, else false.
	Has(key K) bool
	// IsEmpty returns true if this map is empty, else false.
	IsEmpty() bool
	// Remove pops a K-V pair from this map, then returns it.
	Remove(key K) (V, bool)
	// RemoveIf removes the given key from this map based on some condition, then returns the value
	// associated with the removed key and true. If the given key doesn't exist in this map or the
	// key was not removed because of the condition func, a zero-value and false will be returned.
	RemoveIf(key K, conditionFn func(currentVal V, found bool) bool) (V, bool)
	// Clear removes all k-v pairs from this map.
	Clear()
	// Iter returns an iterator which could be used in a for range loop. The capacity of the returned
	// channel is the same as the size of the map at the time Iter() is called.
	Iter() <-chan Tuple[K, V]
	// Items returns all k-v pairs as map[K]V.
	Items() map[K]V
	// ForEach executes the given doEachFn on every k-v pair in this map shard by shard.
	ForEach(doEachFn func(key K, val V))
	// MarshalJSON returns the JSON bytes of this map.
	MarshalJSON() ([]byte, error)
	// UnmarshalJSON consumes a slice of JSON bytes to populate this map.
	UnmarshalJSON(b []byte) error
}
