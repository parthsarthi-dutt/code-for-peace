import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import encoding from 'k6/encoding';
import crypto from 'k6/crypto';

// ──────────────────────────────────────────────
//  CONFIG
// ──────────────────────────────────────────────
const BASE_URL   = __ENV.BASE_URL   || 'http://localhost:8080';
const JWT_SECRET = __ENV.JWT_SECRET || 'super-secret-key';

const PROBLEM_IDS = ['123-A', '124-A', '125-B'];

// Custom metrics
const errorRate          = new Rate('errors');
const publicDuration     = new Trend('public_api_duration', true);
const authDuration       = new Trend('auth_api_duration', true);
const submissionDuration = new Trend('submission_duration', true);
const profileDuration    = new Trend('profile_duration', true);
const queueLatency       = new Trend('redis_queue_latency', true);
const requestCount       = new Counter('total_requests');
const submissionCount    = new Counter('submissions_sent');

// ──────────────────────────────────────────────
//  STRESS TEST — find the breaking point
//  Run:  k6 run --out influxdb=http://localhost:8086/k6 testing/k6.js
// ──────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '30s', target: 20  },   // warm-up
    { duration: '1m',  target: 50  },   // moderate load
    { duration: '1m',  target: 100 },   // heavy load
    { duration: '1m',  target: 200 },   // stress / breaking point
    { duration: '30s', target: 0   },   // recovery
  ],
  thresholds: {
    http_req_duration:    ['p(95)<5000'],
    errors:               ['rate<0.3'],
    redis_queue_latency:  ['p(95)<3000'],
  },
};

// ──────────────────────────────────────────────
//  HELPERS
// ──────────────────────────────────────────────
function b64url(input) {
  return encoding.b64encode(input, 'rawurl');
}

function generateJWT(userId, username) {
  const header  = JSON.stringify({ alg: 'HS256', typ: 'JWT' });
  const payload = JSON.stringify({
    user_id: userId, username: username,
    exp: Math.floor(Date.now() / 1000) + 86400,
  });
  const h   = b64url(header);
  const p   = b64url(payload);
  const sig = b64url(crypto.hmac('sha256', JWT_SECRET, `${h}.${p}`, 'binary'));
  return `${h}.${p}.${sig}`;
}

function authHeaders(token) {
  return { headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' } };
}

function jsonHeaders() {
  return { headers: { 'Content-Type': 'application/json' } };
}

function randomProblem() {
  return PROBLEM_IDS[Math.floor(Math.random() * PROBLEM_IDS.length)];
}

// ──────────────────────────────────────────────
//  DEFAULT — runs per VU iteration
// ──────────────────────────────────────────────
export default function () {
  const userId   = __VU;
  const username = `stress_vu${__VU}_${__ITER}`;
  const token    = generateJWT(userId, username);

  // ── 1. List Problems ─────────────────────────
  group('Stress — List Problems', () => {
    const res = http.get(`${BASE_URL}/api/problems`, jsonHeaders());
    requestCount.add(1);
    publicDuration.add(res.timings.duration);

    const ok = check(res, { 'list problems 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  // ── 2. Get Problem Detail ────────────────────
  group('Stress — Problem Detail', () => {
    const res = http.get(`${BASE_URL}/api/problem?id=${randomProblem()}`, jsonHeaders());
    requestCount.add(1);
    publicDuration.add(res.timings.duration);

    const ok = check(res, { 'problem detail 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  // ── 3. Rapid-fire authenticated requests ─────
  group('Stress — Rapid Auth API calls', () => {
    const endpoints = [
      `${BASE_URL}/user/submissions?id=${userId}`,
      `${BASE_URL}/api/user/profile`,
    ];

    for (const url of endpoints) {
      const res = http.get(url, authHeaders(token));
      requestCount.add(1);
      authDuration.add(res.timings.duration);

      const ok = check(res, {
        [`${url} status OK`]: (r) => r.status === 200,
      });
      errorRate.add(!ok);
    }
  });

  // ── 4. Submit Code (pushes to Redis queue) ───
  group('Stress — Submit Solution (Redis)', () => {
    const payload = JSON.stringify({
      submission_id: `stress-${__VU}-${__ITER}-${Date.now()}`,
      problem_id:    randomProblem(),
      user_id:       String(userId),
      language:      'cpp',
      code:          '#include<bits/stdc++.h>\nusing namespace std;\nint main(){\n  int a,b;\n  cin>>a>>b;\n  cout<<a+b;\n  return 0;\n}',
    });

    const res = http.post(`${BASE_URL}/practice/submit`, payload, authHeaders(token));
    requestCount.add(1);
    submissionCount.add(1);
    submissionDuration.add(res.timings.duration);
    queueLatency.add(res.timings.duration);

    const ok = check(res, { 'submit 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  // ── 5. User Profile (DB heavy) ──────────────
  group('Stress — Profile', () => {
    const res = http.get(`${BASE_URL}/api/user/profile`, authHeaders(token));
    requestCount.add(1);
    profileDuration.add(res.timings.duration);

    const ok = check(res, { 'profile 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(0.2);
}
