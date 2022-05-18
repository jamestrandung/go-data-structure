package cmap

import (
	"encoding/json"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/stretchr/testify/assert"
)

type animal struct {
	Name string
}

func TestNew(t *testing.T) {
	cm := New[string, int]()

	assert.NotNil(t, cm)
	assert.Equal(t, 0, cm.Count(), "new map should be empty.")
	assert.Equal(t, defaultShardCount, len(cm))
}

func TestNewWithConcurrencyLevel(t *testing.T) {
	concurrencyLevel := 64
	cm := NewWithConcurrencyLevel[string, int](concurrencyLevel)

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

func TestGetShard(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(*testing.T)
	}{
		{
			desc: "key is string",
			test: func(t *testing.T) {
				cm := New[string, int]()

				shard := cm.getShard("test")

				assert.True(t, shard == cm[9])
			},
		},
		{
			desc: "key is int",
			test: func(t *testing.T) {
				cm := New[int, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int8",
			test: func(t *testing.T) {
				cm := New[int8, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int16",
			test: func(t *testing.T) {
				cm := New[int16, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int32",
			test: func(t *testing.T) {
				cm := New[int32, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is int64",
			test: func(t *testing.T) {
				cm := New[int64, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint",
			test: func(t *testing.T) {
				cm := New[uint, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint8",
			test: func(t *testing.T) {
				cm := New[uint8, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint16",
			test: func(t *testing.T) {
				cm := New[uint16, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint32",
			test: func(t *testing.T) {
				cm := New[uint32, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key is uint64",
			test: func(t *testing.T) {
				cm := New[uint64, int]()

				shard := cm.getShard(1)

				assert.True(t, shard == cm[14])
			},
		},
		{
			desc: "key implements hasher",
			test: func(t *testing.T) {
				cm := New[dummyHasher, int]()

				shard := cm.getShard(dummyHasher{})

				assert.True(t, shard == cm[dummyHash])
			},
		},
		{
			desc: "hasher panics",
			test: func(t *testing.T) {
				cm := New[panicHasher, int]()

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

				cm := New[struct{}, int]()

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

				cm := New[struct{}, int]()

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

				cm := New[struct{}, int]()

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

func TestSetAll(t *testing.T) {
	cm := New[string, animal]()

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

func TestSet(t *testing.T) {
	cm := New[string, animal]()
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

func TestGet(t *testing.T) {
	cm := New[string, animal]()

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

func TestSetIfAbsent(t *testing.T) {
	cm := New[string, animal]()

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

func TestGetAndSetIf(t *testing.T) {
	cm := New[string, animal]()

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

func TestGetElseCreate(t *testing.T) {
	cm := New[string, animal]()

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

func TestCount(t *testing.T) {
	cm := New[string, animal]()
	for i := 0; i < 100; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	assert.Equal(t, 100, cm.Count())
}

func TestHas(t *testing.T) {
	cm := New[string, animal]()

	assert.False(t, cm.Has("Money"), "element shouldn't exists")

	elephant := animal{"elephant"}
	cm.Set("elephant", elephant)

	assert.True(t, cm.Has("elephant"), "element exists, expecting Has to return True.")
}

func TestIsEmpty(t *testing.T) {
	cm := New[string, animal]()
	assert.True(t, cm.IsEmpty())

	cm.Set("elephant", animal{"elephant"})
	assert.False(t, cm.IsEmpty())
}

func TestRemove(t *testing.T) {
	cm := New[string, animal]()

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

func TestRemoveIf(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(*testing.T)
	}{
		{
			desc: "same value type",
			test: func(t *testing.T) {
				cm := New[string, string]()

				name := "John"
				monkey := "monkey"
				cm.Set(name, monkey)

				actual, removed := cm.RemoveIf(
					name, func(currentVal string, found bool) bool {
						return false
					},
				)

				assert.False(t, removed)
				assert.Equal(t, "", actual)

				actual, ok := cm.Get(name)
				assert.True(t, ok)
				assert.Equal(t, monkey, actual)

				actual, removed = cm.RemoveIf(
					"Cat", func(currentVal string, found bool) bool {
						return true
					},
				)

				assert.False(t, removed)
				assert.Equal(t, "", actual)

				actual, ok = cm.Get(name)
				assert.True(t, ok)
				assert.Equal(t, monkey, actual)

				actual, removed = cm.RemoveIf(
					name, func(currentVal string, found bool) bool {
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
				cm := New[string, chan int]()

				circle := make(chan int)
				cm.Set("dog", circle)

				actual, removed := cm.RemoveIf(
					"cat", func(currentVal chan int, found bool) bool {
						return true
					},
				)

				assert.False(t, removed)
				assert.Nil(t, actual)

				actual, ok := cm.Get("dog")
				assert.True(t, ok)
				assert.Equal(t, circle, actual)

				actual, removed = cm.RemoveIf(
					"dog", func(currentVal chan int, found bool) bool {
						return false
					},
				)

				assert.False(t, removed)
				assert.Nil(t, actual)

				actual, ok = cm.Get("dog")
				assert.True(t, ok)
				assert.Equal(t, circle, actual)

				actual, removed = cm.RemoveIf(
					"dog", func(currentVal chan int, found bool) bool {
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

func TestClear(t *testing.T) {
	cm := New[string, animal]()
	for i := 0; i < 100; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	assert.Equal(t, 100, cm.Count())

	cm.Clear()

	assert.Equal(t, 0, cm.Count())
}

func TestIter(t *testing.T) {
	cm := New[int, animal]()

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

func TestItems(t *testing.T) {
	cm := New[int, animal]()

	for i := 0; i < 100; i++ {
		cm.Set(i, animal{strconv.Itoa(i)})
	}

	items := cm.Items()

	assert.Equal(t, cm.Count(), len(items))

	// Verify all k-v pairs in `cm` are in `items`
	for item := range cm.Iter() {
		_, ok := items[item.Key]
		assert.True(t, ok)
	}
}

func TestForEach(t *testing.T) {
	cm := New[string, int]()

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
		func(key string, val int) {
			a[val] = val
		},
	)

	// Verify all slots in the array carry the expected non-zero value
	for i := 0; i < iterations; i++ {
		assert.Equal(t, i, a[i])
	}
}

func TestForEach_Panic(t *testing.T) {
	cm := New[string, int]()

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
			func(key string, val int) {
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

func TestMarshalJSON(t *testing.T) {
	defer func() { defaultShardCount = 32 }()
	defaultShardCount = 2

	cm := New[string, animal]()
	cm.Set("a", animal{"elephant"})
	cm.Set("b", animal{"cow"})

	expected := "{\"a\":{\"Name\":\"elephant\"},\"b\":{\"Name\":\"cow\"}}"

	j, err := json.Marshal(cm)

	assert.Nil(t, err)
	assert.Equal(t, string(j), expected)
}

func TestUnmarshalJSON(t *testing.T) {
	jsonStr := []byte("{\"a\":{\"Name\":\"elephant\"},\"b\":{\"Name\":\"cow\"}}")

	cm := New[string, animal]()

	err := cm.UnmarshalJSON(jsonStr)

	assert.Nil(t, err)
	assert.Equal(t, 2, cm.Count())
}

func TestConcurrent(t *testing.T) {
	cm := New[string, int]()

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
