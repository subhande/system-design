package main

import (
	"fmt"
	"math/rand"
	"sync"
)

type ConcurrentQueue struct {
	queue []int32
	mu    sync.Mutex
}

var wgE = sync.WaitGroup{}
var wgD = sync.WaitGroup{}

func (cq *ConcurrentQueue) Enqueue(n int32) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	cq.queue = append(cq.queue, n)
}

func (cq *ConcurrentQueue) Dequeue() int32 {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	if len(cq.queue) == 0 {
		panic("Queue is empty")
	}
	item := cq.queue[0]
	cq.queue = cq.queue[1:]
	return item

}

func (cq *ConcurrentQueue) size() int {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return len(cq.queue)
}

func main() {
	NO_OF_ITEMS := 100000

	cq := ConcurrentQueue{}

	for i := 0; i < NO_OF_ITEMS; i++ {
		wgE.Add(1)
		go func(i int) {
			defer wgE.Done()
			cq.Enqueue(rand.Int31())
		}(i)
	}

	for i := 0; i < NO_OF_ITEMS; i++ {
		wgD.Add(1)
		go func(i int) {
			defer wgD.Done()
			cq.Dequeue()
		}(i)
	}

	wgE.Wait()
	wgD.Wait()

	fmt.Printf("Queue size: %d\n", cq.size())
}
