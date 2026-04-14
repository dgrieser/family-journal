# Family Journal

Family Journal is a full-stack application for documenting daily care activities for children with officially recognized care levels. Go/Fiber backend together with React frontend, plus a MySQL database via Docker Compose.

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

If `docker compose` is not available, you can run the containers individually with plain Docker:

The published backend and frontend images do not include MySQL. Run MySQL as a separate container and connect the backend to it over the shared Docker network.

```bash
cp .env.example .env

docker network create familyjournal
docker volume create familyjournal_mysql_data
docker volume create familyjournal_uploads

docker pull ghcr.io/dgrieser/family-journal-backend:latest
docker pull ghcr.io/dgrieser/family-journal-frontend:latest

docker run -d \
  --name familyjournal-mysql \
  --network familyjournal \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=change-me \
  -e MYSQL_DATABASE=familyjournal \
  -v familyjournal_mysql_data:/var/lib/mysql \
  mysql:8.3

docker run -d \
  --name familyjournal-backend \
  --network familyjournal \
  -p 8080:8080 \
  -e MYSQL_DSN="root:${MYSQL_ROOT_PASSWORD:-change-me}@tcp(familyjournal-mysql:3306)/familyjournal?parseTime=true" \
  -e SESSION_SECRET='replace-with-long-random-secret' \
  -e COOKIE_SECURE=false \
  -e UPLOAD_DIR=/app/uploads \
  -e MAX_UPLOAD_MB=25 \
  -v familyjournal_uploads:/app/uploads \
  ghcr.io/dgrieser/family-journal-backend:latest

docker run -d \
  --name familyjournal-frontend \
  --network familyjournal \
  -p 3000:80 \
  -e MAX_UPLOAD_MB=25 \
  ghcr.io/dgrieser/family-journal-frontend:latest
```

The startup order matters: start MySQL first, then the backend, then the frontend. The frontend nginx config proxies `/api`, `/uploads`, and `/healthz` to the backend container over the shared Docker network, so the backend container must be reachable there.

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

The production Docker setup serves the frontend behind nginx and proxies `/api`, `/uploads`, and `/healthz` to the backend container. For local Vite development, the frontend also uses `/api` and `/uploads` proxies (see `frontend/vite.config.ts`).

| Variable | Description | Default |
| --- | --- | --- |
| `BACKEND_UPSTREAM` | Hostname of the backend container/service on the shared network | `familyjournal-backend` |
| `BACKEND_PORT` | Backend port used by nginx upstream proxying | `8080` |
| `MAX_UPLOAD_MB` | nginx upload body-size limit in megabytes; keep this aligned with the backend upload limit | `25` |

The default `BACKEND_UPSTREAM=familyjournal-backend` matches both the documented standalone container name and the Docker Compose service name. Keep `MAX_UPLOAD_MB` aligned between frontend and backend so nginx and the API enforce the same upload limit.

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
The API uses session cookies for authentication. For browser or custom clients, send cookies with requests and include the CSRF cookie value from `csrf_` in the `X-CSRF-Token` header for state-changing requests (`POST`, `PUT`, `PATCH`, `DELETE`).
List endpoints for posts and persons support `page` and `pageSize` query params.
`pageSize` defaults to `20` when omitted and is capped at `100`.
`GET /persons` also supports `search` to filter by partial person name matches.
These endpoints return:

```json
{
  "items": [],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "totalItems": 0,
    "totalPages": 0
  }
}
```

### Auth (`/api/v1/auth`)
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/profile`
- `PUT /auth/profile` (update email and/or password; password change requires `currentPassword` and `newPassword`)

### Posts
- `GET /posts?date=YYYY-MM-DD&startDate=YYYY-MM-DD&endDate=YYYY-MM-DD&hashtags=tag1,tag2&persons=name1,name2&search=query&page=1&pageSize=20`
- `POST /posts`
- `GET /posts/:id`
- `PUT /posts/:id`
- `DELETE /posts/:id`
- `POST /posts/:id/comments`
- `PUT /comments/:id`
- `DELETE /comments/:id`
- `POST /posts/:id/attachments` (multipart form field `files`, supports multiple files)
- `GET /attachments/:id/download`
- `DELETE /attachments/:id`

Post create/update payload:

```json
{
  "date": "2026-04-13",
  "text": "Daily journal entry text"
}
```

### Persons and hashtags
- `GET /persons?page=1&pageSize=20&search=lena`
- `POST /persons`
- `PUT /persons/:id`
- `DELETE /persons/:id`
- `GET /hashtags`
- `POST /hashtags`
- `PUT /hashtags/:id`
- `DELETE /hashtags/:id`

Person payload:

```json
{
  "name": "Lena",
  "description": "Optional description"
}
```

Hashtag payload:

```json
{
  "name": "therapy"
}
```

### Admin
- `GET /admin/users`
- `PATCH /admin/users/:id/role` with `{ "role": "admin" }` or `{ "role": "user" }`
- `PATCH /admin/users/:id/active` with `{ "is_active": true }`

### Other routes
- `GET /healthz`

### Response notes
- Post responses include nested `user`, `hashtags`, `persons`, `comments`, and `attachments` fields.
- Comment responses include a nested `user` object with the author `id` and `email`.
- Person responses include a nested `creator` object.
- Hashtag responses may include a nested `creator` object when creator metadata is available.

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
