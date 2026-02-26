# FamilyJournal — Integration TODO

Tasks required to run the **Gemini frontend** against the **Codex backend**.

Grouped by severity. Within each group, items are ordered by the change location: backend (Codex) first, then frontend (Gemini).

---

## Blockers — features are broken without these

### 1. [BACKEND] Add `GET /api/v1/attachments/:id/download` route

**What:** Gemini's `PostCard.tsx` constructs all attachment URLs as
`${api.defaults.baseURL}/attachments/${a.id}/download`
(i.e. `GET /api/v1/attachments/:id/download`, lookup by database ID).

**Codex today:** serves `GET /uploads/:name` (lookup by filename, registered outside the `/api/v1` group in `main.go`). There is no ID-based attachment route.

**Fix:** In `codex` branch, add a handler `GET /api/v1/attachments/:id` (or `…/:id/download`) that:
1. Parses the attachment ID from the path.
2. Looks up the attachment row by ID (already have `GetAttachmentForUser` by name; need a variant by ID, or expose ID-based lookup in the repository).
3. Performs the same ownership + path-traversal check as the existing handler.
4. Serves the file bytes with the correct `Content-Type`.

**Files to change:** `backend/cmd/server/main.go`, `backend/internal/handlers/posts.go`, `backend/internal/repositories/attachment_repository.go` (add `GetAttachmentByID`).

**Side task:** While touching the attachment model, rename the JSON field from `url` to `storage_path` (`json:"storage_path"`) to match Gemini's `Attachment` type. This requires updating the DB query aliases in `attachment_repository.go` and any place the `URL` field is set (attachment upload handler).

---

### 2. [FRONTEND] Post create/update: switch from FormData to JSON + separate upload

**What:** Gemini's `PostForm.tsx` submits a single multipart `FormData` request that bundles post fields (`text`, `date`) and attachment files together.

Codex's backend:
- `POST /api/v1/posts` and `PUT /api/v1/posts/:id` expect **JSON** (`Content-Type: application/json`) with `{ date, text }` only.
- Files are uploaded separately via `POST /api/v1/posts/:id/attachments` as `FormData` with a repeated `files` field.

Sending FormData to the JSON endpoint causes Codex's `c.BodyParser` to fail and return `400 invalid payload`.

**Fix:** In `gemini` branch in `PostForm.tsx`, change `handleSubmit` to:
1. POST/PUT the post as JSON: `{ text, date }`.
2. If `files.length > 0`, follow up with a multipart POST to `/posts/${postId}/attachments` using a `FormData` where each file is appended under the key `files`.
3. Call `onSuccess()` only after both steps complete.

The returned post object from step 1 contains the `id` needed for step 2.

**File to change:** `frontend/src/components/PostForm.tsx`

---

### 3. [FRONTEND] Admin actions: change `PUT` → `PATCH`

**What:** Gemini's `Admin.tsx` calls:
- `api.put(\`/admin/users/${userId}/role\`, { role })`
- `api.put(\`/admin/users/${userId}/active\`, { is_active })`

Codex registers these as `PATCH` (see `admin.Patch("/users/:id/role", …)` in `main.go`). A `PUT` request returns `405 Method Not Allowed`.

**Fix:** In `gemini` branch in `Admin.tsx`, replace both `api.put(…)` calls with `api.patch(…)`.

**File to change:** `frontend/src/pages/Admin.tsx` (2 lines)

---

### 4. [FRONTEND] Profile update: align password-change field names

**What:** Gemini's `Profile.tsx` sends `PUT /api/v1/auth/profile` with `{ email, password }`.

Codex's `UpdateProfile` handler reads:
```json
{ "email": "…", "currentPassword": "…", "newPassword": "…" }
```
It requires `currentPassword` when `newPassword` is non-empty and returns `400 currentPassword required` otherwise. Gemini's `password` field is ignored entirely, so password changes silently do nothing.

**Fix:** In `gemini` branch, update `Profile.tsx` to:
1. Add a "Current password" input field (bound to a new `currentPassword` state variable).
2. Send `{ email, currentPassword, newPassword }` instead of `{ email, password }` (use `password` state as `newPassword`).
3. Show the error message returned by Codex (e.g. "invalid credentials" when the current password is wrong).

**File to change:** `frontend/src/pages/Profile.tsx`

---

### 5. [FRONTEND] Post card: use `post.persons` instead of `post.mentions`

**What:** Codex's `Post` model serialises the persons/mentions array as `"persons"`:
```go
Persons []Person `json:"persons"`
```
Gemini's `types.ts` declares `mentions: Person[]` and `PostCard.tsx` iterates `post.mentions?.map(…)`. The `mentions` key is always `undefined` in Codex responses, so person badges never render.

**Fix:** in `gemini` branch:
- In `types.ts`: rename `mentions: Person[]` → `persons: Person[]` on the `Post` interface.
- In `PostCard.tsx`: replace `post.mentions?.map` with `post.persons?.map`.
- In `Timeline.tsx`, if `mentions` is referenced anywhere for filters: update similarly (a quick grep shows it is not — filters use `selectedPersons` string array, not the post field).

**Files to change:** `frontend/src/types.ts`, `frontend/src/components/PostCard.tsx`

---

### 6. [BACKEND] Embed author email in comment responses

**What:** Gemini's `PostCard.tsx` renders `c.user?.email` and its `Comment` type includes `user?: User`. Codex's `Comment` model returns a flat `author_email` string field instead of a nested user object, so `c.user` is always `undefined` and comment author names are blank.

**Fix:** In `codex` branch, change the `Comment` model and the query in `comment_repository.go` to embed a minimal user object:
- Add a nested struct (or reuse `models.User`) with at least `id` and `email`, serialised as `"user"`.
- Remove (or keep alongside) the flat `author_email` field — keeping it is fine for backwards compatibility, but the `user` key must be present for the Gemini frontend.
- Update the SQL query in `ListCommentsForPosts` to JOIN `users` and scan the email into the nested struct.

**Files to change:** `backend/internal/models/comment.go`, `backend/internal/repositories/post_repository.go` (comment query).

---

## Code quality — not breaking, but recommended before merging

### 8. [FRONTEND] Move `ProtectedRoute` outside the `App` component

**What:** In `App.tsx`, `ProtectedRoute` is defined as a function inside the `App` component body. React recreates the function on every render of `App`, which means React treats it as a new component type each time and unmounts + remounts the entire subtree it wraps.

**Fix:** In `gemini` branch, move the `ProtectedRoute` function definition outside of `App` (to module scope). It closes over `initialized` and `user` from the store, so switch those to `useAuthStore` calls inside the component itself.

**File to change:** `frontend/src/App.tsx`

---

### 9. [FRONTEND] Migrate i18n translations to separate JSON files

**What:** Gemini inlines all translations in `i18n.ts`. Codex uses separate `locales/en.json` and `locales/de.json`, which is the i18next best practice (easier to hand off to translators, better tree-shaking, no code changes needed to add strings).

**Fix:** In `gemini` branch:
1. Create `frontend/src/locales/en.json` and `frontend/src/locales/de.json` with the existing translation keys from `i18n.ts`.
2. Update `i18n.ts` to load translations from the JSON files using i18next's `resources` or the `i18next-http-backend` / `i18next-resources-to-backend` approach.

**Files to change:** `frontend/src/i18n.ts`; new files `frontend/src/locales/en.json`, `frontend/src/locales/de.json`
