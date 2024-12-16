# Set Up MySQL Read Replica

## Summary

This guide provides a step-by-step approach to setting up a local MySQL read replica using Docker. It covers the process of initializing the primary and replica MySQL containers, configuring replication, and verifying the setup. By the end, you'll have a functional read replica where data from the primary database is automatically synchronized to the replica.

## Table of Contents

- [Set Up MySQL Read Replica](#set-up-mysql-read-replica)
  - [Summary](#summary)
  - [Table of Contents](#table-of-contents)
  - [Resources](#resources)
  - [Setup](#setup)
    - [Files](#files)
    - [1. Connect to Primary](#1-connect-to-primary)
    - [2. Configure the Replica](#2-configure-the-replica)
    - [3. Create a Table in Primary Database](#3-create-a-table-in-primary-database)
    - [4. Test Replication in MySQL Read Replica](#4-test-replication-in-mysql-read-replica)

---

## Resources
- [MySQL Read Replica Setup Guide](https://channaly.medium.com/setting-up-mysql5-7-read-replica-on-macos-61db2cf600cd)
- [Local MySQL Setup Linux](https://medium.com/@neluwah/setting-up-mysql-replication-for-high-availability-a-step-by-step-guide-fa15e8e5b177)
- ChatGPT

---

## Setup

### Files
- `docker-compose.yml`: Contains the configuration for the primary and replica MySQL containers.
- `init_primary.sql`: SQL script to initialize the primary database with replica_user.

Start by tearing down any existing setup and initializing the containers.



```bash
docker-compose down

# Remove existing data by deleting volumes
rm -rf primary-data
rm -rf replica-data

docker-compose up -d
```

---

### 1. Connect to Primary

To create a database named `demo`, follow these steps:

1. **Log in to the Primary MySQL Server:**

   ```bash
   docker exec -it mysql-primary mysql -uroot -p
   ```

   Replace `-p` with the root password specified in your `docker-compose.yml` (e.g., `root_password`).

2. **Create the `demo` Database:**

   ```sql
   CREATE DATABASE demo;
   ```

3. **Verify the Database Creation:**

   ```sql
   SHOW DATABASES;
   ```

   You should see `demo` in the list of databases.

4. **Show Master Status:**

   ```sql
   SHOW MASTER STATUS;
   ```

   Note down the `File` and `Position` values. These will be required for configuring the replica.

---

### 2. Configure the Replica

1. **Log in to the Replica Container:**

   ```bash
   docker exec -it mysql-replica mysql -uroot -p
   ```

   Replace `-p` with the root password specified in your `docker-compose.yml` (e.g., `root_password`).

2. **Configure the Replica:**

   Run the following SQL commands inside the replica's MySQL prompt:

   ```sql
   CHANGE MASTER TO
       MASTER_HOST='mysql-primary',
       MASTER_USER='replica_user',
       MASTER_PASSWORD='replica_password',
       MASTER_LOG_FILE='mysql-bin.00000X',
       MASTER_LOG_POS=XXX;
   ```

   **Example:**
   ```sql
   CHANGE MASTER TO
       MASTER_HOST='mysql-primary',
       MASTER_USER='replica_user',
       MASTER_PASSWORD='replica_password',
       MASTER_LOG_FILE='mysql-bin.000003',
       MASTER_LOG_POS=157;
   ```

   If the above command doesn't work, stop the replica first, then try again:

   ```sql
   STOP REPLICA IO_THREAD;

   START SLAVE;

   SHOW SLAVE STATUS\G
   ```

3. **Verify Slave Status:**

   Execute the command `SHOW SLAVE STATUS\G` and confirm:

   - `Slave_IO_Running: Yes`
   - `Slave_SQL_Running: Yes`

   These statuses indicate that replication is functioning correctly.

---

### 3. Create a Table in Primary Database

1. **Create a Table in the Primary Database:**

   ```sql
   USE demo;

   CREATE TABLE test_table (
       id INT AUTO_INCREMENT PRIMARY KEY,
       column1 VARCHAR(255),
       column2 VARCHAR(255)
   );
   ```

2. **Insert Sample Data (Optional):**

   ```sql
   INSERT INTO test_table (column1, column2) VALUES ('value1', 'value2');
   ```

---

### 4. Test Replication in MySQL Read Replica

1. **Verify the Database Creation:**

   ```sql
   SHOW DATABASES;
   ```

   Ensure `demo` appears in the list of databases.

2. **Check Data Replication:**

   ```sql
   USE demo;

   SELECT * FROM test_table;
   ```

   You should see the records inserted into the `test_table` on the primary database.

