#!/bin/bash

API=http://localhost:8080
DOC_ID="doc_$(date +%s)"

echo "Starting FSM test for document: $DOC_ID"
echo

# Start SSE subscription in background
echo "[SUBSCRIBE] Listening for deltas..."
curl -N "$API/subscribe?doc_id=$DOC_ID" > sse_output.txt &
SSE_PID=$!
sleep 1

# Step 1: Alice inserts "Hello "
echo "[EDIT] Alice inserts 'Hello '"
curl -s -X POST "$API/edit" -H "Content-Type: application/json" -d '{
  "doc_id": "'"$DOC_ID"'",
  "edit": { "user": "alice", "op": "insert", "index": 0, "text": "Hello" }
}' | jq
sleep 1

# Step 2: Bob deletes "lo" from index 3
echo "[EDIT] Bob deletes 'lo'"
curl -s -X POST "$API/edit" -H "Content-Type: application/json" -d '{
  "doc_id": "'"$DOC_ID"'",
  "edit": { "user": "bob", "op": "delete", "index": 3, "text": "lo" }
}' | jq
sleep 1

# Step 3: Alice inserts "lo world!" at index 3
echo "[EDIT] Alice inserts 'lo world!'"
curl -s -X POST "$API/edit" -H "Content-Type: application/json" -d '{
  "doc_id": "'"$DOC_ID"'",
  "edit": { "user": "alice", "op": "insert", "index": 3, "text": "lo world!" }
}' | jq
sleep 1

# Fetch final state
echo
echo "[STATE] Final document:"
curl -s "$API/document?doc_id=$DOC_ID" | jq

# Kill SSE and show deltas
echo
echo "[CLEANUP] Stopping subscription..."
kill $SSE_PID
sleep 1

echo
echo "[SSE OUTPUT] Deltas received:"
cat sse_output.txt

rm -f sse_output.txt
echo
echo "Test complete."
