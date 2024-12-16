# Software Design and Architecture Projects and Experiments

This repository contains a variety of projects that focus on system design, services, and utilities. Below is an organized summary of each project, providing insight into their functionality and goals. 🌟🌟🌟

## Table of Contents

- [Software Design and Architecture Projects and Experiments](#software-design-and-architecture-projects-and-experiments)
  - [Table of Contents](#table-of-contents)
    - [Connection Pool](#connection-pool)
    - [DB Sharding](#db-sharding)
    - [Server Sent Events: Streaming Logs](#server-sent-events-streaming-logs)
    - [Message Brokers](#message-brokers)
    - [MySQL Read Replica Setup](#mysql-read-replica-setup)
    - [Mock EC2 Status Check](#mock-ec2-status-check)
    - [Airline Check-in System](#airline-check-in-system)
    - [SQL Locking](#sql-locking)
    - [RDB Based KV Store](#rdb-based-kv-store)
    - [Load Balancer](#load-balancer)
    - [Remote Lock](#remote-lock)
    - [Distributed ID Generator](#distributed-id-generator)
    - [Comparative Analysis of UUID and Auto-Increment ID Insertion Performance](#comparative-analysis-of-uuid-and-auto-increment-id-insertion-performance)
    - [MySQL LIMIT OFFSET vs Cursor-Based Pagination](#mysql-limit-offset-vs-cursor-based-pagination)
    - [MySQL ON DUPLICATE KEY UPDATE vs REPLACE INTO Performance Analysis](#mysql-on-duplicate-key-update-vs-replace-into-performance-analysis)
    - [Hashtag Service](#hashtag-service)

### [Connection Pool](connection-pool/README.md)

Connection pooling is an essential design pattern for efficient database management. It minimizes the cost of creating and closing connections by maintaining a pool of reusable connections, which significantly enhances performance and optimizes resource use. 🎯🎯🎯

### [DB Sharding](db-sharding/README.md)

Database sharding divides a large database into smaller, more manageable segments called shards. Each shard operates as an independent database containing a portion of the data. Sharding is crucial for improving scalability and performance by distributing data across multiple servers. ⚡⚡⚡

### [Server Sent Events: Streaming Logs](streaming-logs/README.md)

This project demonstrates a log streaming service implemented using FastAPI. Users can initiate deployments and view real-time logs through **Server-Sent Events (SSE)**, enabling continuous updates without requiring the client to repeatedly poll the server. 📝📝📝

### [Message Brokers](message-brokers/README.md)

This project contains examples and implementations of message brokers, including RabbitMQ and Kafka. These tools are explored for their roles in facilitating asynchronous communication and task management between services. 🛠️🛠️🛠️

### [MySQL Read Replica Setup](mysql-read-replica/README.md)

A detailed guide to configuring a MySQL read replica with Docker. This project explains the process of setting up primary and replica MySQL containers, configuring replication, and verifying synchronization between the databases. 📚📚📚

### [Mock EC2 Status Check](mock-ec2-status-check-using-short-and-long-polling/README.md)

This project replicates an EC2 instance status check service using short and long polling. It provides a practical demonstration of polling mechanisms, showcasing how clients can retrieve instance statuses effectively through either method. 🔍🔍🔍

### [Airline Check-in System](airline-checkin-system/README.md)

This system manages seat allocation for multiple airlines and flights, addressing concurrency challenges during booking. Various approaches to seat assignment, such as sequential and parallel strategies with and without locking mechanisms, are evaluated for efficiency. ✈️✈️✈️

### [SQL Locking](sql-locking/README.md)

This project investigates SQL locking mechanisms, focusing on how they maintain data consistency during concurrent access. Understanding these techniques allows developers to design systems that are both reliable and high-performing under concurrent workloads. 🔒🔒🔒

### [RDB Based KV Store](rdb-based-kv-store/README.md)

A relational database is used to implement a Key-Value (KV) store with functionalities like adding key-value pairs with expiration times, retrieving values, and clearing expired keys. The application is built in Go and highlights efficient use of an RDBMS for KV store operations. 🔑🔑🔑

### [Load Balancer](load-balancer/README.md)

This project develops a load balancer that distributes requests across backend servers, ensuring high availability and efficient resource utilization. It focuses on L4 (Transport Layer) load balancing, which directs traffic based on network and transport layer information. ⚖️⚖️⚖️

### [Remote Lock](remote-lock/README.md)

This project demonstrates the implementation of remote locking mechanisms using Redis and Go. It includes both single-instance and quorum-based distributed locks, offering solutions for synchronization in distributed systems. 🔗🔗🔗

### [Distributed ID Generator](distributed-id-generator/README.md#distributed-id-generation)

This project explores methods for generating unique identifiers in distributed systems. Various strategies are compared for their scalability, reliability, and performance under different conditions. 🆔🆔🆔

### [Comparative Analysis of UUID and Auto-Increment ID Insertion Performance](distributed-id-generator/README.md#comparative-analysis-of-uuid-and-auto-increment-id-insertion-performance)

A performance comparison between UUIDs and auto-increment IDs for database insertion operations. The analysis evaluates their suitability for scalability and efficiency in systems requiring high throughput. 📊📊📊

### [MySQL LIMIT OFFSET vs Cursor-Based Pagination](distributed-id-generator/README.md#mysql-limit-offset-vs-cursor-based-pagination)

This experiment contrasts two pagination methods, **LIMIT OFFSET** and **Cursor-Based Pagination**, using a dataset with millions of rows. Results highlight the advantages of cursor-based pagination for efficient data retrieval at scale. 📖📖📖

### [MySQL ON DUPLICATE KEY UPDATE vs REPLACE INTO Performance Analysis](distributed-id-generator/README.md#mysql-on-duplicate-key-update-vs-replace-into-performance-analysis)

This analysis examines two MySQL conflict-resolution statements, `ON DUPLICATE KEY UPDATE` and `REPLACE INTO`, focusing on their performance in high-concurrency environments and evaluating their trade-offs in specific use cases. 🔄🔄🔄

### [Hashtag Service](hashtag-service/README.md)

The Hashtag Service offers tools for managing hashtags in posts, including generating posts with hashtags, extracting and counting hashtag occurrences, and storing the data in MongoDB. The system is designed to handle high-volume hashtag processing efficiently. 🏷️🏷️🏷️

