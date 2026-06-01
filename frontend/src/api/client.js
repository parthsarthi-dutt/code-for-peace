import axios from 'axios';

// We use an empty API_BASE because Vercel will securely proxy our API requests to the AWS backend!
const API_BASE = '';

const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach JWT token to every request if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ─── Problems ────────────────────────────────────
export async function fetchProblems() {
  const res = await api.get('/api/problems');
  return res.data;
}

export async function fetchProblem(id) {
  const res = await api.get(`/api/problem?id=${id}`);
  return res.data;
}

// ─── Submissions ─────────────────────────────────
export async function submitCode(data) {
  const res = await api.post('/practice/submit', data);
  return res.data;
}

export async function runCode(data) {
  const res = await api.post('/api/judge/run', data);
  return res.data;
}

export async function getSubmission(submissionId) {
  const res = await api.get(`/submission?id=${submissionId}`);
  return res.data;
}

export async function getUserSubmissions(userId) {
  const res = await api.get(`/user/submissions?id=${userId}`);
  return res.data;
}

// ─── Profile ─────────────────────────────────────
export async function getUserProfile() {
  const res = await api.get('/api/user/profile');
  return res.data;
}

export async function updateUserProfile(username, avatarURL) {
  const res = await api.post('/api/user/profile/update', { username, avatar_url: avatarURL });
  return res.data;
}

// ─── Auth ────────────────────────────────────────
export function getGoogleLoginURL() {
  return `${API_BASE}/auth/google/login`;
}

// ─── AI Integration ────────────────────────────────
export async function getAIHint(problemId, userCode) {
  const res = await api.post('/api/ai/hint', { problem_id: problemId, user_code: userCode });
  return res.data;
}

export async function getAIFeedback(problemId, userCode) {
  const res = await api.post('/api/ai/feedback', { problem_id: problemId, user_code: userCode });
  return res.data;
}

// ─── AI Interview ──────────────────────────────────
export async function startInterview(level, duration) {
  const res = await api.post('/api/interview/start', { level, duration });
  return res.data;
}

export async function sendInterviewResponse(interviewId, audioBase64, code = "", timeUp = false, systemAction = "") {
  const res = await api.post('/api/interview/respond', {
    interview_id: interviewId,
    audio_base64: audioBase64,
    code: code,
    time_up: timeUp,
    system_action: systemAction,
  });
  return res.data;
}

export async function endInterview(interviewId) {
  const res = await api.post('/api/interview/end', { interview_id: interviewId });
  return res.data;
}

export async function getInterviews() {
  const res = await api.get('/api/interview/list');
  return res.data;
}

export async function getInterviewDetail(id) {
  const res = await api.get(`/api/interview/detail?id=${id}`);
  return res.data;
}

export default api;

