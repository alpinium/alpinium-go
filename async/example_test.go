package Async_test

import "fmt"

import Async "github.com/alpinium/alpinium-go/async"

// Run at most 2 workers at once and wait for all of them.
// WorkStart is called before the goroutine is started: it blocks while
// 2 workers are running, and Wait cannot miss a worker that has not started yet.
func ExampleParallelism() {
    p := Async.NewParallelismWithMaximum(2)
    results := make([]int, 5)
    for i := range results {
        work := p.WorkStart()
        go func() {
            defer p.WorkFinish(work)
            results[i] = i * i
        }()
    }
    p.Wait()
    fmt.Println(results)
    // Output: [0 1 4 9 16]
}

// WorkFinish releases WorkLock if the finishing work still holds it.
func ExampleParallelism_WorkFinish() {
    p := Async.NewParallelism()
    total := 0
    for i := 1; i <= 3; i++ {
        work := p.WorkStart()
        go func() {
            defer p.WorkFinish(work) // also releases WorkLock below
            p.WorkLock(work)
            total += i
        }()
    }
    p.Wait()
    fmt.Println(total)
    // Output: 6
}

// WorkLock is reentrant: the work that holds it may lock it again
// and must then call WorkUnlock the same number of times.
func ExampleParallelism_WorkLock() {
    p := Async.NewParallelism()
    work := p.WorkStart()
    p.WorkLock(work)
    p.WorkLock(work)
    fmt.Println("locked twice")
    p.WorkUnlock(work)
    p.WorkUnlock(work)
    fmt.Println("unlocked")
    p.WorkFinish(work)
    // Output:
    // locked twice
    // unlocked
}

// Each work gets a unique, auto-incremented number.
func ExampleParallelismWork_Id() {
    p := Async.NewParallelism()
    first := p.WorkStart()
    second := p.WorkStart()
    fmt.Println(first.Id(), second.Id())
    p.WorkFinish(first)
    p.WorkFinish(second)
    // Output: 1 2
}
