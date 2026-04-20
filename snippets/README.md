# Snippets (Paste Bin)

A paste bin–style snippet sharing service built with Go. Users can create, store, and retrieve text snippets via a REST API. The system uses PostgreSQL for metadata, local filesystem (mock S3) for content storage, Kafka for analytics event streaming, and Elasticsearch for analytics indexing.

> **Note:** This is a prototype. Authentication and authorization have been skipped for simplicity — `owner_id` is passed directly in the request body and URL params.

## Architecture

```
Client
  │
  ▼
Fiber HTTP Server (:6000)
  │
  ├──▶ PostgreSQL  (snippet metadata)
  ├──▶ Local FS / S3  (snippet content)
  └──▶ Kafka Producer ──▶ Kafka ──▶ Kafka Consumers (x3)
                                          │
                                          ▼
                                    Elasticsearch
                                          │
                                          ▼
                                       Kibana
                                          │
                                          ▼
                                Analytics Dashboard
```

## Complete Architecture Diagram
![Architecture Diagram](snippets.png)


## Tech Stack

| Component        | Technology                 |
| ---------------- | -------------------------- |
| HTTP Framework   | Go Fiber v2                |
| Database         | PostgreSQL 18.3            |
| Object Storage   | Local filesystem (mock S3) |
| Message Broker   | Apache Kafka               |
| Search/Analytics | Elasticsearch 8.x + Kibana |
| ID Generation    | UUID v4                    |

## API Endpoints

| Method | Path                                     | Description            |
| ------ | ---------------------------------------- | ---------------------- |
| GET    | `/`                                      | Welcome / usage info   |
| GET    | `/health`                                | Health check           |
| POST   | `/api/v1/snippets/`                      | Create a new snippet   |
| GET    | `/api/v1/snippets/:snippet_id`           | Get a snippet (public) |
| GET    | `/api/v1/snippets/:snippet_id/:owner_id` | Get a snippet by owner |

### Create Snippet

**Request**

```json
{
  "name": "My Snippet",
  "content": "This is the content of the snippet.",
  "visibility": "public",
  "owner_id": 1,
  "expiry": 30
}
```

- `name` — optional, auto-generated if omitted
- `visibility` — `public` (default) or `private`
- `owner_id` — defaults to `1`
- `expiry` — optional, number of days until expiration

**Response**

```json
{
  "message": "Create a new paste bin snippet",
  "snippet_id": "<uuid>"
}
```

### Get Snippet

```
GET /api/v1/snippets/:snippet_id/:owner_id
```

Returns snippet metadata and content. Returns `410 Gone` if expired, `403 Forbidden` if the snippet is private and the owner doesn't match.

## Data Model

### `snippets` table (PostgreSQL)

| Column     | Type         | Description               |
| ---------- | ------------ | ------------------------- |
| uuid       | UUID (PK)    | Unique snippet identifier |
| name       | VARCHAR(120) | Snippet name              |
| createdAt  | TIMESTAMPTZ  | Creation timestamp        |
| expiredAt  | TIMESTAMPTZ  | Optional expiration time  |
| visibility | VARCHAR(20)  | `public` or `private`     |
| owner_id   | INT          | Owner identifier          |

### Content Storage

Snippet content is stored on the local filesystem under `<BUCKET_NAME>/<owner_id>/<uuid>.txt`, simulating S3 object storage.

## Analytics Pipeline

1. Every create/view event publishes an `AnalyticsEvent` to Kafka (`snippet_analytics` topic).
2. Three concurrent Kafka consumer goroutines read events from the topic.
3. Each event is indexed into Elasticsearch with a daily index pattern (`snippet_analytics-YYYY-MM-DD`).
4. Kibana can be used to visualize and query analytics data.

### Analytics Event Fields

`EventType`, `SnippetID`, `OwnerID`, `Visibility`, `ContentSizeInKB`, `RequestPath`, `UserAgent`, `IP`, `EventTime`

## Project Structure

```
snippets/
├── main.go            # HTTP server, routes, handlers
├── models.go          # Data types and request models
├── db.go              # PostgreSQL connection and schema init
├── analytics.go       # Kafka producer/consumer, Elasticsearch client
├── postgres.sh        # Script to run PostgreSQL in Docker
├── kafka.sh           # Script to run Kafka + Redpanda Console in Docker
├── elasticsearch.sh   # Script to run Elasticsearch + Kibana in Docker
├── .env               # Environment variables
└── go.mod             # Go module definition
```

## Prerequisites

- Go 1.25+
- Docker

## Setup

### 1. Start Infrastructure

```bash
cd snippets

# PostgreSQL
bash postgres.sh

# Kafka + Kafka UI (Redpanda Console on :8080)
bash kafka.sh

# Elasticsearch + Kibana (:9200 / :5601)
bash elasticsearch.sh
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Run the Application

```bash
go run .
```

The server starts on **port 6000**.

## Environment Variables

| Variable                 | Default                                                   |
| ------------------------ | --------------------------------------------------------- |
| `DATABASE_URL`           | `postgres://postgres:postgres@localhost:5432/snippets_db` |
| `BODY_LIMIT_MB`          | `100`                                                     |
| `BUCKET_NAME`            | `snippets-bucket`                                         |
| `KAFKA_BROKERS`          | `localhost:9094`                                          |
| `KAFKA_ANALYTICS_TOPIC`  | `snippet_analytics`                                       |
| `KAFKA_ANALYTICS_GROUP`  | `snippet-analytics-writer`                                |
| `ELASTICSEARCH_URL`      | `http://localhost:9200`                                   |
| `ELASTICSEARCH_USERNAME` | `elastic`                                                 |
| `ELASTICSEARCH_PASSWORD` | `StrongPassword123`                                       |

## Example Usage

```bash
# Create a snippet
curl -X POST http://localhost:6000/api/v1/snippets/ \
  -H "Content-Type: application/json" \
  -d '{"name":"Hello","content":"Hello, World!","visibility":"public","owner_id":1}'

# Retrieve a snippet
curl http://localhost:6000/api/v1/snippets/1/<snippet_uuid>
```