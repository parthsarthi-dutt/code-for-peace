import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import encoding from 'k6/encoding';
import crypto from 'k6/crypto';

// ──────────────────────────────────────────────
//  CONFIG
// ──────────────────────────────────────────────
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const JWT_SECRET = __ENV.JWT_SECRET || 'super-secret-key';
const PROBLEM_IDS = ['123-A', '124-A', '125-B'];

// Custom metrics
const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration', true);
const requestCount = new Counter('total_requests');

// ──────────────────────────────────────────────
//  SPIKE TEST PROFILE
//  Simulates a sudden burst of users to test
//  how the system handles unexpected traffic surges
// ──────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '10s', target: 5   },   // warm up
    { duration: '10s', target: 100 },   // SPIKE — instant surge to 100 users
    { duration: '30s', target: 100 },   // hold the spike
    { duration: '10s', target: 5   },   // crash back down
    { duration: '30s', target: 5   },   // recovery period
    { duration: '10s', target: 0   },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<5000'],  // more lenient — 95th percentile < 5s
    errors:            ['rate<0.3'],    // allow up to 30% errors during spike
  },
};

// ──────────────────────────────────────────────
//  HELPERS
// ──────────────────────────────────────────────
function b64url(input) {
  return encoding.b64encode(input, 'rawurl');
}

function hmacSHA256(key, data) {
  return crypto.hmac('sha256', key, data, 'binary');
}

function generateJWT(userId, username) {
  const header = JSON.stringify({ alg: 'HS256', typ: 'JWT' });
  const payload = JSON.stringify({
    user_id: userId, username: username,
    exp: Math.floor(Date.now() / 1000) + 86400,
  });
  const h = b64url(header);
  const p = b64url(payload);
  const sig = b64url(hmacSHA256(JWT_SECRET, `${h}.${p}`));
  return `${h}.${p}.${sig}`;
}

function authHeaders(token) {
  return { headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` } };
}

function jsonHeaders() {
  return { headers: { 'Content-Type': 'application/json' } };
}

function randomProblemId() {
  return PROBLEM_IDS[Math.floor(Math.random() * PROBLEM_IDS.length)];
}

// ──────────────────────────────────────────────
//  DEFAULT — hot path under spike
// ──────────────────────────────────────────────
export default function () {
  const token = generateJWT(__VU, `k6_spike_${__VU}`);

  // Hit the heaviest endpoints to maximize pressure
  group('Spike — List Problems', () => {
    const res = http.get(`${BASE_URL}/api/problems`);
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { '200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  group('Spike — Get Problem Detail', () => {
    const res = http.get(`${BASE_URL}/api/problem?id=${randomProblemId()}`);
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { '200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  group('Spike — Practice Submit', () => {
    const payload = JSON.stringify({
      submission_id: `spike-${__VU}-${__ITER}-${Date.now()}`,
      problem_id: randomProblemId(),
      user_id: String(__VU),
      language: 'cpp',
      code: '#include<bits/stdc++.h>\nint main(){return 0;}',
    });
    const res = http.post(`${BASE_URL}/practice/submit`, payload, authHeaders(token));
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { '200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  group('Spike — User Profile', () => {
    const res = http.get(`${BASE_URL}/api/user/profile`, authHeaders(token));
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { '200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(0.2); // minimal sleep to maximize pressure
}
