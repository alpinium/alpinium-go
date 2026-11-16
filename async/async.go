package Async

import "fmt"
import "sync"

const DEBUG = false

type Parallelism struct {
    current_count uint64
    maximum_count uint64
    work_counter uint64
    work_active map[uint64]*ParallelismWork
    work_mutex * sync.Mutex
    work_cond * sync.Cond
    lock_mutex * sync.Mutex
    lock_cond * sync.Cond
    lock_owner uint64
    lock_depth uint64
}

// ParallelismWork identifies one unit of work started by WorkStart.
// It must be passed to WorkFinish, WorkLock and WorkUnlock.
type ParallelismWork struct {
    parallelism * Parallelism
    id uint64
}

// Id returns the unique, auto-incremented number of the work (starting at 1).
func (this * ParallelismWork) Id() uint64 {
    return this.id
}

func NewParallelism() *Parallelism {
    result := &Parallelism{}
    result.Init(^uint64(0))
    return result
}

func NewParallelismWithMaximum(maximum_count uint64) *Parallelism {
    result := &Parallelism{}
    result.Init(maximum_count)
    return result
}

func (this * Parallelism) Init(maximum_count uint64) {
    if maximum_count == 0 {
        maximum_count = ^uint64(0) // 0 means unlimited
    }
    this.current_count = 0
    this.maximum_count = maximum_count
    this.work_counter = 0
    this.work_active = make(map[uint64]*ParallelismWork)
    this.work_mutex = &sync.Mutex{}
    this.work_cond = sync.NewCond(this.work_mutex)
    this.lock_mutex = &sync.Mutex{}
    this.lock_cond = sync.NewCond(this.lock_mutex)
    this.lock_owner = 0
    this.lock_depth = 0
}

// checkWork panics if work is nil, belongs to another Parallelism, or is not
// a currently running work returned by WorkStart (finished, copied or forged).
// Caller must hold work_mutex.
func (this * Parallelism) checkWork(work * ParallelismWork, function string) {
    if work == nil {
        this.work_mutex.Unlock()
        panic("Async.Parallelism: " + function + " called with nil work")
    }
    if work.parallelism != this {
        this.work_mutex.Unlock()
        panic("Async.Parallelism: " + function + " called with work of another Parallelism")
    }
    if this.work_active[work.id] != work {
        this.work_mutex.Unlock()
        panic(fmt.Sprintf("Async.Parallelism: %s called with work %d that is not running (finished or not started by WorkStart)", function, work.id))
    }
}

func (this * Parallelism) WorkStart() * ParallelismWork {
    if DEBUG {fmt.Printf("WorkStart begin\n")}
    this.work_mutex.Lock()
    for this.current_count >= this.maximum_count {
        this.work_cond.Wait()
    }
    this.current_count = this.current_count + 1
    this.work_counter = this.work_counter + 1
    result := &ParallelismWork{parallelism: this, id: this.work_counter}
    this.work_active[result.id] = result
    this.work_mutex.Unlock()
    if DEBUG {fmt.Printf("WorkStart end %d\n", result.id)}
    return result
}

func (this * Parallelism) WorkFinish(work * ParallelismWork) {
    if DEBUG {fmt.Printf("WorkFinish begin\n")}
    this.work_mutex.Lock()
    this.checkWork(work, "WorkFinish")
    this.work_mutex.Unlock()

    // Release WorkLock if this work holds it; a lock held by another work is left alone.
    this.lock_mutex.Lock()
    if this.lock_depth != 0 && this.lock_owner == work.id {
        this.lock_depth = 0
        this.lock_owner = 0
        this.lock_cond.Signal()
    }
    this.lock_mutex.Unlock()

    this.work_mutex.Lock()
    this.checkWork(work, "WorkFinish") // WorkFinish of the same work may have run concurrently
    delete(this.work_active, work.id)
    this.current_count = this.current_count - 1
    this.work_cond.Broadcast()
    this.work_mutex.Unlock()
    if DEBUG {fmt.Printf("WorkFinish end %d\n", work.id)}
}

// WorkLock is reentrant: the work holding the lock may call WorkLock
// again, and must then call WorkUnlock the same number of times.
func (this * Parallelism) WorkLock(work * ParallelismWork) {
    if DEBUG {fmt.Printf("WorkLock begin\n")}
    this.work_mutex.Lock()
    this.checkWork(work, "WorkLock")
    this.work_mutex.Unlock()

    this.lock_mutex.Lock()
    for this.lock_depth != 0 && this.lock_owner != work.id {
        this.lock_cond.Wait()
    }
    this.lock_owner = work.id
    this.lock_depth = this.lock_depth + 1
    this.lock_mutex.Unlock()
    if DEBUG {fmt.Printf("WorkLock end %d\n", work.id)}
}

func (this * Parallelism) WorkUnlock(work * ParallelismWork) {
    if DEBUG {fmt.Printf("WorkUnlock begin\n")}
    this.work_mutex.Lock()
    this.checkWork(work, "WorkUnlock")
    this.work_mutex.Unlock()

    this.lock_mutex.Lock()
    if this.lock_depth == 0 || this.lock_owner != work.id {
        this.lock_mutex.Unlock()
        panic("Async.Parallelism: WorkUnlock called with work that does not hold the lock")
    }
    this.lock_depth = this.lock_depth - 1
    if this.lock_depth == 0 {
        this.lock_owner = 0
        this.lock_cond.Signal()
    }
    this.lock_mutex.Unlock()
    if DEBUG {fmt.Printf("WorkUnlock end %d\n", work.id)}
}

func (this * Parallelism) Wait() {
    if DEBUG {fmt.Printf("Wait begin\n")}
    this.work_mutex.Lock()
    for this.current_count != 0 {
        this.work_cond.Wait()
    }
    this.work_mutex.Unlock()
    if DEBUG {fmt.Printf("Wait end\n")}
}
