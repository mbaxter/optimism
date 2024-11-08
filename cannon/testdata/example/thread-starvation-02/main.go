package main

import (
	"fmt"
	"sync"
	"time"
)

type State struct {
	mu                  *sync.Mutex
	cond                *sync.Cond
	wg                  *sync.WaitGroup
	ready               bool
	threadToExecCounter map[int]int
	totalExecCount      int
	targetExecCount     int
}

func main() {
	fmt.Println("Run main")
	threadCount := 5

	var wg sync.WaitGroup
	wg.Add(threadCount)
	var mu sync.Mutex
	state := &State{
		mu:                  &mu,
		cond:                sync.NewCond(&mu),
		wg:                  &wg,
		ready:               false,
		threadToExecCounter: make(map[int]int),
		totalExecCount:      0,
		targetExecCount:     50,
	}

	// Worker threads
	// Run a set of workers
	for i := 0; i < threadCount; i++ {
		go worker(i, state)
	}

	// Coordinator thread
	go coordinator(state)

	wg.Wait()
	fmt.Printf("Thread exec stats: %v\n", state.threadToExecCounter)
}

func worker(id int, state *State) {
	for {
		state.mu.Lock()

		if state.totalExecCount >= state.targetExecCount {
			state.wg.Done()
		}

		for !state.ready {
			state.cond.Wait()
		}

		// Do some work
		state.threadToExecCounter[id] += 1
		state.totalExecCount += 1
		fmt.Printf("Worker [id = %v]: Exec count = %v\n", id, state.totalExecCount)

		state.ready = false
		state.mu.Unlock()
	}
}

func coordinator(state *State) {
	for {
		fmt.Println("Coordinator: Sleep")
		time.Sleep(100 * time.Millisecond)
		state.mu.Lock()
		fmt.Println("Coordinator: Signal")
		state.ready = true
		state.cond.Signal() // Only wakes up one goroutine
		//ready = false
		state.mu.Unlock()
		fmt.Println("Coordinator: Unlock")
	}
}
