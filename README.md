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
2. Run the following command:
   ```bash
   docker-compose up --build
   ```
3. The application will be available at `http://localhost:3000`.
4. The backend API is at `http://localhost:8080/api`.

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
```bash
cd backend
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
