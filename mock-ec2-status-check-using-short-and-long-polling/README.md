## Documentation for Mock EC2 Instance Creation API

### Overview

This Go program simulates a simple EC2 instance creation service. It provides a REST API built with the Gin framework and interacts with a MySQL database to track the status of EC2 instance creation tasks. The program allows users to initiate the creation of an EC2 instance and monitor its progress across several statuses (`TO_DO`, `IN_PROGRESS`, and `DONE`).

### Features

1. **Initialize Database and Dummy Data**
   - On startup, the `init` function connects to a MySQL database, drops any existing `servers` table, creates a new one, and inserts five records with a status of `"NOT_YET_CREATED"`.

2. **Simulate EC2 Creation Process**
   - The `createEC2` function simulates the asynchronous creation of an EC2 instance by updating the instance status in three stages (`TO_DO`, `IN_PROGRESS`, `DONE`) with delays to mimic processing time.

3. **API Endpoints**

   - **Root Endpoint (`GET /`)**
     - Returns a welcome message indicating the purpose of the service.

   - **Start EC2 Instance Creation (`POST /create/:server_id`)**
     - Starts the creation process for a specific EC2 instance with the given `server_id`. This endpoint is asynchronous, meaning the creation process runs in the background.

   - **List All Servers (`GET /servers`)**
     - Returns a list of all EC2 instances along with their `server_id` and current `status`.

   - **Get Short Status of Server (`GET /short/status/:server_id`)**
     - Fetches and returns the current status of a specific EC2 instance based on `server_id`.

   - **Get Long Polling Status of Server (`GET /long/status/:server_id`)**
     - Uses long polling to continuously check for updates to the EC2 instance status until a change is detected or the request times out.

### Tests

To verify the functionality of this API, the following tests can be conducted:

1. **Database Initialization Test**
   - Ensure that on application startup, the `servers` table is created, and five rows with a status of `"NOT_YET_CREATED"` are inserted. Verify this by querying the `GET /servers` endpoint.

2. **EC2 Creation Process Test**
   - Test the `POST /create/:server_id` endpoint by initiating an EC2 instance creation. Confirm that:
     - The initial status is set to `"TO_DO"`.
     - After a delay, the status changes to `"IN_PROGRESS"`.
     - Finally, the status changes to `"DONE"`.
   - Each status can be verified by calling the `GET /short/status/:server_id` or `GET /long/status/:server_id` endpoints.

3. **API Endpoint Tests**
   - **Root Endpoint**: Send a `GET` request to `/` and verify that the response contains the welcome message.
   - **Instance Creation**: Send `POST /create/:server_id` with a valid `server_id` and check if the response confirms the creation has started.
   - **List Servers**: Send a `GET` request to `/servers` and check if the response includes the list of servers with correct `server_id` and `status`.
   - **Short Status Check**: Send `GET /short/status/:server_id` and confirm that the current status of the specified `server_id` is correctly returned.
   - **Long Polling Status Check**: Send `GET /long/status/:server_id` and verify that it keeps polling until there is a status change.

### Installation Guide

Follow these steps to set up and run the EC2 Instance Creation API:

#### Prerequisites

1. **Go**: Make sure Go is installed. You can download it from [golang.org](https://golang.org/dl/).
2. **MySQL**: Install MySQL and ensure the service is running. You’ll also need access credentials.
3. **Gin and MySQL Driver**: These Go packages are required and will be installed as dependencies.

#### Steps

1. **Clone the Repository**
   ```bash
   git clone <repository_url>
   cd <repository_directory>
   ```

2. **Configure MySQL Database**
   - Start MySQL and create a database named `demo`:
     ```sql
     CREATE DATABASE demo;
     ```
   - Update the database connection string in the code (inside the `init` function) with your MySQL username and password:
     ```go
     _db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/demo")
     ```
     Replace `username` and `password` with your actual MySQL credentials.

3. **Install Dependencies**
   - Run the following commands to install the required dependencies:
     ```bash
     go get -u github.com/gin-gonic/gin
     go get -u github.com/go-sql-driver/mysql
     ```

4. **Run the Application**
   - Start the application:
     ```bash
     go run main.go
     ```
   - The server will start on `localhost:8080`.

5. **Verify the Setup**
   - Open a browser or use a tool like `curl` or Postman to test the following endpoint:
     ```bash
     curl http://localhost:8080/
     ```
   - You should receive a response like:
     ```json
     {
       "message": "Welcome to EC2 instance creation service"
     }
     ```

### Example Usage

To test the API, refer to the [Tests](#Tests) section for example API requests. 

With this setup, you should now be able to use the EC2 instance simulation service locally.