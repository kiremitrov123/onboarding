# The API Sandbox

This repository contains several one-day microservices focused on building fast, resilient, and production-aware APIs.
Each service tackles a different concept including caching, real-time syncing, distributed systems, and database performance.

While all services aim to check every box from the requirements, the Commenting API puts greater emphasis on complexity and depth—featuring a more intricate design and broader coverage of edge cases.

---

## Projects

### 1. **OG Tag URL Preview Service**

**Goal:** Basic testing and monitoring

- Provides OpenGraph (`og:title`, `og:description`, `og:image`) tags as JSON for a given URL
- Redis caching for popular URLs
- Circuit breaker to handle flaky URLs
- Tracing included: Powered by OpenTelemetry and viewable in Jaeger
- Focused on speed, reliability, and caching

📁 Folder: `ogpreview/`

---

### 2. **Latest Apple Stock Price Lookup**

**Goal:** Understand client-side caching and network latency

- Runs 3 identical servers using Redis client-side caching (`CLIENT TRACKING`)
- Returns latest Apple stock price with 0–1ms latency from cache
- Includes response time metrics and local in-memory fallback
- Simulates real-time pricing scenarios

📁 Folder: `stockprice/`

---

### 3. **Realtime Text Editor API**

**Goal:** Understand finite state machines and Redis Pub/Sub

- API receives atomic text edits (deltas) instead of full text
- Redis Pub/Sub used to replicate edits to all connected clients
- Returns:
  - Reconstructed full text
  - PubSub stream of deltas
- Demonstrates distributed synchronization

📁 Folder: `texteditor/`

---

### 4. **Commenting API (Threaded)**

**Goal:** Database usage + scalability

- Threaded comments with support for:
  - Upvotes, downvotes, likes
  - Reply count tracking
- Optimized for sorting by:
  - Date
  - Reply count
  - Upvotes
- Built on CockroachDB and Redis
- Supports cursor-based pagination
- High-performance Redis caching using sorted sets with score-based indexing (ZADD) for fast access by upvotes, replies, or date
- Includes:
  - Mock generation
  - Unit tests
  - E2E tests via Hurl

📁 Folder: `commenting/`

---
