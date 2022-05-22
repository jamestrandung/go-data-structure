# Enhanced Map

## Usage

```go
import (
"github.com/jamestrandung/go-data-structure/emap"
)

```

```bash
go get "github.com/jamestrandung/go-data-structure/emap"
```

## Implemented interface

```go
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
    // Count returns size of this map.
    Count() int
    // IsEmpty returns whether this map is empty.
    IsEmpty() bool
    // Has returns whether given key is in this map.
    Has(key K) bool
    // Remove pops a K-V pair from this map, then returns it.
    Remove(key K) (V, bool)
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
    // MarshalJSON returns the JSON bytes of this map.
    MarshalJSON() ([]byte, error)
    // UnmarshalJSON consumes a slice of JSON bytes to populate this map.
    UnmarshalJSON(b []byte) error
    // String returns a string representation of the current state of this map.
    String() string
}
```

## Concurrent Map

The basic `map` type in Go does not support concurrent reads and writes. `emap` provides a high
performance solution by utilizing shards to minimize the time required to acquire a lock to read
or write. It is suitable for cases where many Go routines read & write over the same set of keys.

Do note that since Go 1.9, there's a built-in `sync.Map` that is optimized for 2 common use cases:
(1) when the entry for a given key is only ever written once but read many times, as in caches that
only grow, or (2) when multiple Go routines read, write, and overwrite entries for disjoint sets of
keys. In these cases, using `sync.Map` may significantly reduce lock content compared to `emap`.

Read more: https://pkg.go.dev/sync#Map

### Extra methods

```go
// AsMap returns all k-v pairs as map[K]V.
AsMap() map[K]V
```

### Example

```go
// Create a new map
cm := emap.NewConcurrentMap[string, string]()

// Set item in map
cm.Set("foo", "bar")

// Get item from map without any casting required
if bar, ok := cm.Get("foo"); ok {
    // Do something with "bar"
}

// Remove item under key "foo"
cm.Remove("foo")
```

For more examples, have a look at `cmap_test.go`.