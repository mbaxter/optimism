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
	var stopAtCount int32 = 20
	var count int32 //
	isStarted := func() bool {
		return atomic.LoadInt32(&count) > 0
	}
	isDone := func() bool {
		return atomic.LoadInt32(&count) >= stopAtCount
	}
	increment := func() {
		atomic.AddInt32(&count, 1)
	}

	mu1, mu2 := sync.Mutex{}, sync.Mutex{}
	cond1 := sync.NewCond(&mu1)
	cond2 := sync.NewCond(&mu2)
	ready1, ready2 := true, false // Start with Thread 1 ready

	// Thread 1
	go func() {
		fmt.Println("Start thread 1")
		for {
			if isDone() {
				wg.Done()
				return
			}

			mu1.Lock()
			for !ready1 && !isDone() {
				cond1.Wait()
			}

			// Do some work
			fmt.Println("Thread 1 working")
			increment()

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
			if isDone() {
				wg.Done()
				return
			}

			mu2.Lock()
			for !ready2 && !isDone() {
				cond2.Wait()
			}
			// Do some work
			fmt.Println("Thread 2 working")
			increment()

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
	counter := 0
	go func() {
		fmt.Println("Start thread 3")
		for {
			if isDone() {
				return
			}

			if isStarted() {
				// Only perform thread 3 work concurrently with threads 1 and 2
				fmt.Println("Thread 3 working")
				counter += 1
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
	fmt.Printf("Counter: %v\n", counter)
}
