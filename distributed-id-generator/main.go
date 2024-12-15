package main

import (
	"github.com/distributed_id_generator/benchmark"
)

// const NO_OF_IDS = 1000000 // Number of IDs to generate
const NO_OF_IDS = 1000 // Number of IDs to generate

func main() {

	// // Simple ID generation

	// frequencies := []int{100, 500, 1000, 10000}

	// for _, frequency := range frequencies {
	// 	start := time.Now()

	// 	for i := 0; i < NO_OF_IDS; i++ {
	// 		simple_id.GenerateID(frequency)
	// 	}

	// 	elapsedTime := time.Since(start).Milliseconds()
	// 	fmt.Printf("Generated %d IDs in %d ms | Save Frequency: %d\n", NO_OF_IDS, elapsedTime, frequency)
	// }

	// // Central ID service: Amazon ID generation

	// frequencies = []int{100, 500, 1000, 10000}

	// for _, frequency := range frequencies {
	// 	start := time.Now()
	// 	// Round to the nearest integer
	// 	NO_OF_ITERATIONS := int(math.Round(float64(NO_OF_IDS) / float64(frequency)))
	// 	for i := 0; i < NO_OF_ITERATIONS; i++ {
	// 		central_id_service.GenerateIDAmazon("order", frequency)
	// 	}
	// 	elapsedTime := time.Since(start).Milliseconds()
	// 	fmt.Printf("Generated %d IDs in %d ms | Generation Frequency: %d\n", NO_OF_IDS, elapsedTime, frequency)
	// }

	// // Central ID service: Flicker ID generation

	// MODES := []string{"1", "2", "3"}

	// for _, mode := range MODES {
	// 	start := time.Now()

	// 	for i := 0; i < NO_OF_IDS; i++ {
	// 		central_id_service.GenerateIDFlicker(mode)
	// 	}

	// 	elapsedTime := time.Since(start).Milliseconds()
	// 	fmt.Printf("Generated %d IDs in %d ms | Mode: %s\n", NO_OF_IDS, elapsedTime, mode)
	// }

	// // Central ID service: Snowflake ID generation

	// epoch := "01-01-2015"

	// start := time.Now()

	// for i := 0; i < NO_OF_IDS; i++ {
	// 	central_id_service.GenerateIDSnowFlake(epoch)
	// }

	// elapsedTime := time.Since(start).Milliseconds()

	// fmt.Printf("Generated %d IDs in %d ms | Epoch: %s\n", NO_OF_IDS, elapsedTime, epoch)

	// // Central ID service: Instagram ID generation
	// start = time.Now()
	// central_id_service.GenerateIDSnowFlakeInstagram()
	// elapsedTime = time.Since(start).Milliseconds()

	// fmt.Printf("Generated %d IDs in %d ms | Instagram\n", 15 * 10, elapsedTime)

	// benchmark.InsertBulkTest()
	benchmark.BenchMarkReplaceIntoVsOnDuplicateKeyUpdate()
}
