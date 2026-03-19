package main

import (
	"os"
)

// Service Functions

func main() {
	ConnectDB()
	defer CloseAllRabbitMQClients()

	if len(os.Args) <= 1 {
		return
	}

	switch os.Args[1] {
	case "-controler":
		controlerServiceLoop()
	case "-iterator":
		IteratorService()
	case "-worker":
		WorkerService()
	}
}
