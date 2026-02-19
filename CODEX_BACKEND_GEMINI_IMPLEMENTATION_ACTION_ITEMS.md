# Codex Backend ↔ Gemini Frontend Alignment: Concrete Implementation Action Items

This document turns `CODEX_BACKEND_GEMINI_FRONTEND_ALIGNMENT_PLAN.md` into concrete, implementation-ready actions.

## Scope guardrails

- `category` and `mood` are out of scope for required behavior. They must not be required in backend validation, and frontend must not depend on them.
- Use Codex backend internal models as-is; expose Gemini-compatible DTOs at API boundaries.
- Keep frontend component-level changes minimal and concentrated in API client/types.

---

## 0) Create the contract matrix first (single source of truth)

### Action items

1. Add `docs/api/contract-matrix-codex-gemini.md`.
2. For each endpoint used by Gemini frontend, create a row with:
   - `Endpoint + Method` (e.g., `GET /api/posts`)
   - `Gemini current request shape`
   - `Codex current request shape`
   - `Gemini current response shape`
   - `Codex current response shape`
   - `Final compatibility request/response shape`
   - `Notes (aliases, temporary shims, deprecation date)`
3. Include explicit types in matrix examples:
   - `id: string` vs `id: number`
   - timestamps: RFC3339/ISO8601 string, UTC
   - arrays: never `null`; always `[]`
   - error: `{ "error": string }`
4. Mark each row as one of:
   - `No change`
   - `Backend shim`
   - `Frontend API normalization`
   - `Both`
5. Add an owner + completion checkbox per row.

### Endpoints to include (minimum)

- Auth/session: login, logout, me/profile, csrf/session bootstrap
- Posts: list, detail, create, update, delete
- Comments: create, delete, list-by-post
- Persons: list, create, update(rename), delete
- Attachments: upload, list-by-post, download/view metadata

---

## 1) Backend: define Gemini-compatibility DTOs and mappers

### 1.1 Add DTO package/module structure

Create a dedicated DTO namespace (example naming; adapt to repo conventions):

- `internal/api/dto/auth.go`
- `internal/api/dto/post.go`
- `internal/api/dto/comment.go`
- `internal/api/dto/person.go`
- `internal/api/dto/attachment.go`
- `internal/api/dto/error.go`
- `internal/api/mapper/*.go` (or service-layer mapper functions)

### 1.2 Define explicit response DTO types

Define concrete structs/types for compatibility responses (example fields):

- `type ErrorResponse struct { Error string \`json:"error"\` }`
- `type AuthUserDTO struct { ID string; Email string; Name string; CreatedAt string }`
- `type SessionDTO struct { User AuthUserDTO; CsrfToken string }`
- `type PostDTO struct { ID string; Title string; Content string; Date string; PersonIDs []string; Hashtags []string; Attachments []AttachmentDTO; Comments []CommentDTO; CreatedAt string; UpdatedAt string }`
- `type PersonDTO struct { ID string; Name string; CreatedAt string; UpdatedAt string }`
- `type CommentDTO struct { ID string; PostID string; AuthorID *string; Body string; CreatedAt string; UpdatedAt string }`
- `type AttachmentDTO struct { ID string; FileName string; ContentType string; SizeBytes int64; ViewURL string; DownloadURL string; CreatedAt string }`

> Important: Do not include `category` or `mood` in required contract DTOs.

### 1.3 Add mapper functions with deterministic typing

Implement mapper functions in handler/service layer, not repository layer.

Suggested method signatures (Go-style examples):

- `func ToAuthUserDTO(u domain.User) dto.AuthUserDTO`
- `func ToPostDTO(p domain.Post, comments []domain.Comment, attachments []domain.Attachment) dto.PostDTO`
- `func ToPersonDTO(p domain.Person) dto.PersonDTO`
- `func ToCommentDTO(c domain.Comment) dto.CommentDTO`
- `func ToAttachmentDTO(a domain.Attachment, baseURL string) dto.AttachmentDTO`

Mapper rules:

- Convert IDs to one agreed external type (recommend `string` if frontend currently stringifies keys).
- Format timestamps via one helper: `func FormatAPITime(t time.Time) string` returning UTC RFC3339.
- Convert nil slices to empty slices before JSON encoding.

---

## 2) Backend: request compatibility shims + route parity

### 2.1 Dual-read request decoder for transitional field names

Add request decoding helpers that accept aliases and normalize to canonical internal input.

Examples:

- `personId` and `person_id` → internal `PersonID`
- `postId` and `post_id` → internal `PostID`
- `query` and `search` → internal `SearchText`

Implementation approach:

1. Decode into `map[string]any` (or equivalent intermediate struct).
2. Resolve aliases in priority order.
3. Populate typed internal request struct.
4. Run validation on normalized struct only.

### 2.2 Route alias handlers

If Gemini uses different paths, add temporary alias routes that delegate to canonical handlers.

Examples:

- `GET /api/timeline` → call existing `ListPostsHandler`
- `GET /api/profile` → call existing `GetMeHandler`

Add comment tag on each alias route:

- `// TODO(compat): remove after frontend rollout complete (target: YYYY-MM-DD)`

### 2.3 Pagination parity for all list endpoints

For each list endpoint (`posts`, `persons`, `comments`, `attachments` where applicable):

- Accept `limit` (int), `offset` (int) and optionally `cursor` (string) if already supported.
- Return pagination metadata DTO:
  - `type PageMeta struct { Limit int; Offset int; NextCursor *string; Total *int }`
- Ensure defaults and max limits are server-defined constants.

---

## 3) Backend: auth/session/csrf interoperability + security hardening

### 3.1 CSRF header casing compatibility

In CSRF middleware, accept both headers temporarily:

- `X-CSRF-Token`
- `X-Csrf-Token`

Normalization method:

- Check canonical header first.
- Fallback to alternate casing.
- Use one internal variable `csrfToken` for downstream checks.

### 3.2 Session secret minimum length

At startup config validation, enforce:

- `SESSION_SECRET` length `>= 32`

If invalid:

- Fail fast with explicit message.
- Add a unit test for config validation failure and success paths.

### 3.3 Cookie/session settings parity

Verify and align cookie attributes expected by Gemini frontend:

- `HttpOnly=true`
- `SameSite` value compatible with deployment mode
- `Secure` enabled in production
- path/domain settings match API host usage

Document final values in contract matrix notes.

---

## 4) Backend: validation updates (including category/mood removal)

### 4.1 Post create/update validation changes

In post validator/request struct:

- Remove required checks for `category` and `mood`.
- If fields still accepted for backward compatibility, make them optional and ignored.

Concrete checks to keep:

- title length bounds
- content length bounds
- valid date format
- person IDs exist (if provided)

### 4.2 Input validation hardening across endpoints

Implement/verify validators:

- Registration/Login:
  - email format validator
  - password minimum length + complexity rule (document exact regex/rules)
- Comment/Post bodies:
  - max length limits
  - trim whitespace-only content rejection
- Hashtags/query inputs:
  - allowed characters and max count

Return all validation failures in standardized format with top-level `{ "error": "..." }` (and optional details field if already standard in Codex).

---

## 5) Backend: persons/comments/attachments contract guarantees

### 5.1 Referential integrity for person deletion

Ensure DB and service behavior align with contract:

- On person delete, references in posts/comments become `NULL` (or defined behavior in matrix).
- Do not cascade-delete posts/comments unless explicitly required.

Add migration if needed:

- Foreign key rule: `ON DELETE SET NULL`

### 5.2 Non-null collection guarantees

In serializers/mappers for all endpoints:

- `comments`, `attachments`, `persons`, `hashtags` return `[]` when empty.
- never emit `null` for collection fields consumed by frontend maps.

### 5.3 Attachment URL and metadata normalization

Ensure attachment DTO always includes:

- `viewUrl` (or agreed casing)
- `downloadUrl`
- `fileName`
- `contentType`
- `sizeBytes`

Use one URL builder function so all handlers generate identical paths.

---

## 6) Backend integration tests (contract-critical)

Add/extend integration test suite (example file grouping):

- `tests/integration/auth_flow_test.*`
- `tests/integration/timeline_filters_test.*`
- `tests/integration/post_without_category_mood_test.*`
- `tests/integration/comment_crud_test.*`
- `tests/integration/person_crud_test.*`
- `tests/integration/attachment_contract_test.*`

### Required test cases

1. `login -> me/profile`
   - assert status codes
   - assert DTO field names/types
   - assert csrf/session behavior
2. Timeline fetch with filters
   - date/search/hashtag/person filters and aliases
   - assert pagination metadata
3. Post create/edit without `category`/`mood`
   - request omits fields
   - operation succeeds
4. Comment create/delete
   - relation to post remains consistent
5. Person CRUD minimal flow
   - create, rename, delete
   - verify `SET NULL` behavior for references
6. Attachment upload/list/download metadata
   - verify URL fields and metadata keys

### Assertion conventions

- Add helper assertions for:
  - `assertErrorShape(resp)` → top-level `error` string
  - `assertTimestampRFC3339(value)`
  - `assertNoNullArrays(respBody, ["comments", "attachments", ...])`

---

## 7) Frontend minimal adaptation plan (Gemini)

### 7.1 API client-only compatibility edits

Constrain frontend changes to API boundary files, e.g.:

- `src/api/client.ts`
- `src/api/types.ts`
- `src/api/mappers.ts` (if present)

Concrete actions:

1. Update endpoint paths only where backend cannot provide alias in time.
2. Normalize rare field mismatches in API client response adapters.
3. Keep UI components unchanged unless broken by strict type errors.

### 7.2 Remove category/mood usage end-to-end

- Remove `category`/`mood` from create/edit payload builders.
- Remove any TS required types for these fields in post models.
- Remove rendering references in post list/detail components.

### 7.3 Header and type updates

- Send `X-CSRF-Token` by default from shared request helper.
- Keep compatibility fallback only if needed during rollout.
- Update TS interfaces to match final DTOs exactly (IDs, timestamps, arrays).

### 7.4 ProtectedRoute quality fix

- Move `ProtectedRoute` out of `App` component body.
- Export as top-level component/function to avoid recreation each render.

---

## 8) Rollout sequence with concrete checkpoints

### Phase A: Contract + backend DTO/shims

1. Merge contract matrix with endpoint-by-endpoint signoff.
2. Ship backend DTOs + mappers behind existing endpoints.
3. Add alias routes and request dual-read parsing.
4. Ship CSRF/session compatibility and secret-length enforcement.

Checkpoint A exit criteria:

- All backend integration tests pass.
- Existing frontend still functional.

### Phase B: Frontend API/type minimal updates

1. Update Gemini API client endpoints/types.
2. Remove category/mood dependencies.
3. Apply `ProtectedRoute` extraction.

Checkpoint B exit criteria:

- Gemini smoke checks pass locally/CI.

### Phase C: Stabilization + cleanup

1. Track shim usage via logs/metrics.
2. Remove unused aliases/dual-read params after cutoff date.
3. Update contract matrix status to `final` and close migration tasks.

---

## 9) Definition of done checklist (implementation-level)

- [ ] Contract matrix committed and reviewed.
- [ ] Compatibility DTOs implemented for auth/posts/persons/comments/attachments.
- [ ] Handler/service mappers in place; repositories unchanged.
- [ ] Error payloads standardized to `{ "error": string }`.
- [ ] CSRF accepts both header casings during migration.
- [ ] Session secret minimum length validation enforced (`>= 32`).
- [ ] Post create/update works without `category` and `mood`.
- [ ] List endpoints support agreed pagination parameters + metadata.
- [ ] Collections serialized as `[]` (never `null`).
- [ ] Attachment metadata and URLs match contract.
- [ ] Backend integration tests for all critical flows passing.
- [ ] Gemini frontend smoke checks passing with minimal API-layer changes.
- [ ] Temporary aliases/shims have owners + removal dates.
