# Gowatch – Metrics Collector, Alerting Engine & REST Dashboard
Gowatch is a distributed metrics collector built with gRPC, MySQL, and REST APIs.
It includes three engineering challenges:
Challenge 1 – gRPC Metrics Streaming
Challenge 2 – Real-time Alerting Engine (Fan-Out Workers)
Challenge 3 – REST Dashboard + MySQL Persistence + Graceful Shutdown
This project simulates a real production monitoring agent and backend system.

## Project Structure
```
gowatch/
 ├── cmd/api/main.go          # Orchestrates REST + gRPC + graceful shutdown
 ├── internal/
 │   ├── grpc/                # gRPC server, workers, routes, alert logic
 │   ├── database/            # MySQL implementation
 │   └── proto/               # Generated protobuf code
 ├── proto/metrics.proto      # MetricsService definition
 ├── .env                     # Environment variables
 ├── README.md                # Project documentation (you are here)
 └── go.mod
```

## Prerequisites
✔ Go 1.21+

✔ MySQL 8+

✔ Protocol Buffers

```brew install protobuf```

✔ Stress testing tool (macOS)

```brew install stress-ng```

## Environment Variables (Required)
Create a .env file in the project root:

```
DB_USER=your_db_username
DB_PASSWORD=your_db_password
DB_HOST=localhost_or_ip_addr
DB_PORT=your_db_port
DB_NAME=your_db_name
```
To load env variables, add in go.mod (if not added yet):
```go get github.com/joho/godotenv```

## MySQL Setup
Open MySQL:
```
CREATE DATABASE gowatch;
USE gowatch;

CREATE TABLE alerts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    agent_id VARCHAR(255),
    service_name VARCHAR(255),
    rule_name VARCHAR(255),
    metric VARCHAR(50),
    value DOUBLE,
    threshold DOUBLE,
    timestamp BIGINT
);

CREATE INDEX idx_service_timestamp ON alerts (service_name, timestamp);
```
Verify:
```SELECT * FROM alerts;```

## Running the Application
Run the full system:
```go run cmd/api/main.go```

This launches:

✔ gRPC server on :50051

✔ REST server on :8080

✔ 10 worker goroutines

✔ Graceful shutdown handler

## Challenge 1 — gRPC Metrics Streaming
Proto file (metrics.proto)

The service:
```rpc SendMetrics(stream MetricReport) returns (Summary);```

Agents send metrics continuously:
```stream.Send(&MetricReport{CpuUsage: 92.5})```

Backend receives them and fans them out through a channel:
```metricChan <- metric```

## Challenge 2 — Real-Time Alerting Engine (Fan-Out Workers)
Alert flow:
1. Agent streams metrics → gRPC receives it
2. Backend pushes metric into metricChan
3. Worker pool (StartWorkers(10)) consumes metrics
4. Applies alert rules
5. Logs “Alert Triggered” events
6. Saves to DB (InsertAlert)

To test CPU alerting:
```brew install stress-ng```

Run load:
```stress-ng --cpu 0 --timeout 30```

Your agent should report high CPU → backend should generate alerts.

Check DB:
```SELECT * FROM alerts ORDER BY id DESC;```


## Challenge 3 — REST API + SQL Persistence + Dashboard
REST server automatically exposes:

*GET /health*

Returns DB health:
```{ "database": "UP" }```

*GET /alerts/history*

Returns all stored alerts:

```
[
  {
    "id": 1,
    "agent_id": "agent-123",
    "rule_name": "HighCPU",
    "metric": "cpu_usage",
    "value": 95,
    "threshold": 80,
    "timestamp": 1708350292
  }
]
```
*GET /*

Basic Hello World


## Manual End-to-End Testing
1. Start backend
```go run cmd/api/main.go```
2. Start your agent
```go run cmd/agent/main.go```
3. Load test your machine
```stress-ng --cpu 0 --timeout 30```
4. Check backend logs
You should see:
```[ALERT] High CPU detected from agent-1```
5. Check REST API
```curl localhost:8080/alerts/history```
6. Check DB
```SELECT * FROM alerts;```

## Graceful Shutdown Verification
Press Ctrl + C.

Expected log output:

```
Shutting down...
gRPC server gracefully stopped
DB connection closed
Metric channel closed
Shutdown complete
```


## Development Commands
Start app:
```
go run cmd/api/main.go
# in another terminal
go run cmd/agent/load_test_agents.go
```

