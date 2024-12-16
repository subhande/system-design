# RDB-Based Key-Value Store Implementation

## Overview
This report outlines the development of a Key-Value (KV) store using a relational database management system (RDBMS), specifically MySQL. The application, written in Go, provides essential KV store operations, including adding key-value pairs with a Time-To-Live (TTL), retrieving values, removing keys, and clearing expired entries.

---

## Features

### Key Operations
1. **Set**
   - Adds a new key-value pair to the database with an associated TTL.
   - The expiration time is calculated as the current Unix timestamp plus the TTL.

2. **Get**
   - Retrieves a value from the database by its key.
   - Handles expired or deleted keys by providing clear feedback.

3. **Delete**
   - Marks a key as deleted by setting its expiration time to `-1`.

4. **Cleanup**
   - Automatically removes expired keys from the database.
   - Iterates through expired keys to log which ones will be deleted.

---

## Database Design
The application uses a MySQL table named `kv_store` with the following schema:

| Column Name | Data Type     | Description                          |
|-------------|---------------|--------------------------------------|
| `id`        | INT           | Auto-incremented primary key.        |
| `k`         | VARCHAR(255)  | Stores the key.                      |
| `value`     | TEXT          | Stores the corresponding value.      |
| `expiry`    | INT           | Unix timestamp indicating expiration.|

---

## Implementation Details

### Setup
The `newConn()` function establishes a connection to the MySQL database and ensures that the `kv_store` table is properly initialized. If the table already exists, it is dropped and recreated.

### Functionality
1. **Insert and Update (Set)**
   - The `set(db, key, value, ttl)` function inserts a key-value pair into the database and calculates its expiration time based on the provided TTL.

2. **Retrieve (Get)**
   - The `get(db, key)` function fetches the value associated with a key and checks its expiration status. It returns appropriate messages for expired, deleted, or non-existent keys.

3. **Remove (Delete)**
   - The `delete(db, key)` function logically deletes a key by setting its expiration value to `-1`.

4. **Cleanup**
   - The `cleanup(db)` function removes all entries that have expired based on the current Unix timestamp.

---

## Usage Example
### Initial Setup
To run the program, ensure MySQL is active and the database credentials are correct. The program operates on a database named `store`.

### Running the Program
The `main()` function demonstrates how to use the KV store:

1. Insert key-value pairs with varying TTL values.
2. Retrieve keys to observe the behavior of active, expired, and non-existent entries.
3. Delete a key and verify its status.
4. Perform cleanup to remove expired keys from the database.

### Sample Execution Output
- Inserted keys: `key1`, `key2`, `key3`
- Retrieved values:
  - `key1`: Active.
  - `key2`: Expired after TTL.
  - `key3`: Deleted manually.

---

## Recommendations and Enhancements
1. **Error Handling**: Enhance error messages and recovery mechanisms for better reliability.
2. **Concurrency**: Introduce support for concurrent reads and writes to improve performance.
3. **Indexing**: Add indexes to the `k` and `expiry` columns to accelerate queries.
4. **Scalability**: Explore options such as sharding or transitioning to a NoSQL database for handling large datasets.
5. **Unit Testing**: Create comprehensive unit tests to ensure correctness and stability.

---

## Conclusion
This implementation demonstrates the development of a simple relational database-backed KV store with core functionality. While suitable for small-scale applications, it requires additional features and optimizations to support high-performance, large-scale use cases effectively.

---

## References
- [Go SQL Driver Documentation](https://pkg.go.dev/github.com/go-sql-driver/mysql)
- [MySQL Official Documentation](https://dev.mysql.com/doc/)

