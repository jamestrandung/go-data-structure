package main

import (
    "fmt"
    "time"
)

type testI interface {
    Do()
}

type dummy struct {}

func (*dummy) Do() {

}

func main() {
    c := make(chan testI, 1)

    // var blank testI
    // var blank *dummy
    // c <- blank
    close(c)

    got, ok := <-c
    fmt.Println(ok)
    fmt.Println(got)
}

func testChan() {
    _, closeFn := createChan()

    closeFn()

    <-time.After(2*time.Second)
}

func createChan() (<-chan int, func()) {
    c := make(chan int)

    go func() {
        defer func() {
            recover()
        }()

        for i := 0 ; i < 100 ; i++ {
            c <- i
        }
    }()

    return c, func() {
        close(c)
    }
}
