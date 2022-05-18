# Concurrent Map

The basic `map` type in Go does not support concurrent reads and writes. `cmap` provides a high
performance solution by utilizing shards to minimize the time required to acquire a lock to read
or write. It is suitable for cases where many Go routines read & write over the same set of keys.

Do note that since Go 1.9, there's a built-in `sync.Map` that is optimized for 2 common use cases:
(1) when the entry for a given key is only ever written once but read many times, as in caches that 
only grow, or (2) when multiple Go routines read, write, and overwrite entries for disjoint sets of 
keys. In these cases, using `sync.Map` may significantly reduce lock content compared to `cmap`.

Read more: https://pkg.go.dev/sync#Map

## Usage

```go
import (
    "github.com/jamestrandung/go-data-structure/cmap"
)

```

```bash
go get "github.com/jamestrandung/go-data-structure/cmap"
```

## Example

```go
// Create a new map
cm := cmap.New[string, string]()

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