# 🧵 commenting

A scalable Go API for threaded commenting with voting and reactions.

- CockroachDB for persistence
- Redis for caching and sorting
- Cursor-based pagination
- Full REST API
- Hurl tests for E2E coverage

---

## 🚀 Running Locally

Start the API, Redis, and CockroachDB with:

```bash
docker compose up --build
```

The service will be available at:

```
http://localhost:8080
```

Admin UI for CockroachDB:

```
http://localhost:8081
```

---

## 📬 API Endpoints

### `POST /comments`

Create a new top-level comment or a reply.

**Body:**

```json
{
  "content": "Hello world!",
  "user_id": "kire",
  "parent_id": "optional-parent-id"
}
```

### `GET /comments?thread_id={id}&sort={date|upvotes|replies}&cursor={int}&limit={int}`

List comments in a thread, sorted and paginated.

### `POST /comments/{id}/{like|upvote|downvote}`

Toggle a reaction. Requires `user_id` in body.

**Body:**

```json
{
  "user_id": "kire"
}
```

All reactions are idempotent toggle operations and return `204`.

---

## 🧪 Testing

### Run unit tests:

```bash
go test -cover ./...
```

### E2E tests with Hurl:

```bash
hurl --test e2e.hurl
```

---

## 🔧 Mock Generation

Mocks are generated using [`moq`](https://github.com/matryer/moq):

```bash
cd service
moq -pkg service -out mock_repo.go . CommentRepo
moq -pkg service -out mock_cache.go . CommentCache
```
