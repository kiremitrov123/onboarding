# Create top-level comment A by user Alice
POST http://localhost:8080/comments
Content-Type: application/json

{
  "content": "Comment A from Alice",
  "user_id": "alice"
}
HTTP 201
[Captures]
comment_a_id: jsonpath "$.id"

# Create top-level comment B by Bob
POST http://localhost:8080/comments
Content-Type: application/json

{
  "content": "Comment B from Bob",
  "user_id": "bob"
}
HTTP 201
[Captures]
comment_b_id: jsonpath "$.id"

# Like Comment B (2 times)
POST http://localhost:8080/comments/{{comment_b_id}}/like
Content-Type: application/json

{ "user_id": "charlie" }
HTTP 204

POST http://localhost:8080/comments/{{comment_b_id}}/like
Content-Type: application/json

{ "user_id": "alice" }
HTTP 204

# Upvote Comment A (1 time)
POST http://localhost:8080/comments/{{comment_a_id}}/upvote
Content-Type: application/json

{ "user_id": "bob" }
HTTP 204

# Reply to Comment A
POST http://localhost:8080/comments
Content-Type: application/json

{
  "content": "Reply to A from Charlie",
  "user_id": "charlie",
  "parent_id": "{{comment_a_id}}"
}
HTTP 201

# List by likes, Comment B should come first
GET http://localhost:8080/comments?thread_id={{comment_b_id}}&sort=likes
HTTP 200
[Asserts]
jsonpath "$.comments[0].id" == "{{comment_b_id}}"
jsonpath "$.comments[0].likes" == 2

# List by upvotes, Comment A should come first
GET http://localhost:8080/comments?thread_id={{comment_a_id}}&sort=upvotes
HTTP 200
[Asserts]
jsonpath "$.comments[0].id" == "{{comment_a_id}}"
jsonpath "$.comments[0].upvotes" == 1
