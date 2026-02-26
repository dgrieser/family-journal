# AGENT.md

This file provides guidance when working with code in this branch of the repository.

## What this branch is

This is a **documentation-only branch** — there is no source code here. It tracks the analysis and convergence planning for two rival AI-generated implementations of the **FamilyJournal** application (a full-stack care journal for children with official care levels).

The two implementations live in **separate** branches:
- `codex`
- `gemini`

The agreed convergence strategy: **Codex backend + Gemini frontend**

## Document map

| File | Purpose |
|------|---------|
| `ORIGINAL_PROMPT.md` | Full product specification used to generate both implementations |
| `BRANCH_COMPARISON_REVIEW.md` | Detailed side-by-side comparison of every architectural layer (updated 2026-02-22) |
| `API_CALLS.md` | Per-endpoint API contract comparison table (Codex vs Gemini, request/response fields) |

## Application tech stack

- **Backend:** Go + Fiber, MySQL, session-based auth (HttpOnly cookies, no JWT)
- **Frontend:** React + TypeScript + Vite, TailwindCSS, react-i18next (de/en), Zustand
- **Deployment:** Docker Compose (backend + frontend + MySQL)

## Instructions

When updating any document in this branch, you **MUST** first switch to the respective branches `codex` or `gemini` to retrieve the current state.
