package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	state := NewState(3, 3)

	// Long-running thread
	go heavyComputation(state)

	// Other threads
	for j := 0; j < 3; j++ {
		go worker(j, state)
	}

	state.wg.Wait()
	fmt.Printf("Process exec counter: %v\n", state.idToExecCounter)
	fmt.Printf("Heavy counter: %v\n", state.heavyCounter.Get())
}

func heavyComputation(state *State) {
	fmt.Println("Start heavy computation")
	for i := 0; i < 1e9; i++ {
		// Intensive computation
		_ = i * i
		state.heavyCounter.Increment()
		// Program works if logging is uncommented
		//if i%10_000 == 0 {
		//	fmt.Printf("Heavy comp: %v\n", i)
		//}
	}
}

func worker(id int, state *State) {
	for {
		// Do some work
		fmt.Printf("Thread %d running\n", id)
		//runtime.Gosched() // Attempt to yield
		time.Sleep(100 * time.Millisecond)

		if state.incrementAndGetExecCounter(id) >= state.targetExecCount {
			state.wg.Done()
			return
		}
	}
}

type State struct {
	mu              sync.Mutex
	wg              *sync.WaitGroup
	idToExecCounter map[int]int
	targetExecCount int
	heavyCounter    *AtomicCounter
}

func NewState(workerCount, workerTargetCount int) *State {
	var wg sync.WaitGroup
	wg.Add(workerCount)
	return &State{
		wg:              &wg,
		idToExecCounter: make(map[int]int),
		targetExecCount: workerTargetCount,
		heavyCounter:    NewCounter(),
	}
}

func (s *State) incrementAndGetExecCounter(id int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idToExecCounter[id]++
	return s.idToExecCounter[id]
}

type AtomicCounter struct {
	value uint32
}

func NewCounter() *AtomicCounter {
	return &AtomicCounter{value: 0}
}

func (a *AtomicCounter) Increment() {
	atomic.AddUint32(&a.value, 1)
}

func (a *AtomicCounter) Get() uint32 {
	return atomic.LoadUint32(&a.value)
}
