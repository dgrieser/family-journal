# FamilyJournal

A full-stack web application to document daily care activities for children.

## Features

- **Daily Timeline**: View and create care posts for specific days.
- **Hashtags & Mentions**: Auto-parse `#hashtags` and `@persons` in posts.
- **Attachments**: Upload images and documents to posts.
- **Comments**: Add comments to care posts.
- **Internationalization**: Full support for German and English.
- **User Management**: Role-based access (Admin/User).
- **Responsive Design**: Mobile-first UI.

## Tech Stack

- **Frontend**: React (TypeScript), Vite, TailwindCSS, Zustand, i18next.
- **Backend**: Go, Fiber, GORM.
- **Database**: MySQL.
- **Authentication**: Session-based with HttpOnly cookies.

## Getting Started

### Prerequisites

- Docker and Docker Compose

### Running the Application

1. Clone the repository.
2. Create a `.env` file from the example:
   ```bash
   cp .env.example .env
   ```
   For local Docker use, the compose file now falls back to a development-only session secret if `.env` is missing. You should still set your own `SESSION_SECRET` (at least 32 characters) in `.env` for any real use.
3. Run the following command:
   ```bash
   docker-compose up --build
   ```
4. The application will be available at `http://localhost:3000`.
5. The backend API is at `http://localhost:8080/api/v1`.

### Environment Variables

The application uses the following environment variables (configured in `docker-compose.yml`):

- `DB_HOST`: Database host (default: `db`)
- `DB_PORT`: Database port (default: `3306`)
- `DB_USER`: Database user (default: `user`)
- `DB_PASSWORD`: Database password (default: `password`)
- `DB_NAME`: Database name (default: `familyjournal`)
- `SESSION_SECRET`: Secret for session encryption.
- `PORT`: Backend port (default: `8080`)

## Development

### Backend

To run the backend locally:
1. Ensure you have a MySQL instance running or update `.env` to point to a local database.
2. Set up your environment variables in `.env`.
3. Run the backend:
```bash
cd backend
# The backend uses godotenv to load .env from the root or the backend folder
go run cmd/api/main.go
```

### Frontend

To run the frontend locally:
```bash
cd frontend
npm install
npm run dev
```

### Tests

To run backend tests:
```bash
cd backend
go test ./internal/services/...
```
