# FamilyJournal

FamilyJournal is a full-stack application for documenting daily care activities for children with officially recognized care levels. This branch uses the Codex Go/Fiber backend together with the Gemini React frontend, plus a MySQL database via Docker Compose.

## Tech stack

- **Frontend:** React + TypeScript, Vite, TailwindCSS, Zustand, React Router, react-i18next (de/en), served by nginx.
- **Backend:** Go + Fiber with session-based authentication and CSRF protection.
- **Database:** MySQL with migrations.
- **Deployment:** Docker Compose for local development.

## Quick start (Docker)

Create a local env file first:

```bash
cp .env.example .env
```

Then start the stack:

```bash
docker compose up --build
```

Services:

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080/api/v1`
- MySQL: `localhost:3306`

## Environment variables

### Backend

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | API port | `8080` |
| `MYSQL_DSN` | MySQL connection string | **required** |
| `SESSION_SECRET` | Session secret | **required** |
| `CORS_ALLOW_ORIGINS` | Comma-separated exact origins allowed for cross-origin browser requests | `http://localhost:5173` |
| `COOKIE_SECURE` | Use secure cookies | `false` |
| `UPLOAD_DIR` | Uploads directory | `./uploads` |
| `MAX_UPLOAD_MB` | Maximum upload size | `25` |
| `ALLOWED_UPLOAD_TYPES` | Comma-separated allowed MIME types | `image/jpeg,image/png,application/pdf` |
| `DB_MAX_OPEN` | Max open DB connections | `10` |
| `DB_MAX_IDLE` | Max idle DB connections | `5` |
| `DB_MAX_LIFETIME_MINUTES` | Max connection lifetime (minutes) | `5` |
| `RATE_LIMIT_MAX` | Requests allowed per rate-limit window (`<=0` disables limiter) | `200` |
| `RATE_LIMIT_WINDOW_SECONDS` | Rate-limit window in seconds | `60` |

### Frontend

No frontend-specific environment variables are required for the default Docker setup. The production Docker setup serves the frontend behind nginx and proxies `/api`, `/uploads`, and `/healthz` to the backend container. For local Vite development, the frontend also uses `/api` and `/uploads` proxies (see `frontend/vite.config.ts`).

For cross-origin backend access outside Docker, set `CORS_ALLOW_ORIGINS` to the exact browser origins that should be allowed. Because the app uses session cookies, wildcard origins are not appropriate.

## Database migrations

Run migration SQL files against MySQL in order:

```bash
mysql -u root -p familyjournal < backend/migrations/001_init.sql
mysql -u root -p familyjournal < backend/migrations/002_session_store.sql
```

## API overview

All endpoints are namespaced under `/api/v1`.
Error responses are JSON in the form `{ "error": "message" }`.

### Auth (`/api/v1/auth`)
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/profile`
- `PUT /auth/profile` (update email and/or password; password change requires `currentPassword` and `newPassword`)

### Posts
- `GET /posts?date=YYYY-MM-DD&hashtags=tag1,tag2&persons=name1,name2&search=query`
- `POST /posts`
- `GET /posts/:id`
- `PUT /posts/:id`
- `DELETE /posts/:id`
- `POST /posts/:id/comments`
- `PUT /comments/:id`
- `DELETE /comments/:id`
- `POST /posts/:id/attachments`

### Persons and hashtags
- `GET /persons`
- `POST /persons`
- `PUT /persons/:id`
- `DELETE /persons/:id`
- `GET /hashtags`

### Admin
- `GET /admin/users`
- `PATCH /admin/users/:id/role`
- `PATCH /admin/users/:id/active`

### Other routes
- `GET /healthz`
- `GET /api/v1/attachments/:id/download` (requires authentication)

## Frontend notes

- The integrated frontend is the Gemini UI implementation, mounted at the root path `/`.
- Browser requests should go through the frontend origin in Docker (`http://localhost:3000`), which forwards API and upload traffic to the backend.

### Access scope
- Non-admin users can only read and modify their own posts, comments, persons, and hashtags.
- Admin users can read and manage items across users via the regular content endpoints.

## MySQL schema

See `backend/migrations/001_init.sql` and `backend/migrations/002_session_store.sql` for the full schema, including persistent session storage.

## Tests

Run backend tests:

```bash
cd backend
GOFLAGS=-mod=mod go test ./...
```

The backend test suite now includes both unit-style tests (fake repositories and focused package tests) and app-level integration tests that exercise the production Fiber middleware stack.
