# Notification Service

A small Go-based notification pipeline backed by PostgreSQL and RabbitMQ.

It models a simple campaign flow:

1. A controller seeds users, creates a template, and creates a campaign.
2. The campaign is published to the bulk filter queue.
3. The iterator consumes the campaign, renders notifications for recipients, and fans them out to a priority queue.
4. The worker consumes notification jobs and simulates delivery.
5. Failed jobs are pushed to a dead-letter queue instead of being retried forever.

# Architecture Diagram

![Architecture Diagram](arch.png)

## Components

- `-controler`
  Seeds users and templates, creates a campaign, and enqueues a `BulkFilterJob`.
- `-iterator`
  Consumes from `bulk-filter-queue`, loads recipients, renders message bodies, and publishes `NotificationJob`s to a priority queue.
- `-worker`
  Consumes notification jobs from a priority queue such as `iterator-queue-p1` and simulates sending them.

## Queue Flow

- `bulk-filter-queue`
  Receives campaign-level jobs from the controller.
- `iterator-queue-p1`
- `iterator-queue-p2`
- `iterator-queue-p3`
  Receive recipient-level notification jobs from the iterator.

Dead-letter queues are created automatically when a job fails:

- `bulk-filter-queue.dlq`
- `iterator-queue-p1.dlq`
- `iterator-queue-p2.dlq`
- `iterator-queue-p3.dlq`

Each dead-letter message contains:

- the original queue name
- the job type
- the error message
- the failure timestamp
- the original payload

## Project Files

- [`main.go`](/Users/subhande/Desktop/development/system-design/notification-service/main.go)
  Entry point and service mode selection.
- [`control_service.go`](/Users/subhande/Desktop/development/system-design/notification-service/control_service.go)
  Seeds sample data and creates campaigns.
- [`iterator_service.go`](/Users/subhande/Desktop/development/system-design/notification-service/iterator_service.go)
  Expands campaign jobs into notification jobs.
- [`worker_service.go`](/Users/subhande/Desktop/development/system-design/notification-service/worker_service.go)
  Simulates delivery and updates recipient status.
- [`rabbitmq.go`](/Users/subhande/Desktop/development/system-design/notification-service/rabbitmq.go)
  RabbitMQ connection management, queue publishing, consuming, and dead-letter handling.
- [`service.go`](/Users/subhande/Desktop/development/system-design/notification-service/service.go)
  Database operations for users, templates, campaigns, and recipients.
- [`schema.sql`](/Users/subhande/Desktop/development/system-design/notification-service/schema.sql)
  PostgreSQL schema initialized on startup.

## Prerequisites

- Go
- Docker
- PostgreSQL
- RabbitMQ

This repo includes helper scripts to run PostgreSQL and RabbitMQ in Docker.

## Environment

Create a `.env` file in `notification-service/` with values like:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/notification_service?sslmode=disable
RABBITMQ_URL=amqp://admin:1234@localhost:5672/
RABBITMQ_BULKFILTER_QUEUE=bulk-filter-queue
RABBITMQ_ITERATOR_QUEUE_P1=iterator-queue-p1
RABBITMQ_ITERATOR_QUEUE_P2=iterator-queue-p2
RABBITMQ_ITERATOR_QUEUE_P3=iterator-queue-p3
```

## Start Dependencies

From [`notification-service`](/Users/subhande/Desktop/development/system-design/notification-service):

```bash
sh postgres.sh
sh rabbitmq.sh
```

RabbitMQ management UI:

- URL: `http://localhost:15672`
- Username: `admin`
- Password: `1234`

## Run the Services

Use three terminals and start them in this order:

Terminal 1:

```bash
go run . -iterator
```

Terminal 2:

```bash
go run . -worker
```

Terminal 3:

```bash
go run . -controler
```

Why this order:

- the iterator should be ready before campaign jobs are published
- the worker should be ready before notification jobs are fanned out

## How It Works

### 1. Controller

The controller:

- inserts sample users
- inserts an admin user
- creates a template
- creates a campaign for all regular users
- publishes a `BulkFilterJob` to `bulk-filter-queue`

### 2. Iterator

The iterator:

- consumes a `BulkFilterJob`
- marks the campaign as `running`
- fetches the template and recipients
- renders a message body per recipient
- publishes a `NotificationJob` to a priority queue
- marks recipients as `processing`
- updates campaign status to `queued` after fan-out

Template placeholders such as `{{name}}` and `{{email}}` are supported and normalized before Go template parsing.

### 3. Worker

The worker:

- consumes a `NotificationJob`
- simulates provider delivery
- marks the recipient as `sent` or `failed`
- updates campaign delivery counters

The current implementation intentionally fails every fifth recipient to simulate delivery errors.

## RabbitMQ Connection Handling

RabbitMQ clients are cached globally per process.

That means:

- a service process reuses one RabbitMQ client per queue name
- the iterator does not create a fresh connection for every job
- all cached RabbitMQ clients are closed on process exit

## Database Initialization

Every service calls database initialization on startup.

Important:

- [`schema.sql`](/Users/subhande/Desktop/development/system-design/notification-service/schema.sql) is executed each time a process starts
- keep destructive `DROP TABLE` statements commented unless you explicitly want to reset data

## Build

```bash
go build ./...
```

## Troubleshooting

### Iterator looks stuck

If you see:

```text
iterator service consuming from bulk-filter-queue
```

that usually means the iterator is ready and waiting for messages. It is not necessarily hung.

### Messages go to dead-letter queue

A failed job is dead-lettered and the original message is acknowledged to avoid infinite requeue loops.

Look for logs like:

```text
published failed bulk_filter message from bulk-filter-queue to bulk-filter-queue.dlq
```

or:

```text
published failed notification message from iterator-queue-p1 to iterator-queue-p1.dlq
```

### Template parsing fails

If a stored template uses `{{name}}` instead of `{{.name}}`, the service now normalizes it automatically.

### Worker appears idle

Make sure the worker is consuming the same priority queue used by the campaign, usually:

```bash
go run . -worker
```

which defaults to `p1`.

You can also run:

```bash
go run . -worker p2
go run . -worker p3
```

## Notes

- The service flag is spelled `-controler` in the current code, not `-controller`.
- `run.sh` exists, but running the services in separate terminals is easier for observing logs.


# Future Improvements
[ ] Types of campaign - Marketing, Transactional, etc.
[ ] Trigger One Campaign Multiple Time - With Different User
[ ] Avoid Sending Multiple Same Marketing Campaign To Same User - Use Redis + Bloom Filter
