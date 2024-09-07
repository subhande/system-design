# Load Balancer

## Instructions

```
go mod github.com/load_balancer
go mod tidy

// Install python dependencies
PORT=8081 python main.py
PORT=8082 python main.py
PORT=8083 python main.py
PORT=8084 python main.py

go run *.go

// Test
python load_balancer_test.py
// Results will be saved in loadtest_results.json
```