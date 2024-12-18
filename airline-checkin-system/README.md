# Airline Checkin System


## Overview

The Airline Checkin System is designed to handle the process of seat allocation for multiple airlines, each having multiple flights and trips. The system ensures efficient and concurrent seat booking, addressing the challenges of multiple users trying to book seats simultaneously. This project explores different strategies for seat assignment, including sequential assignment, parallel assignment without locks, with locks, and with skip locks, to evaluate their performance and effectiveness.

---

## Abstract
This document presents an evaluation of various seat assignment methods in an airline check-in system. The study compares four approaches: sequential, parallel without lock, parallel with lock, and parallel with skip lock. The findings demonstrate that the parallel assignment with skip lock method is the most efficient, achieving the fastest time of 24.38 ms to assign 120 seats. A detailed comparison table is provided, summarizing the performance metrics for each method.

---

## Problem Statement

### Constraints
- Multiple Airlines
- Every Airline has multiple planes (flights)
- Each flight has 120 seats
- Every flight has multiple trips
- User books a seat in one trip of a flight
## Requirements
- Handle multiple people trying to pick a seat on the plane at the same time

---

## Schema

#### Trips Table

id | name


#### Users Table

id | name


#### Seats Table

id | name | trip_id | user_id

---


## Experiments

- Sequential assignment of seats
- Parallel assignment of seats without lock
- Parallel assignment of seats with lock
- Parallel assignment of seats with skip lock


## Result

#### Experiment 1: Sequential assignment of seats

Time taken to assign seats sequentially:  **64.775042ms**
Total seats assigned:  **120**

```bash
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
```

---

#### Experiment 2: Parallel assignment of seats without lock

Time taken to assign seats without lock:  **30.4935ms**
Total seats assigned:  **9**

```bash
 *  *  *  *  *  *  *  *  *  *  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
 x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x  x 
```

---

#### Experiment 3: Parallel assignment of seats with lock

Time taken to assign seats with lock:  **54.8835ms**
Total seats assigned:  **120**


 ```bash
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
```

---

#### Experiment 4: Parallel assignment of seats with skip lock

Time taken to assign seats with skip lock:  **24.381041ms**
Total seats assigned:  **120**

```bash
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  * 
 *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *  *
```

### Comparison Table

| Experiment | Time Taken (ms) | Total Seats Assigned |
|------------|------------|----------------------|
| Sequential | 64.78 | 120 |
| Parallel without lock | 30.49 | 9 |
| Parallel with lock | 54.88 | 120 |
| Parallel with skip lock | 24.38 | 120 |