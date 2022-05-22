package set

import (
	"math/rand"
	"testing"

    "github.com/jamestrandung/go-data-structure/ds"
    "github.com/stretchr/testify/assert"
)

func verifySetEquals[T comparable](t *testing.T, a, b ds.Set[T]) {
	assert.True(t, a.Equals(b), "%v != %v\n", a, b)
}

func nrand(n int) []int {
	i := make([]int, n)

	for ind := range i {
		i[ind] = rand.Int()
	}

	return i
}
