package list

import (
    "testing"

    "github.com/jamestrandung/go-data-structure/ds"
)

func TestLinkedListMatchingInterface(t *testing.T) {
    var s ds.Stack[int] = NewLinkedList[int]()
    s.Push(1)

    var q ds.Queue[int] = NewLinkedList[int]()
    q.Offer(1)
}
