# 🔍 ogpreview

A lightweight Go service for fetching and serving OpenGraph metadata (`og:title`, `og:description`, `og:image`) from a given URL.

- Redis-based caching
- Circuit breaker for network fetches
- OpenTelemetry tracing
- JSON API for OpenGraph preview

---

## 🚀 Running Locally

Start the API and Redis and Jaeger using Docker:

```bash
docker compose up --build
```

Service will run at:

```
http://localhost:8080
```

---

## 📬 API Endpoint

### `GET /preview?url={your_url}`

Fetches and returns OpenGraph metadata for a given URL.

**Example:**

```
GET /preview?url=https://example.com
```

**Response:**

```json
{
  "title": "Example Site",
  "description": "This is a sample description",
  "image": "https://example.com/image.png"
}
```

Returns cached result if available. Falls back to live fetch via circuit breaker.

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
cd ogpreview/api
moq -pkg mocks -out mocks/mock_cache.go . Cache
```

---

## 📈 Tracing (Jaeger)

This service exports traces using OpenTelemetry. You can inspect them using the Jaeger UI:

- 🌐 UI: [http://localhost:16686](http://localhost:16686)
- 🔍 Service Name: `og-preview-api`
