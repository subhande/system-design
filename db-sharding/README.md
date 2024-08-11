# Sharding MySQL Databases


## Introduction

Sharding is a database design pattern that breaks a large database into smaller, more manageable parts called shards. Each shard is a separate database that stores a subset of the data. Sharding is used to improve the performance and scalability of the database by distributing the data across multiple servers.


## Experiment

Environment: local

Horizontal sharding used to distribute the data across multiple databases. The data is distributed based on the user_id. The data is distributed across 2 databases.