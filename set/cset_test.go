package set

import (
	"encoding/json"
	"testing"

    "github.com/jamestrandung/go-data-structure/ds"
    "github.com/stretchr/testify/assert"
)

func TestConcurrentSetMatchingInterface(t *testing.T) {
	var s ds.Set[int] = NewConcurrentSet[int]()
	s.Count()
}

func TestNewConcurrentSet(t *testing.T) {
	cs := NewConcurrentSet[int]()
	assert.Equal(t, 0, cs.Count())

	verifySetEquals[int](t, NewConcurrentSet([]int{}...), NewConcurrentSet[int]())
	verifySetEquals[int](t, NewConcurrentSet([]int{1}...), NewConcurrentSet(1))
	verifySetEquals[int](t, NewConcurrentSet([]int{1, 2}...), NewConcurrentSet(1, 2))
	verifySetEquals[string](t, NewConcurrentSet([]string{"a"}...), NewConcurrentSet("a"))
	verifySetEquals[string](t, NewConcurrentSet([]string{"a", "b"}...), NewConcurrentSet("a", "b"))
}

func TestNewConcurrentSetWithConcurrencyLevel(t *testing.T) {
	cs := NewConcurrentSetWithConcurrencyLevel[int](10)

	assert.Equal(t, 0, cs.Count())
}

func TestConcurrentSet_AddAll(t *testing.T) {
	cs := NewConcurrentSet[int](1, 2, 3, 3, 2, 1)

	assert.Equal(t, 3, cs.Count())
}

func TestConcurrentSet_Add(t *testing.T) {
	cs := NewConcurrentSet[int]()

	for i := 0; i < 4; i++ {
		cs.Add(i % 2)
	}

	assert.Equal(t, 2, cs.Count())
}

func TestConcurrentSet_Count(t *testing.T) {
	cs := NewConcurrentSet[int]()

	for i := 0; i < 5; i++ {
		assert.Equal(t, i, cs.Count())
		cs.Add(i)
	}

	for i := 4; i >= 0; i-- {
		cs.Remove(i)
		assert.Equal(t, i, cs.Count())
	}
}

func TestConcurrentSet_IsEmpty(t *testing.T) {
	cs := NewConcurrentSet[int]()

	assert.True(t, cs.IsEmpty())

	cs.Add(1)

	assert.False(t, cs.IsEmpty())

	cs.Remove(1)

	assert.True(t, cs.IsEmpty())
}

func TestConcurrentSet_Has(t *testing.T) {
	cs := NewConcurrentSet[int]()

	assert.False(t, cs.Has(1))

	cs.Add(1)

	assert.True(t, cs.Has(1))

	cs.Remove(1)

	assert.False(t, cs.Has(1))
}

func TestConcurrentSet_Remove(t *testing.T) {
	cs := NewConcurrentSet[int]()

	assert.False(t, cs.Remove(1))

	cs.Add(1)

	assert.True(t, cs.Remove(1))

	assert.False(t, cs.Has(1))
}

func TestConcurrentSet_Pop(t *testing.T) {
	cs := NewConcurrentSet[int]()

	actual, ok := cs.Pop()
	assert.Equal(t, actual, 0)
	assert.False(t, ok)

	cs.Add(1)

	actual, ok = cs.Pop()
	assert.Equal(t, actual, 1)
	assert.True(t, ok)
	assert.Equal(t, 0, cs.Count())

	cs.Add(2)
	cs.Add(3)

	actual, ok = cs.Pop()
	assert.True(t, actual == 2 || actual == 3)
	assert.True(t, ok)
	assert.Equal(t, 1, cs.Count())
}

func TestConcurrentSet_Clear(t *testing.T) {
	cs := NewConcurrentSet[int](1, 2, 3, 3, 2, 1)
	another := cs
	pointer := &cs

	assert.Equal(t, 3, cs.Count())
	assert.Equal(t, 3, another.Count(), "other variable referenced the same set should have the same size")
	assert.Equal(t, 3, pointer.Count(), "other variable referenced the same set should have the same size")

	cs.Clear()

	assert.Equal(t, 0, cs.Count())
	assert.Equal(t, 0, another.Count(), "other variable referenced the same set should be empty as well")
	assert.Equal(t, 0, pointer.Count(), "other variable referenced the same set should be empty as well")
}

func TestConcurrentSet_Difference(t *testing.T) {
	a := NewConcurrentSet[int](1, 3, 4, 5, 6, 99)
	b := NewConcurrentSet[int](1, 2, 3)

	c := a.Difference(b)

	verifySetEquals[int](t, NewConcurrentSet[int](c...), NewConcurrentSet[int](4, 5, 6, 99))
}

func TestConcurrentSet_SymmetricDifference(t *testing.T) {
	a := NewConcurrentSet[int](1, 3, 4, 5, 6, 99)
	b := NewConcurrentSet[int](1, 2, 3)

	c := a.SymmetricDifference(b)

	verifySetEquals[int](t, NewConcurrentSet[int](c...), NewConcurrentSet[int](2, 4, 5, 6, 99))
}

func TestConcurrentSet_Intersect(t *testing.T) {
	a := NewConcurrentSet[int](1, 3, 4, 5, 6, 99)
	b := NewConcurrentSet[int](1, 2, 3)

	c := a.Intersect(b)

	verifySetEquals[int](t, NewConcurrentSet[int](c...), NewConcurrentSet[int](1, 3))
}

func TestConcurrentSet_Union(t *testing.T) {
	a := NewConcurrentSet[int](1, 3, 4, 5, 6, 99)
	b := NewConcurrentSet[int](1, 2, 3)

	c := a.Union(b)

	verifySetEquals[int](t, NewConcurrentSet[int](c...), NewConcurrentSet[int](1, 2, 3, 4, 5, 6, 99))
}

func TestConcurrentSet_Equals(t *testing.T) {
	a := NewConcurrentSet[int](1, 2, 3, 4, 5, 6)
	b := NewConcurrentSet[int](1, 2, 3)
	c := NewConcurrentSet[int](3, 2, 1)

	assert.False(t, a.Equals(b))
	assert.True(t, b.Equals(c))
}

func TestConcurrentSet_IsProperSubset(t *testing.T) {
	a := NewConcurrentSet[int](1, 2, 3, 4, 5, 6)
	b := NewConcurrentSet[int](1, 2, 3)
	c := NewConcurrentSet[int](3, 2, 1)

	assert.False(t, b.IsProperSubset(c))
	assert.True(t, b.IsProperSubset(a))
}

func TestConcurrentSet_Contains(t *testing.T) {
	a := NewConcurrentSet[int](1, 2, 3, 4, 5, 6)
	b := NewConcurrentSet[int](1, 2, 3)
	c := NewConcurrentSet[int](3, 2, 1)

	assert.False(t, b.Contains(a))
	assert.True(t, b.Contains(c))
	assert.True(t, a.Contains(b))
}

func TestConcurrentSet_Iter(t *testing.T) {
	cs := NewConcurrentSet[int](0, 1, 2, 3, 4, 5)

	var arr [6]int
	for element := range cs.Iter() {
		arr[element] = 1
	}

	for _, val := range arr {
		assert.Equal(t, 1, val)
	}
}

func TestConcurrentSet_Items(t *testing.T) {
	cs := NewConcurrentSet[int](0, 1, 2, 3, 4, 5)

	var arr [6]int
	for _, element := range cs.Items() {
		arr[element] = 1
	}

	for _, val := range arr {
		assert.Equal(t, 1, val)
	}
}

func TestConcurrentSet_ForEach(t *testing.T) {
	a := NewConcurrentSet[string]("W", "X", "Y", "Z")

	b := NewConcurrentSet[string]()
	a.ForEach(
		func(element string) bool {
			b.Add(element)
			return false
		},
	)

	verifySetEquals[string](t, a, b)

	var count int
	a.ForEach(
		func(elem string) bool {
			if count == 2 {
				return true
			}

			count++
			return false
		},
	)

	assert.Equal(t, 2, count, "Iteration should stop at 2")
}

func TestConcurrentSet_MarshalJSON(t *testing.T) {
	cs := NewConcurrentSet[string]("elephant", "cow")

	expected1 := "[\"elephant\",\"cow\"]"
	expected2 := "[\"cow\",\"elephant\"]"

	j, err := json.Marshal(cs)

	assert.Nil(t, err)
	assert.True(t, string(j) == expected1 || string(j) == expected2)
}

func TestConcurrentSet_UnmarshalJSON(t *testing.T) {
	jsonStr := []byte("[\"elephant\",\"cow\"]")

	cs := NewConcurrentSet[string]()

	err := cs.UnmarshalJSON(jsonStr)

	assert.Nil(t, err)
	verifySetEquals[string](t, cs, NewConcurrentSet[string]("elephant", "cow"))
}

func TestConcurrentSet_String(t *testing.T) {
	cs := NewConcurrentSet[string]("elephant", "cow")

	expected1 := "[elephant cow]"
	expected2 := "[cow elephant]"

	actual := cs.String()

	assert.True(t, actual == expected1 || actual == expected2)
}
