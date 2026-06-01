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
//  SOAK TEST PROFILE
//  Simulates sustained moderate load over a long
//  period to detect memory leaks, connection pool
//  exhaustion, and gradual degradation.
// ──────────────────────────────────────────────
export const options = {
  stages: [
    { duration: '1m',  target: 20 },   // ramp up
    { duration: '10m', target: 20 },   // sustained load for 10 minutes
    { duration: '1m',  target: 0  },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],  // 95th percentile < 3s
    errors:            ['rate<0.05'],   // stricter — error rate < 5%
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

function randomProblemId() {
  return PROBLEM_IDS[Math.floor(Math.random() * PROBLEM_IDS.length)];
}

// ──────────────────────────────────────────────
//  DEFAULT — realistic user flow over long duration
// ──────────────────────────────────────────────
export default function () {
  const token = generateJWT(__VU, `k6_soak_${__VU}`);

  // Simulate a full user session
  group('Soak — Browse Problems', () => {
    const res = http.get(`${BASE_URL}/api/problems`);
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { 'list problems 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(1);

  group('Soak — View Problem', () => {
    const res = http.get(`${BASE_URL}/api/problem?id=${randomProblemId()}`);
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { 'get problem 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(2); // simulate user reading the problem

  group('Soak — Submit Solution', () => {
    const payload = JSON.stringify({
      submission_id: `soak-${__VU}-${__ITER}-${Date.now()}`,
      problem_id: randomProblemId(),
      user_id: String(__VU),
      language: 'cpp',
      code: '#include<iostream>\nusing namespace std;\nint main(){int a,b;cin>>a>>b;cout<<a+b;return 0;}',
    });
    const res = http.post(`${BASE_URL}/practice/submit`, payload, authHeaders(token));
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { 'submit 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(1);

  group('Soak — Check Profile', () => {
    const res = http.get(`${BASE_URL}/api/user/profile`, authHeaders(token));
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { 'profile 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(1);

  group('Soak — User Submissions', () => {
    const res = http.get(`${BASE_URL}/user/submissions?id=${__VU}`, authHeaders(token));
    requestCount.add(1);
    requestDuration.add(res.timings.duration);
    const ok = check(res, { 'submissions 200': (r) => r.status === 200 });
    errorRate.add(!ok);
  });

  sleep(2); // longer pause between iterations to simulate realistic browsing
}
