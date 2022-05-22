package list

import "container/list"

type LinkedList[T comparable] struct {
    list *list.List
}

func NewLinkedList[T comparable]() LinkedList[T] {
    return LinkedList[T]{
        list: list.New(),
    }
}

func (l LinkedList[T]) Offer(element T) bool {
    l.list.PushBack(element)
    return true
}

func (l LinkedList[T]) Push(element T) {
    l.list.PushFront(element)
}

func (l LinkedList[T]) Pop() (result T, ok bool) {
    if front := l.list.Front(); front != nil {
        result = front.Value.(T)
        ok = true

        l.list.Remove(front)
    }

    return
}

func (l LinkedList[T]) Peek() (result T, ok bool) {
    if front := l.list.Front(); front != nil {
        result = front.Value.(T)
        ok = true
    }

    return
}