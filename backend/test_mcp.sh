#!/bin/bash

# Start the server in background
echo "Starting server..."
export GEMINI_API_KEY=key
./kube-sherlock server --port 8080 &
SERVER_PID=$!

# Wait a moment for server to start
sleep 3

# Test the MCP query endpoint
echo "Testing MCP query..."
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the health of my pods in default namespace?"}' \
  | jq '.'

# Clean up
echo "Stopping server..."
kill $SERVER_PID

echo "Test complete!"
