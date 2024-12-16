# Hashtag Service

## Overview

The Hashtag Service is designed to manage and generate hashtags for posts. It includes functionalities for generating posts with hashtags, extracting hashtags from posts, counting hashtag occurrences, and storing the data in a MongoDB database.

## Table of Contents

- [Hashtag Service](#hashtag-service)
  - [Overview](#overview)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Architecture](#architecture)
  - [Setup](#setup)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
  - [Usage](#usage)
    - [Running the Services](#running-the-services)
    - [Running the API](#running-the-api)
  - [Endpoints](#endpoints)
  - [License](#license)

## Features

- **Post Generation**: Generate posts with random hashtags and send them to Kafka.
- **Hashtag Extraction**: Extract hashtags from posts and send them to Kafka.
- **Counting Service**: Count occurrences of hashtags and update the count in MongoDB.
- **Database Initialization**: Initialize MongoDB with initial data.

## Architecture

The service is composed of the following components:

1. **Post Generator**: Generates posts with random hashtags and sends them to Kafka.
2. **Hashtag Extractor**: Consumes posts from Kafka, extracts hashtags, and sends them to another Kafka topic.
3. **Counting Service**: Consumes hashtags from Kafka, counts their occurrences, and updates MongoDB.
4. **Database Initialization**: Initializes MongoDB with initial data.

## Setup

### Prerequisites

- Go 1.22.3 or later
- Kafka
- MongoDB

### Installation

1. **Clone the Repository**:
    ```sh
    git clone <repository-url>
    cd hashtag-service
    ```

2. **Install Dependencies**:
    ```sh
    go mod tidy
    ```

3. **Start MongoDB**:
    ```sh
    brew services start mongodb/brew/mongodb-community
    ```

4. **Start Kafka**:
    Follow the instructions to start Kafka on your local machine.

## Usage

### Running the Services

1. **Post Generator**:
    ```sh
    go run main.go producer
    ```

2. **Hashtag Extractor**:
    ```sh
    go run main.go extractor
    ```

3. **Counting Service**:
    ```sh
    go run main.go count-svc
    ```

4. **Database Initialization**:
    ```sh
    go run main.go init-db
    ```

### Running the API

1. **Start the FastAPI Server**:
    ```sh
    uvicorn main:app --reload
    ```

2. **Access the API**:
    Open your browser and navigate to `http://localhost:8000` to access the API.

## Endpoints

- **GET /tag/**: Get data for a specific hashtag.
    ```sh
    curl http://localhost:8000/tag/?hashtag=<hashtag>
    ```

- **GET /hashtags**: Get data for all hashtags.
    ```sh
    curl http://localhost:8000/hashtags
    ```

## License

This project is licensed under the MIT License.