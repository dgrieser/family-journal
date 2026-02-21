# Family Journal — API Calls (Codex vs Gemini, side-by-side)

This document compares equivalent frontend API actions (“things”) in **one table per thing**, with concrete request/response field names and error behavior.

**Sources used for this comparison inside this repo:**
- historical detailed API inventory in git history (`API_CALLS.md` from previous commit)
- implementation/contract notes in `BRANCH_COMPARISON_REVIEW.md`
- compatibility plan in `CODEX_BACKEND_GEMINI_FRONTEND_ALIGNMENT_PLAN.md`

---

## 1) API client baseline (cross-cutting)

| Item | Codex | Gemini |
|---|---|---|
| Client file | `frontend/src/api/client.ts` | `frontend/src/api.ts` |
| Base URL | `/api/v1` | `/api` |
| Credentials/cookies | `credentials: 'include'` | `withCredentials: true` |
| CSRF header | `X-CSRF-Token` from `csrf_` cookie | `X-Csrf-Token` from `csrf_` cookie (axios interceptor) |
| Default request body encoding | JSON by default; keeps `FormData` for multipart | axios JSON for objects, multipart for `FormData` |
| Non-2xx behavior | Tries JSON; uses `.error` if present, otherwise text; throws `Error` | axios rejects promise; UI reads `error.response?.data?.error` when available |
| 204 behavior | explicit `null` return from wrapper | axios resolves with empty data/response |

---

## 2) Login

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/auth/login` | `POST /api/login` |
| Request payload | `{ email: string, password: string }` | `{ email: string, password: string }` |
| Success response body | `User { id, email, role, active }` | `User { id, email, role, is_active }` (shape consumed by app state) |
| UI result on success | auth store user set, navigate to app | `setUser(response.data)`, navigate to `/` |
| UI result on error | thrown message displayed | login page shows error message from axios response |

---

## 3) Register

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/auth/register` | `POST /api/register` |
| Request payload | `{ email: string, password: string }` | `{ email: string, password: string }` |
| Success response body | created `User` object (not kept logged in) | success response (UI uses success state more than payload) |
| UI result on success | frontend clears user (`null`), expects admin activation | navigate to `/login` with `registrationSuccess: true` state |
| UI result on error | thrown message surfaced in register form | axios error surfaced in register form |

---

## 4) Logout

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/auth/logout` | `POST /api/logout` |
| Request payload | none | none |
| Success response body | `204 No Content` | success status (typically empty body) |
| UI result on success | local auth state cleared | `setUser(null)`, navigate `/login` |
| UI result on error | thrown message | axios error message |

---

## 5) Session check / current user fetch

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/auth/profile` | `GET /api/me` |
| Request payload | none | none |
| Success response body | `User { id, email, role, active }` | current user object used by app initialization |
| UI result on success | app/store marks user authenticated | `setUser(response.data)` |
| UI result on error | auth boot path treats as logged out | `setUser(null)` during init path |

---

## 6) Update email (profile)

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/auth/profile` | `PUT /api/me` |
| Request payload | `{ email: string }` | `{ email: string, password?: string }` (password can be empty/omitted) |
| Success response body | updated user/profile payload | updated user payload (`setUser(res.data)`) |
| UI result on success | email form shows success + updated state | success notice; user updated |
| UI result on error | thrown message shown in profile page | axios error shown in profile page |

---

## 7) Change password

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/auth/profile` | `PUT /api/me` |
| Request payload | `{ currentPassword: string, newPassword: string }` | `{ email: string, password: string }` (same profile endpoint updates password) |
| Success response body | success/updated user payload | updated user payload |
| UI result on success | password form reset + success message | password input cleared + success message |
| UI result on error | thrown message shown (e.g., wrong current password) | axios error shown |

---

## 8) List posts (timeline)

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/posts` | `GET /api/posts` |
| Query params | `date` required; optional `search`, `hashtags` (CSV), `persons` (CSV) | `date` required; optional `search`, `hashtags` (CSV), `persons` (CSV) |
| Success response body | `Post[]` (used fields include `id`, `text`, `date/created_at`, `hashtags[]`, `persons[]`, `attachments[]`) | `Post[]` for timeline cards including comments/attachments metadata used in `PostCard` |
| UI result on success | timeline list updates | `setPosts(response.data)` |
| UI result on error | thrown message on timeline | axios error on timeline |

---

## 9) Get single post by id

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/posts/:id` | backend supports `GET /api/posts/:id` |
| Request payload | none | none |
| Success response body | `Post { id, text, date, comments[], attachments[] }` | single post payload exists server-side |
| UI usage | used by `PostDetailPage` and edit preload in `PostEditorPage` | not primary UI path (Gemini uses timeline/list flow) |
| UI result on error | thrown message in detail/editor page | axios error if endpoint is called |

---

## 10) Create post

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/posts` | `POST /api/posts` |
| Request payload | JSON: `{ date: string, text: string, category: string|null, mood: string|null }` | `FormData`: `text`, `date`, `attachments` (0..n files) |
| Success response body | created `Post` object (contains `id` used for follow-up attachment upload) | success payload; UI calls `onSuccess()` and refreshes timeline |
| UI result on success | save post then optional separate upload call | clear form + refresh list |
| UI result on error | thrown message in editor | axios error shown in form |

---

## 11) Update post

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/posts/:id` | `PUT /api/posts/:id` |
| Request payload | JSON: `{ date: string, text: string, category: string|null, mood: string|null }` | `FormData`: `text`, `date`, `attachments` (0..n files) |
| Success response body | updated post payload / success status | success payload; UI calls `onSuccess()` |
| UI result on success | return to timeline/detail with updated data | clear edit form + refresh list |
| UI result on error | thrown message in editor | axios error shown in edit form |

---

## 12) Delete post

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | backend route exists, but not exposed in Codex UI (`DELETE /api/v1/posts/:id`) | `DELETE /api/posts/:id` |
| Request payload | none | none |
| Success response body | typically `204` | success status (empty/short body) |
| UI result on success | n/a in current codex frontend | `onUpdate()` refreshes timeline |
| UI result on error | n/a in current codex frontend | axios error shown in card action |

---

## 13) Add comment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/posts/:id/comments` | `POST /api/posts/:id/comments` |
| Request payload | `{ text: string }` | `{ text: string }` |
| Success response body | `Comment { id, text, author_email, created_at }` | success payload; UI refreshes via callback |
| UI result on success | clears input and re-fetches post detail | clears input and calls `onUpdate()` |
| UI result on error | thrown message in detail page | axios error in post card |

---

## 14) Delete comment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | backend route exists, but not exposed in Codex UI (`DELETE /api/v1/comments/:id`) | `DELETE /api/comments/:id` |
| Request payload | none | none |
| Success response body | typically `204` | success status |
| UI result on success | n/a in current codex frontend | calls `onUpdate()` to refresh post card |
| UI result on error | n/a in current codex frontend | axios error shown |

---

## 15) Upload attachments

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/posts/:id/attachments` | no separate upload route in normal UI flow |
| Request payload | `FormData` with repeated `files` field | files sent inline in post create/update `FormData` as `attachments` |
| Success response body | `Attachment[] { id, file_name, file_type, file_size, url }` | attachment metadata included in returned post/list payloads |
| UI result on success | attachments linked after post save | post card shows attachments directly |
| UI result on error | thrown upload error shown in editor | axios error on post form submit |

---

## 16) Download / view attachment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | direct static path via browser (`GET /uploads/:name`) | `GET /api/attachments/:id/download` |
| Request payload | none | none |
| Success response body | binary file stream/static asset | binary file stream via API |
| UI behavior | anchor navigation from post detail | images inline (`<img src=...>`), files via links |
| Error behavior | browser-level 404/403 | failed image/link load or axios/browser error |

---

## 17) List hashtags

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/hashtags` | `GET /api/hashtags` |
| Request payload | none | none |
| Success response body | hashtag collection for filters/autocomplete | hashtag collection for filters/autocomplete |
| UI callers | `TimelinePage.tsx`, `PostEditorPage.tsx` | `Timeline.tsx`, `PostForm.tsx` |
| Error behavior | thrown message in caller page | axios error in caller page/component |

---

## 18) List persons

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/persons` | `GET /api/persons` |
| Request payload | none | none |
| Success response body | person collection for filters/editor/management | person collection for filters/editor/management |
| UI callers | `TimelinePage.tsx`, `PostEditorPage.tsx`, `PersonsPage.tsx` | `Timeline.tsx`, `PostForm.tsx`, `Persons.tsx` |
| Error behavior | thrown message in caller page | axios error in caller page/component |

---

## 19) Create person

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/persons` | `POST /api/persons` |
| Request payload | `{ name: string, description: string|null }` | `{ name: string, description: string }` |
| Success response body | created person payload | created person payload |
| UI result on success | form reset + list refresh | form reset + list refresh (`fetchPersons()`) |
| UI result on error | thrown message (including duplicate-name errors) | axios error shown |

---

## 20) Update person

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/persons/:id` | `PUT /api/persons/:id` |
| Request payload | `{ name: string, description: string|null }` | `{ name: string, description: string }` |
| Success response body | updated person payload | updated person payload |
| UI result on success | inline edit save + list refresh | reset editing state + list refresh |
| UI result on error | thrown message | axios error shown |

---

## 21) Delete person

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `DELETE /api/v1/persons/:id` | `DELETE /api/persons/:id` |
| Request payload | none | none |
| Success response body | `204 No Content` | success status |
| UI result on success | remove item/refresh list | `fetchPersons()` refresh |
| UI result on error | thrown message | axios error shown |

---

## 22) Admin list users

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/admin/users` | `GET /api/admin/users` |
| Request payload | none | none |
| Success response body | User[] { id, email, role, active } | User[] { id, email, role, is_active } |
| UI result on success | admin table load | `setUsers(res.data)` |
| UI result on error | thrown message in admin page | axios error shown |

---

## 23) Admin change role

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PATCH /api/v1/admin/users/:id/role` | `PUT /api/admin/users/:id/role` |
| Request payload | `{ role: "admin" | "user" }` | `{ role: "admin" | "user" }` |
| Success response body | `204 No Content` (UI can optimistic-update) | success status; then `fetchUsers()` |
| UI result on success | role toggled in table | list reloaded |
| UI result on error | thrown message in admin page | axios error shown |

---

## 24) Admin toggle active

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PATCH /api/v1/admin/users/:id/active` | `PUT /api/admin/users/:id/active` |
| Request payload | `{ active: boolean }` | `{ is_active: boolean }` |
| Success response body | `204 No Content` (UI can optimistic-update) | success status; then `fetchUsers()` |
| UI result on success | active badge/button state updated | list reloaded |
| UI result on error | thrown message in admin page | axios error shown |

---

## Error shape comparison (explicit)

| Case | Codex | Gemini |
|---|---|---|
| Backend returns JSON error | expected format `{ error: string }`; wrapper throws that string | axios exposes `error.response.data` (typically `{ error: string }`) |
| Backend returns plain text error | wrapper throws text body | axios error falls back to message/response text |
| Unauthorized session (`401`) | call rejects; auth bootstrap/profile logic resets user to logged-out | call rejects; app init/guards set user `null` and redirect as needed |
| Validation error (`400`) | thrown validation message from `.error` | axios validation message shown in page/component |
| Server error (`5xx`) | wrapper throws parsed/fallback message | axios generic or backend-provided message |

