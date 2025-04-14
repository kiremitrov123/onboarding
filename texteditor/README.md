# 📝 texteditor

A realtime collaborative text editor backend built in Go, designed to demonstrate:

- Finite State Machine (FSM) logic for applying text deltas
- Redis Pub/Sub for live edit replication
- In-memory state tracking per `docID`
- Server-Sent Events (SSE) for real-time updates
- Unit-tested FSM and service layer

---

## 🚀 Getting Started

Start the API and Redis using Docker:

```bash
docker compose up --build
```

The service will be available at:

```
http://localhost:8080
```

---

## 📬 API Endpoints

### `GET /subscribe?doc_id=doc123`

Streams deltas in real-time via SSE.

```
data: {"user":"alice","op":"insert","index":0,"text":"Hello "}
data: {"user":"bob","op":"insert","index":6,"text":"world!"}
```

---

### `GET /document?doc_id=doc123`

Returns the current full document text.

```json
{
  "text": "Hello world!"
}
```

---

### `POST /edit`

Submit a text edit.

```json
{
  "doc_id": "doc123",
  "edit": {
    "user": "alice",
    "op": "insert", // or "delete"
    "index": 0,
    "text": "Hello "
  }
}
```

**Response:**

```json
{
  "status": "ok"
}
```

---

## 🧪 Tests

Run unit tests with coverage:

```bash
go test -cover ./...
```

Mocks are generated with [`moq`](https://github.com/matryer/moq):

```bash
cd service
moq -pkg mocks -out mocks/mock_pubsub.go ../redis PubSub
```

---

## 🎯 Realtime Test Script

Run the scripted demo of insert/delete edits and SSE output:

- Note that the servers need to be up and running

```bash
./realtime_test.sh
```

This sends edits, streams deltas via SSE, and shows the final document state.
