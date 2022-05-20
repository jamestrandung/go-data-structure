package cmap

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

func BenchmarkGetShard(b *testing.B) {
	cm := New[string, string]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.getShard(strconv.Itoa(i))
	}
}

func BenchmarkConcurrentMap_SetAll(b *testing.B) {
	data := make(map[int]int)
	for i := 0; i < 1000; i++ {
		data[i] = i
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cm := New[int, int]()

		b.StartTimer()
		cm.SetAll(data)
	}
}

func BenchmarkSingleInsertAbsent(b *testing.B) {
	cm := New[string, string]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.Set(strconv.Itoa(i), "value")
	}
}

func BenchmarkSingleInsertPresent(b *testing.B) {
	cm := New[string, string]()
	cm.Set("key", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.Set("key", "value")
	}
}

func benchmarkMultiInsertDifferent(b *testing.B, shardCount int) {
	cm := NewWithConcurrencyLevel[string, string](shardCount)

	finished := make(chan struct{}, b.N)
	_, set := getSet(cm, finished)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i), "value")
	}

	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiInsertDifferent_1_Shard(b *testing.B) {
	benchmarkMultiInsertDifferent(b, 1)
}
func BenchmarkMultiInsertDifferent_16_Shard(b *testing.B) {
	benchmarkMultiInsertDifferent(b, 16)
}
func BenchmarkMultiInsertDifferent_32_Shard(b *testing.B) {
	benchmarkMultiInsertDifferent(b, 32)
}
func BenchmarkMultiInsertDifferent_256_Shard(b *testing.B) {
	benchmarkMultiInsertDifferent(b, 256)
}

func BenchmarkMultiInsertSame(b *testing.B) {
	cm := New[string, string]()

	finished := make(chan struct{}, b.N)
	_, set := getSet(cm, finished)
	cm.Set("key", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go set("key", "value")
	}

	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSame(b *testing.B) {
	cm := New[string, string]()

	finished := make(chan struct{}, b.N)
	get, _ := getSet(cm, finished)
	cm.Set("key", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go get("key", "value")
	}

	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func benchmarkMultiGetSetDifferent(b *testing.B, shardCount int) {
	cm := NewWithConcurrencyLevel[string, string](shardCount)

	finished := make(chan struct{}, 2*b.N)
	get, set := getSet(cm, finished)
	cm.Set("-1", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i-1), "value")
		go get(strconv.Itoa(i), "value")
	}

	for i := 0; i < 2*b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSetDifferent_1_Shard(b *testing.B) {
	benchmarkMultiGetSetDifferent(b, 1)
}
func BenchmarkMultiGetSetDifferent_16_Shard(b *testing.B) {
	benchmarkMultiGetSetDifferent(b, 16)
}
func BenchmarkMultiGetSetDifferent_32_Shard(b *testing.B) {
	benchmarkMultiGetSetDifferent(b, 32)
}
func BenchmarkMultiGetSetDifferent_256_Shard(b *testing.B) {
	benchmarkMultiGetSetDifferent(b, 256)
}

func benchmarkMultiGetSetBlock(b *testing.B, shardCount int) {
	cm := NewWithConcurrencyLevel[string, string](shardCount)

	finished := make(chan struct{}, 2*b.N)
	get, set := getSet(cm, finished)

	for i := 0; i < b.N; i++ {
		cm.Set(strconv.Itoa(i%100), "value")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go set(strconv.Itoa(i%100), "value")
		go get(strconv.Itoa(i%100), "value")
	}

	for i := 0; i < 2*b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSetBlock_1_Shard(b *testing.B) {
	benchmarkMultiGetSetBlock(b, 1)
}
func BenchmarkMultiGetSetBlock_16_Shard(b *testing.B) {
	benchmarkMultiGetSetBlock(b, 16)
}
func BenchmarkMultiGetSetBlock_32_Shard(b *testing.B) {
	benchmarkMultiGetSetBlock(b, 32)
}
func BenchmarkMultiGetSetBlock_256_Shard(b *testing.B) {
	benchmarkMultiGetSetBlock(b, 256)
}

func BenchmarkCreateItems(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.Items()
	}
}

func BenchmarkCreateIter(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.Iter()
	}
}

func BenchmarkLoopItems(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, value := range cm.Items() {
			fmt.Sprintf("%v", value)
		}
	}
}

func BenchmarkLoopIter(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for item := range cm.Iter() {
			fmt.Sprintf("%v", item.Val)
		}
	}
}

func BenchmarkLoopForEach(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.ForEach(
			func(key string, val animal) bool {
				fmt.Sprintf("%v", val)
				return false
			},
		)
	}
}

func BenchmarkLoopAsMap(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := cm.AsMap()
		for _, val := range m {
			fmt.Sprintf("%v", val)
		}
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	cm := New[string, animal]()

	for i := 0; i < 1000; i++ {
		cm.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cm.MarshalJSON()
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	jsonStr := []byte("{\"a\":{\"Name\":\"elephant\"},\"b\":{\"Name\":\"cow\"}}")

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		cm := New[string, animal]()

		b.StartTimer()
		err := cm.UnmarshalJSON(jsonStr)
		if err != nil {
			b.FailNow()
		}
	}
}

func getSet[K comparable, V any](cm ConcurrentMap[K, V], finished chan struct{}) (set func(key K, value V), get func(key K, value V)) {
	return func(key K, value V) {
			for i := 0; i < 10; i++ {
				cm.Get(key)
			}

			finished <- struct{}{}

		}, func(key K, value V) {
			for i := 0; i < 10; i++ {
				cm.Set(key, value)
			}

			finished <- struct{}{}
		}
}
