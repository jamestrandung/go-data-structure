package set

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSetMatchingInterface(t *testing.T) {
	var s Set[int] = NewHashSet[int]()
	s.Count()
}

func TestNewHashSet(t *testing.T) {
	hs := NewHashSet[int]()
	assert.Equal(t, 0, hs.Count())

	verifySetEquals[int](t, NewHashSet([]int{}...), NewHashSet[int]())
	verifySetEquals[int](t, NewHashSet([]int{1}...), NewHashSet(1))
	verifySetEquals[int](t, NewHashSet([]int{1, 2}...), NewHashSet(1, 2))
	verifySetEquals[string](t, NewHashSet([]string{"a"}...), NewHashSet("a"))
	verifySetEquals[string](t, NewHashSet([]string{"a", "b"}...), NewHashSet("a", "b"))
}

func TestNewHashSetWithInitialSize(t *testing.T) {
	hs := NewHashSetWithInitialSize[int](10)

	assert.Equal(t, 0, hs.Count())
}

func TestHashSet_AddAll(t *testing.T) {
	hs := NewHashSet[int](1, 2, 3, 3, 2, 1)

	assert.Equal(t, 3, hs.Count())
}

func TestHashSet_Add(t *testing.T) {
	hs := NewHashSet[int]()

	for i := 0; i < 4; i++ {
		hs.Add(i % 2)
	}

	assert.Equal(t, 2, hs.Count())
}

func TestHashSet_Count(t *testing.T) {
	hs := NewHashSet[int]()

	for i := 0; i < 5; i++ {
		assert.Equal(t, i, hs.Count())
		hs.Add(i)
	}

	for i := 4; i >= 0; i-- {
		hs.Remove(i)
		assert.Equal(t, i, hs.Count())
	}
}

func TestHashSet_IsEmpty(t *testing.T) {
	hs := NewHashSet[int]()

	assert.True(t, hs.IsEmpty())

	hs.Add(1)

	assert.False(t, hs.IsEmpty())

	hs.Remove(1)

	assert.True(t, hs.IsEmpty())
}

func TestHashSet_Has(t *testing.T) {
	hs := NewHashSet[int]()

	assert.False(t, hs.Has(1))

	hs.Add(1)

	assert.True(t, hs.Has(1))

	hs.Remove(1)

	assert.False(t, hs.Has(1))
}

func TestHashSet_Remove(t *testing.T) {
	hs := NewHashSet[int]()

	assert.False(t, hs.Remove(1))

	hs.Add(1)

	assert.True(t, hs.Remove(1))

	assert.False(t, hs.Has(1))
}

func TestHashSet_Pop(t *testing.T) {
	hs := NewHashSet[int]()

	actual, ok := hs.Pop()
	assert.Equal(t, actual, 0)
	assert.False(t, ok)

	hs.Add(1)

	actual, ok = hs.Pop()
	assert.Equal(t, actual, 1)
	assert.True(t, ok)
	assert.Equal(t, 0, hs.Count())

	hs.Add(2)
	hs.Add(3)

	actual, ok = hs.Pop()
	assert.True(t, actual == 2 || actual == 3)
	assert.True(t, ok)
	assert.Equal(t, 1, hs.Count())
}

func TestHashSet_Clear(t *testing.T) {
	hs := NewHashSet[int](1, 2, 3, 3, 2, 1)
	another := hs
	pointer := &hs

	assert.Equal(t, 3, hs.Count())
	assert.Equal(t, 3, another.Count(), "other variable referenced the same set should have the same size")
	assert.Equal(t, 3, pointer.Count(), "other variable referenced the same set should have the same size")

	hs.Clear()

	assert.Equal(t, 0, hs.Count())
	assert.Equal(t, 0, another.Count(), "other variable referenced the same set should be empty as well")
	assert.Equal(t, 0, pointer.Count(), "other variable referenced the same set should be empty as well")
}

func TestHashSet_Difference(t *testing.T) {
	a := NewHashSet[int](1, 3, 4, 5, 6, 99)
	b := NewHashSet[int](1, 2, 3)

	c := a.Difference(b)

	verifySetEquals[int](t, NewHashSet[int](c...), NewHashSet[int](4, 5, 6, 99))
}

func TestHashSet_SymmetricDifference(t *testing.T) {
	a := NewHashSet[int](1, 3, 4, 5, 6, 99)
	b := NewHashSet[int](1, 2, 3)

	c := a.SymmetricDifference(b)

	verifySetEquals[int](t, NewHashSet[int](c...), NewHashSet[int](2, 4, 5, 6, 99))
}

func TestHashSet_Intersect(t *testing.T) {
	a := NewHashSet[int](1, 3, 4, 5, 6, 99)
	b := NewHashSet[int](1, 2, 3)

	c := a.Intersect(b)

	verifySetEquals[int](t, NewHashSet[int](c...), NewHashSet[int](1, 3))
}

func TestHashSet_Union(t *testing.T) {
	a := NewHashSet[int](1, 3, 4, 5, 6, 99)
	b := NewHashSet[int](1, 2, 3)

	c := a.Union(b)

	verifySetEquals[int](t, NewHashSet[int](c...), NewHashSet[int](1, 2, 3, 4, 5, 6, 99))
}

func TestHashSet_Equals(t *testing.T) {
	a := NewHashSet[int](1, 2, 3, 4, 5, 6)
	b := NewHashSet[int](1, 2, 3)
	c := NewHashSet[int](3, 2, 1)

	assert.False(t, a.Equals(b))
	assert.True(t, b.Equals(c))
}

func TestHashSet_IsProperSubset(t *testing.T) {
	a := NewHashSet[int](1, 2, 3, 4, 5, 6)
	b := NewHashSet[int](1, 2, 3)
	c := NewHashSet[int](3, 2, 1)

	assert.False(t, b.IsProperSubset(c))
	assert.True(t, b.IsProperSubset(a))
}

func TestHashSet_Contains(t *testing.T) {
	a := NewHashSet[int](1, 2, 3, 4, 5, 6)
	b := NewHashSet[int](1, 2, 3)
	c := NewHashSet[int](3, 2, 1)

	assert.False(t, b.Contains(a))
	assert.True(t, b.Contains(c))
	assert.True(t, a.Contains(b))
}

func TestHashSet_Iter(t *testing.T) {
	hs := NewHashSet[int](0, 1, 2, 3, 4, 5)

	var arr [6]int
	for element := range hs.Iter() {
		arr[element] = 1
	}

	for _, val := range arr {
		assert.Equal(t, 1, val)
	}
}

func TestHashSet_Items(t *testing.T) {
	hs := NewHashSet[int](0, 1, 2, 3, 4, 5)

	var arr [6]int
	for _, element := range hs.Items() {
		arr[element] = 1
	}

	for _, val := range arr {
		assert.Equal(t, 1, val)
	}
}

func TestHashSet_ForEach(t *testing.T) {
	a := NewHashSet[string]("W", "X", "Y", "Z")

	b := NewHashSet[string]()
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

func TestHashSet_MarshalJSON(t *testing.T) {
	hs := NewHashSet[string]("elephant", "cow")

	expected1 := "[\"elephant\",\"cow\"]"
	expected2 := "[\"cow\",\"elephant\"]"

	j, err := json.Marshal(hs)

	assert.Nil(t, err)
	assert.True(t, string(j) == expected1 || string(j) == expected2)
}

func TestHashSet_UnmarshalJSON(t *testing.T) {
	jsonStr := []byte("[\"elephant\",\"cow\"]")

	hs := NewHashSet[string]()

	err := hs.UnmarshalJSON(jsonStr)

	assert.Nil(t, err)
	verifySetEquals[string](t, hs, NewHashSet[string]("elephant", "cow"))
}

func TestHashSet_String(t *testing.T) {
	hs := NewHashSet[string]("elephant", "cow")

	expected1 := "[elephant cow]"
	expected2 := "[cow elephant]"

	actual := hs.String()

	assert.True(t, actual == expected1 || actual == expected2)
}
