
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
    - [Hashtag Service](#hashtag-service)

### [Connection Pool](connection-pool/README.md)

Connection Pooling is a widely-used design pattern aimed at optimizing database connection management. Instead of creating a new connection every time a database operation is performed, a pool of reusable connections is maintained. This approach minimizes the overhead associated with opening and closing connections, leading to better performance and resource utilization.

### [DB Sharding](db-sharding/README.md)

Sharding is a database design pattern that breaks a large database into smaller, more manageable parts called shards. Each shard is a separate database that stores a subset of the data. Sharding is used to improve the performance and scalability of the database by distributing the data across multiple servers.


### [Server Sent Events: Streaming Logs](streaming-logs/README.md)

A service for streaming logs in real-time.


### [Message Brokers](message-brokers/README.md)

Examples and implementations of various message brokers.


### [MySQL Read Replica Setup](mysql-local-read-replica/README.md)

A guide and setup for creating a local MySQL read replica.


### [Mock EC2 Status Check](mock-ec2-status-check-using-short-and-long-polling/README.md)

A mock service for checking EC2 instance statuses using short and long polling.


### [Airline Check-in System](airline-checkin-system/README.md)

The Airline Checkin System is designed to handle the process of seat allocation for multiple airlines, each having multiple flights and trips. The system ensures efficient and concurrent seat booking, addressing the challenges of multiple users trying to book seats simultaneously. This project explores different strategies for seat assignment, including sequential assignment, parallel assignment without locks, with locks, and with skip locks, to evaluate their performance and effectiveness.


### [SQL Locking](sql-locking/README.md)

Examples and explanations of SQL locking mechanisms.


### [RDB Based KV Store](rdb-based-kv-store/README.md)

A key-value store based on a relational database.


### [Load Balancer](load-balancer/README.md)

A basic implementation of a load balancer.


### [Remote Lock](remote-lock/README.md)

A project for implementing remote locking mechanisms.


### [Distributed ID Generator](distributed-id-generator/README.md)

A service for generating unique IDs in a distributed system.


### [Hashtag Service](hashtag-service/README.md)

A service for managing and generating hashtags.








