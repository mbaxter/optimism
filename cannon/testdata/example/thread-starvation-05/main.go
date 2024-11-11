package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	dataCh := make(chan int, 100)

	// Producer
	wg.Add(1)
	go producer(dataCh, &wg)

	// Consumers
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go consumer(i, dataCh, &wg)
	}

	wg.Wait()
}

func producer(dataCh chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Start producer")
	for i := 0; i < 1000; i++ {
		dataCh <- i
		// Producer is fast
	}
	fmt.Println("Producer done")
	close(dataCh)
}

func consumer(id int, dataCh chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Start consumer %v\n", id)
	count := 0
	for data := range dataCh {
		// Simulate slow processing
		_ = data
		// No sleep or yield here
		count++
	}
	fmt.Printf("Consumer %v done after %v steps\n", id, count)
}
