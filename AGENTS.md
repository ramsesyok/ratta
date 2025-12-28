# AGENTS.md

## Purpose

This file guides Codex (and other coding agents) to implement the ratta desktop issue-tracking application consistently with the approved design.

## Authoritative documents

When making implementation decisions, treat the following English documents as the source of truth:

- docs/requirements_specification.md
- docs/requirements_definition.md
- docs/basic_design.md
- docs/detailed_design.md

If an instruction here conflicts with those documents, follow the documents.

## Product summary

rattа is a Windows desktop application for managing issues shared between Contractor and Vendor in a single project. Data is stored as JSON files in a project root folder. Synchronization and conflict resolution are out of scope and are handled operationally via git.

## Non-goals

- No user management or login system
- No in-app sync, merge, or locking
- No backup/restore features (operational responsibility)
- No multi-level category hierarchy (categories are flat folders under project root)

## Tech stack and constraints

- Desktop: Wails
- Backend: Go
- Frontend: Vue 3 + Vuetify 3
- State management: Pinia
- Unit tests: Go testing, Vitest
- E2E tests: Playwright
- Offline-first: no external cloud APIs; avoid network-dependent build steps at runtime
- Distribution: zip archive (no installer)

## Development method (t_wada-style TDD)

Implement using test-driven development as the default workflow:

- Red: write a small failing test describing the next behavior
- Green: implement the minimum code to pass
- Refactor: improve structure while keeping tests green
- Prefer small, fast unit tests for domain and infrastructure logic
- Keep dependencies injectable; avoid hard-to-test code paths
- Add E2E tests only for representative cross-layer flows, not for everything

## Architecture and layering

Follow the layering from docs/detailed_design.md (IDs may appear as DD-...):

- internal/domain
  - Domain types, validation, status transitions, invariants
- internal/app
  - Use-cases orchestrating domain + infra
- internal/infra
  - File I/O, atomic writes, schema validation, logging, crypto
- internal/present
  - Wails-exposed DTOs and API surface, error DTO mapping

Rules:

- Frontend must not depend on Go internal structures; only DTOs
- Backend enforces final authorization and invariants; frontend may only narrow UI choices
- All update operations must use atomic write (tmp write then rename)
- Schema-invalid issues are read-only (viewable, not updatable)

## Data and file rules (implementation-critical)

Project root:

- <PROJECT_ROOT>/<category>/<issue_id>.json
- <PROJECT_ROOT>/<category>/<issue_id>.files/<attachment_file>

Exclusions during scanning:

- Ignore .git and dot-prefixed dirs, except .tmp_rename (special handling)
- Do not treat nested subfolders as categories

JSON formatting:

- UTF-8, LF
- Indent: 2 spaces
- Key order: fixed; decide the canonical order in code and keep it stable

Timestamps:

- Persist: ISO 8601 with timezone, second precision
- UI display: Japanese format "YYYY年MM月DD日 hh時mm分ss秒"

IDs:

- issue_id: nanoid 9 chars
- comment_id: UUID v7
- attachment_id: nanoid 9 chars

Attachments:

- Stored under <issue_id>.files
- stored_name format: <attachment_id>_<sanitized_original_name>
- Sanitize Windows-invalid characters and trailing dot/space; trim to 255 chars
- Comment attachments: up to 5 per comment
- If attachment saving fails, do not update issue JSON; clean up partial files

Modes:

- Vendor mode by default
- Contractor mode requires auth/contractor.json and password verification
- Vendor cannot transition to Closed/Rejected
- Closed/Rejected issues are not editable by either mode

## Error handling contract

- Backend returns a response DTO that includes ok/data/error (or equivalent)
- Map internal errors to a common ApiErrorDTO:
  - error_code, message, detail, target_path, hint
- Frontend aggregates errors into a single store (stores/errors) and displays via an error detail dialog

## How to work with tasks.md

- Implement tasks in order from tasks.md
- Keep changes small and coherent; update tasks.md checkboxes as you complete items
- For each task:
  - add/adjust unit tests first
  - run the relevant test suites locally before marking complete
  - do not change design docs unless explicitly requested

## Common commands (expected)

Backend:

- go test ./...

Frontend:

- npm test
- npm run build

E2E:

- npx playwright test

Wails:

- wails dev
- wails build

Adjust to the repository’s actual scripts, but keep commands documented and stable.
