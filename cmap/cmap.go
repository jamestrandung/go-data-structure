package cmap

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jamestrandung/go-data-structure/core"
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

// HashString converts the given string into a hash value.
func HashString(s string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)

	keyLength := len(s)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(s[i])
	}

	return hash
}

type BasicKeyType interface {
	string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

// ConcurrentMap is a thread safe map for K -> V. To avoid lock bottlenecks this map is divided
// into to several concurrentMapShard.
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
type ConcurrentMap[K comparable, V any] []*concurrentMapShard[K, V]

// concurrentMapShard is 1 shard in a ConcurrentMap
type concurrentMapShard[K comparable, V any] struct {
	sync.RWMutex
	items map[K]V
}

// New returns a new instance of ConcurrentMap.
func New[K comparable, V any]() ConcurrentMap[K, V] {
	return NewWithConcurrencyLevel[K, V](defaultShardCount)
}

// NewWithConcurrencyLevel returns a new instance of ConcurrentMap with the given amount of shards.
func NewWithConcurrencyLevel[K comparable, V any](concurrencyLevel int) ConcurrentMap[K, V] {
	m := make([]*concurrentMapShard[K, V], concurrencyLevel)
	for i := 0; i < concurrencyLevel; i++ {
		m[i] = &concurrentMapShard[K, V]{
			items: make(map[K]V),
		}
	}

	return m
}

func (cm ConcurrentMap[K, V]) getShard(key K) *concurrentMapShard[K, V] {
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

// SetAll sets all k-v pairs from the given map in this map.
func (cm ConcurrentMap[K, V]) SetAll(data map[K]V) {
	for key, value := range data {
		shard := cm.getShard(key)
		shard.Lock()

		shard.items[key] = value

		shard.Unlock()
	}
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

// RemoveIf removes the given key from this map based on some condition, then returns the value
// associated with the removed key and true. If the given key doesn't exist in this map or the
// key was not removed because of the condition func, a zero-value and false will be returned.
//
// Note: Condition func is called while lock is held. Hence, it must NOT access this map as it
// may lead to deadlock since sync.RWLock is not reentrant.
func (cm ConcurrentMap[K, V]) RemoveIf(key K, conditionFn func(currentVal V, found bool) bool) (V, bool) {
	shard := cm.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	val, ok := shard.items[key]
	if ok && conditionFn(val, ok) {
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
		go func(s *concurrentMapShard[K, V]) {
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

// Returns an array of channels that contains all k-v pairs in each shard, which is likely
// a snapshot of this map. It returns once the size of each buffered channel is determined,
// before all the channels are populated using Go routines.
func (cm ConcurrentMap[K, V]) takeSnapshot() []<-chan core.Tuple[K, V] {
	// When this map is not initialized
	if len(cm) == 0 {
		panic(`cmap.ConcurrentMap is not initialized. Should run New() before usage.`)
	}

	shardCount := len(cm)

	var wg sync.WaitGroup
	wg.Add(shardCount)

	result := make([]<-chan core.Tuple[K, V], shardCount)

	for idx, shard := range cm {
		go func(i int, s *concurrentMapShard[K, V]) {
			s.RLock()
			defer s.RUnlock()

			// A buffered channel that is big enough to hold all k-v pairs in this
			// shard will prevent this Go routine from being blocked indefinitely
			// in case reader crashes
			channel := make(chan core.Tuple[K, V], len(s.items))
			defer close(channel)

			result[i] = channel

			// Done early before populating channel to let the main routine proceeds
			wg.Done()

			for key, val := range s.items {
				channel <- core.Tuple[K, V]{key, val}
			}
		}(idx, shard)
	}

	wg.Wait()

	return result
}

// fanInBufferedChannels reads elements from channels `bufferedIns` into channel `out`.
func fanInBufferedChannels[K comparable, V any](bufferedIns []<-chan core.Tuple[K, V]) <-chan core.Tuple[K, V] {
	totalBufferSize := 0
	for _, c := range bufferedIns {
		totalBufferSize += cap(c)
	}

	out := make(chan core.Tuple[K, V], totalBufferSize)

	go func() {
		defer close(out)

		wg := sync.WaitGroup{}
		wg.Add(len(bufferedIns))

		for _, c := range bufferedIns {
			go func(channel <-chan core.Tuple[K, V]) {
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

// Iter returns an iterator which could be used in a for range loop. The capacity of the returned
// channel is the same as the size of the map at the time Iter() is called.
func (cm ConcurrentMap[K, V]) Iter() <-chan core.Tuple[K, V] {
	return fanInBufferedChannels[K, V](cm.takeSnapshot())
}

// Items returns all k-v pairs as map[K]V.
func (cm ConcurrentMap[K, V]) Items() map[K]V {
	channel := cm.Iter()

	result := make(map[K]V, cap(channel))

	for item := range channel {
		result[item.Key] = item.Val
	}

	return result
}

// ForEach executes the given doEachFn on every k-v pair in this map shard by shard.
//
// Note: doEachFn is called while lock is held. Hence, it must NOT access this map as it
// may lead to deadlock since sync.RWLock is not reentrant.
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
	return json.Marshal(cm.Items())
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
