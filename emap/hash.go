package emap

import (
	"fmt"
	"github.com/mitchellh/hashstructure/v2"
)

// Hasher can be implemented by the type to be used as keys in case clients want to provide
// their own hashing algo. Note that this hash does NOT control which bucket will hold a key
// in the actual map in each shard. Instead, it is used to distribute keys into the underlying
// shards which can be locked/unlocked independently by many goroutines.
//
// To guarantee good performance for ConcurrentMap, clients must make sure their hashing
// algo can evenly distribute keys across shards. Otherwise, a hot shard may become a
// bottleneck when many goroutines try to lock it at the same time.
//
// Note: we prioritize availability over correctness. If clients' hashing algo panics, the
// library will automatically recover and return 0, meaning all panicked keys will go into
// the 1st shard. As a consequence, subsequent concurrent accesses to these keys will happen
// on the same shard and performance may suffer. However, client applications will still work.
type Hasher interface {
	Hash() uint64
}

// HashString converts the given string into a hash value.
func HashString(s string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)

	keyLength := len(s)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(s[i])
	}

	return hash
}

func hashHasher(key Hasher) uint64 {
	defer func() {
		// Fall back to 0 if panics
		if r := recover(); r != nil {
			fmt.Println("ConcurrentMap - Panic recovered:", r)
		}
	}()

	return key.Hash()
}

var hashFn = hashstructure.Hash

func hashAny(key interface{}) uint64 {
	defer func() {
		// Fall back to 0 if panics
		if r := recover(); r != nil {
			fmt.Println("ConcurrentMap - Panic recovered:", r)
		}
	}()

	hash, err := hashFn(key, hashstructure.FormatV2, &hashstructure.HashOptions{UseStringer: true})
	if err != nil {
		fmt.Println("ConcurrentMap - Error encountered:", err)

		// Use the 1st shard as fallback in case hashing fails
		return 0
	}

	return hash
}
