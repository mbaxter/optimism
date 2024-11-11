package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	procCount := runtime.GOMAXPROCS(0)
	workerCount := procCount
	if workerCount < 3 {
		workerCount = 3
	}
	fmt.Printf("Procs: %v | Workers: %v\n", procCount, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Start three goroutines, each locked to an OS thread
	for i := 1; i <= workerCount; i++ {
		go worker(i, &wg)
	}

	wg.Wait()
	fmt.Println("All goroutines finished.")
}

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	runtime.LockOSThread() // Lock this goroutine to the current OS thread
	defer runtime.UnlockOSThread()

	for i := 0; i < 5; i++ {
		fmt.Printf("Goroutine %d running on thread\n", id)
		time.Sleep(500 * time.Millisecond)
	}
}
