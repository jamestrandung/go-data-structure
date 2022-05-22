package emap

import (
	"encoding/json"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/jamestrandung/go-data-structure/ds"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/stretchr/testify/assert"
)

type animal struct {
	Name string
}

func TestConcurrentMapMatchingInterface(t *testing.T) {
	var m ds.Map[string, int] = NewConcurrentMap[string, int]()
	m.Count()
}

func TestNewConcurrentMap(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	assert.NotNil(t, cm)
	assert.Equal(t, 0, cm.Count(), "new map should be empty.")
	assert.Equal(t, defaultShardCount, len(cm))
}

func TestNewConcurrentMapWithConcurrencyLevel(t *testing.T) {
	concurrencyLevel := 64
	cm := NewConcurrentMapWithConcurrencyLevel[string, int](concurrencyLevel)

	assert.NotNil(t, cm)
	assert.Equal(t, 0, cm.Count(), "new map should be empty.")
	assert.Equal(t, concurrencyLevel, len(cm))
}

type dummyHasher struct{}

const dummyHash uint64 = 1

func (dummyHasher) Hash() uint64 {
	return dummyHash
}

type panicHasher struct{}

func (panicHasher) Hash() uint64 {
	panic("panicHasher")
}

func TestConcurrentMap_GetShard(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(*testing.T)
	}{
		{
			desc: "key is string",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[string, int]()

				shard := cm.getShard("test")

				assert.True(t, shard == cm[9])
			},
		},
		{
			desc: "key is int",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[int, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int8",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[int8, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int16",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[int16, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int32",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[int32, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int64",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[int64, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[uint, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint8",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[uint8, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint16",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[uint16, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint32",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[uint32, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint64",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[uint64, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key implements hasher",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[dummyHasher, int]()

				shard := cm.getShard(dummyHasher{})

				assert.True(t, shard == cm[dummyHash])
			},
		},
		{
			desc: "hasher panics",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[panicHasher, int]()

				shard := cm.getShard(panicHasher{})

				assert.True(t, shard == cm[0])
			},
		},
		{
			desc: "any keys",
			test: func(t *testing.T) {
				defer func(original func(v interface{}, format hashstructure.Format, opts *hashstructure.HashOptions) (uint64, error)) {
					hashFn = original
				}(hashFn)

				hashFn = func(v interface{}, format hashstructure.Format, opts *hashstructure.HashOptions) (uint64, error) {
					return 1, nil
				}

				cm := NewConcurrentMap[struct{}, int]()

				shard := cm.getShard(struct{}{})

				assert.True(t, shard == cm[1])
			},
		},
		{
			desc: "hashFn panics",
			test: func(t *testing.T) {
				defer func(original func(v interface{}, format hashstructure.Format, opts *hashstructure.HashOptions) (uint64, error)) {
					hashFn = original
				}(hashFn)

				hashFn = func(v interface{}, format hashstructure.Format, opts *hashstructure.HashOptions) (uint64, error) {
					panic("hashFn")
				}

				cm := NewConcurrentMap[struct{}, int]()

				shard := cm.getShard(struct{}{})

				assert.True(t, shard == cm[0])
			},
		},
		{
			desc: "hashFn errors",
			test: func(t *testing.T) {
				defer func(original func(v interface{}, format hashstructure.Format, opts *hashstructure.HashOptions) (uint64, error)) {
					hashFn = original
				}(hashFn)

				hashFn = func(v interface{}, format hashstructure.Format, opts *hashstructure.HashOptions) (uint64, error) {
					return 1000, assert.AnError
				}

				cm := NewConcurrentMap[struct{}, int]()

				shard := cm.getShard(struct{}{})

				assert.True(t, shard == cm[0])
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(
			sc.desc, func(t *testing.T) {
				sc.test(t)
			},
		)
	}
}

func TestConcurrentMap_Set(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()
	elephant := animal{"elephant"}
	monkey := animal{"monkey"}

	prev, ok := cm.Set("elephant", elephant)
	assert.Equal(t, animal{}, prev)
	assert.False(t, ok)

	prev, ok = cm.Set("monkey", monkey)
	assert.Equal(t, animal{}, prev)
	assert.False(t, ok)

	assert.Equal(t, 2, cm.Count(), "map should contain exactly two elements.")
}

func TestConcurrentMap_SetAll(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	monkey := animal{"monkey"}

	data := make(map[string]animal)
	data["elephant"] = elephant
	data["monkey"] = monkey

	cm.SetAll(data)

	assert.Equal(t, 2, cm.Count(), "map should contain exactly two elements.")

	// Retrieve inserted element.
	actual, ok := cm.Get("elephant")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", actual.Name)

	actual, ok = cm.Get("monkey")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, monkey, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "monkey", actual.Name)
}

func TestConcurrentMap_Get(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	// Get a missing element.
	val, ok := cm.Get("Dog")
	assert.Equal(t, animal{}, val, "Missing values should return as zero-value.")
	assert.False(t, ok, "ok should be false when item is missing from map.")

	elephant := animal{"elephant"}
	cm.Set("elephant", elephant)

	// Retrieve inserted element.
	actual, ok := cm.Get("elephant")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", actual.Name)
}

func TestConcurrentMap_SetIfAbsent(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	cm.Set("elephant", elephant)

	absent := cm.SetIfAbsent("elephant", elephant)
	assert.False(t, absent, "absent should be false for item stored within the map.")

	cow := animal{"cow"}
	absent = cm.SetIfAbsent("cow", cow)
	assert.True(t, absent, "absent should be true for item newly inserted to the map.")

	actual, ok := cm.Get("cow")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.Name)
}

func TestConcurrentMap_GetAndSetIf(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	cm.Set("elephant", elephant)

	// Retrieve inserted element.
	cow := animal{"cow"}
	oldVal, found, newVal, shouldSet := cm.GetAndSetIf(
		"elephant", func(currentVal animal, found bool) (animal, bool) {
			return cow, true
		},
	)

	assert.True(t, found, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, oldVal, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", oldVal.Name)
	assert.True(t, shouldSet)
	assert.Equal(t, cow, newVal, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", newVal.Name, "expected element to be different.")

	// Re-retrieve inserted element.
	newVal, found = cm.Get("elephant")
	assert.True(t, found, "ok should be true for item stored within the map.")
	assert.Equal(t, cow, newVal, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", newVal.Name, "expected element to be different.")
}

func TestConcurrentMap_GetElseCreate(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	cm.Set("elephant", elephant)

	elephantFactory := func() animal {
		return elephant
	}

	// Retrieve inserted element.
	actual, ok := cm.GetElseCreate("elephant", elephantFactory)
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", actual.Name)

	cow := animal{"cow"}
	cowFactory := func() animal {
		return cow
	}

	actual, ok = cm.GetElseCreate("cow", cowFactory)
	assert.False(t, ok, "ok should be false for item newly inserted to the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.Name)

	actual, ok = cm.Get("cow")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.Name)
}

func TestConcurrentMap_Count(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()
	for i := 0; i < 100; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	assert.Equal(t, 100, cm.Count())
}

func TestConcurrentMap_IsEmpty(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()
	assert.True(t, cm.IsEmpty())

	cm.Set("elephant", animal{"elephant"})
	assert.False(t, cm.IsEmpty())
}

func TestConcurrentMap_Has(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	assert.False(t, cm.Has("Money"), "element shouldn't exists")

	elephant := animal{"elephant"}
	cm.Set("elephant", elephant)

	assert.True(t, cm.Has("elephant"), "element exists, expecting Has to return True.")
}

func TestConcurrentMap_Remove(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()

	monkey := animal{"monkey"}
	cm.Set("monkey", monkey)

	actual, found := cm.Remove("noone")
	assert.False(t, found)
	assert.Equal(t, animal{}, actual)

	assert.Equal(t, 1, cm.Count())

	actual, found = cm.Remove("monkey")
	assert.True(t, found)
	assert.Equal(t, monkey, actual)

	assert.Equal(t, 0, cm.Count())

	actual, ok := cm.Get("monkey")
	assert.False(t, ok)
	assert.Equal(t, animal{}, actual)
}

func TestConcurrentMap_RemoveAll(t *testing.T) {
	cm := NewConcurrentMap[string, string]()

	cm.Set("monkey", "monkey")
	cm.Set("elephant", "elephant")
	cm.Set("dog", "dog")

	actual := cm.RemoveAll([]string{"noone", "nobody"})

	assert.False(t, actual)
	assert.Equal(t, 3, cm.Count())

	actual = cm.RemoveAll([]string{"monkey", "elephant"})

	assert.True(t, actual)
	assert.Equal(t, 1, cm.Count())
}

func TestConcurrentMap_RemoveIf(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(*testing.T)
	}{
		{
			desc: "same value type",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[string, string]()

				name := "John"
				monkey := "monkey"
				cm.Set(name, monkey)

				actual, removed := cm.RemoveIf(
					name, func(currentVal string) bool {
						return false
					},
				)

				assert.False(t, removed)
				assert.Equal(t, "", actual)

				actual, ok := cm.Get(name)
				assert.True(t, ok)
				assert.Equal(t, monkey, actual)

				actual, removed = cm.RemoveIf(
					"Cat", func(currentVal string) bool {
						return true
					},
				)

				assert.False(t, removed)
				assert.Equal(t, "", actual)

				actual, ok = cm.Get(name)
				assert.True(t, ok)
				assert.Equal(t, monkey, actual)

				actual, removed = cm.RemoveIf(
					name, func(currentVal string) bool {
						return true
					},
				)

				assert.True(t, removed)
				assert.Equal(t, monkey, actual)

				actual, ok = cm.Get(name)
				assert.False(t, ok)
				assert.Equal(t, actual, "")
			},
		},
		{
			desc: "different value types",
			test: func(t *testing.T) {
				cm := NewConcurrentMap[string, chan int]()

				circle := make(chan int)
				cm.Set("dog", circle)

				actual, removed := cm.RemoveIf(
					"cat", func(currentVal chan int) bool {
						return true
					},
				)

				assert.False(t, removed)
				assert.Nil(t, actual)

				actual, ok := cm.Get("dog")
				assert.True(t, ok)
				assert.Equal(t, circle, actual)

				actual, removed = cm.RemoveIf(
					"dog", func(currentVal chan int) bool {
						return false
					},
				)

				assert.False(t, removed)
				assert.Nil(t, actual)

				actual, ok = cm.Get("dog")
				assert.True(t, ok)
				assert.Equal(t, circle, actual)

				actual, removed = cm.RemoveIf(
					"dog", func(currentVal chan int) bool {
						return true
					},
				)

				assert.True(t, removed)
				assert.Equal(t, circle, actual)

				actual, ok = cm.Get("dog")
				assert.False(t, ok)
				assert.Nil(t, actual)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario
		t.Run(
			sc.desc, func(t *testing.T) {
				sc.test(t)
			},
		)
	}
}

func TestConcurrentMap_Clear(t *testing.T) {
	cm := NewConcurrentMap[string, animal]()
	for i := 0; i < 100; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	assert.Equal(t, 100, cm.Count())

	cm.Clear()

	assert.Equal(t, 0, cm.Count())
}

func TestConcurrentMap_Iter(t *testing.T) {
	cm := NewConcurrentMap[int, animal]()

	for i := 0; i < 100; i++ {
		cm.Set(i, animal{strconv.Itoa(i)})
	}

	arr := [100]int{}
	for item := range cm.Iter() {
		arr[item.Key] = 1
	}

	// Verify we got back all keys
	for _, val := range arr {
		assert.Equal(t, 1, val)
	}
}

func TestConcurrentMap_Items(t *testing.T) {
	cm := NewConcurrentMap[int, animal]()

	for i := 0; i < 100; i++ {
		cm.Set(i, animal{strconv.Itoa(i)})
	}

	arr := [100]int{}
	for _, item := range cm.Items() {
		arr[item.Key] = 1
	}

	// Verify we got back all keys
	for _, val := range arr {
		assert.Equal(t, 1, val)
	}
}

func TestConcurrentMap_ForEach(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < iterations/2; i++ {
			cm.Set(strconv.Itoa(i), i)
		}

		wg.Done()
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			cm.Set(strconv.Itoa(i), i)
		}

		wg.Done()
	}()

	wg.Wait()

	// Verify all values have been inserted
	assert.Equal(t, iterations, cm.Count())

	// Extract all inserted values into an array
	var a [iterations]int
	cm.ForEach(
		func(key string, val int) bool {
			a[val] = val
			return false
		},
	)

	// Verify all slots in the array carry the expected non-zero value
	for i := 0; i < iterations; i++ {
		assert.Equal(t, i, a[i])
	}
}

func TestConcurrentMap_ForEach_Panic(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	const iterations = 1000
	for i := 0; i < iterations; i++ {
		cm.Set(strconv.Itoa(i), i)
	}

	// Verify all values have been inserted
	assert.Equal(t, iterations, cm.Count())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
				assert.Equal(t, assert.AnError, e)
			}
		}()

		cm.ForEach(
			func(key string, val int) bool {
				panic(assert.AnError)
			},
		)
	}()

	wg.Wait()

	for _, shard := range cm {
		ok := shard.TryLock()
		assert.True(t, ok, "panics must not cause any shards to be locked indefinitely")
		if ok {
			shard.Unlock()
		}
	}
}

func TestConcurrentMap_AsMap(t *testing.T) {
	cm := NewConcurrentMap[int, animal]()

	for i := 0; i < 100; i++ {
		cm.Set(i, animal{strconv.Itoa(i)})
	}

	m := cm.AsMap()

	assert.Equal(t, cm.Count(), len(m))

	// Verify all k-v pairs in `cm` are in `items`
	for item := range cm.Iter() {
		_, ok := m[item.Key]
		assert.True(t, ok)
	}
}

func TestConcurrentMap_MarshalJSON(t *testing.T) {
	defer func() { defaultShardCount = 32 }()
	defaultShardCount = 2

	cm := NewConcurrentMap[string, animal]()
	cm.Set("a", animal{"elephant"})
	cm.Set("b", animal{"cow"})

	expected := "{\"a\":{\"Name\":\"elephant\"},\"b\":{\"Name\":\"cow\"}}"

	j, err := json.Marshal(cm)

	assert.Nil(t, err)
	assert.Equal(t, expected, string(j))
}

func TestConcurrentMap_UnmarshalJSON(t *testing.T) {
	jsonStr := []byte("{\"a\":{\"Name\":\"elephant\"},\"b\":{\"Name\":\"cow\"}}")

	cm := NewConcurrentMap[string, animal]()

	err := cm.UnmarshalJSON(jsonStr)

	assert.Nil(t, err)
	assert.Equal(t, 2, cm.Count())
}

func TestConcurrentMap_String(t *testing.T) {
	defer func() { defaultShardCount = 32 }()
	defaultShardCount = 2

	cm := NewConcurrentMap[string, animal]()
	cm.Set("a", animal{"elephant"})
	cm.Set("b", animal{"cow"})

	expected := "map[a:{elephant} b:{cow}]"

	assert.Equal(t, expected, cm.String())
}

func TestConcurrent(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	ch := make(chan int)
	const iterations = 1000
	var a [iterations]int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < iterations/2; i++ {
			cm.Set(strconv.Itoa(i), i)

			val, _ := cm.Get(strconv.Itoa(i))

			ch <- val
		}

		wg.Done()
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			cm.Set(strconv.Itoa(i), i)

			val, _ := cm.Get(strconv.Itoa(i))

			ch <- val
		}

		wg.Done()
	}()

	go func() {
		defer close(ch)
		wg.Wait()
	}()

	counter := 0
	for elem := range ch {
		a[counter] = elem
		counter++
	}

	assert.Equal(t, iterations, cm.Count())
	assert.Equal(t, iterations, counter)

	// Sorts array to make it easier to verify all inserted values are there
	sort.Ints(a[0:iterations])

	for i := 0; i < iterations; i++ {
		assert.Equal(t, i, a[i])
	}
}
