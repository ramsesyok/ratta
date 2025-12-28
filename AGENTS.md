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

ratta is a Windows desktop application for managing issues shared between Contractor and Vendor in a single project. Data is stored as JSON files in a project root folder. Synchronization and conflict resolution are out of scope and are handled operationally via git.

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

Comment language and traceability:

- Source code comments must be written in Japanese.
- To ensure traceability with the detailed design document, include the ID of the relevant detailed design section (e.g., DD-...) in function comments.
- Test code must include Japanese comments that clearly explain intent and purpose.

## Commenting and documentation rules (enforced)

These rules exist to make the codebase readable and maintainable when implemented by agents. If a change does not meet these rules, revise the code and comments until it does.

### Language and tone

- Comments are Japanese by default.
- Use consistent terminology across the project (do not alternate synonyms arbitrarily).
- Avoid vague wording (e.g., "maybe", "roughly", "somehow"). Be specific and testable.

### What must be commented

Project-level and file-level:

- For each new or modified file, add 1 to 3 lines near the top describing the file responsibility (責務) and the boundary of what it does not do.

Exported identifiers (GoDoc):

- All exported types, functions, methods, constants, and variables must have GoDoc-style comments.
- The first sentence must start with the identifier name.

Functions and methods (required items):

For every new or modified function or method, add a Japanese comment block immediately above it that covers the following, as applicable:

- 目的: what the function guarantees and why it exists
- 入力: meaning of parameters, units, assumptions, preconditions
- 出力: meaning of return values and possible states
- エラー: when errors occur and how callers are expected to handle them
- 副作用: I/O, file writes, logging, metrics, global state changes
- 並行性: whether it is thread-safe, how shared state is protected, locking assumptions
- 不変条件: invariants that must hold (e.g., sorted, normalized, validated)
- 関連DD: the relevant detailed design ID (DD-...)

Non-obvious logic:

Add a short Japanese rationale comment immediately above code that is likely to confuse a reader, including:

- Non-trivial conditionals, boundary handling, early returns with hidden intent
- Regular expressions, bit operations, complex formatting, parsing logic
- Magic numbers or constants with implicit meaning (always state unit and origin)
- Workarounds or compatibility logic (what is avoided, why, and removal condition)

Performance-related code:

- If you introduce an optimization, explain why it is needed, what bottleneck it addresses, and what to measure.
- If complexity matters, state expected time/space characteristics at a high level.

### What must not be done

- Do not write comments that simply restate the code in words.
- Do not let comments drift from implementation. If the code changes behavior, update the comment in the same change.

### Tests

- Tests must include Japanese comments explaining intent and purpose.
- Table-driven tests must make the intent obvious per case:
  - Put the intent in the case name and/or add a comment per case.
- For boundary cases, error cases, and regression tests, explain why the case exists and what failure it prevents.
- When an expected value is non-trivial, document the basis (前提 or 根拠) so failures are diagnosable.

### Definition of done (commenting)

A change is complete only if:

- New/modified functions include the required Japanese comment block (purpose, inputs/outputs, errors, side effects, concurrency, invariants, and DD-... reference).
- Non-obvious logic has a rationale comment (why, not what).
- Any workaround/compatibility logic includes a removal condition.
- Tests include intent/purpose comments and cases are understandable without reading the implementation.

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
