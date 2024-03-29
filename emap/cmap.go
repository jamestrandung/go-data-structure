package emap

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jamestrandung/go-data-structure/ds"
)

var (
	defaultShardCount = 32
)

type BasicKeyType interface {
	string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// ConcurrentMap is a thread safe map for K -> V. To avoid lock bottlenecks this map is divided
// into to several ConcurrentMapShard.
//
// To distribute keys into the underlying shards, when K is one of the BasicKeyType, the library
// will convert provided keys into string and then hash it using Go built-in fnv.New64(). On the
// other hand, if K is NOT one of the BasicKeyType, the library will use hashstructure.Hash with
// hashstructure.FormatV2 to hash provided keys. Clients should take a look at the documentation
// from hashstructure to understand how it works and how to customize its behaviors at runtime
// (e.g. ignore some struct fields).
//
// Alternatively, clients can implement the Hasher core to provide their own hashing algo.
// This is preferred for optimizing performance since clients can choose fields to hash based
// on the actual type of keys without relying on heavy reflections inside hashstructure.Hash.
type ConcurrentMap[K comparable, V any] []*ConcurrentMapShard[K, V]

// ConcurrentMapShard is 1 shard in a ConcurrentMap
type ConcurrentMapShard[K comparable, V any] struct {
	sync.RWMutex
	items map[K]V
}

// UnlockFn unlocks a shard
type UnlockFn func()

// GetItemsToRead returns the items in this shard and an UnlockFn after getting RLock
func (s *ConcurrentMapShard[K, V]) GetItemsToRead() (map[K]V, UnlockFn) {
	s.RLock()

	return s.items, func() {
		s.RUnlock()
	}
}

// GetItemsToWrite returns the items in this shard and an UnlockFn after getting Lock
func (s *ConcurrentMapShard[K, V]) GetItemsToWrite() (map[K]V, UnlockFn) {
	s.Lock()

	return s.items, func() {
		s.Unlock()
	}
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
		hash := HashString(castedKey)
		return cm[hash%uint32(len(cm))]
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		hash := HashString(fmt.Sprintf("%v", castedKey))
		return cm[hash%uint32(len(cm))]
	default:
	}

	if h, ok := tmp.(Hasher); ok {
		return cm[hashHasher(h)%uint64(len(cm))]
	}

	return cm[hashAny(key)%uint64(len(cm))]
}

// Set sets a new k-v pair in this map, then returns the previous value associated with
// key, and whether such value exists.
func (cm ConcurrentMap[K, V]) Set(key K, value V) (V, bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	prevValue, ok := shard.items[key]

	shard.items[key] = value

	return prevValue, ok
}

// SetAll sets all k-v pairs from the given map in this map.
func (cm ConcurrentMap[K, V]) SetAll(data map[K]V) {
	for key, value := range data {
		shard := cm.getShard(key)
		shard.Lock()

		shard.items[key] = value

		shard.Unlock()
	}
}

// Get gets a value based on the given key.
func (cm ConcurrentMap[K, V]) Get(key K) (V, bool) {
	shard := cm.getShard(key)
	shard.RLock()
	defer shard.RUnlock()

	val, ok := shard.items[key]
	return val, ok
}

// SetIfAbsent sets a new k-v pair in this map if it doesn't contain this key, then returns
// whether such k-v pair was absent.
func (cm ConcurrentMap[K, V]) SetIfAbsent(key K, value V) bool {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	if _, found := shard.items[key]; found {
		return false
	}

	shard.items[key] = value

	return true
}

// GetAndSetIf gets a value based on the given key and sets a new value based on some condition,
// returning all inset and outset params in the condition func.
//
// Note: Condition func is called while lock is held. Hence, it must NOT access this map as it
// may lead to deadlock since sync.RWLock is not reentrant.
func (cm ConcurrentMap[K, V]) GetAndSetIf(
	key K,
	conditionFn func(currentVal V, found bool) (newVal V, shouldSet bool),
) (currentVal V, found bool, newVal V, shouldSet bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	currentVal, found = shard.items[key]
	newVal, shouldSet = conditionFn(currentVal, found)
	if shouldSet {
		shard.items[key] = newVal
	}

	return
}

// GetElseCreate return the value associated with the given key and true. If the key doesn't
// exist in this map, create a new value, set it in the map and then return the new value
// together with false instead.
//
// Note: newValueFn is called while lock is held. Hence, it must NOT access this map as it may
// lead to deadlock since sync.RWLock is not reentrant.
func (cm ConcurrentMap[K, V]) GetElseCreate(key K, newValueFn func() V) (V, bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	if val, found := shard.items[key]; found {
		return val, true
	}

	newValue := newValueFn()
	shard.items[key] = newValue
	return newValue, false
}

// Count returns size of this map.
func (cm ConcurrentMap[K, V]) Count() int {
	count := 0
	for _, shard := range cm {
		shard.RLock()

		count += len(shard.items)

		shard.RUnlock()
	}

	return count
}

// IsEmpty returns whether this map is empty
func (cm ConcurrentMap[K, V]) IsEmpty() bool {
	return cm.Count() == 0
}

// Has returns whether given key is in this map.
func (cm ConcurrentMap[K, V]) Has(key K) bool {
	shard := cm.getShard(key)
	shard.RLock()
	defer shard.RUnlock()

	_, ok := shard.items[key]
	return ok
}

// Remove pops a K-V pair from this map, then returns it.
func (cm ConcurrentMap[K, V]) Remove(key K) (V, bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	val, ok := shard.items[key]
	if ok {
		delete(shard.items, key)
	}

	return val, ok
}

// RemoveAll removes all given keys from this map, then returns whether this map changed as a
// result of the call.
func (cm ConcurrentMap[K, V]) RemoveAll(keys []K) bool {
	isMapChanged := false

	var wg sync.WaitGroup
	wg.Add(len(cm))

	for _, shard := range cm {
		go func(s *ConcurrentMapShard[K, V]) {
			s.Lock()
			defer wg.Done()
			defer s.Unlock()

			for _, key := range keys {
				_, ok := s.items[key]
				if ok {
					delete(s.items, key)
					isMapChanged = true
				}
			}
		}(shard)
	}

	wg.Wait()

	return isMapChanged
}

// RemoveIf removes the given key from this map based on some condition, then returns the value
// associated with the removed key and true. If the given key doesn't exist in this map or the
// key was not removed because of the condition func, a zero-value and false will be returned.
//
// Note: Condition func is called while lock is held. Hence, it must NOT access this map as it
// may lead to deadlock since sync.RWLock is not reentrant.
func (cm ConcurrentMap[K, V]) RemoveIf(key K, conditionFn func(currentVal V) bool) (V, bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	val, ok := shard.items[key]
	if ok && conditionFn(val) {
		delete(shard.items, key)
		return val, true
	}

	var tmp V
	return tmp, false
}

// Clear removes all k-v pairs from this map.
func (cm ConcurrentMap[K, V]) Clear() {
	var wg sync.WaitGroup
	wg.Add(len(cm))

	for _, shard := range cm {
		go func(s *ConcurrentMapShard[K, V]) {
			s.Lock()
			defer wg.Done()
			defer s.Unlock()

			for key := range s.items {
				delete(s.items, key)
			}
		}(shard)
	}

	wg.Wait()
}

// takeSnapshot returns an array of channels that contains all k-v pairs in each shard, which is
// likely a snapshot of this map. It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines.
func (cm ConcurrentMap[K, V]) takeSnapshot() []<-chan ds.Tuple[K, V] {
	// When this map is not initialized
	if len(cm) == 0 {
		panic(`emap.ConcurrentMap is not initialized. Should run NewConcurrentMap() before usage.`)
	}

	shardCount := len(cm)

	var wg sync.WaitGroup
	wg.Add(shardCount)

	result := make([]<-chan ds.Tuple[K, V], shardCount)

	for idx, shard := range cm {
		go func(i int, s *ConcurrentMapShard[K, V]) {
			s.RLock()
			defer s.RUnlock()

			// A buffered channel that is big enough to hold all k-v pairs in this
			// shard will prevent this goroutine from being blocked indefinitely
			// in case reader crashes
			channel := make(chan ds.Tuple[K, V], len(s.items))
			defer close(channel)

			result[i] = channel

			// Done early before populating channel to let the main routine proceeds
			wg.Done()

			for key, val := range s.items {
				channel <- ds.Tuple[K, V]{key, val}
			}
		}(idx, shard)
	}

	wg.Wait()

	return result
}

// fanInBufferedChannels reads elements from channels `bufferedIns` into channel `out`.
func fanInBufferedChannels[K comparable, V any](bufferedIns []<-chan ds.Tuple[K, V]) <-chan ds.Tuple[K, V] {
	totalBufferSize := 0
	for _, c := range bufferedIns {
		totalBufferSize += cap(c)
	}

	out := make(chan ds.Tuple[K, V], totalBufferSize)

	go func() {
		defer close(out)

		wg := sync.WaitGroup{}
		wg.Add(len(bufferedIns))

		for _, c := range bufferedIns {
			go func(channel <-chan ds.Tuple[K, V]) {
				defer wg.Done()

				for t := range channel {
					out <- t
				}
			}(c)
		}

		wg.Wait()
	}()

	return out
}

// Iter returns a channel which could be used in a for range loop. The capacity of the returned
// channel is the same as the size of the map at the time Iter() is called.
func (cm ConcurrentMap[K, V]) Iter() <-chan ds.Tuple[K, V] {
	return fanInBufferedChannels[K, V](cm.takeSnapshot())
}

// Items returns all k-v pairs as a slice of core.Tuple.
func (cm ConcurrentMap[K, V]) Items() []ds.Tuple[K, V] {
	var result []ds.Tuple[K, V]

	for _, shard := range cm {
		shard.RLock()

		for key, val := range shard.items {
			result = append(result, ds.Tuple[K, V]{key, val})
		}

		shard.RUnlock()
	}

	return result
}

// ForEach executes the given doEachFn on every element in this map. If `doEachFn` returns true,
// stop execution immediately.
//
// Note: doEachFn is called while lock is held. Hence, it must NOT access this map as it may
// lead to deadlock since sync.RWLock is not reentrant.
func (cm ConcurrentMap[K, V]) ForEach(doEachFn func(key K, val V) bool) {
	for _, shard := range cm {
		mustStop := func() bool {
			// Execute in a func with defer so that if doEachFn panics,
			// we will still unlock the shard properly
			// https://stackoverflow.com/questions/54291236/how-to-wait-for-a-panicking-goroutine
			shard.RLock()
			defer shard.RUnlock()

			for key, val := range shard.items {
				if doEachFn(key, val) {
					return true
				}
			}

			return false
		}()

		if mustStop {
			return
		}
	}
}

// AsMap returns all k-v pairs as map[K]V.
func (cm ConcurrentMap[K, V]) AsMap() map[K]V {
	result := make(map[K]V)

	for _, shard := range cm {
		shard.RLock()

		for key, val := range shard.items {
			result[key] = val
		}

		shard.RUnlock()
	}

	return result
}

// MarshalJSON returns the JSON bytes of this map.
func (cm ConcurrentMap[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(cm.AsMap())
}

// UnmarshalJSON consumes a slice of JSON bytes to populate this map.
func (cm ConcurrentMap[K, V]) UnmarshalJSON(b []byte) error {
	var tmp map[K]V

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	cm.SetAll(tmp)

	return nil
}

// String returns a string representation of the current state of this map.
func (cm ConcurrentMap[K, V]) String() string {
	return fmt.Sprintf("%v", cm.AsMap())
}
