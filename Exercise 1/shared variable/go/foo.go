// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    //"time"
)

var i = 0


func incrementing(ch chan int, finish_ch chan int) {
    //TODO: increment i 1000000 times
    for k := 0; k < 1000000; k++{
        ch <- 1
    }
    finish_ch <- 1
}

func decrementing(ch chan int, finish_ch chan int) {
    //TODO: decrement i 1000000 times
    for k := 0; k < 1000000; k++{
        ch <- -1
    }
    finish_ch <- 1
}

func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(3)    

    ch := make(chan int) 
    finish_ch := make(chan int)
	
    // TODO: Spawn both functions as goroutines
    go incrementing(ch, finish_ch)
    go decrementing(ch, finish_ch)
	
    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    finished_threads := 0
    
    for(finished_threads < 2){
        select{
        case v := <- ch:
            i += v
        case <- finish_ch:
            finished_threads++
        }
    }
    Println("The magic number is:", i)
}
