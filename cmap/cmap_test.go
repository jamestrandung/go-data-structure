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
	name string
}

func TestNewConcurrentMap(t *testing.T) {
	m := NewConcurrentMap[string, int]()

	assert.NotNil(t, m)
	assert.Equal(t, 0, m.Count(), "new map should be empty.")
	assert.Equal(t, defaultShardCount, len(m))
}

func TestNewConcurrentMapWithConcurrencyLevel(t *testing.T) {
	concurrencyLevel := 64
	m := NewConcurrentMapWithConcurrencyLevel[string, int](concurrencyLevel)

	assert.NotNil(t, m)
	assert.Equal(t, 0, m.Count(), "new map should be empty.")
	assert.Equal(t, concurrencyLevel, len(m))
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
				m := NewConcurrentMap[string, int]()

				shard := m.getShard("test")

				assert.True(t, shard == m[9])
			},
		},
		{
			desc: "key is int",
			test: func(t *testing.T) {
				m := NewConcurrentMap[int, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is int8",
			test: func(t *testing.T) {
				m := NewConcurrentMap[int8, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is int16",
			test: func(t *testing.T) {
				m := NewConcurrentMap[int16, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is int32",
			test: func(t *testing.T) {
				m := NewConcurrentMap[int32, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is int64",
			test: func(t *testing.T) {
				m := NewConcurrentMap[int64, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is uint",
			test: func(t *testing.T) {
				m := NewConcurrentMap[uint, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is uint8",
			test: func(t *testing.T) {
				m := NewConcurrentMap[uint8, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is uint16",
			test: func(t *testing.T) {
				m := NewConcurrentMap[uint16, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is uint32",
			test: func(t *testing.T) {
				m := NewConcurrentMap[uint32, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key is uint64",
			test: func(t *testing.T) {
				m := NewConcurrentMap[uint64, int]()

				shard := m.getShard(1)

				assert.True(t, shard == m[14])
			},
		},
		{
			desc: "key implements hasher",
			test: func(t *testing.T) {
				m := NewConcurrentMap[dummyHasher, int]()

				shard := m.getShard(dummyHasher{})

				assert.True(t, shard == m[dummyHash])
			},
		},
		{
			desc: "hasher panics",
			test: func(t *testing.T) {
				m := NewConcurrentMap[panicHasher, int]()

				shard := m.getShard(panicHasher{})

				assert.True(t, shard == m[0])
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

				m := NewConcurrentMap[struct{}, int]()

				shard := m.getShard(struct{}{})

				assert.True(t, shard == m[1])
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

				m := NewConcurrentMap[struct{}, int]()

				shard := m.getShard(struct{}{})

				assert.True(t, shard == m[0])
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

				m := NewConcurrentMap[struct{}, int]()

				shard := m.getShard(struct{}{})

				assert.True(t, shard == m[0])
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

func TestSet(t *testing.T) {
	m := NewConcurrentMap[string, animal]()
	elephant := animal{"elephant"}
	monkey := animal{"monkey"}

	prev, ok := m.Set("elephant", elephant)
	assert.Equal(t, animal{}, prev)
	assert.False(t, ok)

	prev, ok = m.Set("monkey", monkey)
	assert.Equal(t, animal{}, prev)
	assert.False(t, ok)

	assert.Equal(t, 2, m.Count(), "map should contain exactly two elements.")
}

func TestGet(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	// Get a missing element.
	val, ok := m.Get("Dog")
	assert.Equal(t, animal{}, val, "Missing values should return as zero-value.")
	assert.False(t, ok, "ok should be false when item is missing from map.")

	elephant := animal{"elephant"}
	m.Set("elephant", elephant)

	// Retrieve inserted element.
	actual, ok := m.Get("elephant")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", actual.name)
}

func TestGetElseSet(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	m.Set("elephant", elephant)

	actual, ok := m.GetElseSet("elephant", elephant)
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", actual.name)

	cow := animal{"cow"}
	actual, ok = m.GetElseSet("cow", cow)
	assert.False(t, ok, "ok should be false for item newly inserted to the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.name)

	actual, ok = m.Get("cow")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.name)
}

func TestGetAndSetIf(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	m.Set("elephant", elephant)

	// Retrieve inserted element.
	cow := animal{"cow"}
	oldVal, found, newVal, shouldSet := m.GetAndSetIf(
		"elephant", func(currentVal animal, found bool) (animal, bool) {
			return cow, true
		},
	)

	assert.True(t, found, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, oldVal, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", oldVal.name)
	assert.True(t, shouldSet)
	assert.Equal(t, cow, newVal, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", newVal.name, "expected element to be different.")

	// Re-retrieve inserted element.
	newVal, found = m.Get("elephant")
	assert.True(t, found, "ok should be true for item stored within the map.")
	assert.Equal(t, cow, newVal, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", newVal.name, "expected element to be different.")
}

func TestGetElseCreate(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	elephant := animal{"elephant"}
	m.Set("elephant", elephant)

	elephantFactory := func() animal {
		return elephant
	}

	// Retrieve inserted element.
	actual, ok := m.GetElseCreate("elephant", elephantFactory)
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, elephant, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "elephant", actual.name)

	cow := animal{"cow"}
	cowFactory := func() animal {
		return cow
	}

	actual, ok = m.GetElseCreate("cow", cowFactory)
	assert.False(t, ok, "ok should be false for item newly inserted to the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.name)

	actual, ok = m.Get("cow")
	assert.True(t, ok, "ok should be true for item stored within the map.")
	assert.Equal(t, cow, actual, "expecting an element, not zero-value.")
	assert.Equal(t, "cow", actual.name)
}

func TestCount(t *testing.T) {
	m := NewConcurrentMap[string, animal]()
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	assert.Equal(t, 100, m.Count())
}

func TestHas(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	assert.False(t, m.Has("Money"), "element shouldn't exists")

	elephant := animal{"elephant"}
	m.Set("elephant", elephant)

	assert.True(t, m.Has("elephant"), "element exists, expecting Has to return True.")
}

func TestIsEmpty(t *testing.T) {
	m := NewConcurrentMap[string, animal]()
	assert.True(t, m.IsEmpty())

	m.Set("elephant", animal{"elephant"})
	assert.False(t, m.IsEmpty())
}

func TestRemove(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	monkey := animal{"monkey"}
	m.Set("monkey", monkey)
	m.Remove("monkey")

	assert.Equal(t, 0, m.Count(), "Expecting count to be zero once item was removed.")

	actual, ok := m.Get("monkey")
	assert.False(t, ok)
	assert.Equal(t, animal{}, actual)

	// Remove a none existing element.
	m.Remove("noone")
}

func TestRemoveIfValue(t *testing.T) {
	// same value type
	m := NewConcurrentMap[string, string]()

	name := "John"
	monkey := "monkey"
	m.Set(name, monkey)

	ok, err := m.RemoveIfValue(name, "tiger")
	assert.False(t, ok)
	assert.Nil(t, err)

	actual, ok := m.Get(name)
	assert.True(t, ok)
	assert.Equal(t, monkey, actual)

	ok, err = m.RemoveIfValue("Cat", monkey)
	assert.False(t, ok)
	assert.Nil(t, err)

	actual, ok = m.Get(name)
	assert.True(t, ok)
	assert.Equal(t, monkey, actual)

	ok, err = m.RemoveIfValue(name, monkey)
	assert.True(t, ok)
	assert.Nil(t, err)

	actual, ok = m.Get(name)
	assert.False(t, ok)
	assert.Equal(t, actual, "")

	// different value types
	m2 := NewConcurrentMap[string, chan int]()

	circle := make(chan int)
	m2.Set("dog", circle)

	ok, err = m2.RemoveIfValue("cat", circle)
	assert.False(t, ok)
	assert.Nil(t, err)

	actual2, ok := m2.Get("dog")
	assert.True(t, ok)
	assert.Equal(t, circle, actual2)

	ok, err = m2.RemoveIfValue("dog", make(chan int))
	assert.False(t, ok)
	assert.Nil(t, err)

	actual2, ok = m2.Get("dog")
	assert.True(t, ok)
	assert.Equal(t, circle, actual2)

	ok, err = m2.RemoveIfValue("dog", circle)
	assert.True(t, ok)
	assert.Nil(t, err)

	actual2, ok = m2.Get("dog")
	assert.False(t, ok)
	assert.Nil(t, actual2)
}

func TestIter(t *testing.T) {
	m := NewConcurrentMap[string, animal]()

	// Insert 100 elements.
	for i := 0; i < 100; i++ {
		m.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	counter := 0
	for _, item := range m.Iter() {
		val := item.Val
		assert.NotNil(t, val)
		counter++
	}

	assert.Equal(t, 100, counter)
}

func TestForEach(t *testing.T) {
	m := NewConcurrentMap[string, int]()

	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < iterations/2; i++ {
			m.Set(strconv.Itoa(i), i)
		}

		wg.Done()
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			m.Set(strconv.Itoa(i), i)
		}

		wg.Done()
	}()

	wg.Wait()

	// Verify all values have been inserted
	assert.Equal(t, iterations, m.Count())

	// Extract all inserted values into an array
	var a [iterations]int
	m.ForEach(
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
	m := NewConcurrentMap[string, int]()

	const iterations = 1000
	for i := 0; i < iterations; i++ {
		m.Set(strconv.Itoa(i), i)
	}

	// Verify all values have been inserted
	assert.Equal(t, iterations, m.Count())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
				assert.Equal(t, assert.AnError, e)
			}
		}()

		m.ForEach(
			func(key string, val int) {
				panic(assert.AnError)
			},
		)
	}()

	wg.Wait()

	for _, shard := range m {
		ok := shard.TryLock()
		assert.True(t, ok, "panics must not cause any shards to be locked indefinitely")
		if ok {
			shard.Unlock()
		}
	}
}

func TestJsonMarshal(t *testing.T) {
	defer func() { defaultShardCount = 32 }()
	defaultShardCount = 2

	expected := "{\"a\":1,\"b\":2}"

	m := NewConcurrentMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)

	j, err := json.Marshal(m)
	assert.Nil(t, err)
	assert.Equal(t, string(j), expected)
}

func TestConcurrent(t *testing.T) {
	m := NewConcurrentMap[string, int]()

	ch := make(chan int)
	const iterations = 1000
	var a [iterations]int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i := 0; i < iterations/2; i++ {
			m.Set(strconv.Itoa(i), i)

			val, _ := m.Get(strconv.Itoa(i))

			ch <- val
		}

		wg.Done()
	}()

	go func() {
		for i := iterations / 2; i < iterations; i++ {
			m.Set(strconv.Itoa(i), i)

			val, _ := m.Get(strconv.Itoa(i))

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

	assert.Equal(t, iterations, m.Count())
	assert.Equal(t, iterations, counter)

	// Sorts array to make it easier to verify all inserted values are there
	sort.Ints(a[0:iterations])

	for i := 0; i < iterations; i++ {
		assert.Equal(t, i, a[i])
	}
}
