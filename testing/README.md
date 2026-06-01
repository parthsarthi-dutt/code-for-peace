# 🧪 Online Judge — k6 Stress Testing + Redis Monitoring

Production-grade load testing for the Online Judge backend using [Grafana k6](https://k6.io/) with **real-time Grafana dashboards**, **Redis metrics**, and **RPS analysis**.

---

## 📁 Directory Structure

```
testing/
├── k6.js                          # Main stress test (100 VU + 50 RPS)
├── k6-spike.js                    # Spike test (sudden 100 VU burst)
├── k6-soak.js                     # Soak test (sustained 20 VU for 12 min)
├── docker-compose.yml             # Full observability stack
├── prometheus/
│   └── prometheus.yml             # Prometheus scrape config
├── grafana/
│   ├── dashboards/
│   │   └── k6-dashboard.json      # 30+ panel dashboard
│   └── provisioning/
│       ├── dashboards/
│       │   └── dashboards.yml
│       └── datasources/
│           └── influxdb.yml       # InfluxDB + Prometheus datasources
└── README.md
```

---

## 🚀 Quick Start

### Prerequisites

- [k6](https://k6.io/docs/getting-started/installation/) installed
- [Docker & Docker Compose](https://docs.docker.com/get-docker/) installed
- Online Judge backend running on `localhost:8080`
- Redis running on `localhost:6379`

### 1. Start the Full Observability Stack

```bash
cd testing
docker-compose up -d
```

This starts **4 containers**:

| Service | Port | Purpose |
|---------|------|---------|
| **InfluxDB** | `8086` | k6 time-series metrics |
| **Prometheus** | `9090` | Redis metrics scraping |
| **Redis Exporter** | `9121` | Exposes Redis metrics to Prometheus |
| **Grafana** | `3000` | Dashboards (auto-login, no password) |

### 2. Run the 100 VU Stress Test

```bash
# Full stress test: 100 VU ramp + 50 RPS constant load
k6 run --out influxdb=http://localhost:8086/k6 k6.js
```

### 3. View the Dashboard

Open **http://localhost:3000** → **k6** folder → **Online Judge — 100 VU Stress Test + Redis**

---

## 📊 What You'll See in Grafana

### Row 1: Live Overview (6 stat panels)
- Avg Response Time | P95 Latency | P99 Latency | Error Rate | Peak VUs | Total Requests

### Row 2: VUs, RPS & Throughput
- **Active Virtual Users** — 0→20→50→100 ramp
- **Requests Per Second** — real-time RPS graph
- **Network Throughput** — bytes sent/received

### Row 3: Response Times by Endpoint
- **P50/P95/P99 distribution** across all endpoints
- **Per-group breakdown**: Public vs Auth vs Submission vs Profile

### Row 4: Submissions & Redis Queue
- **Submission → Redis Queue Latency** (Avg + P95)
- **Submissions Per Second** — how fast code is being queued
- **Request Rate by Category** — public vs authenticated

### Row 5: Redis Server Metrics (from Prometheus)
- Redis Memory % | Connected Clients | Memory Used | Keys in DB | Blocked Clients | Uptime
- **Commands/sec** — ops throughput during stress
- **Memory Usage** (used vs RSS) — detect leaks
- **Network I/O** — input/output bytes per second
- **Client Connections** — connected vs blocked
- **Cache Hit Rate** — keyspace hits vs misses

### Row 6: Errors & HTTP Status Codes
- **Error Rate** over time
- **2xx / 4xx / 5xx** response codes per second
- **Request Timing Breakdown** — DNS, connecting, TTFB

---

## 📋 Load Profile

The main `k6.js` runs **2 scenarios back-to-back**:

### Scenario 1: Stress Ramp (100 VUs)
```
0s    → 30s:   0 → 20 VUs    (warm up)
30s   → 1m:    20 → 50 VUs   (medium load)
1m    → 2m:    50 → 100 VUs  (ramp to peak)
2m    → 4m:    100 VUs       (sustain peak)
4m    → 4m30s: 100 → 50 VUs  (step down)
4m30s → 5m:    50 → 0 VUs    (ramp down)
```

### Scenario 2: Constant 50 RPS (starts at 5m30s)
- Fires exactly **50 requests/second** for 2 minutes
- Weighted endpoint selection:
  - 30% — List Problems
  - 20% — Problem Detail
  - 20% — Submit Code (Redis queue pressure)
  - 15% — User Submissions
  - 15% — User Profile

**Total test duration: ~7m30s**

---

## ⚡ Other Test Variants

```bash
# Spike test — sudden 100 VU burst
k6 run --out influxdb=http://localhost:8086/k6 k6-spike.js

# Soak test — 20 VU for 12 minutes (detect memory leaks)
k6 run --out influxdb=http://localhost:8086/k6 k6-soak.js

# Terminal-only (no Grafana)
k6 run k6.js
```

---

## 🔧 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:8080` | Backend URL |
| `JWT_SECRET` | `super-secret-key` | JWT secret (must match backend) |

```bash
k6 run --out influxdb=http://localhost:8086/k6 -e BASE_URL=http://myserver:8080 k6.js
```

---

## 📐 Thresholds

| Metric | Limit | Meaning |
|--------|-------|---------|
| P50 Response | < 500ms | Median response under half a second |
| P95 Response | < 2s | 95th percentile under 2 seconds |
| P99 Response | < 5s | 99th percentile under 5 seconds |
| Error Rate | < 10% | Less than 10% errors |
| Public API P95 | < 1s | Fast public endpoints |
| Submission P95 | < 3s | Submission + Redis queue under 3s |
| Redis Queue P95 | < 500ms | Queue push latency under 500ms |

---

## 🧹 Teardown

```bash
cd testing
docker-compose down -v    # -v removes volumes (clears all data)
```
