
# Projects

This repository contains multiple projects related to various system designs, services, and utilities. Below is a brief overview of each project and its purpose.

## Table Of Contents

- [Projects](#projects)
  - [Table Of Contents](#table-of-contents)
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

Connection Pooling is a widely-used design pattern aimed at optimizing database connection management. Instead of creating a new connection every time a database operation is performed, a pool of reusable connections is maintained. This approach minimizes the overhead associated with opening and closing connections, leading to better performance and resource utilization.

### [DB Sharding](db-sharding/README.md)

Sharding is a database design pattern that breaks a large database into smaller, more manageable parts called shards. Each shard is a separate database that stores a subset of the data. Sharding is used to improve the performance and scalability of the database by distributing the data across multiple servers.


### [Server Sent Events: Streaming Logs](streaming-logs/README.md)

This project provides a service to stream logs in real-time using FastAPI. It is a mock deployment service that allows users to trigger new deployments and view logs for each deployment. The service is built using FastAPI, a modern web framework for building APIs with Python. This uses Server-Sent Events (SSE) to stream logs in real-time to the client.

### [Message Brokers](message-brokers/README.md)

Examples and implementations of various message brokers i.e. RabbitMQ, Kafka, etc.


### [MySQL Read Replica Setup](mysql-read-replica/README.md)

A Step-by-step guide to setting up a MySQL read replica using Docker. It covers the process of initializing the primary and replica MySQL containers, configuring replication, and verifying the setup. By the end, you'll have a functional read replica where data from the primary database is automatically synchronized to the replica.


### [Mock EC2 Status Check](mock-ec2-status-check-using-short-and-long-polling/README.md)

This project demonstrates how to implement a mock EC2 status check service using short and long polling. The service simulates the behavior of an EC2 instance status check, allowing clients to check the status of an instance using either short polling (regular HTTP requests) or long polling (HTTP requests that wait for a response). This project provides a simple and effective way to understand and implement polling mechanisms in web services.


### [Airline Check-in System](airline-checkin-system/README.md)

The Airline Checkin System is designed to handle the process of seat allocation for multiple airlines, each having multiple flights and trips. The system ensures efficient and concurrent seat booking, addressing the challenges of multiple users trying to book seats simultaneously. This project explores different strategies for seat assignment, including sequential assignment, parallel assignment without locks, with locks, and with skip locks, to evaluate their performance and effectiveness.


### [SQL Locking](sql-locking/README.md)

In this experiment, we will explore the concept of SQL locking mechanisms and their impact on database transactions. SQL locking is a crucial aspect of database management systems that ensures data consistency and integrity during concurrent access. By understanding how different locking strategies work, we can design more efficient and reliable database applications.


### [RDB Based KV Store](rdb-based-kv-store/README.md)

This project outlines the development of a Key-Value (KV) store using a relational database management system (RDBMS), specifically MySQL. The application, written in Go, provides essential KV store operations, including adding key-value pairs with a Time-To-Live (TTL), retrieving values, removing keys, and clearing expired entries.


### [Load Balancer](load-balancer/README.md)

This project implements a load balancer that distributes incoming requests across multiple backend servers to ensure efficient resource utilization and high availability. This is L4 (Transport Layer) load balancing, which operates at the network transport layer and forwards requests based on network and transport layer information.


### [Remote Lock](remote-lock/README.md)

This repository explores the implementation of remote locks using Redis 🛠️ in Go 🐹. The system supports single-instance locks 🔑 and quorum-based distributed locks 🌐, providing mechanisms for reliable ✅ and efficient ⚡ synchronization in both standalone 🖥️ and distributed environments. By leveraging Redis’s atomic operations ⚙️, the locks ensure consistency 📏 and robustness 💪.



### [Distributed ID Generator](distributed-id-generator/README.md#distributed-id-generation)

This repository explores different approaches to generating unique identifiers in a distributed system. The goal is to provide a comprehensive overview of the different strategies and their trade-offs.


### [Comparative Analysis of UUID and Auto-Increment ID Insertion Performance](distributed-id-generator/README.md#comparative-analysis-of-uuid-and-auto-increment-id-insertion-performance)

This project aims to compare the performance of UUIDs and auto-increment IDs for insertion operations in a MySQL database. By analyzing the time taken to insert records using both types of identifiers, we can gain insights into the efficiency and scalability of each approach.

### [MySQL LIMIT OFFSET vs Cursor-Based Pagination](distributed-id-generator/README.md#mysql-limit-offset-vs-cursor-based-pagination)

This experiment compares the performance of two pagination techniques in MySQL: **LIMIT OFFSET Pagination** and **Cursor-Based Pagination**, using a dataset of 6,000,000 rows. The test utilizes an auto-increment or monotonically increasing ID for efficient retrieval in Cursor-Based Pagination.

### [MySQL ON DUPLICATE KEY UPDATE vs REPLACE INTO Performance Analysis](distributed-id-generator/README.md#mysql-on-duplicate-key-update-vs-replace-into-performance-analysis)

This report presents a comparative analysis of two MySQL statements, `ON DUPLICATE KEY UPDATE` and `REPLACE INTO`, for handling insertions with potential conflicts. The study evaluates the performance of these statements in a high-concurrency environment with a focus on efficiency and scalability.

### [Hashtag Service](hashtag-service/README.md)

The Hashtag Service is designed to manage and generate hashtags for posts. It includes functionalities for generating posts with hashtags, extracting hashtags from posts, counting hashtag occurrences, and storing the data in a MongoDB database.









