# FamilyJournal

FamilyJournal is a full-stack application for documenting daily care activities for children with officially recognized care levels. It ships with a Go/Fiber backend, a React (Vite + TypeScript) frontend, and a MySQL database via Docker Compose.

## Tech stack

- **Frontend:** React + TypeScript, Vite, TailwindCSS, Zustand, React Router, react-i18next (de/en).
- **Backend:** Go + Fiber with session-based authentication and CSRF protection.
- **Database:** MySQL with migrations.
- **Deployment:** Docker Compose for local development.

## Quick start (Docker)

```bash
docker compose up --build
```

Services:

- Frontend: `http://localhost:5173`
- Backend API: `http://localhost:8080/api/v1`
- MySQL: `localhost:3306`

## Environment variables

### Backend

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | API port | `8080` |
| `MYSQL_DSN` | MySQL connection string | **required** |
| `SESSION_SECRET` | Session secret | **required** |
| `COOKIE_SECURE` | Use secure cookies | `false` |
| `UPLOAD_DIR` | Uploads directory | `./uploads` |
| `MAX_UPLOAD_MB` | Maximum upload size | `10` |

### Frontend

The frontend uses `/api` and `/uploads` proxies in development (see `vite.config.ts`).

## Database migrations

Run the migration SQL file against MySQL:

```bash
mysql -u root -p familyjournal < backend/migrations/001_init.sql
```

## API overview

All endpoints are namespaced under `/api/v1`.

### Auth
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/logout`
- `GET /auth/profile`
- `PUT /auth/profile`

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

## MySQL schema

See `backend/migrations/001_init.sql` for the full schema, including indexes and foreign keys.

## Tests

Run backend tests:

```bash
cd backend
GOFLAGS=-mod=mod go test ./...
```
