package hset

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

func nrand(n int) []int {
	i := make([]int, n)

	for ind := range i {
		i[ind] = rand.Int()
	}

	return i
}

type testSetInt struct {
	set  HashSet[int]
	nums []int
}

func newSetInt(n int) testSetInt {
	nums := nrand(n)

	return testSetInt{
		set:  New[int](nums...),
		nums: nums,
	}
}

func BenchmarkHashSet_AddAll(b *testing.B) {
	nums := nrand(1000)

	for i := 0; i < b.N; i++ {
		hs := New[int]()
		hs.AddAll(nums)
	}
}

func BenchmarkHashSet_Add(b *testing.B) {
	nums := nrand(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs := New[int]()

		for _, v := range nums {
			hs.Add(v)
		}
	}
}

func BenchmarkHashSet_Count(b *testing.B) {
	hs := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.Count()
	}
}

func BenchmarkHashSet_IsEmpty(b *testing.B) {
	hs := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.IsEmpty()
	}
}

func BenchmarkHashSet_Has(b *testing.B) {
	hs := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.Has(i)
	}
}

func BenchmarkHashSet_Remove(b *testing.B) {
	hs := newSetInt(b.N)

	b.ResetTimer()

	for _, val := range hs.nums {
		hs.set.Remove(val)
	}
}

func BenchmarkHashSet_Pop(b *testing.B) {
	hs := newSetInt(b.N)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.Pop()
	}
}

func BenchmarkHashSet_Clear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hs := newSetInt(1000)

		b.StartTimer()
		hs.set.Clear()
	}
}

func BenchmarkHashSet_Difference(b *testing.B) {
	this := newSetInt(1000)
	that := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.Difference(that.set)
	}
}

func BenchmarkHashSet_SymmetricDifference(b *testing.B) {
	this := newSetInt(1000)
	that := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.SymmetricDifference(that.set)
	}
}

func BenchmarkHashSet_Intersect(b *testing.B) {
	this := newSetInt(1000)
	that := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.Intersect(that.set)
	}
}

func BenchmarkHashSet_Union(b *testing.B) {
	this := newSetInt(1000)
	that := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.Union(that.set)
	}
}

func BenchmarkHashSet_Equals(b *testing.B) {
	nums := nrand(1000)

	this := New[int](nums...)
	that := New[int](nums...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.Equals(that)
	}
}

func BenchmarkHashSet_IsProperSubset(b *testing.B) {
	this := newSetInt(1000)
	that := newSetInt(1001)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.IsProperSubset(that.set)
	}
}

func BenchmarkHashSet_Contains(b *testing.B) {
	nums := nrand(1000)

	this := New[int](nums...)
	that := New[int](nums...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.Contains(that)
	}
}

func BenchmarkCreateIter(b *testing.B) {
	hs := New[string]()

	for i := 0; i < 10000; i++ {
		hs.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.Iter()
	}
}

func BenchmarkCreateItems(b *testing.B) {
	hs := New[string]()

	for i := 0; i < 10000; i++ {
		hs.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.Items()
	}
}

func BenchmarkLoopIter(b *testing.B) {
	hs := New[string]()

	for i := 0; i < 10000; i++ {
		hs.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for item := range hs.Iter() {
			fmt.Sprintf("%v", item)
		}
	}
}

func BenchmarkLoopItems(b *testing.B) {
	hs := New[string]()

	for i := 0; i < 10000; i++ {
		hs.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, item := range hs.Items() {
			fmt.Sprintf("%v", item)
		}
	}
}

func BenchmarkLoopForEach(b *testing.B) {
	hs := New[string]()

	for i := 0; i < 10000; i++ {
		hs.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.ForEach(
			func(element string) bool {
				fmt.Sprintf("%v", element)
				return false
			},
		)
	}
}

func BenchmarkHashSet_MarshalJSON(b *testing.B) {
	hs := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := hs.set.MarshalJSON()
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkHashSet_UnmarshalJSON(b *testing.B) {
	jsonStr := []byte("[\"elephant\",\"cow\"]")

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hs := New[string]()

		b.StartTimer()
		err := hs.UnmarshalJSON(jsonStr)
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkHashSet_String(b *testing.B) {
	hs := newSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.String()
	}
}
