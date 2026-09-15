# LingoLift — Adaptive English Practice

A small, playable English-learning proof of concept. LingoLift gives a prompt, accepts one learner choice, grades it immediately, explains the rule, and offers a retry or an adaptive next item. It runs deterministically with no account, API key, model download, or network connection after the images are built.

## What is included

- 16 exercises across grammar agreement, past tense, contextual vocabulary, sentence building, articles, and prepositions
- Multiple-choice, usage, and sentence-building interactions
- Immediate correctness, rule explanation, correct-answer reveal, XP, hearts, and session progress
- A skill map that explains which skills are growing or need focus
- Deterministic adaptation: an incorrect answer lowers that skill score; the next unseen exercise is selected from that same weakest skill
- Explicit handling for empty answers, malformed requests, missing sessions, and unknown exercises
- Responsive, keyboard-friendly UI (`1`–`3` selects; `Enter` checks)

## Run with Docker (recommended)

Prerequisite: Docker Desktop with Compose.

```bash
docker compose up --build
```

Open **http://localhost:3000**. Stop with `docker compose down`.

## Run locally

Prerequisites: Go 1.23+ and Python 3 (only used as a static file server).

Terminal 1:

```bash
cd backend
go run .
```

Terminal 2:

```bash
cd frontend
python -m http.server 3000
```

Open **http://localhost:3000/?api=http://localhost:8080**. The explicit, allowlisted `api` query parameter is only needed for this two-server development setup; Docker uses Nginx's same-origin `/api` proxy by default.

### Environment

No environment variables are required. The optional backend variable `PORT` changes the API port from `8080`. If you change it for local development, update `API_BASE` in `frontend/app.js` to match. No API key is accepted or stored.

## Complete the learning interaction

1. Read the instruction and prompt.
2. Pick an answer (or press `1`, `2`, or `3`).
3. Select **Check answer** or press `Enter`.
4. Read the immediate rule-based feedback.
5. If incorrect, choose **Try again** to answer the same item, or **Continue** for a different item in the weak skill.
6. Open **Skill map** to see accuracy, per-skill signals, and why the next focus was selected.

A session is intentionally in-memory. Refreshing or restarting creates a clean session.

## API

| Method | Route | Purpose |
|---|---|---|
| `GET` | `/api/health` | Liveness check |
| `POST` | `/api/sessions` | Start a five-heart session |
| `GET` | `/api/sessions/{id}/exercise` | Get the next exercise (never includes the answer) |
| `POST` | `/api/sessions/{id}/answers` | Grade `{ "exerciseId": "...", "answer": "..." }` |
| `GET` | `/api/sessions/{id}/progress` | Get XP, hearts, accuracy inputs, and skill scores |

## Tests and quality checks

```bash
cd backend
gofmt -w .
go test ./...
go vet ./...

cd ../frontend
npm install
npm test
# With the Docker app running on port 3000:
npm run test:e2e

cd ..
docker compose config
```

Backend behavior was developed in red/green slices covering playable session creation, grading and XP, empty input, adaptive follow-up, exercise variety, and the HTTP learning flow. The frontend check uses only Node built-ins and verifies syntax, core controls, API wiring, and the absence of remote runtime assets.

## Learning design

The interaction maps **practice testing** to repeated retrieval through short questions and immediate attempts. Dunlosky et al. rated practice testing as a high-utility learning technique across varied conditions. Feedback is task-focused: it says whether the response is correct, supplies the relevant rule, and indicates what to do next, following Hattie and Timperley’s account of feedback as information that helps reduce the gap between current and desired performance.

See [TECHNICAL_NOTE.md](TECHNICAL_NOTE.md) for the source mapping, architecture, adaptation rules, and limitations.

## Privacy and proof-of-concept limits

- No login, cookies, analytics, third-party scripts, LLM, microphone, or external runtime requests.
- Session answers live only in bounded backend memory and disappear on restart; random session IDs, a local-origin CORS policy, and loopback-only Docker ports reduce accidental local exposure. Do not use this PoC for sensitive or identifying learner data.
- Scores are directional practice signals, **not** a validated proficiency assessment, CEFR placement, disability diagnosis, or high-stakes decision tool.
- The finite authored bank is transparent and reliable offline, but it has limited coverage and can become familiar. Production use would need expert content review, accessibility testing, localization, persistence with consent and deletion controls, and validated measurement.

## Repository structure

```text
backend/
  api/          HTTP transport
  learning/     deterministic exercise and adaptation domain
frontend/       dependency-free web client served by Nginx
docker-compose.yml
TECHNICAL_NOTE.md
DEMO_SCRIPT.md
```
