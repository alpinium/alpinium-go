# Lock example: 10 goroutines and `WorkLock`

Starts 10 goroutines, at most 3 at a time. Each goroutine first does its work
in parallel, and at the end takes `WorkLock` to safely update a shared
variable and print a message. The lock is released with `defer WorkUnlock`.

```go
package main

import "fmt"
import "time"

import Async "github.com/alpinium/alpinium-go/async"

func main() {
    parallelism := Async.NewParallelismWithMaximum(3)
    finished := 0

    for i := 1; i <= 10; i++ {
        work := parallelism.WorkStart()
        go func() {
            defer parallelism.WorkFinish(work)

            // Parallel part: runs in up to 3 goroutines at once.
            time.Sleep(100 * time.Millisecond)

            // Locked part: only one goroutine at a time.
            parallelism.WorkLock(work)
            func() {
                defer parallelism.WorkUnlock(work)
                finished = finished + 1
                fmt.Printf("goroutine %2d (work %2d): done, finished %2d of 10\n", i, work.Id(), finished)
            }()
        }()
    }

    parallelism.Wait()
    fmt.Println("all goroutines finished")
}
```

## What happens

1. `WorkStart()` before `go` registers a work and returns `work`. If 3 works
   are already running, it waits (see [example_simple.md](example_simple.md)).
2. The parallel part (`time.Sleep`) runs in up to 3 goroutines at once.
3. `WorkLock(work)` – only one work at a time may enter the locked part.
   The others wait until the holder releases it.
4. The locked part is inside a lambda function `func() { ... }()` that is
   called immediately. `defer WorkUnlock(work)` inside it releases the lock
   at the end of the lambda function, not at the end of the goroutine.
5. Inside the lock, the shared `finished` counter is safely incremented and
   a message is printed. Without the lock this would be a data race
   (`go run -race` would report it).
6. `Wait()` waits for all works.

## Why a lambda function

In Go, `defer` runs at the end of the **function**, not at the end of a `{}` block.
The lambda function therefore defines how long the lock is held:

```go
defer parallelism.WorkFinish(work) // 2. finishes the work at the end of the goroutine
...
parallelism.WorkLock(work)
func() {
    defer parallelism.WorkUnlock(work) // 1. releases the lock at the end of the lambda function
    ...
}()
// the lock is already released here
```

Code written after the lambda function runs without the lock. If
`defer WorkUnlock` were placed directly in the goroutine (or in a plain `{}`
block), the lock would stay held until the end of the goroutine.

## Possible output

The order of goroutines depends on the scheduler, but the `finished` counter
always increases by 1 – two goroutines never run the locked part at the same time:

```
goroutine  2 (work  2): done, finished  1 of 10
goroutine  3 (work  3): done, finished  2 of 10
goroutine  1 (work  1): done, finished  3 of 10
goroutine  6 (work  6): done, finished  4 of 10
goroutine  5 (work  5): done, finished  5 of 10
goroutine  4 (work  4): done, finished  6 of 10
goroutine  7 (work  7): done, finished  7 of 10
goroutine  9 (work  9): done, finished  8 of 10
goroutine  8 (work  8): done, finished  9 of 10
goroutine 10 (work 10): done, finished 10 of 10
all goroutines finished
```

## Lock rules

- `WorkLock` belongs to a work (`work`), not to a goroutine.
- `WorkLock` is reentrant: the same `work` may take it again, but must then
  call `WorkUnlock` the same number of times.
- `WorkUnlock` with a `work` other than the one holding the lock panics.
- If `WorkUnlock` is missing, `WorkFinish(work)` releases the lock itself.
  An explicit `defer WorkUnlock` is more readable, though, and releases the
  lock right away instead of when the work finishes.
- One work may call `WorkFinish` while another work holds the lock.
- Do not call `WorkStart` while a work holds `WorkLock` – it can deadlock.
