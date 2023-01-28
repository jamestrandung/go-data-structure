package emap

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jamestrandung/go-data-structure/ds"
)

// ConcurrentMapUnsafeKey ...
type ConcurrentMapUnsafeKey[V any] []*ConcurrentMapShardUnsafeKey[V]

// ConcurrentMapShardUnsafeKey is 1 shard in a ConcurrentMapUnsafeKey
type ConcurrentMapShardUnsafeKey[V any] struct {
	sync.RWMutex
	items map[interface{}]V
}

// GetItemsToRead returns the items in this shard and an UnlockFn after getting RLock
func (s *ConcurrentMapShardUnsafeKey[V]) GetItemsToRead() (map[interface{}]V, UnlockFn) {
	s.RLock()

	return s.items, func() {
		s.RUnlock()
	}
}

// GetItemsToWrite returns the items in this shard and an UnlockFn after getting Lock
func (s *ConcurrentMapShardUnsafeKey[V]) GetItemsToWrite() (map[interface{}]V, UnlockFn) {
	s.Lock()

	return s.items, func() {
		s.Unlock()
	}
}

// NewConcurrentMapUnsafeKey returns a new instance of ConcurrentMapUnsafeKey.
func NewConcurrentMapUnsafeKey[V any]() ConcurrentMapUnsafeKey[V] {
	return NewNewConcurrentMapUnsafeKeyWithConcurrencyLevel[V](defaultShardCount)
}

// NewNewConcurrentMapUnsafeKeyWithConcurrencyLevel returns a new instance of ConcurrentMapUnsafeKey with the given amount of shards.
func NewNewConcurrentMapUnsafeKeyWithConcurrencyLevel[V any](concurrencyLevel int) ConcurrentMapUnsafeKey[V] {
	m := make([]*ConcurrentMapShardUnsafeKey[V], concurrencyLevel)
	for i := 0; i < concurrencyLevel; i++ {
		m[i] = &ConcurrentMapShardUnsafeKey[V]{
			items: make(map[interface{}]V),
		}
	}

	return m
}

func (cm ConcurrentMapUnsafeKey[V]) getShard(key interface{}) *ConcurrentMapShardUnsafeKey[V] {
	switch castedKey := key.(type) {
	case string:
		hash := HashString(castedKey)
		return cm[hash%uint32(len(cm))]
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		hash := HashString(fmt.Sprintf("%v", castedKey))
		return cm[hash%uint32(len(cm))]
	default:
	}

	if h, ok := key.(Hasher); ok {
		return cm[hashHasher(h)%uint64(len(cm))]
	}

	return cm[hashAny(key)%uint64(len(cm))]
}

// Set sets a new k-v pair in this map, then returns the previous value associated with
// key, and whether such value exists.
func (cm ConcurrentMapUnsafeKey[V]) Set(key interface{}, value V) (V, bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	prevValue, ok := shard.items[key]

	shard.items[key] = value

	return prevValue, ok
}

// SetAll sets all k-v pairs from the given map in this map.
func (cm ConcurrentMapUnsafeKey[V]) SetAll(data map[interface{}]V) {
	for key, value := range data {
		shard := cm.getShard(key)
		shard.Lock()

		shard.items[key] = value

		shard.Unlock()
	}
}

// Get gets a value based on the given key.
func (cm ConcurrentMapUnsafeKey[V]) Get(key interface{}) (V, bool) {
	shard := cm.getShard(key)
	shard.RLock()
	defer shard.RUnlock()

	val, ok := shard.items[key]
	return val, ok
}

// SetIfAbsent sets a new k-v pair in this map if it doesn't contain this key, then returns
// whether such k-v pair was absent.
func (cm ConcurrentMapUnsafeKey[V]) SetIfAbsent(key interface{}, value V) bool {
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
func (cm ConcurrentMapUnsafeKey[V]) GetAndSetIf(
	key interface{},
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
func (cm ConcurrentMapUnsafeKey[V]) GetElseCreate(key interface{}, newValueFn func() V) (V, bool) {
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
func (cm ConcurrentMapUnsafeKey[V]) Count() int {
	count := 0
	for _, shard := range cm {
		shard.RLock()

		count += len(shard.items)

		shard.RUnlock()
	}

	return count
}

// IsEmpty returns whether this map is empty
func (cm ConcurrentMapUnsafeKey[V]) IsEmpty() bool {
	return cm.Count() == 0
}

// Has returns whether given key is in this map.
func (cm ConcurrentMapUnsafeKey[V]) Has(key interface{}) bool {
	shard := cm.getShard(key)
	shard.RLock()
	defer shard.RUnlock()

	_, ok := shard.items[key]
	return ok
}

// Remove pops a K-V pair from this map, then returns it.
func (cm ConcurrentMapUnsafeKey[V]) Remove(key interface{}) (V, bool) {
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
func (cm ConcurrentMapUnsafeKey[V]) RemoveAll(keys []interface{}) bool {
	isMapChanged := false

	var wg sync.WaitGroup
	wg.Add(len(cm))

	for _, shard := range cm {
		go func(s *ConcurrentMapShardUnsafeKey[V]) {
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
func (cm ConcurrentMapUnsafeKey[V]) RemoveIf(key interface{}, conditionFn func(currentVal V) bool) (V, bool) {
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
func (cm ConcurrentMapUnsafeKey[V]) Clear() {
	var wg sync.WaitGroup
	wg.Add(len(cm))

	for _, shard := range cm {
		go func(s *ConcurrentMapShardUnsafeKey[V]) {
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
func (cm ConcurrentMapUnsafeKey[V]) takeSnapshot() []<-chan ds.Tuple[interface{}, V] {
	// When this map is not initialized
	if len(cm) == 0 {
		panic(`emap.ConcurrentMap is not initialized. Should run NewConcurrentMap() before usage.`)
	}

	shardCount := len(cm)

	var wg sync.WaitGroup
	wg.Add(shardCount)

	result := make([]<-chan ds.Tuple[interface{}, V], shardCount)

	for idx, shard := range cm {
		go func(i int, s *ConcurrentMapShardUnsafeKey[V]) {
			s.RLock()
			defer s.RUnlock()

			// A buffered channel that is big enough to hold all k-v pairs in this
			// shard will prevent this goroutine from being blocked indefinitely
			// in case reader crashes
			channel := make(chan ds.Tuple[interface{}, V], len(s.items))
			defer close(channel)

			result[i] = channel

			// Done early before populating channel to let the main routine proceeds
			wg.Done()

			for key, val := range s.items {
				channel <- ds.Tuple[interface{}, V]{key, val}
			}
		}(idx, shard)
	}

	wg.Wait()

	return result
}

// fanInBufferedUnsafeChannels reads elements from channels `bufferedIns` into channel `out`.
func fanInBufferedUnsafeChannels[V any](bufferedIns []<-chan ds.Tuple[interface{}, V]) <-chan ds.Tuple[interface{}, V] {
	totalBufferSize := 0
	for _, c := range bufferedIns {
		totalBufferSize += cap(c)
	}

	out := make(chan ds.Tuple[interface{}, V], totalBufferSize)

	go func() {
		defer close(out)

		wg := sync.WaitGroup{}
		wg.Add(len(bufferedIns))

		for _, c := range bufferedIns {
			go func(channel <-chan ds.Tuple[interface{}, V]) {
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
func (cm ConcurrentMapUnsafeKey[V]) Iter() <-chan ds.Tuple[interface{}, V] {
	return fanInBufferedUnsafeChannels[V](cm.takeSnapshot())
}

// Items returns all k-v pairs as a slice of core.Tuple.
func (cm ConcurrentMapUnsafeKey[V]) Items() []ds.Tuple[interface{}, V] {
	var result []ds.Tuple[interface{}, V]

	for _, shard := range cm {
		shard.RLock()

		for key, val := range shard.items {
			result = append(result, ds.Tuple[interface{}, V]{key, val})
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
func (cm ConcurrentMapUnsafeKey[V]) ForEach(doEachFn func(key interface{}, val V) bool) {
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
func (cm ConcurrentMapUnsafeKey[V]) AsMap() map[interface{}]V {
	result := make(map[interface{}]V)

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
func (cm ConcurrentMapUnsafeKey[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(cm.AsMap())
}

// UnmarshalJSON consumes a slice of JSON bytes to populate this map.
func (cm ConcurrentMapUnsafeKey[V]) UnmarshalJSON(b []byte) error {
	var tmp map[interface{}]V

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	cm.SetAll(tmp)

	return nil
}

// String returns a string representation of the current state of this map.
func (cm ConcurrentMapUnsafeKey[V]) String() string {
	return fmt.Sprintf("%v", cm.AsMap())
}
