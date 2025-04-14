# 🔍 stockprice

A fast and scalable Go service for tracking the **latest Apple stock price** using:

- Redis **client-side caching**
- In-memory **local cache**
- Graceful cache invalidation
- Prometheus metrics for latency

---

## 🚀 Running Locally

Start the API and Redis using Docker:

```bash
docker compose up --build
```

By default, the service runs on:

```
http://localhost:8080
```

---

## 📬 API Endpoints

### `GET /price`

Fetch the latest Apple stock price (`AAPL`).

**Response:**

```json
{
  "symbol": "AAPL",
  "price": 184.12,
  "timestamp": "2025-04-12T17:00:00Z"
}
```

Returns `404` if no price is found.

---

### `POST /price`

Set or update the Apple stock price.

**Request Body:**

```json
{
  "price": 184.12
}
```

**Response:**  
HTTP 204 No Content on success.

---

### `GET /metrics`

Prometheus-compatible metrics endpoint for request counts, durations, and cache stats.

**View in browser:**

```
http://localhost:8080/metrics
```

(or `8081`, `8082`, etc. depending on your instance)

---

## 🧪 Running Tests

Run unit tests with coverage:

```bash
go test -cover ./...
```

---

## 🔧 Mock Generation

Mocks are generated using [`moq`](https://github.com/matryer/moq).

To create or update a mock for the `Cache` interface:

```bash
cd stockprice/api
moq -pkg mocks -out mocks/mock_cache.go . Cache
```

---

## 📈 Metrics (Prometheus)

Latency and request counts are tracked using Prometheus.

Visit:

```
http://localhost:8081/metrics
```

(or `8082`, `8083`)
