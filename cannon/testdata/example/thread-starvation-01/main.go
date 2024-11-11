package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	execTracker := NewExecTracker(50)

	mu1, mu2 := sync.Mutex{}, sync.Mutex{}
	cond1 := sync.NewCond(&mu1)
	cond2 := sync.NewCond(&mu2)
	ready1, ready2 := false, false

	// Thread 1
	go func() {
		fmt.Println("Start thread 1")
		for {
			if execTracker.IsDone() {
				wg.Done()
				return
			}

			mu1.Lock()
			for !ready1 && !execTracker.IsDone() {
				cond1.Wait()
			}

			// Do some work
			fmt.Println("Thread 1 working")
			execTracker.Increment()

			// Signal Thread 2
			mu2.Lock()
			ready2 = true
			cond2.Signal()
			mu2.Unlock()

			ready1 = false
			mu1.Unlock()
		}
	}()

	// Thread 2
	go func() {
		fmt.Println("Start thread 2")
		for {
			if execTracker.IsDone() {
				wg.Done()
				return
			}

			mu2.Lock()
			for !ready2 && !execTracker.IsDone() {
				cond2.Wait()
			}
			// Do some work
			fmt.Println("Thread 2 working")
			execTracker.Increment()

			// Signal Thread 1
			mu1.Lock()
			ready1 = true
			cond1.Signal()
			mu1.Unlock()

			ready2 = false
			mu2.Unlock()
		}
	}()

	// Thread 3
	otherThreadCount := 0
	go func() {
		fmt.Println("Start thread 3")
		for {
			// Only run this thread while threads 1 and 2 are executing to see if this thread gets scheduled
			// concurrently
			if execTracker.IsDone() {
				return
			}

			if execTracker.IsStarted() {
				// Only perform thread 3 work concurrently with threads 1 and 2
				fmt.Println("Thread 3 working")
				otherThreadCount += 1
			}

			time.Sleep(0)
		}
	}()

	// Start the ping-pong
	mu1.Lock()
	ready1 = true
	cond1.Signal()
	mu1.Unlock()

	wg.Wait()
	fmt.Printf("Exec count: %v\n", execTracker.GetCount())
	fmt.Printf("Other thread counter: %v\n", otherThreadCount)

	if otherThreadCount == 0 {
		panic("Other thread did not run!")
	}
}

type ExecTracker struct {
	execCount   uint32
	doneAtCount uint32
}

func NewExecTracker(terminationCount uint32) *ExecTracker {
	return &ExecTracker{
		execCount:   0,
		doneAtCount: terminationCount,
	}
}

func (a *ExecTracker) IsStarted() bool {
	return a.GetCount() > 0
}

func (a *ExecTracker) IsDone() bool {
	return a.GetCount() >= a.doneAtCount
}

func (a *ExecTracker) Increment() {
	atomic.AddUint32(&a.execCount, 1)
}

func (a *ExecTracker) GetCount() uint32 {
	return atomic.LoadUint32(&a.execCount)
}
