# Event-Driven Order System

A production-inspired **event-driven microservices system** built with **Golang**, designed to demonstrate backend engineering fundamentals aligned with real-world requirements.

This project focuses on **system design, decoupling, reliability, and observability** rather than UI or excessive features.

---

##  Project Goals

* Demonstrate **microservices architecture** using Golang
* Apply **event-driven design** with RabbitMQ
* Use **Redis** for caching and **MySQL** for persistent storage
* Practice **clean architecture**, testing, and service isolation
* Showcase **production-minded thinking** suitable for backend interviews

---

##  High-Level Architecture

```
Client
  │
  │ REST API
  ▼
Order Service ── publish ──▶ RabbitMQ ── consume ──▶ Analytics Service
      │                                         │
      │                                         ▼
   MySQL + Redis                           MySQL + Redis
```

### Architecture Principles

* Services are **loosely coupled** via events
* Each service owns its **own database**
* Communication is **async-first** for scalability
* Failure in one service **does not block** others

---

##  Services Overview

### 1️⃣ Order Service

**Responsibilities:**

* Handle order creation and retrieval
* Persist order data
* Publish domain events

**Tech:**

* REST API (Golang)
* MySQL (orders)
* Redis (cache order by ID)
* RabbitMQ (event producer)

**Key Endpoints:**

* `POST /orders`
* `GET /orders/{id}`

---

### 2️⃣ Analytics Service

**Responsibilities:**

* Consume order events
* Aggregate business metrics
* Provide analytics APIs

**Tech:**

* RabbitMQ (event consumer)
* MySQL (aggregates)
* Redis (cache summary)

**Key Endpoints:**

* `GET /analytics/summary`

---

### 3️⃣ Notification Worker (Optional)

**Responsibilities:**

* Consume order events
* Simulate email / notification sending

**Purpose:**

* Demonstrate fan-out consumers
* Show extensibility of event-driven design

---

##  Project Structure

```
services/
├── order-service/
│   ├── cmd/
│   ├── internal/
│   │   ├── handler/
│   │   ├── service/
│   │   ├── repository/
│   │   ├── mq/
│   │   └── cache/
│   ├── migrations/
│   ├── tests/
│   └── main.go
│
├── analytics-service/
└── notification-worker/
```

---

##  Event Flow

1. Client sends `POST /orders`
2. Order Service:

   * Validates input
   * Saves order in MySQL (transaction)
   * Publishes `OrderCreated` event
3. Analytics Service:

   * Consumes event
   * Updates aggregates
   * Caches results in Redis

---

##  Testing Strategy

* **Unit Tests**

  * Service layer logic
  * Repository interfaces (mocked)

* **Integration Tests**

  * MySQL migrations
  * Redis caching behavior

---

##  Observability (Basic)

* Structured logging
* Metrics endpoint (`/metrics`)
* Example metrics:

  * `orders_created_total`
  * `http_request_duration_seconds`

---

##  Running Locally

```bash
docker-compose up --build
```

Services exposed:

* Order Service: `http://localhost:8080`
* Analytics Service: `http://localhost:8081`

---

##  Design Decisions

* **Why RabbitMQ?**

  * Decouple services
  * Avoid synchronous dependencies

* **Why Redis Cache-Aside?**

  * Reduce DB load
  * Simple and predictable behavior

* **Why separate databases?**

  * Service autonomy
  * Independent scaling and migrations

---

##  Future Improvements

* gRPC for internal service communication
* Dead-letter queue (DLQ)
* Distributed tracing
* Rate limiting

---

## 📌 Author
[anvndev](github.com/andev0x)
> Built as a backend-focused side project to demonstrate **Golang microservices, event-driven architecture, and production-ready thinking**.

