# FamilyJournal — Original Implementation Prompt

The following prompt was used to generate the two rivaling implementations (Codex and Gemini branches) of the FamilyJournal application.

> **Current-state note (2026-03-01):** This is the original generation prompt. For the latest implementation status of the `codex` and `gemini` branches, see `BRANCH_COMPARISON_REVIEW.md` and `API_CALLS.md`.

---

Create a full-stack web application called **FamilyJournal** to document daily care activities for children with officially recognized care levels.

### Technical stack

- **Frontend:** React (TypeScript), built with Vite or Create React App, using React Router for routing and a lightweight state management solution (Zustand or Redux). Use TailwindCSS or another simple utility-first CSS framework. The UI must support **both German and English**, with language switch (e.g. i18n solution such as react-i18next). All labels, buttons, and messages must be localized (de/en).
- **Backend:** Go with Fiber as the HTTP framework, using session-based authentication with secure, HttpOnly cookies. Do not use JWT for authentication.
- **Database:** MySQL. Do not use field-level encryption in MySQL.
- **Deployment:** Provide Docker Compose configuration with services for backend, frontend, and MySQL.

### Functional requirements

1. **User management and roles**
   - Users can register and log in with email and password.
   - Passwords are stored with secure hashing (bcrypt).
   - Session-based authentication using cookies:
     - On successful login, create a server-side session that stores at least `user_id` and `role`.
     - Send a secure, HttpOnly session cookie to the client.
   - Roles:
     - `admin`: can manage users (view all users, change roles, deactivate users).
     - `user`: regular application user.
   - Provide endpoints and UI flows for:
     - Registration.
     - Login.
     - Logout.
     - Viewing and updating own profile.

2. **Domain entities and data model**

   Model a small social-style system for care events with the following tables (you may refine names and fields but keep the semantics):

   - `users`
     - `id`
     - `email` (unique)
     - `password_hash`
     - `role` (enum: `admin`, `user`)
     - `created_at`
     - `updated_at`
   - `persons` (care recipients, e.g. children)
     - `id`
     - `name` (display name, must be unique per user or globally; choose a consistent rule)
     - `description` (optional)
     - `created_by_user_id`
     - `created_at`
     - `updated_at`
   - `posts`
     - `id`
     - `user_id` (author)
     - `date` (the care day, separate from creation time, e.g. `DATE`)
     - `text` (main content)
     - Optional fields like `category` or `mood` are allowed.
     - `created_at`
     - `updated_at`
   - `comments`
     - `id`
     - `post_id`
     - `user_id` (author)
     - `text`
     - `created_at`
     - `updated_at`
   - `hashtags`
     - `id`
     - `name` (unique, store in lowercase)
     - `created_at`
   - `post_hashtags` (many-to-many between posts and hashtags)
     - `post_id`
     - `hashtag_id`
   - `mentions` (many-to-many between posts and persons)
     - `post_id`
     - `person_id`
   - `attachments`
     - `id`
     - `post_id`
     - `file_name`
     - `file_type`
     - `file_size`
     - `storage_path` or `url`
     - `created_at`

   Add appropriate foreign keys and indexes for efficient queries on:
   - `posts.date`
   - `posts.user_id`
   - `post_hashtags.hashtag_id`
   - `mentions.person_id`

3. **Timeline view and navigation**

   - The main view after login is a **timeline for the current day**:
     - Shows all posts for today for the logged-in user (or for all users in the same family/team, if you introduce such grouping -- but keep it simple initially).
   - Allow navigation to other days (e.g. date picker or previous/next day controls).
   - Support creating **multiple posts per day**.
   - Sort posts by creation time (newest first).

4. **Posts, hashtags, persons and comments**

   - Users can:
     - Create, edit, and delete their own posts.
     - Add comments to posts, edit/delete their own comments.
   - The post text allows:
     - `#hashtags`: when the user types `#` and starts typing, show an autocomplete dropdown of existing hashtags.
     - Unknown hashtags are automatically created when a post is saved.
   - The post text also allows:
     - `@Person` mentions: when the user types `@` and starts typing, show an autocomplete dropdown of existing `persons`.
     - If a person with that name does not exist yet, create a new `person` record when the post is saved and link it.
   - Implement backend logic to:
     - Parse the post text for `#hashtags` and `@names`.
     - Resolve or create `hashtags` and `persons`.
     - Populate `post_hashtags` and `mentions` accordingly on every post create/update.
   - Comments belong to posts and can be created by authenticated users.

5. **Attachments (images and files)**

   - Each post can have one or more attachments (images and/or documents).
   - Implement a file upload endpoint in the backend:
     - Validate file type (e.g. allow jpeg, png, pdf).
     - Validate maximum file size (configurable).
     - Store files on disk (e.g. in a volume) or another simple storage, and store the file metadata and path/URL in the `attachments` table.
   - The frontend must support:
     - Selecting multiple files for a post.
     - Showing a list/preview of attached files.
     - Download/open attachments from the post detail view.

6. **Filtering and search**

   - Implement API endpoints and UI to filter posts by:
     - One or more hashtags.
     - One or more persons.
     - Optional combination of hashtags AND persons together.
   - Implement a text search over post text (and optionally comment text), using simple `LIKE` queries in MySQL.
   - Define an endpoint pattern like:
     - `GET /api/posts?date=YYYY-MM-DD&hashtags=tag1,tag2&persons=person1,person2&search=query`
   - The timeline UI should have:
     - Filters for hashtags (multi-select).
     - Filters for persons (multi-select).
     - A search field for text.

7. **Frontend UX and internationalization (German/English)**

   - Implement the frontend with the following views:
     - Login
     - Registration
     - Timeline view (with date selection, filters, list of posts for the selected date)
     - Post creation/edit view (with text area, hashtag and person autocomplete, file upload)
     - Post detail view (showing comments and attachments)
     - Persons management view (list, create, edit, delete persons)
     - Admin view (user list, role management, deactivate/activate users)
   - All UI texts (labels, buttons, validation errors, messages, headings) must be available in **German and English**:
     - Implement an i18n mechanism (e.g. JSON translation files for `de` and `en`).
     - Provide at least example translations for typical UI texts (e.g. "Login", "Neuer Pflegeeintrag", "Save", "Speichern", etc.).
     - Add a language switcher in the UI (e.g. toggle or dropdown).
   - The React components should be designed **mobile-first**, with responsive layout that works well on smartphones, tablets, and desktop.

8. **Security and sessions**

   - Use Fiber's session capabilities to implement session-based authentication:
     - Store sessions in a configurable storage (e.g. in-memory for development, Redis or database for production).
     - Attach session data to incoming requests via middleware.
   - Protect all modifying endpoints (create/update/delete) so they are only accessible to authenticated users.
   - Implement role checks where needed (e.g. admin actions).
   - Set cookies with `HttpOnly`, `Secure` (configurable by environment), and `SameSite` attributes.
   - Implement CSRF protection for state-changing requests.

9. **Project structure and quality**

   - Backend:
     - Separate layers: HTTP handlers (controllers), services (business logic), repositories (database access), models (data structures).
     - Provide database migrations for MySQL.
   - Frontend:
     - Organize code into pages/routes, shared components (inputs, buttons, tag inputs), hooks for data fetching, and a small API client wrapper.
   - Add basic tests:
     - For registration, login, and session handling.
     - For creating posts with hashtags and persons.
     - For filter/search behavior on the backend.
   - Provide a README with:
     - Setup instructions.
     - Environment variables.
     - How to run migrations.
     - How to start all services with Docker Compose.

10. **Extensibility**

   - Design the code so that new modules (e.g. medication plan, doctor appointments, family calendar) can be added later:
     - Use a clear and modular API structure (e.g. `/api/v1/...`).
     - Keep domain logic for care posts in its own package/module so additional domains can be added alongside.

---

Use these requirements to generate:

1. The MySQL schema (CREATE TABLE statements and indexes).
2. The Go/Fiber backend project (including session-based auth, role handling, and all required endpoints).
3. The React frontend project (with German/English UI, routing, timeline, filters, post editor with autocomplete and file upload).
4. A Docker Compose setup to run the complete stack in a development environment.

---

## Current implementation status against this prompt (summary, 2026-03-01)

- Both branches implement the requested stack shape (Fiber + React + MySQL + Docker Compose), with different architectural tradeoffs.
- Codex aligns better with backend robustness and modularity goals in this prompt (interface-based repositories, tracked SQL migrations, centralized error handling, AccessScope authorization pattern).
- Gemini aligns better with frontend polish and modern tooling (React 19, Vite 7, Tailwind 4, icons, image previews, responsive layout).
- All integration blockers between the Codex backend and Gemini frontend have been resolved — the two code lines are fully API-compatible.
- The Codex implementation has been merged into `main`; the `codex` branch no longer exists.
- The active convergence strategy is: **Codex backend (`main`) + Gemini frontend**. See `TODO.md` for the full resolution log.
