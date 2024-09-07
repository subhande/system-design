package main

import "time"

func main() {
	InitLB()
	go lb.Run()
	for {
		// sleep 1s
		time.Sleep(1 * time.Second)
	}
}
