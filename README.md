# Concurrent Ticket Booking System

A backend-focused ticket booking system built to demonstrate **safe concurrent bookings**, transactional database operations, and race-condition handling.

The core problem: multiple users can attempt to book the same seat at the same time. The system guarantees that a seat can be booked by **only one request**, even under concurrent load.

## Architecture

```text
                    ┌──────────────┐
                    │    Client    │
                    │  Web / HTTP   │
                    └───────┬──────┘
                            │
                            ▼
                    ┌──────────────┐
                    │   Go API     │
                    │   HTTP       │
                    └───────┬──────┘
                            │
                            ▼
                    ┌──────────────┐
                    │ PostgreSQL   │
                    │ Transaction  │
                    │ + Locking    │
                    └──────────────┘
```

## The Concurrency Problem

Consider two users attempting to book seat `A10` simultaneously:

```text
User A ────────┐
               ├──► Book A10 ──► PostgreSQL
User B ────────┘
```

A naive implementation can produce:

```text
User A: Check A10 → available
User B: Check A10 → available

User A: Book A10
User B: Book A10   ❌
```

This creates a **double booking**.

The system instead makes the availability check and booking operation atomic at the database level.

```text
Request
   │
   ▼
BEGIN TRANSACTION
   │
   ▼
Lock / validate seat
   │
   ├── unavailable ──► ROLLBACK
   │
   └── available
          │
          ▼
      Create booking
          │
          ▼
        COMMIT
```

## Concurrency Guarantee

For `N` simultaneous booking attempts targeting the same seat:

```text
100 booking attempts
        │
        ▼
┌─────────────────────┐
│ PostgreSQL          │
│ transaction/locking │
└──────────┬──────────┘
           │
           ▼
   1 successful booking
   99 rejected
   0 double bookings
```

The database is treated as the source of truth rather than relying on application-level synchronization.

## Tech Stack

- **Go** — HTTP backend
- **PostgreSQL** — persistent storage and concurrency control
- **pgx** — PostgreSQL driver
- **Docker / Docker Compose** — local development and deployment
- **Vite** — frontend
- **SQL migrations** — schema versioning

## Project Structure

```text
.
├── main.go
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── db/
│   └── migrations/
├── handlers/
├── models/
├── repository/
└── frontend/
```

The exact structure may vary as the project evolves.

## Running Locally

### 1. Clone

```bash
git clone https://github.com/77Gaurav/concurrent-ticket-booking.git
cd concurrent-ticket-booking
```

### 2. Start PostgreSQL

Using Docker Compose:

```bash
docker compose up -d
```

### 3. Configure the database

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/tickets
```

### 4. Run the backend

```bash
go run .
```

The API starts on:

```text
http://localhost:8080
```

### 5. Run the frontend

```bash
cd frontend
npm install
npm run dev
```

## Concurrency Testing

The most important test is attempting to book the same seat concurrently.

Example:

```text
100 concurrent requests
        ↓
     Same seat
        ↓
┌─────────────────┐
│ Booking API     │
└────────┬────────┘
         ↓
   PostgreSQL
         ↓
┌─────────────────┐
│ 1 succeeds      │
│ 99 are rejected │
└─────────────────┘
```

Run the Go race detector:

```bash
go test -race ./...
```

A concurrency stress test can additionally launch many goroutines against the same booking endpoint and verify that:

```text
successful bookings == 1
duplicate bookings == 0
database inconsistencies == 0
```

## Why Database-Level Concurrency Control?

A Go mutex can protect concurrent requests inside one application process, but it is not sufficient for a production distributed system.

For example:

```text
             Load Balancer
              /         \
             ▼           ▼
        Go Server 1   Go Server 2
             │           │
             └─────┬─────┘
                   ▼
              PostgreSQL
```

A mutex on Server 1 cannot prevent Server 2 from booking the same seat.

Database transactions and constraints provide a shared consistency boundary across application instances.

## API

Example booking request:

```http
POST /bookings
Content-Type: application/json
```

```json
{
  "event_id": 1,
  "seat_id": 42
}
```

Successful booking:

```json
{
  "id": 123,
  "event_id": 1,
  "seat_id": 42,
  "status": "confirmed"
}
```

If another request already owns the seat:

```http
409 Conflict
```

```json
{
  "error": "seat already booked"
}
```

## Design Goals

This project focuses on backend engineering problems rather than CRUD functionality:

- Concurrent request handling
- Race-condition prevention
- Database transactions
- Atomic state changes
- PostgreSQL constraints
- HTTP API design
- Error handling
- Integration testing
- Dockerized development
- Reproducible local setup

## Failure Cases Considered

The booking operation should remain correct when:

- Multiple users book the same seat simultaneously
- Requests arrive milliseconds apart
- A transaction fails midway
- A client retries a failed request
- Multiple application instances access the database
- A seat has already been booked

The key invariant is:

> **A seat must never have more than one confirmed booking.**

## Future Improvements

Potential production extensions:

- Redis-based rate limiting
- Idempotency keys for safe request retries
- Authentication and authorization
- Event-driven booking notifications
- Kafka/RabbitMQ for asynchronous workflows
- Distributed tracing
- Prometheus metrics
- Load testing with k6
- Kubernetes deployment
- Horizontal API scaling

## License

MIT