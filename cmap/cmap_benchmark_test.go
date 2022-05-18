package cmap

import (
	"strconv"
	"testing"
)

func BenchmarkGetShard(b *testing.B) {
	m := New[string, string]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.getShard(strconv.Itoa(i))
	}
}

func BenchmarkSingleInsertAbsent(b *testing.B) {
	m := New[string, string]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Set(strconv.Itoa(i), "value")
	}
}

func BenchmarkSingleInsertPresent(b *testing.B) {
	m := New[string, string]()
	m.Set("key", "value")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Set("key", "value")
	}
}

func benchmarkMultiInsertDifferent(b *testing.B, shardCount int) {
	m := NewWithConcurrencyLevel[string, string](shardCount)

	finished := make(chan struct{}, b.N)
	_, set := getSet(m, finished)

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
	m := New[string, string]()

	finished := make(chan struct{}, b.N)
	_, set := getSet(m, finished)
	m.Set("key", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go set("key", "value")
	}

	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func BenchmarkMultiGetSame(b *testing.B) {
	m := New[string, string]()

	finished := make(chan struct{}, b.N)
	get, _ := getSet(m, finished)
	m.Set("key", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go get("key", "value")
	}

	for i := 0; i < b.N; i++ {
		<-finished
	}
}

func benchmarkMultiGetSetDifferent(b *testing.B, shardCount int) {
	m := NewWithConcurrencyLevel[string, string](shardCount)

	finished := make(chan struct{}, 2*b.N)
	get, set := getSet(m, finished)
	m.Set("-1", "value")

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
	m := NewWithConcurrencyLevel[string, string](shardCount)

	finished := make(chan struct{}, 2*b.N)
	get, set := getSet(m, finished)

	for i := 0; i < b.N; i++ {
		m.Set(strconv.Itoa(i%100), "value")
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

func BenchmarkItems(b *testing.B) {
	m := New[string, animal]()

	for i := 0; i < 10000; i++ {
		m.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	for i := 0; i < b.N; i++ {
		m.Items()
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	m := New[string, animal]()

	for i := 0; i < 10000; i++ {
		m.Set(strconv.Itoa(i), animal{strconv.Itoa(i)})
	}

	for i := 0; i < b.N; i++ {
		_, err := m.MarshalJSON()
		if err != nil {
			b.FailNow()
		}
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	m := New[string, animal]()

	jsonStr := []byte("{\"a\":{\"Name\":\"elephant\"},\"b\":{\"Name\":\"cow\"}}")

	for i := 0; i < b.N; i++ {
		err := m.UnmarshalJSON(jsonStr)
		if err != nil {
			b.FailNow()
		}
	}
}

func getSet[K comparable, V any](m ConcurrentMap[K, V], finished chan struct{}) (set func(key K, value V), get func(key K, value V)) {
	return func(key K, value V) {
			for i := 0; i < 10; i++ {
				m.Get(key)
			}

			finished <- struct{}{}

		}, func(key K, value V) {
			for i := 0; i < 10; i++ {
				m.Set(key, value)
			}

			finished <- struct{}{}
		}
}
