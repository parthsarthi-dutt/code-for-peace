# Online Judge System

## Overview

This project is a **distributed online judge platform** designed to securely evaluate programming submissions at scale. It allows users to submit code for problems, executes the code inside isolated Docker containers, and returns the results after running against predefined test cases.

The system is built with a **scalable backend architecture** using Go, Redis job queues, worker pools, Docker sandboxing, and PostgreSQL for persistent storage.

The primary goal of the project is to simulate a production‑grade competitive programming judge similar to platforms like Codeforces, LeetCode, or HackerRank.

---

## Key Features

- Secure code execution using Docker sandboxing
- Distributed worker system for parallel evaluation
- Redis job queue for asynchronous submission processing
- PostgreSQL database for persistent storage
- OAuth authentication for user login
- Resource isolation with CPU and memory limits
- Support for multiple programming languages
- Automatic compilation and execution
- Test case based evaluation

---

## System Architecture

The system follows a **producer–consumer architecture**.

1. A user submits code through the API.
2. The backend server validates the request.
3. The submission is pushed to a Redis queue.
4. Worker processes consume jobs from the queue.
5. Workers compile and execute the code inside Docker containers.
6. The output is compared with expected results.
7. Results are stored in PostgreSQL and returned to the user.

### Core Components

**API Server**

Handles user requests including authentication, problem retrieval, and code submissions.

**Redis Queue**

Acts as a message broker between the API server and worker pool, enabling asynchronous and scalable job processing.

**Worker Pool**

Multiple workers fetch submissions from Redis and process them concurrently.

**Docker Sandbox**

Each submission runs inside an isolated container with strict resource limits to ensure security.

**PostgreSQL Database**

Stores users, problems, submissions, and evaluation results.

---

## Tech Stack

Backend

- Go (Golang)
- REST APIs

Infrastructure

- Docker
- Redis

Database

- PostgreSQL

Authentication

- OAuth

---

## Project Structure

```
online-judge/

api/
    get_all_submissions.go      # Endpoint to fetch all submissions
    get_submissions.go          # Endpoint to fetch user submissions
    problem_endpoint.go         # Problem related APIs
    submission_endpoint.go      # Code submission API

problems/
    123-A/                      # Problem folder with testcases
    124-A/

server/

    auth/
        google.go               # Google OAuth implementation
        handler.go              # OAuth login handlers
        jwt.go                  # JWT token generation and validation
        middleware.go           # Authentication middleware

    database/
        postgres/
            migrations/         # Database schema migrations
            postgres.go        # PostgreSQL connection setup

        problems.go            # Problem database operations
        redis.go               # Redis queue configuration
        submissions.go         # Submission database operations

    docker/
        Dockerfile             # Docker image used for sandbox execution

    judge/
        fileops.go             # File creation and management for submissions
        runner.go              # Code compilation and execution logic
        sandbox.go             # Docker sandbox configuration

    models/
        judgeResult.go         # Judge result structure
        output.go              # Program output structure
        problem.go             # Problem model
        submission.go          # Submission model
        user.go                # User model

    queue/                    # Redis queue logic

    repository/
        submissions.go        # Submission repository
        user.go               # User repository

    workers/                  # Worker pool for processing submissions


.env                          # Environment variables
go.mod                        # Go module definition
go.sum                        # Go dependencies checksum
main.go                       # Application entry point
```
---

## Submission Flow

1. User submits code for a problem.
2. The server stores submission metadata.
3. Submission is pushed into Redis queue.
4. Worker retrieves job from queue.
5. Worker prepares execution environment.
6. Code is compiled if required.
7. Program runs against test cases.
8. Output is validated.
9. Result is stored in database.

---

## Security Measures

To ensure safe execution of untrusted code, the platform implements:

- Docker container isolation
- Disabled network access
- Memory limits
- CPU usage limits
- Temporary filesystem execution

These restrictions prevent malicious code from affecting the host system.

---

## Running the Project

### 1. Start PostgreSQL

```
docker run -d \
--name judge-postgres \
-e POSTGRES_USER=judgeuser \
-e POSTGRES_PASSWORD=judgepass \
-e POSTGRES_DB=onlinejudge \
-p 5435:5432 \
postgres:16
```

### 2. Start Redis

```
docker run -d -p 6379:6379 redis
```

### 3. Run the Server

```
go run cmd/server/main.go
```

### 4. Start Workers

```
go run cmd/worker/main.go
```

---

## Future Improvements

- Web frontend for problem solving
- Leaderboards and contests
- Multiple language runtimes
- Code plagiarism detection
- Rate limiting and monitoring
- Horizontal scaling using Kubernetes

---

## Learning Outcomes

This project demonstrates:

- Distributed system design
- Secure code execution
- Job queue architecture
- Containerization with Docker
- Backend development using Go
- Scalable worker processing

---

## License

This project is created for educational and portfolio purposes.

