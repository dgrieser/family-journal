# Plan: Align Codex Backend for Gemini Frontend Compatibility

> **Current status snapshot:** 2026-02-22
>
> This plan remains the recommended path from the latest branch review in this repository.

## Implementation status (based on latest documented branch state)

- ✅ Direction confirmed: Codex backend as canonical, Gemini frontend as compatibility target.
- 🔄 In progress: endpoint/DTO contract normalization (`/api/v1` vs `/api`).
- 🔄 In progress: auth/profile field-name parity (`active` vs `is_active`, similar deltas).
- 🔄 In progress: CSRF header casing interoperability.
- ⏳ Next: finalize contract matrix and add cross-branch integration tests for the compatibility surface.

---

## Context and assumptions

- This plan is derived from the existing branch comparison notes in this repository.
- `category` and `mood` are explicitly out of scope and should be ignored by both backend and frontend.
- Goal: keep Gemini frontend changes minimal by adapting Codex backend response/request shapes where practical.

---

## 1) Target integration strategy

Use **Codex backend as the source of truth** and introduce a **Gemini-compatibility API surface** (DTO + handler translation layer) so Gemini frontend can continue using near-current assumptions.

### Why this approach

- Codex backend is more mature in architecture and security.
- Gemini frontend can be migrated with smaller, lower-risk changes.
- A translation layer avoids weakening Codex internal domain models.

---

## 2) API contract deltas to close (backend-first)

Create a short “contract matrix” in implementation docs mapping:

- Gemini frontend currently expected field names/types
- Codex backend current field names/types
- Final compatibility contract

At minimum, align these areas:

1. **Auth/session payloads**
   - login/me/profile response fields
   - error payload shape (`{ error: string }` consistently)
   - CSRF header normalization (accept both `X-CSRF-Token` and `X-Csrf-Token` temporarily)

2. **Posts list/detail payloads**
   - Ensure fields used by Gemini UI are always present and consistently typed.
   - Exclude `category` and `mood` from required validation and from frontend rendering contract.
   - Add pagination support (e.g., cursor or offset-based) to all list endpoints to ensure scalability.

3. **Persons and comments payloads**
   - Keep identifiers and timestamp formats stable.
   - Return empty arrays instead of `null` where frontend maps collections.
   - Document and align on referential integrity rules (e.g., using `SET NULL` on person deletion to prevent data loss).

4. **Filtering/query params**
   - Support Gemini query format for date/search/hashtag/person filters.
   - Keep backward-compatible aliases for any renamed params during migration.

5. **Attachments shape**
   - Ensure download/view URLs and metadata keys match what Gemini UI expects.

---

## 3) Codex backend changes required

### 3.1 Add compatibility DTO layer (non-breaking)

- Add dedicated response DTOs for auth, posts, persons, comments, attachments.
- Map internal entities -> DTO in handlers/services, not in repositories.
- Guarantee stable JSON tags matching Gemini frontend expectations.

### 3.2 Relax post validation for optional fields

- Remove any “required” logic for `category` and `mood`.
- If columns still exist in DB, treat them as nullable/optional and do not enforce UI presence.

### 3.3 Request/response compatibility shims

- Accept both old/new field names where branches differ (temporary dual-read logic).
- Normalize output to one canonical compatibility format.
- Maintain centralized JSON errors so frontend gets predictable `{ "error": "..." }`.

### 3.4 Route-level parity pass

- Verify endpoint paths/methods used by Gemini frontend exist on Codex backend.
- If paths differ, add lightweight alias routes that delegate to canonical handlers.

### 3.5 CSRF/session interoperability hardening

- Temporarily accept both CSRF header casings.
- Enforce a minimum session secret length (e.g., 32 characters) in backend configuration.
- Confirm cookie attributes and session lifecycle work with Gemini frontend request flow.

### 3.6 Integration checks

Add/update backend integration tests for:

- login -> profile/me flow
- timeline fetch with filters
- create/edit post without `category` and `mood`
- comment create/delete
- person CRUD minimal flow
- attachment upload/list/download metadata contract

### 3.7 Input Validation Hardening

- Review and strengthen input validation across all endpoints, especially for registration (e.g., email format, password complexity) and user-generated content to improve security and data integrity.

---

## 4) Minimal Gemini frontend changes required

### 4.1 API client minimal normalization

- Update only the API client layer to match Codex-compatible endpoints if needed.
- Add tiny response normalizers only where field names differ and cannot be shimmed backend-side quickly.

### 4.2 Ignore category/mood end-to-end

- Do not render `category`/`mood` inputs.
- Do not send them in create/update payloads.
- Do not expect/display them in post cards/detail.

### 4.3 Header compatibility

- Prefer sending `X-CSRF-Token`.
- Keep existing behavior temporarily if backend shim accepts both.

### 4.4 Type adjustments only where needed

- Update shared TS types for the final compatibility contract.
- Avoid component rewrites; confine changes to types + API mapping.

### 4.5 Regression smoke checks

- Login/logout
- Timeline load by date
- Create/edit/delete post (without category/mood)
- Person create/rename/delete
- Comment add/delete
- Attachment upload/open

### 4.6 Code quality fixes

- Move `ProtectedRoute` component definition outside of `App` so it is not recreated on every render.

---

## 5) Suggested execution order

1. Write/agree contract matrix (single source of truth).
2. Implement Codex backend compatibility DTO + shims.
3. Add backend integration tests for contract-critical flows.
4. Apply minimal Gemini frontend API/type updates.
5. Apply targeted Gemini frontend code quality fixes (`ProtectedRoute` extraction).
6. Run full integration smoke test.
7. Remove temporary aliases/shims later (optional cleanup phase).

---

## 6) Definition of done

Integration is complete when:

- Gemini frontend works against Codex backend without UI-level rewrites.
- Posts and comments workflows operate normally without `category`/`mood`.
- API errors are consistently shaped and frontend-handled.
- All agreed integration tests and smoke checks pass.
- Backend enforces minimum session secret length per agreed security baseline.
- Contract matrix is checked into repo and reflects actual API behavior.
