package cmap

import (
    "encoding/json"
    "fmt"
    "hash/fnv"
    "reflect"
    "sync"

    "github.com/mitchellh/hashstructure/v2"
)

var (
    defaultShardCount = 32
)

// Hasher can be implemented by the type to be used as keys in case clients want to provide
// their own hashing algo. Note that this hash does NOT control which bucket will hold a key
// in the actual map in each shard. Instead, it is used to distribute keys into the underlying
// shards which can be locked/unlocked independently by many Go routines.
//
// To guarantee good performance for ConcurrentMap, clients must make sure their hashing
// algo can evenly distribute keys across shards. Otherwise, a hot shard may become a
// bottleneck when many Go routines try to lock it at the same time.
//
// Note: we prioritize availability over correctness. If clients' hashing algo panics, the
// library will automatically recover and return 0, meaning all panicked keys will go into
// the 1st shard. As a consequence, subsequent concurrent accesses to these keys will happen
// on the same shard and performance may suffer. However, client applications will still work.
type Hasher interface {
    Hash() uint64
}

// HashString converts the given string into a hash value. If the hashing algo panics or
// throws an error, the library will also return an error to clients together with a zero
// hash. Clients can choose to use the zero hash and force all failing keys into the 1st
// shard. Subsequent concurrent accesses to these keys will happen on the same shard and
// performance may suffer. However, client applications will still work.
func HashString(s string) (result uint64, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("HashString - Panic encountered: %v", r)
        }
    }()

    hash := fnv.New64()

    if _, err = hash.Write([]byte(s)); err != nil {
        return 0, err
    }

    return hash.Sum64(), nil
}

type BasicKeyType interface {
    string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// ConcurrentMap is a thread safe map for K -> V.
//
// When K is one of the BasicKeyType, the library will convert keys into string and then hash it
// for distribution into the underlying shards. Otherwise, if K is NOT one of the BasicKeyType,
// the library will use hashstructure.Hash with hashstructure.FormatV2 to convert keys into a
// hash for distribution into the underlying shards. Clients can take a look at the documentation
// from hashstructure to understand how it works and how they can customize its behaviors at
// runtime (e.g. ignore some struct fields).
//
// Alternatively, clients can implement the Hasher interface to provide their own hashing algo.
// This is preferred for optimizing performance since clients can choose fields to hash based
// on the actual type of keys without relying on heavy reflections inside hashstructure.Hash.
type ConcurrentMap[K comparable, V any] []*ConcurrentMapShard[K, V]

// ConcurrentMapShard is 1 shard in a ConcurrentMap
type ConcurrentMapShard[K comparable, V any] struct {
    sync.RWMutex
    items map[K]V
}

// NewConcurrentMap returns a new instance of ConcurrentMap.
func NewConcurrentMap[K comparable, V any]() ConcurrentMap[K, V] {
    return NewConcurrentMapWithConcurrencyLevel[K, V](defaultShardCount)
}

// NewConcurrentMapWithConcurrencyLevel returns a new instance of ConcurrentMap with the given amount of shards.
func NewConcurrentMapWithConcurrencyLevel[K comparable, V any](concurrencyLevel int) ConcurrentMap[K, V] {
    m := make([]*ConcurrentMapShard[K, V], concurrencyLevel)
    for i := 0; i < concurrencyLevel; i++ {
        m[i] = &ConcurrentMapShard[K, V]{
            items: make(map[K]V),
        }
    }

    return m
}

func (cm ConcurrentMap[K, V]) getShard(key K) *ConcurrentMapShard[K, V] {
    tmp := any(key)

    switch castedKey := tmp.(type) {
    case string:
        hash, _ := HashString(castedKey)
        return cm[hash%uint64(len(cm))]
    case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
        hash, _ := HashString(fmt.Sprintf("%v", castedKey))
        return cm[hash%uint64(len(cm))]
    default:
    }

    if h, ok := any(key).(Hasher); ok {
        return cm[cm.hash(h)%uint64(len(cm))]
    }

    return cm[cm.hashAny(key)%uint64(len(cm))]
}

func (cm ConcurrentMap[K, V]) hash(key Hasher) uint64 {
    defer func() {
        // Fall back to 0 if panics
        if r := recover(); r != nil {
            fmt.Println("ConcurrentMap - Panic recovered:", r)
        }
    }()

    return key.Hash()
}

var hashFn = hashstructure.Hash

func (cm ConcurrentMap[K, V]) hashAny(key K) uint64 {
    defer func() {
        // Fall back to 0 if panics
        if r := recover(); r != nil {
            fmt.Println("ConcurrentMap - Panic recovered:", r)
        }
    }()

    hash, err := hashFn(key, hashstructure.FormatV2, &hashstructure.HashOptions{UseStringer: true})
    if err != nil {
        fmt.Println("ConcurrentMap - Error encountered:", err)

        // Use the 1st shard as fallback in case hashing fails
        return 0
    }

    return hash
}

// Set sets a new k-v pair in the map, then returns the previous value associated with key,
// and whether such value exists.
func (cm *ConcurrentMap[K, V]) Set(key K, value V) (V, bool) {
    shard := cm.getShard(key)
    shard.Lock()
    defer shard.Unlock()

    prevValue, ok := shard.items[key]

    shard.items[key] = value

    return prevValue, ok
}

// Get gets value based on given key in the map.
func (cm ConcurrentMap[K, V]) Get(key K) (V, bool) {
    shard := cm.getShard(key)
    shard.RLock()
    defer shard.RUnlock()

    val, ok := shard.items[key]
    return val, ok
}

// GetElseSet return the value associated with the given key and true. If the key doesn't
// exist in this map, set and return the new value together with false instead.
func (cm ConcurrentMap[K, V]) GetElseSet(key K, newValue V) (V, bool) {
    shard := cm.getShard(key)
    shard.Lock()
    defer shard.Unlock()

    if val, found := shard.items[key]; found {
        return val, true
    }

    shard.items[key] = newValue
    return newValue, false
}

// GetAndSetIf gets a value based on the given key and sets a new value based on some condition,
// returning all inset and outset params in the condition func.
func (cm ConcurrentMap[K, V]) GetAndSetIf(
    key K,
    condition func(currentVal V, found bool) (newVal V, shouldSet bool),
) (currentVal V, found bool, newVal V, shouldSet bool) {
    shard := cm.getShard(key)
    shard.Lock()
    defer shard.Unlock()

    currentVal, found = shard.items[key]
    newVal, shouldSet = condition(currentVal, found)
    if shouldSet {
        shard.items[key] = newVal
    }

    return
}

// GetElseCreate return the value associated with the given key and true. If the key doesn't
// exist in this map, create a new value, set it in the map and then return the new value
// together with false instead.
func (cm ConcurrentMap[K, V]) GetElseCreate(key K, newValueFactory func() V) (V, bool) {
    shard := cm.getShard(key)
    shard.Lock()
    defer shard.Unlock()

    if val, found := shard.items[key]; found {
        return val, true
    }

    newValue := newValueFactory()
    shard.items[key] = newValue
    return newValue, false
}

// Count returns the count of all items in this map.
func (cm ConcurrentMap[K, V]) Count() int {
    count := 0
    for _, shard := range cm {
        shard.RLock()

        count += len(shard.items)

        shard.RUnlock()
    }

    return count
}

// Has returns true if the given key is in this map, else false.
func (cm ConcurrentMap[K, V]) Has(key K) bool {
    shard := cm.getShard(key)
    shard.RLock()
    defer shard.RUnlock()

    _, ok := shard.items[key]
    return ok
}

// IsEmpty returns true if this map is empty, else false.
func (cm ConcurrentMap[K, V]) IsEmpty() bool {
    return cm.Count() == 0
}

// Remove removes the given key from this map.
func (cm ConcurrentMap[K, V]) Remove(key K) {
    shard := cm.getShard(key)
    shard.Lock()
    defer shard.Unlock()

    delete(shard.items, key)
}

// RemoveIfValue removes key ONLY when value matches.
func (cm ConcurrentMap[K, V]) RemoveIfValue(key K, value V) (ret bool, err error) {
    shard := cm.getShard(key)
    shard.Lock()
    defer shard.Unlock()

    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("RemoveIfValue - Panic encountered: %v", r)
        }
    }()

    if val, ok := shard.items[key]; ok {
        if reflect.DeepEqual(val, value) {
            delete(shard.items, key)
            ret = true
        }
    }

    return
}

// Tuple used by the Iter function to wrap key and value in a pair.
type Tuple[K comparable, V any] struct {
    Key K
    Val V
}

// Iter returns an iterator which could be used in a for range loop.
func (cm ConcurrentMap[K, V]) Iter() []Tuple[K, V] {
    result := make([]Tuple[K, V], 0, cm.Count())

    for _, shard := range cm {
        shard.RLock()

        for key, val := range shard.items {
            result = append(result, Tuple[K, V]{key, val})
        }

        shard.RUnlock()
    }

    return result
}

// ForEach executes the given doEachFn on every k-v pair in this map 1-by-1.
func (cm ConcurrentMap[K, V]) ForEach(doEachFn func(key K, val V)) {
    for _, shard := range cm {
        func() {
            // Execute in a func with defer so that if doEachFn panics,
            // we will still unlock the shard properly
            // https://stackoverflow.com/questions/54291236/how-to-wait-for-a-panicking-goroutine
            shard.RLock()
            defer shard.RUnlock()

            for key, val := range shard.items {
                doEachFn(key, val)
            }
        }()
    }
}

// MarshalJSON returns the JSON bytes of this map.
func (cm ConcurrentMap[K, V]) MarshalJSON() ([]byte, error) {
    // Create a temporary map, which will hold all item spread across shards.
    tmp := make(map[K]V)

    for _, item := range cm.Iter() {
        tmp[item.Key] = item.Val
    }

    return json.Marshal(tmp)
}
