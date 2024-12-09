package simple_id

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	PATH = "id.txt" // Path to the file where the last saved ID is stored
)

type StaticCounter struct {
	counter int
	mu      sync.Mutex // Mutex to ensure thread-safe counter updates
}

var staticCounter = StaticCounter{counter: 0}

// restoreLastSavedID restores the last saved ID from the file and increments it by the offset.
// If the file does not exist, it returns the offset.
func restoreLastSavedID(offset int) int {
	data, err := os.ReadFile(PATH)
	if err != nil {
		if os.IsNotExist(err) {
			return offset
		}
		fmt.Println("Error reading file:", err)
		return offset
	}

	lastSavedID, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Println("Error converting file data to int:", err)
		return offset
	}

	return lastSavedID + offset
}

// generateID generates a unique ID using the current time, a random machine ID, and a counter.
func GenerateID(saveFrequency int) string {
	staticCounter.mu.Lock()
	defer staticCounter.mu.Unlock()

	// If counter is 0 and the file exists, restore the last saved ID
	if staticCounter.counter == 0 {
		staticCounter.counter = restoreLastSavedID(saveFrequency)
	}

	// Get the current time in milliseconds since epoch
	epochMs := time.Now().UnixMilli()

	// Generate a random machine ID between 0 and 10, padded with zeros to 2 digits
	machineID := fmt.Sprintf("%02d", rand.Intn(11))

	// Create the ID string
	id := fmt.Sprintf("%d%s%04d", epochMs, machineID, staticCounter.counter)

	// If the counter is a multiple of saveFrequency, save the counter to the file
	if staticCounter.counter%saveFrequency == 0 {
		err := os.WriteFile(PATH, []byte(strconv.Itoa(staticCounter.counter)), 0644)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}

	// Increment the counter
	staticCounter.counter++
	return id
}
