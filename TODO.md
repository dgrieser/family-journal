# FamilyJournal — Integration TODO

Tasks required to run the **Gemini frontend** against the **Codex backend**.

> **Status as of 2026-03-01:** All blockers and code-quality items have been resolved. The Gemini frontend and Codex backend are now fully compatible.

---

## ✅ Resolved — Blockers

### 1. [BACKEND] Add `GET /api/v1/attachments/:id/download` route ✅ DONE

**What was needed:** Gemini's `PostCard.tsx` constructs attachment URLs as
`${api.defaults.baseURL}/attachments/${a.id}/download`
(i.e. `GET /api/v1/attachments/:id/download`, lookup by database ID).

**Codex fix (PR #21):** Added `GET /api/v1/attachments/:id/download` handler (`DownloadAttachmentByID`) in `posts.go`. The `url` field was removed from the `Attachment` model entirely (no longer stored or returned); the ID-based download route is now the canonical way to serve files. A corresponding DB migration was added. Security headers are applied to download responses (PR #22).

**Side task (URL→storage_path rename):** Superseded — the `url`/`storage_path` field was removed from the model altogether in favour of ID-based downloads. The Gemini `Attachment` type in `types.ts` was aligned to match (no `url` or `storage_path` field).

---

### 2. [FRONTEND] Post create/update: switch from FormData to JSON + separate upload ✅ DONE

**What was needed:** Gemini's `PostForm.tsx` was sending a single multipart `FormData` request bundling post fields and files together. Codex's `POST /api/v1/posts` and `PUT /api/v1/posts/:id` expect JSON.

**Gemini fix:** `handleSubmit` in `PostForm.tsx` now:
1. POSTs/PUTs the post as JSON `{ text, date }` (with explicit `Content-Type: application/json` header).
2. If files are selected, follows up with a multipart `POST /posts/${postId}/attachments` using `FormData` with `files` keys.
3. Calls `onSuccess()` only after both steps complete.

**File changed:** `frontend/src/components/PostForm.tsx`

---

### 3. [FRONTEND] Admin actions: change `PUT` → `PATCH` ✅ DONE

**What was needed:** Gemini's `Admin.tsx` was calling `api.put()` for role and active-state changes. Codex registers these as `PATCH`.

**Gemini fix (PR #27):** Both `api.put(…/role, …)` and `api.put(…/active, …)` replaced with `api.patch(…)`.

**File changed:** `frontend/src/pages/Admin.tsx`

---

### 4. [FRONTEND] Profile update: align password-change field names ✅ DONE

**What was needed:** Gemini's `Profile.tsx` was sending `{ email, password }`. Codex's `UpdateProfile` handler reads `{ email, currentPassword, newPassword }`.

**Gemini fix (PR #28):** Added a "Current password" input. The payload now sends `{ email, currentPassword, newPassword }`. Backend error messages (e.g. "invalid credentials") are surfaced to the user.

**File changed:** `frontend/src/pages/Profile.tsx`

---

### 5. [FRONTEND] Post card: use `post.persons` instead of `post.mentions` ✅ DONE

**What was needed:** Codex serialises the persons array as `"persons"`. Gemini had `mentions: Person[]` in `types.ts` and `post.mentions?.map` in `PostCard.tsx`.

**Gemini fix:**
- `types.ts`: renamed `mentions: Person[]` → `persons: Person[]` on the `Post` interface.
- `PostCard.tsx`: replaced `post.mentions?.map` with `post.persons?.map`.

**Files changed:** `frontend/src/types.ts`, `frontend/src/components/PostCard.tsx`

---

### 6. [BACKEND] Embed author email in comment responses ✅ DONE

**What was needed:** Gemini's `PostCard.tsx` renders `c.user?.email` and its `Comment` type includes `user?: User`. Codex's `Comment` model previously returned a flat `author_email` string.

**Codex fix (PR #25):** Added `CommentUser` nested struct with `id` and `email` fields, serialised as `"user"`. The flat `author_email` field is kept in the DB scan but hidden from JSON (`json:"-"`). A `HydrateUser()` method populates the nested struct from the flat fields. The hydration is called on create, update, and list operations.

**Files changed:** `backend/internal/models/comment.go`, `backend/internal/repositories/post_repository.go`

---

## ✅ Resolved — Code Quality

### 8. [FRONTEND] Move `ProtectedRoute` outside the `App` component ✅ DONE

**Gemini fix (PR #29):** `ProtectedRoute` extracted to `frontend/src/components/ProtectedRoute.tsx` (module scope). A separate `AdminRoute` component was also extracted. `App.tsx` imports both and no longer defines any routing components inline.

**Files changed:** `frontend/src/App.tsx`; new `frontend/src/components/ProtectedRoute.tsx`

---

### 9. [FRONTEND] Migrate i18n translations to separate JSON files ✅ DONE

**Gemini fix (PR #30):** `i18n.ts` now uses a dynamic `import(`./locales/${language}.json`)` backend (lazy-loaded). `frontend/src/locales/de.json` and `frontend/src/locales/en.json` contain all translation keys.

**Files changed:** `frontend/src/i18n.ts`; new `frontend/src/locales/de.json`, `frontend/src/locales/en.json`
