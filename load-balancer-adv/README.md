# Load Balancer

## Overview

This project implements a load balancer that distributes incoming requests across multiple backend servers to ensure efficient resource utilization and high availability. This is L4 (Transport Layer) load balancing, which operates at the network transport layer and forwards requests based on network and transport layer information.

## Instructions

### Setup

1. **Initialize Go Modules:**
    ```sh
    go mod init github.com/load_balancer
    go mod tidy
    ```

2. **Install Python Dependencies:**
    ```sh
    pip install -r requirements.txt
    ```

3. **Start Backend Servers:**
    ```sh
    PORT=8081 python main.py
    PORT=8082 python main.py
    PORT=8083 python main.py
    PORT=8084 python main.py
    ```

4. **Run the Load Balancer:**
    ```sh
    go run *.go
    ```

5. **Run Tests:**
    ```sh
    python load_balancer_test.py
    ```

    Results will be saved in `loadtest_results.json`.

## Features

- [x] **Round Robin Load Balancing:** Distributes requests evenly across all backend servers.