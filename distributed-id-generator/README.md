# Distributed ID Generation


### Introduction
This document explores different approaches to generating unique identifiers in a distributed system. The goal is to provide a comprehensive overview of the different strategies and their trade-offs.


### Requirements
The requirements for a distributed ID generation system are as follows:
1. **Uniqueness**: IDs must be unique across the entire system.
2. **Consistency**: IDs must be generated in a consistent manner.
3. **Scalability**: The system must be able to generate IDs at a high rate.
4. **Availability**: The system must be available even in the presence of failures.
5. **Efficiency**: The system must be efficient in terms of time and space complexity.
6. **Deterministic**: IDs should be generated deterministically based on some input parameters.
7. **Compactness**: IDs should be as compact as possible to reduce storage requirements. 


### Approaches

#### 1. Simple ID Generation (Epoch MS + Machine ID + Static Counter)
- **Description**: This approach generates IDs by combining the current timestamp in milliseconds, a unique machine ID, and a static counter. The machine ID can be derived from the network address or some other unique identifier.
- Save the static counter with a frequency. e.g. After every 1000 IDs, save the counter to the local disk.
- Sequential Benchmark
No of IDs | Counter Store Frequency | Duration (ms)
--- | --- | ---
1 Milllion | 100 | 681
1 Milllion | 500 | 357
1 Milllion | 1000 | 315
1 Milllion | 10000 | 275
10 Milllion | 100 | 6899
10 Milllion | 1000 | 3222
10 Milllion | 10000 | 2880

#### 2. Central ID Generation Service: Amazon's Way

- **Description**: Using microservice with sql database to generate unique IDs.

- DB Schema
  -  ID | Service Name | Counter
  -  1  |  user-service | 1000
  -  2  |  order-service | 500

#### 3. Fliker's Odd Even ID Generation
#### 4. Twitter's Snowflake
#### 5. Benchmark UUID VS MongoDB's ObjectID VS UUID VS Snowflake
#### 6. Snowflake At Instagram
#### 7. Benchmark Pagination: Limit Offset VS Cursor Based Pagination


