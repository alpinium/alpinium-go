# Simple example: 10 goroutines

Starts 10 goroutines, at most 3 at a time, and waits until all of them finish.

```go
package main

import "fmt"
import "time"

import Async "github.com/alpinium/alpinium-go/async"

func main() {
    parallelism := Async.NewParallelismWithMaximum(3)

    for i := 1; i <= 10; i++ {
        work := parallelism.WorkStart()
        go func() {
            defer parallelism.WorkFinish(work)
            fmt.Printf("goroutine %2d: start (work %d)\n", i, work.Id())
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("goroutine %2d: end\n", i)
        }()
    }

    parallelism.Wait()
    fmt.Println("all goroutines finished")
}
```

## What happens

1. `NewParallelismWithMaximum(3)` – at most 3 works run at the same time.
   `NewParallelism()` or a maximum of `0` means no limit.
2. `WorkStart()` registers a new work and returns a `*ParallelismWork`
   with a unique number (`work.Id()` = 1, 2, 3, …). If 3 works are already
   running, `WorkStart()` waits until one of them finishes.
3. `defer parallelism.WorkFinish(work)` finishes the work even if the goroutine
   exits early (`return`) or panics. This frees a slot for the next
   `WorkStart()`.
4. `Wait()` waits until all running works have finished.

## Possible output

The order depends on the scheduler, but no more than 3 goroutines ever run at once:

```
goroutine  2: start (work 2)
goroutine  1: start (work 1)
goroutine  3: start (work 3)
goroutine  3: end
goroutine  4: start (work 4)
goroutine  2: end
goroutine  1: end
goroutine  5: start (work 5)
goroutine  6: start (work 6)
...
goroutine 10: end
all goroutines finished
```

## Why `WorkStart()` before `go`, not inside the goroutine

```go
// WRONG
for i := 1; i <= 10; i++ {
    go func() {
        work := parallelism.WorkStart()
        defer parallelism.WorkFinish(work)
        // ...
    }()
}
parallelism.Wait()
```

`Wait()` only waits for works that have already called `WorkStart()`. If the
main goroutine reaches `Wait()` before the other goroutines have even started,
`Wait()` sees 0 running works and returns immediately. This is a race that
depends on the scheduler, so it may not show up in tests – but it can in
production. The same rule applies to `sync.WaitGroup.Add`.

Calling `WorkStart()` before `go` has another benefit: the loop does not start
a new goroutine until a slot is free, so you don't end up with 10 goroutines
of which 7 are just waiting.

## Rules for `work`

- Every `work` returned by `WorkStart()` must be passed to `WorkFinish()` exactly once.
- `WorkFinish()` may be called from a different goroutine than the one that called `WorkStart()`.
- `WorkFinish()` panics when given `nil`, a `work` from another `Parallelism`,
  an already finished `work`, or a `work` that was not created by `WorkStart()`.
