# Code For Peace - 60 Interview Questions & Answers

This document provides detailed answers to the 60 interview questions for the "Code For Peace" Online Judge project. It covers the current implementation and suggests improvements for a production-grade system.

## 🏛️ Architecture

**1. Describe Code For Peace in one paragraph — what problem does it solve and who is it for?**
**Answer:** Code For Peace is a distributed online judge platform built for competitive programming and technical assessments. It allows users to submit source code for algorithmic problems, securely evaluates the code inside isolated Docker containers against predefined test cases, and asynchronously returns verdicts (e.g., AC, WA, TLE). It is designed for students, interviewers, and competitive programmers who need a scalable and secure execution environment.
**Improvement:** Expand the target audience to include B2B usage (e.g., companies conducting coding interviews).

**2. Walk me through the end-to-end flow of a single code submission — from the moment the API receives it to the moment the user gets a result.**
**Answer:** 
1. Client sends code to the `/submit` API endpoint.
2. The API server stores the submission metadata in PostgreSQL (status: `pending`) via the repository layer.
3. The server pushes the `submissionID` into a Redis List (`practiceSubmission` or `contestSubmission`) using `LPUSH`.
4. A Go worker goroutine, blocked on `BRPOP`, picks up the job from Redis.
5. The worker fetches problem constraints and creates temporary files on the host.
6. The worker spawns a Docker container (`judge-sandbox`) to compile and execute the code.
7. The worker compares the container's output against expected test cases.
8. The worker updates the result and verdict in PostgreSQL.
9. The client (currently) polls the `/submission` endpoint to retrieve the final result.

**3. Why did you choose a queue-based architecture instead of executing code synchronously in the API handler?**
**Answer:** To decouple the fast API server from the slow code execution process. Synchronous execution would block HTTP handlers, leading to connection timeouts, massive memory usage, and inevitable API crashes under high load. A queue provides backpressure, ensuring the worker pool only processes what it can handle (fault isolation), while the API remains highly available.

**4. What are the main components of your system and what is each component's single responsibility?**
**Answer:** 
- **API Server (Go):** Handles HTTP routing, OAuth auth, and enqueuing jobs.
- **Redis (Queue):** Acts as an asynchronous message broker to buffer submissions.
- **Worker Pool (Go):** Consumes queue jobs, manages the lifecycle of sandboxes, and evaluates results.
- **Docker Sandbox:** Provides a secure, isolated runtime with strict resource constraints for untrusted user code.
- **PostgreSQL (DB):** Serves as the persistent source of truth for users, problems, and submission history.

**5. How does your system behave when all four workers are busy and a new submission arrives?**
**Answer:** The API server continues to accept the submission, saves it to PostgreSQL as `pending`, and pushes it to Redis. Because all workers are busy executing code, the new job waits in the Redis List. The API responds immediately to the user. Once a worker finishes its current job, it calls `BRPOP` and immediately picks up the waiting submission.
**Improvement:** Implement WebSocket/SSE so the user is actively notified when their submission goes from queue -> processing -> done.

**6. What would break first if you had to handle 2000 concurrent users instead of 200?**
**Answer:** The Docker spawn rate and the Worker Pool capacity. Creating 2000 containers on a single host would exhaust the host's CPU and memory, causing severe latency or Out-Of-Memory (OOM) kills. The Redis queue depth would spike since the 4 workers cannot keep up.
**Improvement:** Introduce horizontal scaling. Deploy the worker pool across multiple VMs (e.g., using Kubernetes) to distribute the Docker load.

**7. Why did you use four workers specifically? How would you decide the right number in production?**
**Answer:** Four workers were chosen to match the typical number of CPU cores on a standard development machine, avoiding context-switching overhead while maximizing parallel execution. In production, I would benchmark the host machine. If a server has 16 vCPUs and the code execution is CPU-bound, I'd configure ~14-16 workers per node to leave some capacity for the OS and Docker daemon.

**8. How would you add a second API server instance without breaking anything in your current design?**
**Answer:** The current design is stateless at the API layer. The database (PostgreSQL) and the queue (Redis) are external. To add a second API instance, I simply need to run a second Go server process and place a Load Balancer (like Nginx, HAProxy, or AWS ALB) in front of them.
**Improvement:** Ensure OAuth callbacks use sticky sessions if any local state is kept, though JWTs make this fully stateless.

---

## 🗄️ Redis & Queuing

**9. Why did you choose Redis as your message queue instead of a database table with a polling loop?**
**Answer:** Redis operates entirely in memory, offering sub-millisecond latency. A database polling loop requires workers to continuously run `SELECT` queries, which burns CPU, thrashes the database, and introduces polling latency. Redis provides blocking commands (`BRPOP`) which let workers sleep without consuming CPU until a job arrives.

**10. What Redis data structure did you use for the queue — a List, a Stream, or something else? Why?**
**Answer:** I used Redis Lists. The producer uses `LPUSH` (Left Push) and the consumer uses `BRPOP` (Blocking Right Pop). Lists are incredibly simple and perfect for a basic FIFO task queue where each job needs to be processed exactly once by any available worker.

**11. Explain LPUSH and BRPOP in plain English. What does the 'B' in BRPOP mean and why is it important?**
**Answer:** `LPUSH` adds a new submission to the left side of a list. `BRPOP` removes and reads a submission from the right side. The 'B' stands for Blocking. If the list is empty, the worker connection sleeps (blocks) for a specified timeout (e.g., 1 second) rather than returning immediately. This avoids a busy-wait loop that would eat up CPU cycles.

**12. What happens to jobs in the queue if Redis crashes and restarts? Did you account for this?**
**Answer:** By default, Redis holds data in memory. If it crashes, the queued jobs are lost. 
**Improvement:** Currently, we rely on the fact that submissions are stored in PostgreSQL *before* pushing to Redis. To recover, we would need a startup script to find `pending` submissions in PG and re-queue them. To improve Redis directly, we can enable AOF (Append Only File) persistence.

**13. How do you prevent a job from being lost if a worker picks it up from Redis but crashes before completing it?**
**Answer:** In the current `BRPOP` implementation, the job is removed from Redis the moment the worker receives it. If the worker panics, the job is lost (stuck in `pending` in the DB forever).
**Improvement:** Use Redis Streams (with Consumer Groups) or the `BRPOPLPUSH` (Reliable Queue pattern) command. This moves the job to a "processing" list. If the worker completes it, we remove it from the processing list. If the worker crashes, a monitoring service can push it back.

**14. What is the difference between Redis as a cache and Redis as a message queue? Which role are you using here?**
**Answer:** As a cache, Redis stores ephemeral key-value data with TTLs (Time-to-Live) and an eviction policy (like LRU) to speed up reads. As a message queue, Redis routes jobs from producers to consumers using Lists or Streams, and data should *not* be evicted. I am using it purely as a message broker.

**15. How would you monitor queue depth in Redis? What alert would you set and at what threshold?**
**Answer:** I would use the `LLEN <queue_name>` command. If the length of `practiceSubmission` grows continuously, it means workers are saturated. I would set up a Grafana alert (via Prometheus Redis Exporter) to trigger if queue depth > 100 for more than 2 minutes, signaling the need to scale workers.

**16. If two workers both call BRPOP on the same queue, can they pick up the same job? Why or why not?**
**Answer:** No, they cannot. Redis commands are atomic because Redis executes them on a single thread. When multiple clients are blocked on `BRPOP`, Redis guarantees that only one client receives the popped element.

**17. What is a Redis dead-letter queue and would you add one to this system? When would a job end up there?**
**Answer:** A Dead-Letter Queue (DLQ) is a secondary queue for messages that fail to process multiple times (e.g., a toxic submission that keeps crashing the worker). 
**Improvement:** I would add a DLQ. If a worker recovers from a panic during execution, instead of throwing the job away, it should push the `submissionID` to a `dead_letter` list so an admin can investigate the bug in the judge logic.

---

## 🐳 Docker & Sandbox

**18. Why do you run user code inside a Docker container instead of directly on the host machine?**
**Answer:** Running `system("rm -rf /")` directly on the host would destroy the server. Docker provides **Isolation** (PID, Network, Mount namespaces), **Resource Limits** (CPU/RAM cgroups), and **Security** (dropped capabilities).

**19. What is the difference between a Docker image and a Docker container in the context of your worker?**
**Answer:** The Docker Image (`judge-sandbox`) is the immutable blueprint containing the GCC compiler, runtime libraries, and a non-root user. The Container is the active, running instance of that image where a specific user's code is executed for a few milliseconds and then destroyed (`--rm`).

**20. What specific resource limits did you configure on each container — CPU, memory, network, filesystem?**
**Answer:** 
- Network: `--network=none` (No internet access).
- Memory: `--memory=256m` (Max 256MB RAM).
- CPU: `--cpus=1` (Limited to 1 CPU core).
- Filesystem: `--read-only` root filesystem with a small writable `--tmpfs=/tmp:size=64m`.
- Filesystem Mount: `-v absPath:/sandbox:ro` (Host directory mounted as Read-Only).

**21. How does Docker actually enforce resource limits under the hood?**
**Answer:** It uses Linux `cgroups` (control groups). The kernel tracks the container's processes. If the memory cgroup limit is exceeded, the Linux OOM-killer terminates the process. CPU shares slice the CPU scheduler time proportionately.

**22. What happens if a user submits an infinite loop? How does your system detect and terminate it?**
**Answer:** In Go, I wrap the `exec.Command` in a `context.WithTimeout` (e.g., based on constraints). If the context deadline is exceeded, Go sends a `SIGKILL` to the `docker run` process, which stops the container and returns a "Time Limit Exceeded" error.

**23. How long does it take to spin up a Docker container for each submission? Does that affect your 80ms average?**
**Answer:** A cold `docker run` takes about 100-300ms, heavily impacting the execution latency.
**Improvement:** Use a pre-warmed container pool. Instead of `docker run` per submission, keep N containers running, pass the code via volume/stdin, run it via `docker exec`, and clean the directory afterwards. This reduces overhead to <10ms.

**24. How do you pass the user's code into the container and get the output back?**
**Answer:** I write the code to a host directory (`server/workers/submissions/<id>`). I use a bind mount (`-v`) to attach this directory into the container. The compilation output and runtime `stdout` are captured directly by Go via the `cmd.CombinedOutput()` byte buffer.

**25. What is a Docker namespace and what does it isolate — PID, network, filesystem?**
**Answer:** Namespaces provide isolation at the kernel level.
- **PID Namespace:** The container cannot see host processes (e.g., process 1 in the container is the sandbox script).
- **Network Namespace:** The container gets its own network stack (`--network=none` means no eth0 interface).
- **Mount Namespace:** Isolates the filesystem tree.

**26. How would you prevent a user from running \`docker run\` inside their submitted code to escape the sandbox?**
**Answer:** Docker daemon socket (`/var/run/docker.sock`) is not mounted. The container runs as a non-root user (`--user=1001:1001`), and privilege escalation is blocked. 
**Improvement:** Add `--security-opt no-new-privileges` to explicitly forbid `setuid` binaries.

**27. What is a fork bomb and would your current Docker setup survive one?**
**Answer:** A fork bomb (`:(){ :|:& };:`) recursively spawns processes until the system crashes.
**Improvement:** The current setup might hit the memory limit and OOM, but a safer approach is to use Docker's `--pids-limit=64` flag to prevent the user from spawning more than 64 processes.

---

## 🐹 Go Concurrency

**28. Why did you choose Go for this project over Node.js or Python?**
**Answer:** Go's concurrency model (Goroutines) allows massive parallel processing with very low memory overhead (~2KB per goroutine). It compiles to a static binary, is blazingly fast, and has a fantastic standard library for handling OS processes (`os/exec`) which is essential for managing Docker sandboxes.

**29. How does your worker pool work in Go? What concurrency primitives did you use to manage 4 workers?**
**Answer:** The worker pool loops indefinitely, polling Redis. I use a buffered channel as a semaphore (`sem := make(chan struct{}, 4)`). Before spawning a worker goroutine, I write to the channel. If 4 goroutines are active, the channel blocks. Inside the goroutine, `defer func() { <-sem }()` frees the slot when finished.

**30. What is a goroutine and how is it different from an OS thread? Why does this matter for your worker pool?**
**Answer:** A goroutine is a lightweight thread managed by the Go runtime, not the OS. Go multiplexes thousands of goroutines onto a few OS threads (M:N scheduling). This matters because spawning OS threads is expensive; spawning goroutines is practically free, allowing highly efficient asynchronous task execution.

**31. How did you handle the case where a worker goroutine panics while executing a sandboxed job?**
**Answer:** Currently, there is no `recover()` block, meaning a panic in the judge logic would crash the entire worker process!
**Improvement:** Add a `defer func() { if r := recover(); r != nil { log.Println("Recovered:", r) } }()` at the start of the worker goroutine to ensure the worker pool stays alive even if a single job panics.

**32. What is context.Context and why is it critical for your Docker execution timeout?**
**Answer:** `context.Context` is Go's mechanism for managing deadlines, cancellation signals, and request-scoped values. Using `context.WithTimeout` attached to `exec.CommandContext` guarantees that if the code loops infinitely, the OS process is forcibly killed by Go when the timer expires.

**33. How do your 4 worker goroutines coordinate access to shared state without race conditions?**
**Answer:** They don't have shared state! Each worker receives a unique `submissionID`, writes to a unique directory (`submissions/<id>`), and runs a unique Docker container. The database handles concurrent updates. This "Share memory by communicating" philosophy prevents race conditions.

**34. What is Go's memory model and what is a data race? How did you check for races in your code?**
**Answer:** Go's memory model specifies the conditions under which reads of a variable in one goroutine can be guaranteed to observe values produced by writes to the same variable in a different goroutine. A data race occurs when two goroutines access the same variable concurrently and at least one is a write. 
**Improvement:** Run tests with `go run -race cmd/worker/main.go` to ensure no hidden race conditions exist.

---

## 🐘 PostgreSQL

**35. What data do you store in PostgreSQL? Why not store everything in Redis?**
**Answer:** PostgreSQL stores Users, Problems, Submissions, and Verdicts. Redis is an in-memory data store; if the server reboots, everything is lost. PostgreSQL provides ACID compliance, persistent storage, and complex querying (e.g., "Find all TLE submissions for User X on Problem Y").

**36. Describe the schema for storing a code submission and its result. What columns would the table have?**
**Answer:** The `submissions` table includes:
- `submission_id` (UUID, Primary Key)
- `problem_id` (String, Foreign Key)
- `user_id` (String, Foreign Key)
- `language` (String)
- `code` (Text)
- `verdict` (String - pending, AC, WA, TLE, RE)
- `execution_time` (Integer ms)
- `memory_used` (Integer KB)
- `message` (Text for compile errors)
- `created_at` (Timestamp)

**37. What indexes did you add to your submissions table and why?**
**Answer:** 
**Improvement:** We should index `user_id` (to quickly fetch a user's submission history), `problem_id` (for leaderboards), and `(user_id, problem_id)` for displaying "Solved" statuses on the problem list.

**38. How do you handle a DB write failure after a job completes — does the user get their result or not?**
**Answer:** Currently, if `repository.UpdateSubmissionResult` fails, the error is logged, and the state remains `pending`. The user never gets their result.
**Improvement:** Implement a retry mechanism with exponential backoff for database updates. If it permanently fails, write it to a DLQ or a fallback log file so it can be reconciled later.

**39. What is a PostgreSQL transaction and did you use one when persisting a submission result?**
**Answer:** A transaction is a sequence of operations performed as a single logical unit of work (Atomicity). Currently, the result update is a single SQL `UPDATE` statement, which is implicitly atomic.

**40. How did you achieve under-100ms DB query times? What did you measure and optimize?**
**Answer:** The DB schema is simple, and lookups are primarily based on Primary Keys (`submission_id`). 
**Improvement:** I would use `EXPLAIN ANALYZE` on queries fetching user history. Connection pooling via `pgxpool` guarantees connections are reused, avoiding the ~30ms TCP/SSL handshake penalty per query.

**41. What is connection pooling and why is it critical when you have 200 concurrent users hitting PostgreSQL?**
**Answer:** Connection pooling maintains a pool of active database connections to be reused by the application. Every new PostgreSQL connection spawns a new OS process, taking significant memory and time. Without pooling (`pgxpool`), 200 users hitting the DB simultaneously could exhaust PostgreSQL's `max_connections` (default 100), causing the API to crash.

---

## 🔒 Security & Auth

**42. What is OAuth 2.0 and why did you use it to secure your API instead of simple API keys?**
**Answer:** OAuth 2.0 is an authorization framework. I used Google OAuth so users don't have to create passwords on my platform (delegated identity). It is more secure, provides a better UX, and I don't have to worry about securely hashing and salting passwords.

**43. What is the difference between authentication and authorization? Which does OAuth 2.0 handle?**
**Answer:** Authentication (AuthN) proves *who* you are (e.g., logging in). Authorization (AuthZ) defines *what* you can do. OAuth 2.0 is strictly an Authorization protocol, though OpenID Connect (OIDC), which sits on top of OAuth, handles the Authentication part via ID tokens.

**44. Which OAuth 2.0 grant type did you implement — Authorization Code, Client Credentials, or something else? Why?**
**Answer:** I implemented the **Authorization Code flow**. The user clicks login, is redirected to Google, authenticates, and Google sends an authorization code to my backend callback (`/auth/google/callback`). The backend exchanges this code for an access token securely.

**45. What is an access token and a refresh token? How does your API validate an incoming access token?**
**Answer:** In this system, after OAuth, the backend generates its own JWT (Access Token). The API validates it in the `AuthMiddleware` by checking the signature using an HMAC secret. A refresh token (currently not implemented) would be a long-lived token used to get new access tokens when they expire.

**46. How do you prevent a user from submitting malicious code that reads environment variables or host files?**
**Answer:** Docker isolates the filesystem. The container runs with its own root filesystem (`/`), and the host environment variables are not injected. The bind mount (`/sandbox`) is explicitly set to Read-Only (`:ro`).

**47. What is SQL injection and how did you prevent it in your PostgreSQL queries?**
**Answer:** SQL Injection is when user input is executed as SQL commands. In Go, I use parameterized queries (via `pgxpool` or an ORM). The database driver sends the SQL string and the arguments separately to PostgreSQL, neutralizing any malicious input.

**48. How do you prevent a user from calling external network APIs inside their submitted code?**
**Answer:** I enforce `--network=none` in the `docker run` command. The container has no loopback interface mapped to the host and cannot resolve DNS or route packets to the internet.

**49. What is the principle of least privilege and how is it applied in your Docker sandbox configuration?**
**Answer:** Least privilege means giving a process only the permissions it strictly needs to function. It is applied by creating a non-root user (`useradd sandbox`) inside the Docker image and running the container as that user (`--user=1001:1001`), combined with a read-only filesystem.

**50. What is a container escape vulnerability? What steps in your setup reduce this risk?**
**Answer:** Container escape is when a process breaks out of the Docker sandbox and gains root access to the host machine. I mitigate this by NOT using the `--privileged` flag, running as a non-root user, and completely isolating the network.

---

## 🚀 Scalability & Ops

**51. You processed 165,000 requests. How did you generate this load for testing? What tool did you use?**
**Answer:** I used a load-testing tool (like `k6` or `wrk`). I wrote a script simulating Virtual Users (VUs) making POST requests to the `/submit` endpoint and concurrently checking the `/submission` endpoint over a sustained period.

**52. What does 80ms average response time mean? Is average a good metric for latency? Why or why not?**
**Answer:** 80ms average means the total time / number of requests. Average (mean) is a **terrible** metric because it hides outliers. If 99 requests take 10ms, and 1 request takes 8000ms, the average looks fine, but one user had an awful experience.

**53. What is the 95th percentile latency and what does your P95 of 275ms mean in plain English?**
**Answer:** It means that 95% of all users received their code evaluation result in 275ms or less. Only the worst 5% experienced slower times (due to GC pauses, Docker spin-up, or Redis queue blocking).

**54. What does 'zero failures' mean in your benchmark — HTTP 5xx errors, timeouts, or something else?**
**Answer:** It means zero HTTP 5xx Server Errors and zero dropped TCP connections during the load test. Every submission was successfully recorded in PostgreSQL and processed by a worker.

**55. How would you horizontally scale your worker pool if 200 concurrent users became 2000?**
**Answer:** Because Redis is decoupled from the API, I can deploy the `worker` binary to 10 separate physical servers (or AWS EC2 instances). All 10 machines will point to the same Redis URL and blindly pop jobs. The API server doesn't need to know how many workers exist.

**56. What is fault tolerance and how did you build it into your worker pool?**
**Answer:** Fault tolerance is the ability to continue operating despite failures. My worker pool isolates failures: if a user submits a C++ program that segfaults, only that specific Docker container dies. The Go worker routine reports a Runtime Error and continues pulling new jobs.

**57. What monitoring would you add to this system in production — what metrics matter most?**
**Answer:** I would integrate Prometheus and Grafana to track:
1. **Queue Depth:** Redis `LLEN`.
2. **Worker Utilization:** How many of the 4 goroutines are currently busy.
3. **P95 / P99 Latency:** HTTP response times.
4. **Error Rates:** DB connection failures, compilation failures.
5. **Host Resources:** CPU and RAM usage to prevent OOM.

**58. How would you support multiple programming languages — Python, Java, C++ — in the same system?**
**Answer:** 
1. Add a `language` field to the API payload.
2. Build separate Docker images (`judge-sandbox-python`, `judge-sandbox-java`).
3. In the worker's `sandbox.go`, use a switch statement to determine the execution command (`python3 code.py` vs `java Main`) and which Docker image to pull.

**59. What is backpressure and how would you implement it so your API doesn't accept more jobs than workers can handle?**
**Answer:** Backpressure tells the client "I am too busy, slow down." 
**Improvement:** Currently, the queue grows infinitely. I would implement a check in the API: if Redis `LLEN` > 10,000, the API returns `503 Service Unavailable` or `429 Too Many Requests`, protecting the database and Redis from running out of memory.

**60. If you had to redesign this system from scratch knowing what you know now, what would you change and why?**
**Answer:**
1. **Pre-warm Docker Containers:** To drop execution overhead from ~200ms to ~10ms.
2. **WebSocket Support:** To push results to the frontend instantly, removing the need for clients to poll the DB.
3. **Redis Streams / Kafka:** For persistent messaging, acknowledging messages so jobs are never lost if a worker dies.
4. **Structured Logging (JSON):** Using `slog` or `logrus` to make debugging distributed logs in Kibana or Grafana Loki much easier.
