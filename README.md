# CodeForPeace - Advanced Agentic Online Judge & AI Interview Platform

CodeForPeace is a next-generation, **distributed online judge platform** featuring a full competitive programming environment intertwined with an **Autonomous AI Technical Interviewer**. 

It allows users to submit algorithmic code securely, execute it inside highly isolated Docker containers, track their streaks and tokens, and undergo real-time spoken technical interviews conducted by advanced LLMs (Groq & Gemini) and STT/TTS models.

## ✨ Core Features

### Competitive Programming Judge
*   **Docker-Sandboxed Execution**: Submissions run securely in isolated containers (`judge-sandbox`) to prevent malicious execution while enforcing tight Memory & CPU limits.
*   **Redis-Backed Distributed Queue**: All problem submissions are published to Redis Streams. Background worker nodes consume and process these submissions asynchronously.
*   **Multi-Language Support**: Complete compilation and execution pipelines for C++, Java, and Python 3.
*   **Token Gamification System**: Users earn 🪙 tokens dynamically as they solve problems and hit milestones (10, 50, 100+ problems).

### Autonomous AI Interview System
*   **Real-time Voice Interaction**: Uses Groq Whisper (Speech-to-Text) and Piper TTS (Text-to-Speech) for incredibly fast, spoken interviews directly in the browser.
*   **Intelligent Agentic Interviewer**: Driven by Llama-3.1 (via Groq) and Gemini-2.5-Flash with smart API key rotation. The AI dynamically adapts to user skill levels (Easy, Medium, Hard).
*   **Final Performance Metrics**: Provides a comprehensive markdown-styled post-interview evaluation with specific strengths, weaknesses, and study recommendations.
*   **AI Hints & Feedback**: Users can spend earned tokens to get precise algorithm hints and complexity optimizations without revealing the direct solution.

### Modern Gamified UI
*   **Vite + React + Monaco Editor**: Blazing fast frontend with full syntax highlighting, problem statements, real-time submission verdicts, and a beautiful Dark Mode aesthetic.
*   **React Portals & Modals**: Centered UI components ensuring pixel-perfect UX for modals and alerts.

---

## 🏗 System Architecture

1.  **React Frontend**: Handles the gamified UI, Monaco Code Editor, and the Audio Recording pipeline for interviews.
2.  **Go Backend**: The central API layer handling OAuth, PostgreSQL interactions, Rate Limiting (Redis), and distributing code payloads.
3.  **Go Worker Nodes**: Consumes code submissions from Redis Streams. Compiles, mounts files into Docker sandboxes, and verifies inputs vs outputs.
4.  **Python AI Service (gRPC)**: A microservice running on port `50051`. Receives requests from the Go backend. It rotates API keys to generate hints, interview questions, and parse audio data via Whisper.

---

## 🚀 Deployment (Docker Compose)

The project includes a ready-to-run `docker-compose.yml` to instantly spin up the Postgres database, Redis cache, the Python AI Service, and the Go Backend.

### Prerequisites
*   Docker & Docker Compose installed.
*   `.env` file created in the root directory (containing DB credentials, Groq, and Gemini API keys).

### Running Locally
```bash
# Start all services in detached mode
docker-compose up -d --build

# The React frontend can be started independently:
cd frontend
npm install
npm run dev
```

### Important Deployment Note (Docker-in-Docker)
The Go backend **requires access to the Docker daemon** to spawn isolated sandboxes for user code execution. 
*   If deploying to a PaaS like **Render** or **Heroku**, standard Web Services *do not* grant access to the host's Docker socket. 
*   **Recommendation**: Deploy the Go backend on a VPS (like an AWS EC2 instance or DigitalOcean Droplet) where the backend has root access to `docker run`.

---

## 🚦 Rate Limiting & Usage Restraints
To maintain infrastructure stability, the following Redis-backed rate limits are enforced:
*   **Code Run**: 1 execution per 10 seconds.
*   **Code Submit**: 1 submission per 15 seconds.
*   **AI Hints & Feedback**: Limited to exactly 5 requests per day, resetting on a rolling 24-hour window.

---

## 🛠 Tech Stack

*   **Backend**: Go (Golang), gRPC
*   **AI Microservice**: Python 3, Groq (Llama-3, Whisper), Google Gemini
*   **Frontend**: React, Vite, Monaco Editor, Lucide Icons, React Markdown
*   **Infrastructure**: Docker, Redis Streams
*   **Database**: PostgreSQL
*   **Auth**: Google OAuth 2.0
