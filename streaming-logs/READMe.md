# Streaming Logs

This project provides a service to stream logs in real-time using FastAPI. It is a mock deployment service that allows users to trigger new deployments and view logs for each deployment. The service is built using FastAPI, a modern web framework for building APIs with Python.

## Features

- **Trigger New Deployments**: Easily initiate new deployments through a simple API call.
- **Real-Time Log Streaming**: Stream logs for each deployment in real-time, allowing for immediate feedback and monitoring.
- **Deployment Management**: View and manage all deployments from a central interface.

## Setup

### Prerequisites

- Python 3.11
- FastAPI
- Uvicorn
- Jinja2
- Faker

### Installation

1. **Clone the Repository**:
    ```sh
    git clone <repository-url>
    cd streaming-logs
    ```

2. **Install the Required Dependencies**:
    ```sh
    pip install -r requirements.txt
    ```

### Running the Service

1. **Start the FastAPI Server**:
    ```sh
    ./run.sh
    ```

2. **Access the Service**:
    Open your browser and navigate to `http://localhost:8000` to access the service.

## Endpoints

- **GET /**: Home page listing all deployments.
- **POST /deployments**: Trigger a new deployment.
- **GET /deployments/{deployment_id}**: View logs for a specific deployment.
- **GET /logs/{deployment_id}**: Stream logs for a specific deployment.

### Example Usage

- **Trigger a New Deployment**:
    ```sh
    curl -X POST http://localhost:8000/deployments
    ```

- **View Logs for a Specific Deployment**:
    ```sh
    curl http://localhost:8000/deployments/{deployment_id}
    ```

- **Stream Logs for a Specific Deployment**:
    ```sh
    curl http://localhost:8000/logs/{deployment_id}
    ```

## License

This project is licensed under the MIT License.