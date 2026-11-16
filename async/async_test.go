package Async

import "sync/atomic"
import "testing"
import "time"

// How long a goroutine must stay blocked to be considered "blocked".
const blockedFor = 50 * time.Millisecond

// How long to wait for something that is expected to finish.
const finishWithin = 5 * time.Second

// run starts f in a new goroutine and returns a channel closed when f returns.
func run(f func()) <-chan struct{} {
    done := make(chan struct{})
    go func() {
        defer close(done)
        f()
    }()
    return done
}

func assertBlocked(t *testing.T, done <-chan struct{}, what string) {
    t.Helper()
    select {
    case <-done:
        t.Fatalf("%s: expected to block, but it returned", what)
    case <-time.After(blockedFor):
    }
}

func assertFinished(t *testing.T, done <-chan struct{}, what string) {
    t.Helper()
    select {
    case <-done:
    case <-time.After(finishWithin):
        t.Fatalf("%s: expected to finish, but it is still blocked", what)
    }
}

// recoverFrom runs f and returns the recovered panic value (nil if none).
func recoverFrom(f func()) (result any) {
    defer func() { result = recover() }()
    f()
    return nil
}

func assertPanics(t *testing.T, f func(), what string) {
    t.Helper()
    if recoverFrom(f) == nil {
        t.Fatalf("%s: expected panic", what)
    }
}

// --- Init / maximum ---------------------------------------------------------

func TestNewParallelismIsUnlimited(t *testing.T) {
    p := NewParallelism()
    works := make([]*ParallelismWork, 0, 10000)
    done := run(func() {
        for i := 0; i < 10000; i++ {
            works = append(works, p.WorkStart())
        }
    })
    assertFinished(t, done, "10000x WorkStart")
    for _, work := range works {
        p.WorkFinish(work)
    }
    p.Wait()
}

func TestMaximumZeroMeansUnlimited(t *testing.T) {
    p := NewParallelismWithMaximum(0)
    works := make([]*ParallelismWork, 0, 10000)
    done := run(func() {
        for i := 0; i < 10000; i++ {
            works = append(works, p.WorkStart())
        }
    })
    assertFinished(t, done, "10000x WorkStart with maximum 0")
    for _, work := range works {
        p.WorkFinish(work)
    }
    p.Wait()
}

// --- WorkStart / WorkFinish ---------------------------------------------------

func TestWorkStartReturnsIncrementingIds(t *testing.T) {
    p := NewParallelism()
    for want := uint64(1); want <= 5; want++ {
        work := p.WorkStart()
        if work.Id() != want {
            t.Fatalf("Id() = %d, want %d", work.Id(), want)
        }
        p.WorkFinish(work)
    }
}

func TestWorkStartIdsAreUniqueAcrossGoroutines(t *testing.T) {
    const workers = 1000
    p := NewParallelism()
    ids := make(chan uint64, workers)
    for i := 0; i < workers; i++ {
        go func() {
            work := p.WorkStart()
            ids <- work.Id()
            p.WorkFinish(work)
        }()
    }
    seen := make(map[uint64]bool)
    for i := 0; i < workers; i++ {
        id := <-ids
        if seen[id] {
            t.Fatalf("duplicate id %d", id)
        }
        seen[id] = true
    }
}

func TestWorkStartBlocksAtMaximum(t *testing.T) {
    p := NewParallelismWithMaximum(2)
    first := p.WorkStart()
    second := p.WorkStart()

    var third *ParallelismWork
    started := run(func() { third = p.WorkStart() })
    assertBlocked(t, started, "WorkStart over maximum")

    p.WorkFinish(first)
    assertFinished(t, started, "WorkStart after WorkFinish")

    p.WorkFinish(second)
    p.WorkFinish(third)
    p.Wait()
}

func TestMaximumLimitsConcurrency(t *testing.T) {
    const maximum = 4
    const workers = 200
    p := NewParallelismWithMaximum(maximum)
    var active, peak int32
    for i := 0; i < workers; i++ {
        work := p.WorkStart()
        go func() {
            n := atomic.AddInt32(&active, 1)
            for {
                old := atomic.LoadInt32(&peak)
                if n <= old || atomic.CompareAndSwapInt32(&peak, old, n) {
                    break
                }
            }
            time.Sleep(time.Millisecond)
            atomic.AddInt32(&active, -1)
            p.WorkFinish(work)
        }()
    }
    p.Wait()
    if active != 0 {
        t.Fatalf("Wait returned with %d workers still active", active)
    }
    if peak > maximum {
        t.Fatalf("peak concurrency %d exceeds maximum %d", peak, maximum)
    }
}

func TestWorkFinishInAnotherGoroutine(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    assertFinished(t, run(func() { p.WorkFinish(work) }), "WorkFinish in another goroutine")
    p.Wait()
}

func TestWorkFinishTwicePanics(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    p.WorkFinish(work)
    assertPanics(t, func() { p.WorkFinish(work) }, "second WorkFinish")
}

func TestWorkFinishWithNilPanics(t *testing.T) {
    p := NewParallelism()
    assertPanics(t, func() { p.WorkFinish(nil) }, "WorkFinish(nil)")
}

func TestWorkFinishWithWorkOfAnotherParallelismPanics(t *testing.T) {
    p := NewParallelism()
    other := NewParallelism()
    work := other.WorkStart()
    assertPanics(t, func() { p.WorkFinish(work) }, "WorkFinish with foreign work")
    other.WorkFinish(work)
}

func TestForgedWorkPanics(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    forged := &ParallelismWork{parallelism: p, id: work.Id()} // same id, not from WorkStart
    assertPanics(t, func() { p.WorkLock(forged) }, "WorkLock with forged work")
    assertPanics(t, func() { p.WorkUnlock(forged) }, "WorkUnlock with forged work")
    assertPanics(t, func() { p.WorkFinish(forged) }, "WorkFinish with forged work")
    p.WorkFinish(work)
}

func TestCopiedWorkPanics(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    copied := *work
    assertPanics(t, func() { p.WorkFinish(&copied) }, "WorkFinish with copied work")
    p.WorkFinish(work)
}

func TestUnknownWorkIdPanics(t *testing.T) {
    p := NewParallelism()
    unknown := &ParallelismWork{parallelism: p, id: 42}
    assertPanics(t, func() { p.WorkLock(unknown) }, "WorkLock with unknown id")
    assertPanics(t, func() { p.WorkUnlock(unknown) }, "WorkUnlock with unknown id")
    assertPanics(t, func() { p.WorkFinish(unknown) }, "WorkFinish with unknown id")
}

func TestWorkFinishRemovesWorkFromRegistry(t *testing.T) {
    p := NewParallelism()
    first := p.WorkStart()
    second := p.WorkStart()
    if len(p.work_active) != 2 {
        t.Fatalf("running works = %d, want 2", len(p.work_active))
    }
    p.WorkFinish(first)
    p.WorkFinish(second)
    if len(p.work_active) != 0 {
        t.Fatalf("running works = %d, want 0", len(p.work_active))
    }
}

func TestWorkUnlockWithFinishedWorkPanics(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    p.WorkFinish(work)
    assertPanics(t, func() { p.WorkUnlock(work) }, "WorkUnlock with finished work")
}

// --- Wait --------------------------------------------------------------------

func TestWaitWithoutWorkReturnsImmediately(t *testing.T) {
    p := NewParallelism()
    assertFinished(t, run(p.Wait), "Wait without work")
}

func TestWaitWaitsForAllWork(t *testing.T) {
    p := NewParallelism()
    first := p.WorkStart()
    second := p.WorkStart()

    waiting := run(p.Wait)
    assertBlocked(t, waiting, "Wait with 2 running")

    p.WorkFinish(first)
    assertBlocked(t, waiting, "Wait with 1 running")

    p.WorkFinish(second)
    assertFinished(t, waiting, "Wait after all finished")
}

func TestMultipleWaiters(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    first := run(p.Wait)
    second := run(p.Wait)
    assertBlocked(t, first, "first Wait")
    assertBlocked(t, second, "second Wait")
    p.WorkFinish(work)
    assertFinished(t, first, "first Wait")
    assertFinished(t, second, "second Wait")
}

// --- WorkLock / WorkUnlock ----------------------------------------------------

func TestWorkLockIsReentrant(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    done := run(func() {
        p.WorkLock(work)
        p.WorkLock(work)
        p.WorkLock(work)
        p.WorkUnlock(work)
        p.WorkUnlock(work)
        p.WorkUnlock(work)
    })
    assertFinished(t, done, "3x WorkLock + 3x WorkUnlock")
    p.WorkFinish(work)
}

func TestWorkLockIsReentrantAcrossGoroutinesOfSameWork(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    p.WorkLock(work)
    assertFinished(t, run(func() {
        p.WorkLock(work)
        p.WorkUnlock(work)
    }), "WorkLock of the same work in another goroutine")
    p.WorkUnlock(work)
    p.WorkFinish(work)
}

func TestWorkLockHeldUntilLastWorkUnlock(t *testing.T) {
    p := NewParallelism()
    owner := p.WorkStart()
    other := p.WorkStart()
    p.WorkLock(owner)
    p.WorkLock(owner)

    locked := run(func() {
        p.WorkLock(other)
        p.WorkUnlock(other)
    })
    assertBlocked(t, locked, "WorkLock while held twice")

    p.WorkUnlock(owner)
    assertBlocked(t, locked, "WorkLock while held once")

    p.WorkUnlock(owner)
    assertFinished(t, locked, "WorkLock after last WorkUnlock")

    p.WorkFinish(owner)
    p.WorkFinish(other)
}

func TestWorkLockMutualExclusion(t *testing.T) {
    const workers = 500
    p := NewParallelismWithMaximum(8)
    var inside, violations int32
    counter := 0
    for i := 0; i < workers; i++ {
        work := p.WorkStart()
        go func() {
            p.WorkLock(work)
            p.WorkLock(work)
            if atomic.AddInt32(&inside, 1) != 1 {
                atomic.AddInt32(&violations, 1)
            }
            counter++ // protected only by WorkLock; -race reports if exclusion fails
            atomic.AddInt32(&inside, -1)
            p.WorkUnlock(work)
            p.WorkUnlock(work)
            p.WorkFinish(work)
        }()
    }
    p.Wait()
    if violations != 0 {
        t.Fatalf("%d works were inside WorkLock at the same time", violations)
    }
    if counter != workers {
        t.Fatalf("counter = %d, want %d", counter, workers)
    }
}

func TestWorkUnlockWithoutWorkLockPanics(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    assertPanics(t, func() { p.WorkUnlock(work) }, "WorkUnlock without WorkLock")
    p.WorkFinish(work)
}

func TestWorkUnlockWithOtherWorkPanics(t *testing.T) {
    p := NewParallelism()
    owner := p.WorkStart()
    other := p.WorkStart()
    p.WorkLock(owner)
    assertPanics(t, func() { p.WorkUnlock(other) }, "WorkUnlock with other work")
    p.WorkUnlock(owner) // the owner can still unlock
    p.WorkFinish(owner)
    p.WorkFinish(other)
}

func TestWorkLockWithFinishedWorkPanics(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    p.WorkFinish(work)
    assertPanics(t, func() { p.WorkLock(work) }, "WorkLock with finished work")
}

func TestWorkLockWithNilPanics(t *testing.T) {
    p := NewParallelism()
    assertPanics(t, func() { p.WorkLock(nil) }, "WorkLock(nil)")
    assertPanics(t, func() { p.WorkUnlock(nil) }, "WorkUnlock(nil)")
}

// --- WorkFinish + WorkLock ----------------------------------------------------

func TestWorkFinishReleasesOwnWorkLock(t *testing.T) {
    p := NewParallelism()
    work := p.WorkStart()
    p.WorkLock(work)
    p.WorkLock(work)
    p.WorkFinish(work) // no WorkUnlock: WorkFinish releases the lock completely

    next := p.WorkStart()
    assertFinished(t, run(func() {
        p.WorkLock(next)
        p.WorkUnlock(next)
    }), "WorkLock after WorkFinish released it")
    p.WorkFinish(next)
    p.Wait()
}

func TestWorkFinishWhileOtherWorkHoldsWorkLock(t *testing.T) {
    p := NewParallelism()
    owner := p.WorkStart()
    other := p.WorkStart()
    p.WorkLock(owner)

    // Allowed: other work finishes (in another goroutine) while owner holds the lock.
    assertFinished(t, run(func() { p.WorkFinish(other) }), "WorkFinish while other work holds lock")

    // The lock is still held by owner.
    next := p.WorkStart()
    locked := run(func() {
        p.WorkLock(next)
        p.WorkUnlock(next)
    })
    assertBlocked(t, locked, "WorkLock while owner still holds it")

    p.WorkUnlock(owner)
    assertFinished(t, locked, "WorkLock after owner unlocked")
    p.WorkFinish(owner)
    p.WorkFinish(next)
    p.Wait()
}

// --- Benchmarks ---------------------------------------------------------------

func BenchmarkWorkStartFinish(b *testing.B) {
    p := NewParallelism()
    for i := 0; i < b.N; i++ {
        p.WorkFinish(p.WorkStart())
    }
}

func BenchmarkWorkLockUnlock(b *testing.B) {
    p := NewParallelism()
    work := p.WorkStart()
    for i := 0; i < b.N; i++ {
        p.WorkLock(work)
        p.WorkUnlock(work)
    }
    p.WorkFinish(work)
}
