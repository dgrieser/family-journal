# Family Journal — API Calls (Codex vs Gemini)

This file compares API usage **thing-by-thing** so each behavior has one table with Codex and Gemini side-by-side, including payloads, responses, and error handling.

> Notes:
> - `:id` means a path parameter.
> - Field names are shown exactly as used by each frontend.
> - Some endpoints are the same backend capability surfaced in different UI locations.

---

## 1) API client behavior (global)

| Aspect | Codex | Gemini |
|---|---|---|
| Frontend API file | `frontend/src/api/client.ts` | `frontend/src/api.ts` |
| Base path | `/api/v1` | `/api` |
| Auth/session transport | Cookie session (`credentials: 'include'`) | Cookie session via axios instance |
| CSRF | Reads `csrf_` cookie and sends `X-CSRF-Token` | Not documented as explicit custom header in this comparison file |
| JSON behavior | `application/json` default unless `FormData` | axios defaults (JSON for objects, multipart for `FormData`) |
| Non-2xx handling | Parses response body and throws `Error` (JSON `.error` preferred, else raw text) | axios promise rejection (error object with backend message when present) |
| 204 handling | Returns `null` | axios resolves with empty body as applicable |

---

## 2) Login

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/auth/login` | `POST /api/login` |
| Request payload | `{ email: string, password: string }` | `{ email: string, password: string }` |
| Success result | `User { id, email, role, active }` stored in auth store | Auth/session established; user loaded via `/api/me` flow/store |
| Frontend caller(s) | `LoginPage.tsx` → auth store `login` | `Login.tsx` |
| Error result | Thrown `Error(message)` from parsed `.error` or text body | axios error surfaced to UI; backend message shown when available |

---

## 3) Register

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/auth/register` | `POST /api/register` |
| Request payload | `{ email: string, password: string }` | `{ email: string, password: string }` |
| Success result | Returns `User` payload but frontend resets local user to `null` (await admin activation) | Registration success state/message, then login flow |
| Frontend caller(s) | `RegisterPage.tsx` | `Register.tsx` |
| Error result | Thrown `Error(message)` from API wrapper | axios error surfaced to UI |

---

## 4) Logout

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/auth/logout` | `POST /api/logout` |
| Request payload | none | none |
| Success result | `204 No Content`, clears local auth state | Session ended; UI auth state cleared |
| Frontend caller(s) | `Layout.tsx` logout action via auth store | Layout/logout action |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 5) Session/profile fetch (who am I)

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/auth/profile` | `GET /api/me` |
| Request payload | none | none |
| Success result | `User { id, email, role, active }` | Current user object for session check |
| Frontend caller(s) | `App.tsx` on mount; auth store `fetchProfile`; also profile page preload | App/store bootstrap and protected-route flows |
| Error result | Thrown `Error(message)`; unauth usually drives logged-out state | axios error handled as unauth/logged-out |

---

## 6) Update email (profile)

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/auth/profile` | `PUT /api/me` |
| Request payload | `{ email: string }` | `{ email: string }` |
| Success result | Updated profile data reflected in UI/state | Updated user/profile reflected in UI/state |
| Frontend caller(s) | `ProfilePage.tsx` email form submit | `Profile.tsx` |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 7) Change password

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/auth/profile` | `PUT /api/me` |
| Request payload | `{ currentPassword: string, newPassword: string }` | `{ password: string }` (or branch-specific equivalent password field in same `/me` update route) |
| Success result | Password changed (with current-password verification in Codex backend) | Password updated |
| Frontend caller(s) | `ProfilePage.tsx` password form | `Profile.tsx` |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 8) List posts (timeline + filters)

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/posts` with query params | `GET /api/posts` with query params |
| Query/payload fields | `date` required; optional `hashtags`, `persons`, `search` | same functional filters used in Gemini timeline |
| Success result | `Post[]` with fields including `id`, `text`, `created_at`, `hashtags[]`, `persons[]`, `attachments[]` | `Post[]` for timeline cards (plus attachment/comment info used by UI) |
| Frontend caller(s) | `TimelinePage.tsx` on query changes | `Timeline.tsx` |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 9) Get single post (detail load)

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/posts/:id` | Backend supports `GET /api/posts/:id`, but Gemini UI does not depend on a dedicated detail page |
| Request payload | none | none |
| Success result | One `Post` with comments/attachments for detail page and editor preload | Available backend capability, not primary UI path |
| Frontend caller(s) | `PostDetailPage.tsx`, `PostEditorPage.tsx` (edit preload) | N/A (timeline-centric UI) |
| Error result | Thrown `Error(message)` | axios error if called |

---

## 10) Create post

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/posts` | `POST /api/posts` |
| Request payload | JSON includes `date`, `text`, optional `category`, optional `mood` | `FormData`/payload with `date`, `text`, and attachments inline |
| Success result | New post created; then Codex may call attachment-upload endpoint separately | New post created (attachments included in same create flow) |
| Frontend caller(s) | `PostEditorPage.tsx` | `PostForm.tsx` / timeline create flow |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 11) Update post

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/posts/:id` | `PUT /api/posts/:id` |
| Request payload | JSON post fields (same schema as create for editable fields) | Updated post payload used by edit UI |
| Success result | Post updated | Post updated |
| Frontend caller(s) | `PostEditorPage.tsx` | Post edit action in timeline/card flow |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 12) Delete post

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | (Not exposed in Codex UI in documented flow) | `DELETE /api/posts/:id` |
| Request payload | n/a | none |
| Success result | n/a | Post removed from timeline after refresh/state update |
| Frontend caller(s) | n/a | `PostCard.tsx` delete action |
| Error result | n/a | axios error surfaced to UI |

---

## 13) Add comment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/posts/:id/comments` | `POST /api/posts/:id/comments` |
| Request payload | Comment text payload (field name as used by UI form) | Comment text payload |
| Success result | Comment persisted; post detail re-fetched | Comment persisted; post list/detail state refreshed |
| Frontend caller(s) | `PostDetailPage.tsx` comment form | `PostCard.tsx` / timeline comment action |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 14) Delete comment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | (Not exposed in Codex UI in documented flow) | `DELETE /api/comments/:id` |
| Request payload | n/a | none |
| Success result | n/a | Comment removed from UI after refresh/state update |
| Frontend caller(s) | n/a | `PostCard.tsx` delete-comment action |
| Error result | n/a | axios error surfaced to UI |

---

## 15) Upload attachment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/posts/:id/attachments` | No separate endpoint in frontend flow (attachments uploaded in create/update post payload) |
| Request payload | `FormData` file upload after post creation | Included inline with post submission |
| Success result | Attachment linked to post | Attachment linked to post |
| Frontend caller(s) | `PostEditorPage.tsx` attachment upload flow | `PostForm.tsx` inline file flow |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 16) Download/view attachment

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | Direct static URL navigation (e.g., `/uploads/:name`) | `GET /api/attachments/:id/download` |
| Request payload | none | none |
| Success result | Browser opens/downloads file | Browser opens/downloads file via API route |
| Frontend caller(s) | Attachment link click in post detail/list | Attachment action in post card/list |
| Error result | Browser-level 404/403 behavior | axios/browser request error surfaced as failed fetch/download |

---

## 17) List hashtags

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/hashtags` | `GET /api/hashtags` |
| Request payload | none | none |
| Success result | Hashtag list for filter UI/autocomplete | Hashtag list for filter UI/autocomplete |
| Frontend caller(s) | Timeline/editor filter controls | Timeline/filter controls |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 18) List persons

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/persons` | `GET /api/persons` |
| Request payload | none | none |
| Success result | Person list for filters/forms and persons management page | Same purpose |
| Frontend caller(s) | persons page + filter/editor UIs | `Persons.tsx` + timeline/form filters |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 19) Create person

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `POST /api/v1/persons` | `POST /api/persons` |
| Request payload | `{ name: string, description: string }` | `{ name: string, description: string }` |
| Success result | Person created; list refresh | Person created; list refresh |
| Frontend caller(s) | `PersonsPage.tsx` | `Persons.tsx` |
| Error result | Thrown `Error(message)` (includes duplicate-name style backend errors) | axios error surfaced to UI |

---

## 20) Update person

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PUT /api/v1/persons/:id` | `PUT /api/persons/:id` |
| Request payload | `{ name: string, description: string }` | `{ name: string, description: string }` |
| Success result | Person updated; list refresh | Person updated; list refresh |
| Frontend caller(s) | `PersonsPage.tsx` edit submit | `Persons.tsx` edit submit |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 21) Delete person

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `DELETE /api/v1/persons/:id` | `DELETE /api/persons/:id` |
| Request payload | none | none |
| Success result | Person deleted; list refresh | Person deleted; list refresh |
| Frontend caller(s) | `PersonsPage.tsx` delete action | `Persons.tsx` delete action |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 22) Admin: list users

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `GET /api/v1/admin/users` | `GET /api/admin/users` |
| Request payload | none | none |
| Success result | `User[]` for admin table | `User[]` for admin table |
| Frontend caller(s) | `AdminPage.tsx` | `Admin.tsx` |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 23) Admin: change user role

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PATCH /api/v1/admin/users/:id/role` | `PUT /api/admin/users/:id/role` |
| Request payload | `{ role: "admin" | "user" }` | `{ role: "admin" | "user" }` |
| Success result | Role changed; list refresh | Role changed; list refresh |
| Frontend caller(s) | `AdminPage.tsx` role action | `Admin.tsx` role action |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## 24) Admin: toggle active status

| Item | Codex | Gemini |
|---|---|---|
| Method + endpoint | `PATCH /api/v1/admin/users/:id/active` | `PUT /api/admin/users/:id/active` |
| Request payload | `{ active: boolean }` | `{ is_active: boolean }` |
| Success result | Active state changed; list refresh | Active state changed; list refresh |
| Frontend caller(s) | `AdminPage.tsx` active toggle action | `Admin.tsx` active toggle action |
| Error result | Thrown `Error(message)` | axios error surfaced to UI |

---

## Error mapping quick-reference

| Error source | Codex behavior | Gemini behavior |
|---|---|---|
| API returns JSON error body | Uses `.error` field as thrown message | axios error contains backend message in response data |
| API returns text error body | Uses raw text as thrown message | axios error message / response text surfaced |
| Non-JSON unexpected body | Generic parse fallback to text-based message | axios generic error if unable to parse structured payload |
| Unauthorized/expired session | Caller typically falls back to logged-out state after thrown error | Caller typically handles 401 via state reset / route guard |

