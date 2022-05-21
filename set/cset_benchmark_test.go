package set

import (
	"fmt"
	"strconv"
	"testing"
)

type testCSetInt struct {
	set  ConcurrentSet[int]
	nums []int
}

func newCSetInt(n int) testCSetInt {
	nums := nrand(n)

	return testCSetInt{
		set:  NewConcurrentSet[int](nums...),
		nums: nums,
	}
}

func BenchmarkConcurrentSet_AddAll(b *testing.B) {
	nums := nrand(1000)

	cs := NewConcurrentSet[int]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.AddAll(nums)

		b.StopTimer()
		cs.Clear()
		b.StartTimer()
	}
}

func BenchmarkConcurrentSet_Add(b *testing.B) {
	nums := nrand(1000)

	cs := NewConcurrentSet[int]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, v := range nums {
			cs.Add(v)
		}

		b.StopTimer()
		cs.Clear()
		b.StartTimer()
	}
}

func BenchmarkConcurrentSet_Count(b *testing.B) {
	cs := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.set.Count()
	}
}

func BenchmarkConcurrentSet_IsEmpty(b *testing.B) {
	cs := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.set.IsEmpty()
	}
}

func BenchmarkConcurrentSet_Has(b *testing.B) {
	cs := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.set.Has(i)
	}
}

func BenchmarkConcurrentSet_Remove(b *testing.B) {
	cs := newCSetInt(b.N)

	b.ResetTimer()

	for _, val := range cs.nums {
		cs.set.Remove(val)
	}
}

func BenchmarkConcurrentSet_Pop(b *testing.B) {
	cs := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ok := cs.set.Pop()
		if ok {
			b.StopTimer()
			cs = newCSetInt(1000)
			b.StartTimer()
		}
	}
}

func BenchmarkConcurrentSet_Clear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cs := newCSetInt(1000)

		b.StartTimer()
		cs.set.Clear()
	}
}

func BenchmarkConcurrentSet_Difference(b *testing.B) {
	this := newCSetInt(1000)
	that := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.Difference(that.set)
	}
}

func BenchmarkConcurrentSet_SymmetricDifference(b *testing.B) {
	this := newCSetInt(1000)
	that := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.SymmetricDifference(that.set)
	}
}

func BenchmarkConcurrentSet_Intersect(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "that is ConcurrentSet",
			test: func(b *testing.B) {
				this := newCSetInt(1000)
				that := newCSetInt(1000)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.set.Intersect(that.set)
				}
			},
		},
		{
			desc: "that is HashSet",
			test: func(b *testing.B) {
				this := newCSetInt(1000)
				that := newHSetInt(1000)

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

func BenchmarkConcurrentSet_Union(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "that is ConcurrentSet",
			test: func(b *testing.B) {
				this := newCSetInt(1000)
				that := newCSetInt(1000)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.set.Union(that.set)
				}
			},
		},
		{
			desc: "that is HashSet",
			test: func(b *testing.B) {
				this := newCSetInt(1000)
				that := newHSetInt(1000)

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

func BenchmarkConcurrentSet_Equals(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "that is ConcurrentSet",
			test: func(b *testing.B) {
				nums := nrand(1000)

				this := NewConcurrentSet[int](nums...)
				that := NewConcurrentSet[int](nums...)

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					this.Equals(that)
				}
			},
		},
		{
			desc: "that is HashSet",
			test: func(b *testing.B) {
				nums := nrand(1000)

				this := NewConcurrentSet[int](nums...)
				that := NewHashSet[int](nums...)

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

func BenchmarkConcurrentSet_IsProperSubset(b *testing.B) {
	this := newCSetInt(1000)
	that := newCSetInt(1001)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.set.IsProperSubset(that.set)
	}
}

func BenchmarkConcurrentSet_Contains(b *testing.B) {
	nums := nrand(1000)

	this := NewConcurrentSet[int](nums...)
	that := NewConcurrentSet[int](nums...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		this.Contains(that)
	}
}

func BenchmarkConcurrentSet_Iter(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "create",
			test: func(b *testing.B) {
				cs := NewConcurrentSet[string]()

				for i := 0; i < 10000; i++ {
					cs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					cs.Iter()
				}
			},
		},
		{
			desc: "loop",
			test: func(b *testing.B) {
				cs := NewConcurrentSet[string]()

				for i := 0; i < 10000; i++ {
					cs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					for item := range cs.Iter() {
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

func BenchmarkConcurrentSet_Items(b *testing.B) {
	scenarios := []struct {
		desc string
		test func(*testing.B)
	}{
		{
			desc: "create",
			test: func(b *testing.B) {
				cs := NewConcurrentSet[string]()

				for i := 0; i < 10000; i++ {
					cs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					cs.Items()
				}
			},
		},
		{
			desc: "loop",
			test: func(b *testing.B) {
				cs := NewConcurrentSet[string]()

				for i := 0; i < 10000; i++ {
					cs.Add(strconv.Itoa(i))
				}

				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					for _, item := range cs.Items() {
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

func BenchmarkConcurrentSet_ForEach(b *testing.B) {
	cs := NewConcurrentSet[string]()

	for i := 0; i < 10000; i++ {
		cs.Add(strconv.Itoa(i))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.ForEach(
			func(element string) bool {
				fmt.Sprintf("%v", element)
				return false
			},
		)
	}
}

func BenchmarkConcurrentSet_MarshalJSON(b *testing.B) {
	cs := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cs.set.MarshalJSON()
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkConcurrentSet_UnmarshalJSON(b *testing.B) {
	jsonStr := []byte("[\"elephant\",\"cow\"]")

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cs := NewConcurrentSet[string]()

		b.StartTimer()
		err := cs.UnmarshalJSON(jsonStr)
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkConcurrentSet_String(b *testing.B) {
	cs := newCSetInt(1000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.set.String()
	}
}
