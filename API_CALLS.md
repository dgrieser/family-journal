# Family Journal — UI API Calls

This document catalogs every API call made by the frontend UI in both the **codex** and **gemini** branches.

---

## Table of Contents

- [Codex Branch](#codex-branch)
  - [API Client](#codex-api-client)
  - [Authentication](#codex-authentication)
  - [Profile](#codex-profile)
  - [Posts](#codex-posts)
  - [Comments](#codex-comments)
  - [Attachments](#codex-attachments)
  - [Hashtags](#codex-hashtags)
  - [Persons](#codex-persons)
  - [Admin](#codex-admin)
  - [Summary](#codex-summary)
  - [Backend-Only Endpoints](#codex-backend-only-endpoints)
- [Gemini Branch](#gemini-branch)
  - [API Client](#gemini-api-client)
  - [Authentication](#gemini-authentication)
  - [Profile](#gemini-profile)
  - [Posts](#gemini-posts)
  - [Comments](#gemini-comments)
  - [Attachments](#gemini-attachments)
  - [Hashtags](#gemini-hashtags)
  - [Persons](#gemini-persons)
  - [Admin](#gemini-admin)
  - [Summary](#gemini-summary)
  - [Backend-Only Endpoints](#gemini-backend-only-endpoints)
- [Branch Comparison](#branch-comparison)

---

## Codex Branch

### Codex API Client

**File:** `frontend/src/api/client.ts`

All API calls go through a single `apiFetch()` wrapper function:

- **Base URL:** `/api/v1`
- **Credentials:** `credentials: 'include'` (session cookies sent with every request)
- **CSRF:** Reads a `csrf_` cookie and sends it as the `X-CSRF-Token` header
- **Content-Type:** Defaults to `application/json` unless the body is `FormData`
- **Error handling:** Non-OK responses have their body parsed (JSON `.error` field or raw text) and thrown as `Error`
- **204 responses:** Return `null` (no body parsing)

---

### Codex Authentication

#### 1. Login

| Property | Value |
|---|---|
| **File** | `frontend/src/stores/authStore.ts` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/auth/login` |
| **Purpose** | Authenticate user and establish a session |
| **Request Body** | `{ email: string, password: string }` |
| **Response** | `User { id, email, role, active }` — stored in Zustand auth store |
| **Called From** | `LoginPage.tsx` via `handleSubmit` |

#### 2. Register

| Property | Value |
|---|---|
| **File** | `frontend/src/stores/authStore.ts` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/auth/register` |
| **Purpose** | Create a new user account (starts inactive, must be activated by admin) |
| **Request Body** | `{ email: string, password: string }` |
| **Response** | `User` object (ignored; user set to `null` after registration) |
| **Called From** | `RegisterPage.tsx` via `handleSubmit` |

#### 3. Logout

| Property | Value |
|---|---|
| **File** | `frontend/src/stores/authStore.ts` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/auth/logout` |
| **Purpose** | Destroy server-side session |
| **Request Body** | None |
| **Response** | `204 No Content` |
| **Called From** | `Layout.tsx` logout button |

#### 4. Fetch Profile (Session Check)

| Property | Value |
|---|---|
| **File** | `frontend/src/stores/authStore.ts` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/auth/profile` |
| **Purpose** | Check for active session on app load |
| **Request Body** | None |
| **Response** | `User { id, email, role, active }` |
| **Called From** | `App.tsx` in `useEffect` on mount |

---

### Codex Profile

#### 5. Get Profile (for Edit Form)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/ProfilePage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/auth/profile` |
| **Purpose** | Load current user's email into the profile edit form |
| **Called From** | `ProfilePage.tsx` on component mount |

#### 6. Update Email

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/ProfilePage.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/v1/auth/profile` |
| **Purpose** | Update the user's email address |
| **Request Body** | `{ email: string }` |
| **Called From** | `ProfilePage.tsx` profile form submit |

#### 7. Change Password

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/ProfilePage.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/v1/auth/profile` |
| **Purpose** | Change the user's password |
| **Request Body** | `{ currentPassword: string, newPassword: string }` |
| **Called From** | `ProfilePage.tsx` password form submit |

---

### Codex Posts

#### 8. List Posts (with Filters)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/TimelinePage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/posts?date=YYYY-MM-DD[&hashtags=a,b][&persons=a,b][&search=term]` |
| **Purpose** | Fetch posts for selected date with optional filters |
| **Query Params** | `date` (required), `hashtags` (comma-separated), `persons` (comma-separated), `search` |
| **Response** | `Post[] { id, text, created_at, hashtags[], persons[], attachments[] }` |
| **Called From** | `TimelinePage.tsx` when query changes |

#### 9. Get Single Post

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostDetailPage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/posts/:id` |
| **Purpose** | Load a single post with comments and attachments |
| **Response** | `Post { id, text, date, comments[], attachments[] }` |
| **Called From** | `PostDetailPage.tsx` on mount and after adding a comment |

#### 10. Load Post for Editing

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostEditorPage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/posts/:id` |
| **Purpose** | Load post data into form fields when editing |
| **Called From** | `PostEditorPage.tsx` when `id` URL param is present |

#### 11. Create Post

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostEditorPage.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/posts` |
| **Purpose** | Create a new journal post |
| **Request Body** | `{ date: string, text: string, category: string|null, mood: string|null }` |
| **Response** | Created `Post { id, ... }` — `id` used for attachment upload |
| **Called From** | `PostEditorPage.tsx` form submit (new post mode) |

#### 12. Update Post

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostEditorPage.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/v1/posts/:id` |
| **Purpose** | Update an existing journal post |
| **Request Body** | `{ date: string, text: string, category: string|null, mood: string|null }` |
| **Called From** | `PostEditorPage.tsx` form submit (edit mode) |

---

### Codex Comments

#### 13. Add Comment

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostDetailPage.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/posts/:id/comments` |
| **Purpose** | Add a comment to a post |
| **Request Body** | `{ text: string }` |
| **Response** | `Comment { id, text, author_email, created_at }` |
| **Called From** | `PostDetailPage.tsx` comment form submit |

---

### Codex Attachments

#### 14. Upload Attachments

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostEditorPage.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/posts/:id/attachments` |
| **Purpose** | Upload files as attachments to a post |
| **Request Body** | `FormData` with `files` field (`multipart/form-data`) |
| **Response** | `Attachment[] { id, file_name, file_type, file_size, url }` |
| **Called From** | `PostEditorPage.tsx` after post save, only if files selected |

---

### Codex Hashtags

#### 15. List Hashtags (Timeline Filters)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/TimelinePage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/hashtags` |
| **Purpose** | Load hashtags for filter UI |
| **Called From** | `TimelinePage.tsx` on mount |

#### 16. List Hashtags (Editor Autocomplete)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostEditorPage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/hashtags` |
| **Purpose** | Load hashtags for `#` autocomplete in text editor |
| **Called From** | `PostEditorPage.tsx` on mount |

---

### Codex Persons

#### 17. List Persons (Timeline Filters)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/TimelinePage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/persons` |
| **Purpose** | Load persons for filter UI |
| **Called From** | `TimelinePage.tsx` on mount |

#### 18. List Persons (Editor Autocomplete)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PostEditorPage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/persons` |
| **Purpose** | Load persons for `@` autocomplete in text editor |
| **Called From** | `PostEditorPage.tsx` on mount |

#### 19. List Persons (Management Page)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PersonsPage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/persons` |
| **Purpose** | Load persons for management table |
| **Called From** | `PersonsPage.tsx` on mount and after CRUD operations |

#### 20. Create Person

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PersonsPage.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/v1/persons` |
| **Request Body** | `{ name: string, description: string|null }` |
| **Called From** | `PersonsPage.tsx` add form submit |

#### 21. Update Person

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PersonsPage.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/v1/persons/:id` |
| **Request Body** | `{ name: string, description: string|null }` |
| **Called From** | `PersonsPage.tsx` inline edit save |

#### 22. Delete Person

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/PersonsPage.tsx` |
| **Method** | `DELETE` |
| **Endpoint** | `/api/v1/persons/:id` |
| **Response** | `204 No Content` |
| **Called From** | `PersonsPage.tsx` delete button |

---

### Codex Admin

#### 23. List Users

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/AdminPage.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/v1/admin/users` |
| **Purpose** | Load all users for admin management panel |
| **Response** | `User[] { id, email, role, active }` |
| **Called From** | `AdminPage.tsx` on mount |

#### 24. Update User Role

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/AdminPage.tsx` |
| **Method** | `PATCH` |
| **Endpoint** | `/api/v1/admin/users/:id/role` |
| **Request Body** | `{ role: string }` (`"admin"` or `"user"`) |
| **Response** | `204 No Content` (optimistic local update) |
| **Called From** | `AdminPage.tsx` role toggle button |

#### 25. Update User Active Status

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/AdminPage.tsx` |
| **Method** | `PATCH` |
| **Endpoint** | `/api/v1/admin/users/:id/active` |
| **Request Body** | `{ active: boolean }` |
| **Response** | `204 No Content` (optimistic local update) |
| **Called From** | `AdminPage.tsx` activate/deactivate button |

---

### Codex Summary

#### By HTTP Method

| Method | Count | Endpoints |
|---|---|---|
| `GET` | 11 | `/auth/profile` ×2, `/posts`, `/posts/:id` ×2, `/hashtags` ×2, `/persons` ×3, `/admin/users` |
| `POST` | 7 | `/auth/login`, `/auth/register`, `/auth/logout`, `/posts`, `/posts/:id/comments`, `/posts/:id/attachments`, `/persons` |
| `PUT` | 4 | `/auth/profile` ×2, `/posts/:id`, `/persons/:id` |
| `PATCH` | 2 | `/admin/users/:id/role`, `/admin/users/:id/active` |
| `DELETE` | 1 | `/persons/:id` |

#### Unique Endpoints

| # | Method | Endpoint | Description |
|---|---|---|---|
| 1 | `GET` | `/api/v1/auth/profile` | Get current user profile |
| 2 | `POST` | `/api/v1/auth/login` | Log in |
| 3 | `POST` | `/api/v1/auth/register` | Register new account |
| 4 | `POST` | `/api/v1/auth/logout` | Log out |
| 5 | `PUT` | `/api/v1/auth/profile` | Update email or change password |
| 6 | `GET` | `/api/v1/posts` | List posts with filters |
| 7 | `POST` | `/api/v1/posts` | Create a new post |
| 8 | `GET` | `/api/v1/posts/:id` | Get a single post |
| 9 | `PUT` | `/api/v1/posts/:id` | Update a post |
| 10 | `POST` | `/api/v1/posts/:id/comments` | Add a comment |
| 11 | `POST` | `/api/v1/posts/:id/attachments` | Upload attachments |
| 12 | `GET` | `/api/v1/hashtags` | List all hashtags |
| 13 | `GET` | `/api/v1/persons` | List all persons |
| 14 | `POST` | `/api/v1/persons` | Create a person |
| 15 | `PUT` | `/api/v1/persons/:id` | Update a person |
| 16 | `DELETE` | `/api/v1/persons/:id` | Delete a person |
| 17 | `GET` | `/api/v1/admin/users` | List all users (admin) |
| 18 | `PATCH` | `/api/v1/admin/users/:id/role` | Update user role (admin) |
| 19 | `PATCH` | `/api/v1/admin/users/:id/active` | Activate/deactivate user (admin) |

### Codex Backend-Only Endpoints

These backend routes exist but have **no corresponding frontend API call**:

| Method | Endpoint | Notes |
|---|---|---|
| `DELETE` | `/api/v1/posts/:id` | Post deletion supported server-side but no delete button in UI |
| `PUT` | `/api/v1/comments/:id` | Comment editing supported server-side but not in UI |
| `DELETE` | `/api/v1/comments/:id` | Comment deletion supported server-side but not in UI |
| `GET` | `/uploads/:name` | Attachments linked via `<a href>` (direct browser navigation, not `apiFetch()`) |
| `GET` | `/healthz` | Health check endpoint, not intended for UI |

---

## Gemini Branch

### Gemini API Client

**File:** `frontend/src/api.ts`

Uses **axios** with a shared instance:

- **Base URL:** `/api` (proxied to `http://localhost:8080` in dev via Vite, served via nginx in production)
- **Credentials:** `withCredentials: true` (sends cookies with every request)
- **CSRF:** Request interceptor reads `csrf_` cookie and attaches as `X-Csrf-Token` header
- **Content-Type:** Automatically handled by axios (`application/json` or `multipart/form-data`)

---

### Gemini Authentication

#### 1. Check Current User (Session Validation)

| Property | Value |
|---|---|
| **File** | `frontend/src/App.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/me` |
| **Purpose** | Check if user is authenticated on app load |
| **Response** | On success: `setUser(response.data)`. On error: `setUser(null)`. Both: `setInitialized(true)` |

#### 2. Login

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Login.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/login` |
| **Purpose** | Authenticate user with email and password |
| **Request Body** | `{ email: string, password: string }` |
| **Response** | On success: `setUser(response.data)`, navigate to `/`. On error: display error message |

#### 3. Register

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Register.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/register` |
| **Purpose** | Register a new user account |
| **Request Body** | `{ email: string, password: string }` |
| **Response** | On success: navigate to `/login` with `{ state: { registrationSuccess: true } }` |

#### 4. Logout

| Property | Value |
|---|---|
| **File** | `frontend/src/components/Layout.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/logout` |
| **Purpose** | Log out and destroy server-side session |
| **Response** | On success: `setUser(null)`, navigate to `/login` |

---

### Gemini Profile

#### 5. Update Profile

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Profile.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/me` |
| **Purpose** | Update user's email and/or password |
| **Request Body** | `{ email: string, password: string }` (password can be empty to leave unchanged) |
| **Response** | On success: `setUser(res.data)`, show success message, clear password field |

---

### Gemini Posts

#### 6. Fetch Posts (with Filters)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Timeline.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/posts` |
| **Purpose** | Fetch posts for timeline with filters |
| **Query Params** | `date`, `search`, `hashtags` (comma-separated), `persons` (comma-separated) |
| **Response** | `setPosts(response.data)` |

#### 7. Create Post

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostForm.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/posts` |
| **Purpose** | Create a new journal post with optional attachments |
| **Request Body** | `FormData` with `text`, `date`, `attachments` (multipart/form-data) |
| **Response** | Clears form, calls `onSuccess()` callback |

#### 8. Update Post

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostForm.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/posts/:id` |
| **Purpose** | Update an existing post |
| **Request Body** | `FormData` with `text`, `date`, `attachments` (multipart/form-data) |
| **Response** | Clears form, calls `onSuccess()` callback |

#### 9. Delete Post

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostCard.tsx` |
| **Method** | `DELETE` |
| **Endpoint** | `/api/posts/:id` |
| **Purpose** | Delete a post (with confirmation dialog) |
| **Response** | Calls `onUpdate()` callback to refresh list |

---

### Gemini Comments

#### 10. Add Comment

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostCard.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/posts/:id/comments` |
| **Purpose** | Add a comment to a post |
| **Request Body** | `{ text: string }` |
| **Response** | Clears comment input, calls `onUpdate()` |

#### 11. Delete Comment

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostCard.tsx` |
| **Method** | `DELETE` |
| **Endpoint** | `/api/comments/:id` |
| **Purpose** | Delete a comment |
| **Response** | Calls `onUpdate()` to refresh |

---

### Gemini Attachments

#### 12. Download/View Attachment

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostCard.tsx` |
| **Method** | `GET` (browser-initiated via `src`/`href` attributes) |
| **Endpoint** | `/api/attachments/:id/download` |
| **Purpose** | Display images inline or offer file downloads |
| **Notes** | Images rendered in `<img>` tags, other files linked with `<a target="_blank">` |

---

### Gemini Hashtags

#### 13. List Hashtags (Timeline Filters)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Timeline.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/hashtags` |
| **Purpose** | Load hashtags for filter panel |

#### 14. List Hashtags (PostForm Autocomplete)

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostForm.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/hashtags` |
| **Purpose** | Load hashtags for `#` autocomplete |

---

### Gemini Persons

#### 15. List Persons (Timeline Filters)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Timeline.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/persons` |
| **Purpose** | Load persons for filter panel |

#### 16. List Persons (PostForm Autocomplete)

| Property | Value |
|---|---|
| **File** | `frontend/src/components/PostForm.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/persons` |
| **Purpose** | Load persons for `@` autocomplete |

#### 17. List Persons (Management Page)

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Persons.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/persons` |
| **Purpose** | Load persons for management table |

#### 18. Create Person

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Persons.tsx` |
| **Method** | `POST` |
| **Endpoint** | `/api/persons` |
| **Request Body** | `{ name: string, description: string }` |

#### 19. Update Person

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Persons.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/persons/:id` |
| **Request Body** | `{ name: string, description: string }` |

#### 20. Delete Person

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Persons.tsx` |
| **Method** | `DELETE` |
| **Endpoint** | `/api/persons/:id` |

---

### Gemini Admin

#### 21. List Users

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Admin.tsx` |
| **Method** | `GET` |
| **Endpoint** | `/api/admin/users` |
| **Purpose** | Fetch all users for admin management table |

#### 22. Update User Role

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Admin.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/admin/users/:id/role` |
| **Request Body** | `{ role: string }` (`"admin"` or `"user"`) |
| **Response** | Calls `fetchUsers()` to refresh |

#### 23. Toggle User Active Status

| Property | Value |
|---|---|
| **File** | `frontend/src/pages/Admin.tsx` |
| **Method** | `PUT` |
| **Endpoint** | `/api/admin/users/:id/active` |
| **Request Body** | `{ is_active: boolean }` |
| **Response** | Calls `fetchUsers()` to refresh |

---

### Gemini Summary

#### By HTTP Method

| Method | Count | Endpoints |
|---|---|---|
| `GET` | 6 | `/me`, `/posts`, `/hashtags`, `/persons`, `/attachments/:id/download`, `/admin/users` |
| `POST` | 6 | `/login`, `/register`, `/logout`, `/posts`, `/posts/:id/comments`, `/persons` |
| `PUT` | 5 | `/me`, `/posts/:id`, `/persons/:id`, `/admin/users/:id/role`, `/admin/users/:id/active` |
| `DELETE` | 3 | `/posts/:id`, `/persons/:id`, `/comments/:id` |

#### Unique Endpoints

| # | Method | Endpoint | Description |
|---|---|---|---|
| 1 | `GET` | `/api/me` | Get current user / session check |
| 2 | `POST` | `/api/login` | Log in |
| 3 | `POST` | `/api/register` | Register new account |
| 4 | `POST` | `/api/logout` | Log out |
| 5 | `PUT` | `/api/me` | Update email or password |
| 6 | `GET` | `/api/posts` | List posts with filters |
| 7 | `POST` | `/api/posts` | Create a new post |
| 8 | `PUT` | `/api/posts/:id` | Update a post |
| 9 | `DELETE` | `/api/posts/:id` | Delete a post |
| 10 | `POST` | `/api/posts/:id/comments` | Add a comment |
| 11 | `DELETE` | `/api/comments/:id` | Delete a comment |
| 12 | `GET` | `/api/attachments/:id/download` | Download/view attachment |
| 13 | `GET` | `/api/hashtags` | List all hashtags |
| 14 | `GET` | `/api/persons` | List all persons |
| 15 | `POST` | `/api/persons` | Create a person |
| 16 | `PUT` | `/api/persons/:id` | Update a person |
| 17 | `DELETE` | `/api/persons/:id` | Delete a person |
| 18 | `GET` | `/api/admin/users` | List all users (admin) |
| 19 | `PUT` | `/api/admin/users/:id/role` | Update user role (admin) |
| 20 | `PUT` | `/api/admin/users/:id/active` | Activate/deactivate user (admin) |

### Gemini Backend-Only Endpoints

| Method | Endpoint | Notes |
|---|---|---|
| `GET` | `/api/posts/:id` | Fetches single post by ID; the UI always fetches via list endpoint |

---

## Branch Comparison

### Key Differences

| Aspect | Codex | Gemini |
|---|---|---|
| **HTTP Client** | Custom `apiFetch()` wrapper around `fetch()` | **axios** shared instance |
| **Base URL** | `/api/v1` | `/api` (no version prefix) |
| **Auth Profile Endpoint** | `GET /api/v1/auth/profile` | `GET /api/me` |
| **Login Endpoint** | `POST /api/v1/auth/login` | `POST /api/login` |
| **Register Endpoint** | `POST /api/v1/auth/register` | `POST /api/register` |
| **Logout Endpoint** | `POST /api/v1/auth/logout` | `POST /api/logout` |
| **Profile Update** | `PUT /api/v1/auth/profile` (separate email/password calls) | `PUT /api/me` (single call for both) |
| **Post Delete** | Not exposed in UI | Exposed in UI via `PostCard.tsx` |
| **Comment Delete** | Not exposed in UI | Exposed in UI via `PostCard.tsx` |
| **Single Post View** | Dedicated `PostDetailPage.tsx` with `GET /posts/:id` | No single-post view; posts only in timeline list |
| **Attachment Upload** | Separate `POST /posts/:id/attachments` after post creation | Inline with post creation via `FormData` |
| **Attachment Download** | Direct browser navigation to `/uploads/:name` | API route `GET /api/attachments/:id/download` |
| **Post Fields** | `date`, `text`, `category`, `mood` | `date`, `text` (no category/mood) |
| **Admin Role Update** | `PATCH /admin/users/:id/role` | `PUT /admin/users/:id/role` |
| **Admin Active Update** | `PATCH /admin/users/:id/active` | `PUT /admin/users/:id/active` |
| **Admin Active Body** | `{ active: boolean }` | `{ is_active: boolean }` |
| **State Management** | Zustand store (`authStore.ts`) | Zustand store (inline in `App.tsx`) |

### Endpoint Count Comparison

| Method | Codex (unique) | Gemini (unique) |
|---|---|---|
| `GET` | 6 | 6 |
| `POST` | 7 | 6 |
| `PUT` | 3 | 5 |
| `PATCH` | 2 | 0 |
| `DELETE` | 1 | 3 |
| **Total** | **19** | **20** |

### Features Only in Codex
- Single post detail view (`PostDetailPage.tsx`)
- Separate attachment upload endpoint
- Post `category` and `mood` fields
- Comment add (via detail page)

### Features Only in Gemini
- Post deletion from UI
- Comment deletion from UI
- Attachment download via API route (vs. static file serving)
- Inline file upload with post creation
