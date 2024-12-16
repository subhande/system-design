
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

A key-value store based on a relational database.


### [Load Balancer](load-balancer/README.md)

A basic implementation of a load balancer.


### [Remote Lock](remote-lock/README.md)

A project for implementing remote locking mechanisms.


### [Distributed ID Generator](distributed-id-generator/README.md)

A service for generating unique IDs in a distributed system.


### [Hashtag Service](hashtag-service/README.md)

A service for managing and generating hashtags.








