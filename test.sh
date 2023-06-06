#!/bin/bash

# Find the process ID of the process listening on port 8080
PID=$(lsof -t -i:8080)

# If a process is found, kill it
if [ -n "$PID" ]; then
    kill -9 $PID
    # Wait for a few seconds to make sure the process has been killed
    sleep 2
fi

# Run the Go server
echo "Starting server..."
go run main.go

