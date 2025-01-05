package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const MAX_INT = 1000000000
const CONCURRENCY = 8

var totalPrimeNumbers int32 = 0
var currentNumber int32 = 2

func checkPrime(x int) {

	if x&1 == 0 {
		return
	}

	for i := 3; i*i <= x; i++ {
		if x%i == 0 {
			return
		}
	}

	atomic.AddInt32(&totalPrimeNumbers, 1)

}

func countPrimeNumbersSequentially() {
	start := time.Now()
	for i := 0; i < MAX_INT; i++ {
		checkPrime(i)
	}
	fmt.Println("Time taken: ", time.Since(start))
	fmt.Println("Total prime numbers: ", totalPrimeNumbers)
}

func doBatch(name string, wg *sync.WaitGroup, nstart int, nend int) {
	defer wg.Done()
	start := time.Now()
	for i := nstart; i < nend; i++ {
		checkPrime(i)
	}
	fmt.Printf("Thread Id %s, Time taken: %v\n", name, time.Since(start))
}

func doWork(name string, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	for {
		n := atomic.AddInt32(&currentNumber, 1)
		if n > MAX_INT {
			break
		}
		checkPrime(int(n))
	}
	fmt.Printf("Thread Id %s, Time taken: %v\n", name, time.Since(start))
}

func countPrimeNumbersParallelly() {
	start := time.Now()
	var wg sync.WaitGroup
	batchSize := MAX_INT / CONCURRENCY

	for i := 0; i < CONCURRENCY; i++ {
		wg.Add(1)
		go doBatch(fmt.Sprintf("%d", i), &wg, i*batchSize, (i+1)*batchSize)
	}

	wg.Wait()
	fmt.Println("Time taken: ", time.Since(start))
}

func countPrimeNumbersFairParallelly() {
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < CONCURRENCY; i++ {
		wg.Add(1)
		go doWork(fmt.Sprintf("%d", i), &wg)
	}

	wg.Wait()
	fmt.Println("Time taken: ", time.Since(start))
}

func main() {
	// fmt.Println("=========================")
	// countPrimeNumbersSequentially()
	fmt.Println("=========================")
	countPrimeNumbersParallelly()
	fmt.Println("=========================")
	countPrimeNumbersFairParallelly()
}
