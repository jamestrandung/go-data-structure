package set

import (
	"fmt"
	"strconv"
	"testing"
)

type testHSetInt struct {
	set  HashSet[int]
	nums []int
}

func newHSetInt(n int) testHSetInt {
	nums := nrand(n)

	return testHSetInt{
		set:  NewHashSet[int](nums...),
		nums: nums,
	}
}

func BenchmarkHashSet_AddAll(b *testing.B) {
	nums := nrand(1000)

	hs := NewHashSet[int]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.AddAll(nums)

		b.StopTimer()
		hs.Clear()
		b.StartTimer()
	}
}

func BenchmarkHashSet_Add(b *testing.B) {
	nums := nrand(1000)

	hs := NewHashSet[int]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, v := range nums {
			hs.Add(v)
		}

		b.StopTimer()
		hs.Clear()
		b.StartTimer()
	}
}

func BenchmarkHashSet_Count(b *testing.B) {
	hs := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.Count()
	}
}

func BenchmarkHashSet_IsEmpty(b *testing.B) {
	hs := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.IsEmpty()
	}
}

func BenchmarkHashSet_Has(b *testing.B) {
	hs := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.Has(i)
	}
}

func BenchmarkHashSet_Remove(b *testing.B) {
	hs := newHSetInt(b.N)

	b.ResetTimer()

	for _, val := range hs.nums {
		hs.set.Remove(val)
	}
}

func BenchmarkHashSet_Pop(b *testing.B) {
	hs := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ok := hs.set.Pop()
		if !ok {
			b.StopTimer()
			hs = newHSetInt(1000)
			b.StartTimer()
		}
	}
}

func BenchmarkHashSet_Clear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		hs := newHSetInt(1000)

		b.StartTimer()
		hs.set.Clear()
	}
}

func BenchmarkHashSet_Difference(b *testing.B) {
	this := newHSetInt(1000)
	that := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.Difference(that.set)
	}
}

func BenchmarkHashSet_SymmetricDifference(b *testing.B) {
	this := newHSetInt(1000)
	that := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.SymmetricDifference(that.set)
	}
}

func BenchmarkHashSet_Intersect(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "that is HashSet",
			test: func(b *testing.B) {
				this := newHSetInt(1000)
				that := newHSetInt(1000)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.set.Intersect(that.set)
				}
			},
		},
		{
			desc: "that is ConcurrentSet",
			test: func(b *testing.B) {
				this := newHSetInt(1000)
				that := newCSetInt(1000)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.set.Intersect(that.set)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		b.Run(
			sc.desc, func(b *testing.B) {
				sc.test(b)
			},
		)
	}
}

func BenchmarkHashSet_Union(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "that is HashSet",
			test: func(b *testing.B) {
				this := newHSetInt(1000)
				that := newHSetInt(1000)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.set.Union(that.set)
				}
			},
		},
		{
			desc: "that is ConcurrentSet",
			test: func(b *testing.B) {
				this := newHSetInt(1000)
				that := newCSetInt(1000)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.set.Union(that.set)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		b.Run(
			sc.desc, func(b *testing.B) {
				sc.test(b)
			},
		)
	}
}

func BenchmarkHashSet_Equals(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "that is HashSet",
			test: func(b *testing.B) {
				nums := nrand(1000)

				this := NewHashSet[int](nums...)
				that := NewHashSet[int](nums...)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.Equals(that)
				}
			},
		},
		{
			desc: "that is ConcurrentSet",
			test: func(b *testing.B) {
				nums := nrand(1000)

				this := NewHashSet[int](nums...)
				that := NewConcurrentSet[int](nums...)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.Equals(that)
				}
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		b.Run(
			sc.desc, func(b *testing.B) {
				sc.test(b)
			},
		)
	}
}

func BenchmarkHashSet_IsProperSubset(b *testing.B) {
	this := newHSetInt(1000)
	that := newHSetInt(1001)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.IsProperSubset(that.set)
	}
}

func BenchmarkHashSet_Contains(b *testing.B) {
	nums := nrand(1000)

	this := NewHashSet[int](nums...)
	that := NewHashSet[int](nums...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.Contains(that)
	}
}

func BenchmarkHashSet_Iter(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "create",
			test: func(b *testing.B) {
				hs := NewHashSet[string]()

				for i := 0; i < 10000; i++ {
					hs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					hs.Iter()
				}
			},
		},
		{
			desc: "loop",
			test: func(b *testing.B) {
				hs := NewHashSet[string]()

				for i := 0; i < 10000; i++ {
					hs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					for item := range hs.Iter() {
						fmt.Sprintf("%v", item)
					}
				}
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		b.Run(
			sc.desc, func(b *testing.B) {
				sc.test(b)
			},
		)
	}
}

func BenchmarkHashSet_Items(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "create",
			test: func(b *testing.B) {
				hs := NewHashSet[string]()

				for i := 0; i < 10000; i++ {
					hs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					hs.Items()
				}
			},
		},
		{
			desc: "loop",
			test: func(b *testing.B) {
				hs := NewHashSet[string]()

				for i := 0; i < 10000; i++ {
					hs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					for _, item := range hs.Items() {
						fmt.Sprintf("%v", item)
					}
				}
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		b.Run(
			sc.desc, func(b *testing.B) {
				sc.test(b)
			},
		)
	}
}

func BenchmarkHashSet_ForEach(b *testing.B) {
	hs := NewHashSet[string]()

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
	hs := newHSetInt(1000)

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
		hs := NewHashSet[string]()

		b.StartTimer()
		err := hs.UnmarshalJSON(jsonStr)
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkHashSet_String(b *testing.B) {
	hs := newHSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hs.set.String()
	}
}
