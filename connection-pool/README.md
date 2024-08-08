# Connection Pool


## Introduction

Connection Pool is a design pattern used to manage a pool of connections to a database. It is used to reduce the overhead of opening and closing connections to the database. It is a creational pattern that provides a way to reuse the existing connection without creating a new one every time.


## Test

Environment: local

Test: Number of Connections local MYSQL databale can handle without connection pool and with connection pool.

QUERY: `SELECT SLEEP(0.01)`

### Results

| N Connections | Non Pool (ms)               | Pool (ms) |
|---------------|-----------------------------|-----------|
| 10            | 45.56                       | 46.59     |
| 100           | 72.40                       | 364.41    |
| 200           | Error: Too Many Connections | 766.42    |
| 500           | Error: Too Many Connections | 1796      |
| 1000          | Error: Too Many Connections | 3663      |
| 10000         | Error: Too Many Connections | 35.776    |

